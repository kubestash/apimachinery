/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package blob

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
)

const testBucket = "worm"

type fakeRetention struct {
	Mode            string    `xml:"Mode"`
	RetainUntilDate time.Time `xml:"RetainUntilDate"`
}

// fakeObjectLockS3 is a minimal S3 server that implements the Object Lock APIs used by the Blob.
type fakeObjectLockS3 struct {
	mu         sync.Mutex
	lockConfig string
	objects    map[string]*fakeRetention
	puts       map[string]fakeRetention
}

func (f *fakeObjectLockS3) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, "/"+testBucket), "/")
	q := r.URL.Query()
	switch {
	case key == "" && q.Has("object-lock"):
		if f.lockConfig == "" {
			writeS3Error(w, http.StatusNotFound, "ObjectLockConfigurationNotFoundError")
			return
		}
		_, _ = io.WriteString(w, f.lockConfig)
	case key == "" && q.Get("list-type") == "2":
		var b strings.Builder
		b.WriteString(`<ListBucketResult><IsTruncated>false</IsTruncated>`)
		for k := range f.objects {
			if strings.HasPrefix(k, q.Get("prefix")) {
				fmt.Fprintf(&b, "<Contents><Key>%s</Key></Contents>", k)
			}
		}
		b.WriteString(`</ListBucketResult>`)
		_, _ = io.WriteString(w, b.String())
	case q.Has("retention") && r.Method == http.MethodGet:
		ret, ok := f.objects[key]
		if !ok || ret == nil {
			writeS3Error(w, http.StatusNotFound, "NoSuchObjectLockConfiguration")
			return
		}
		fmt.Fprintf(w, "<Retention><Mode>%s</Mode><RetainUntilDate>%s</RetainUntilDate></Retention>", ret.Mode, ret.RetainUntilDate.Format(time.RFC3339))
	case q.Has("retention") && r.Method == http.MethodPut:
		var ret fakeRetention
		if err := xml.NewDecoder(r.Body).Decode(&ret); err != nil {
			writeS3Error(w, http.StatusBadRequest, "MalformedXML")
			return
		}
		f.puts[key] = ret
	default:
		writeS3Error(w, http.StatusNotImplemented, "NotImplemented")
	}
}

func writeS3Error(w http.ResponseWriter, status int, code string) {
	w.WriteHeader(status)
	fmt.Fprintf(w, "<Error><Code>%s</Code><Message>%s</Message></Error>", code, code)
}

func newObjectLockTestBlob(t *testing.T, f *fakeObjectLockS3, provider storageapi.StorageProvider) *Blob {
	t.Helper()
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)

	bs := &storageapi.BackupStorage{
		Spec: storageapi.BackupStorageSpec{
			Storage: storageapi.Backend{
				Provider: provider,
				S3: &storageapi.S3Spec{
					Endpoint:   srv.URL,
					Bucket:     testBucket,
					Region:     "us-east-1",
					Prefix:     "prefix",
					SecretName: "s3-secret",
				},
			},
		},
	}
	return &Blob{
		backupStorage: bs,
		prefix:        bs.Spec.Storage.S3.Prefix,
		s3Secret: &core.Secret{Data: map[string][]byte{
			AWSAccessKeyId:     []byte("id"),
			AWSSecretAccessKey: []byte("key"),
		}},
	}
}

func TestGetObjectLockConfig(t *testing.T) {
	tests := []struct {
		name       string
		lockConfig string
		expected   *ObjectLockConfig
	}{
		{
			name:     "Object Lock not configured",
			expected: &ObjectLockConfig{},
		},
		{
			name:       "Object Lock enabled without default retention",
			lockConfig: `<ObjectLockConfiguration><ObjectLockEnabled>Enabled</ObjectLockEnabled></ObjectLockConfiguration>`,
			expected:   &ObjectLockConfig{Enabled: true},
		},
		{
			name:       "Compliance mode with days",
			lockConfig: `<ObjectLockConfiguration><ObjectLockEnabled>Enabled</ObjectLockEnabled><Rule><DefaultRetention><Mode>COMPLIANCE</Mode><Days>3</Days></DefaultRetention></Rule></ObjectLockConfiguration>`,
			expected:   &ObjectLockConfig{Enabled: true, Mode: storageapi.ObjectLockModeCompliance, DefaultRetention: 72 * time.Hour},
		},
		{
			name:       "Governance mode with years",
			lockConfig: `<ObjectLockConfiguration><ObjectLockEnabled>Enabled</ObjectLockEnabled><Rule><DefaultRetention><Mode>GOVERNANCE</Mode><Years>1</Years></DefaultRetention></Rule></ObjectLockConfiguration>`,
			expected:   &ObjectLockConfig{Enabled: true, Mode: storageapi.ObjectLockModeGovernance, DefaultRetention: 365 * 24 * time.Hour},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newObjectLockTestBlob(t, &fakeObjectLockS3{lockConfig: tt.lockConfig}, storageapi.ProviderS3)
			cfg, err := b.GetObjectLockConfig(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestGetObjectLockConfigNotSupported(t *testing.T) {
	b := newObjectLockTestBlob(t, &fakeObjectLockS3{}, storageapi.ProviderGCS)
	_, err := b.GetObjectLockConfig(context.Background())
	assert.ErrorIs(t, err, ErrObjectLockNotSupported)
}

func TestExtendObjectRetention(t *testing.T) {
	now := time.Now().UTC()
	day := 24 * time.Hour
	f := &fakeObjectLockS3{
		objects: map[string]*fakeRetention{
			// expires soon, must be extended keeping its mode
			"prefix/repo/data/aa/expiring": {Mode: "COMPLIANCE", RetainUntilDate: now.Add(time.Hour)},
			// locked for more than half of the period, must be skipped
			"prefix/repo/data/bb/fresh": {Mode: "COMPLIANCE", RetainUntilDate: now.Add(6 * day)},
			// already expired, must be extended
			"prefix/repo/index/expired": {Mode: "GOVERNANCE", RetainUntilDate: now.Add(-time.Hour)},
			// no retention, must be locked with the given mode
			"prefix/repo/snapshots/unlocked": nil,
			// restic lock files, must be skipped
			"prefix/repo/locks/lock1":        nil,
			"prefix/repo/v1/dump/locks/lock": nil,
			// other repository, must be skipped
			"prefix/other/data/cc/x": nil,
		},
		puts: map[string]fakeRetention{},
	}
	b := newObjectLockTestBlob(t, f, storageapi.ProviderS3)

	stats, err := b.ExtendObjectRetention(context.Background(), "repo", storageapi.ObjectLockModeCompliance, 7*day)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), stats.Scanned)
	assert.Equal(t, int64(3), stats.Extended)

	assert.Len(t, f.puts, 3)
	assert.Equal(t, "COMPLIANCE", f.puts["prefix/repo/data/aa/expiring"].Mode)
	assert.Equal(t, "GOVERNANCE", f.puts["prefix/repo/index/expired"].Mode)
	assert.Equal(t, "COMPLIANCE", f.puts["prefix/repo/snapshots/unlocked"].Mode)
	for k, ret := range f.puts {
		assert.WithinDuration(t, now.Add(7*day), ret.RetainUntilDate, time.Minute, k)
	}
}

func TestExtendObjectRetentionInvalidPeriod(t *testing.T) {
	b := newObjectLockTestBlob(t, &fakeObjectLockS3{}, storageapi.ProviderS3)
	_, err := b.ExtendObjectRetention(context.Background(), "repo", storageapi.ObjectLockModeCompliance, 0)
	assert.Error(t, err)
}

func TestToS3RetentionMode(t *testing.T) {
	assert.Equal(t, "COMPLIANCE", string(toS3RetentionMode(storageapi.ObjectLockModeCompliance)))
	assert.Equal(t, "GOVERNANCE", string(toS3RetentionMode(storageapi.ObjectLockModeGovernance)))
	assert.Equal(t, "GOVERNANCE", string(toS3RetentionMode("")))
}
