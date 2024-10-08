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
	"os"
	"path/filepath"
	"testing"

	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"

	"github.com/stretchr/testify/assert"
	"gocloud.dev/gcerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	region            = "us-east-1"
	bucket            = "kubestash"
	endpoint          = "us-east-1.linodeobjects.com"
	prefix            = "unitTest"
	s3AccessKeyId     = "AWS_ACCESS_KEY_ID"
	s3SecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	sampleData        = "sample data"
	testPath          = "data"
	sampleFile        = "sample.txt"
)

// Set the necessary environment variables before executing the test
// Go to Run | Edit Configurations...| Environment:
func TestExistsShouldReturnTrueIfObjectExists(t *testing.T) {
	skipTestIfCredentialsNotFound(t)
	storage, err := getSampleStorage()
	assert.Nil(t, err)
	err = storage.Upload(context.Background(), filepath.Join(testPath, sampleFile), []byte(sampleData), "")
	assert.Nil(t, err)
	exists, err := storage.Exists(context.Background(), filepath.Join(testPath, sampleFile))
	assert.Nil(t, err)
	assert.True(t, exists)
	cleanupTestData(storage, t)
}

func TestExistsShouldReturnFalseIfObjectDoesNotExist(t *testing.T) {
	skipTestIfCredentialsNotFound(t)
	storage, err := getSampleStorage()
	assert.Nil(t, err)
	exists, err := storage.Exists(context.Background(), filepath.Join(testPath, sampleFile))
	assert.Nil(t, err)
	assert.False(t, exists)
}

func TestGetAfterUploadShouldReturnTheObject(t *testing.T) {
	skipTestIfCredentialsNotFound(t)
	storage, err := getSampleStorage()
	assert.Nil(t, err)
	err = storage.Upload(context.Background(), filepath.Join(testPath, sampleFile), []byte(sampleData), "")
	assert.Nil(t, err)
	d, err := storage.Get(context.Background(), filepath.Join(testPath, sampleFile))
	assert.Nil(t, err)
	receivedData := string(d)
	assert.Equal(t, sampleData, receivedData)
	cleanupTestData(storage, t)
}

func TestGetWithoutUploadShouldReturnNotFoundError(t *testing.T) {
	skipTestIfCredentialsNotFound(t)
	storage, err := getSampleStorage()
	assert.Nil(t, err)
	_, err = storage.Get(context.Background(), filepath.Join(testPath, sampleFile))
	assert.NotNil(t, err)
	assert.True(t, isNotFound(err))
}

func TestGetAfterDeleteShouldReturnNotFoundError(t *testing.T) {
	skipTestIfCredentialsNotFound(t)
	storage, err := getSampleStorage()
	assert.Nil(t, err)
	err = storage.Upload(context.Background(), filepath.Join(testPath, sampleFile), []byte(sampleData), "")
	assert.Nil(t, err)
	err = storage.Delete(context.Background(), filepath.Join(testPath, sampleFile), false)
	assert.Nil(t, err)
	_, err = storage.Get(context.Background(), filepath.Join(testPath, sampleFile))
	assert.NotNil(t, err)
	assert.True(t, isNotFound(err))
}

func TestListAfterUploadShouldReturnTheObjectList(t *testing.T) {
	skipTestIfCredentialsNotFound(t)
	storage, err := getSampleStorage()
	assert.Nil(t, err)
	paths := []string{"sample1.txt", "sample2.txt", "sample3.txt"}
	data := []string{"sample data 1", "sample data 2", "sample data 3"}
	for i := range paths {
		err = storage.Upload(context.Background(), filepath.Join(testPath, paths[i]), []byte(data[i]), "")
		assert.Nil(t, err)
	}
	receivedData, err := storage.List(context.Background(), testPath)
	assert.Nil(t, err)
	for i, d := range receivedData {
		assert.Equal(t, data[i], string(d))
	}
	cleanupTestData(storage, t)
}

func isNotFound(err error) bool {
	return gcerrors.Code(err) == gcerrors.NotFound
}

func cleanupTestData(storage *Blob, t *testing.T) {
	if err := storage.Delete(context.Background(), testPath, true); err != nil {
		t.Fatal(err)
	}
}

func getSampleStorage() (*Blob, error) {
	bs := sampleBackupStorage()
	fakeClient, err := getFakeClient()
	if err != nil {
		return nil, err
	}
	return NewBlob(context.Background(), fakeClient, bs)
}

func sampleBackupStorage(transformFuncs ...func(*storageapi.BackupStorage)) *storageapi.BackupStorage {
	bs := &storageapi.BackupStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-backup-storage",
			Namespace: "db",
		},
	}
	bs.Spec = storageapi.BackupStorageSpec{
		Storage: storageapi.Backend{
			Provider: storageapi.ProviderS3,
			S3: &storageapi.S3Spec{
				Region:   region,
				Bucket:   bucket,
				Endpoint: endpoint,
				Prefix:   prefix,
			},
		},
	}
	for _, fn := range transformFuncs {
		fn(bs)
	}
	return bs
}

func getFakeClient(initObjs ...kubeClient.Object) (kubeClient.WithWatch, error) {
	scheme := runtime.NewScheme()
	if err := storageapi.AddToScheme(scheme); err != nil {
		return nil, err
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(initObjs...).Build()
	return fakeClient, nil
}

func skipTestIfCredentialsNotFound(t *testing.T) {
	if _, ok := os.LookupEnv(s3AccessKeyId); !ok {
		t.Skip("Credential does not exist.")
	}
	if _, ok := os.LookupEnv(s3SecretAccessKey); !ok {
		t.Skip("Credential does not exist.")
	}
}
