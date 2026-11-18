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

// fakeObjectLockS3 is a minimal S3 server that implements the Object Lock APIs used by the Blob.
type fakeObjectLockS3 struct {
	mu         sync.Mutex
	lockConfig string
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
