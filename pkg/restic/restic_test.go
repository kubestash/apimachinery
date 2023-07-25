/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package restic

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	addonapi "kubestash.dev/apimachinery/apis/addons/v1alpha1"
	coreapi "kubestash.dev/apimachinery/apis/core/v1alpha1"
	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"

	"github.com/stretchr/testify/assert"
	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	storage "kmodules.xyz/objectstore-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	localRepoDir string
	scratchDir   string
	targetPath   string
	password     = "password"

	fileName          = "some-file"
	fileContent       = "hello stash"
	stdinPipeCommand  = Command{Name: "echo", Args: []interface{}{"hello"}}
	stdoutPipeCommand = Command{Name: "cat"}
)

func setupTest(tempDir string) (*ResticWrapper, error) {
	localRepoDir = filepath.Join(tempDir, "repo")
	scratchDir = filepath.Join(tempDir, "scratch")
	targetPath = filepath.Join(tempDir, "target")

	if err := os.MkdirAll(localRepoDir, 0o777); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(scratchDir, 0o777); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(targetPath, 0o777); err != nil {
		return nil, err
	}
	err := os.WriteFile(filepath.Join(targetPath, fileName), []byte(fileContent), os.ModePerm)
	if err != nil {
		return nil, err
	}

	bs := sampleBackupStorage()
	// ss := sampleSecret(&kmapi.ObjectReference{Name: bs.Spec.Storage.GCS.Secret, Namespace: bs.Namespace})
	es := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-encryptionsecret",
			Namespace: bs.Namespace,
		},
		Data: map[string][]byte{
			"RESTIC_PASSWORD": []byte(password),
		},
	}

	setupOpt := SetupOptions{
		backend: backend{
			provider: storage.ProviderLocal,
			bucket:   localRepoDir,
		},
		EncryptionSecret: &kmapi.ObjectReference{
			Name:      es.Name,
			Namespace: es.Namespace,
		},
		ScratchDir:  scratchDir,
		EnableCache: false,
		BackupStorage: &kmapi.TypedObjectReference{
			Name:      bs.Name,
			Namespace: bs.Namespace,
		},
	}

	setupOpt.Client, err = getFakeClient(bs, es)
	if err != nil {
		return nil, err
	}

	w, err := NewResticWrapper(setupOpt)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func cleanup(tempDir string) {
	if err := os.RemoveAll(tempDir); err != nil {
		klog.Errorln(err)
	}
}

func TestInitializeRepository(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "stash-unit-test-")
	if err != nil {
		t.Error(err)
		return
	}
	w, err := setupTest(tempDir)
	if err != nil {
		t.Error(err)
		return
	}
	defer cleanup(tempDir)
	err = w.InitializeRepository()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestRepositoryAlreadyExist_AfterInitialization(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "stash-unit-test-")
	if err != nil {
		t.Error(err)
		return
	}
	w, err := setupTest(tempDir)
	if err != nil {
		t.Error(err)
		return
	}
	defer cleanup(tempDir)
	err = w.InitializeRepository()
	if err != nil {
		t.Error(err)
		return
	}

	repoExist := w.RepositoryAlreadyExist()
	assert.Equal(t, true, repoExist)
}

func TestRepositoryAlreadyExist_WithoutInitialization(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "stash-unit-test-")
	if err != nil {
		t.Error(err)
		return
	}
	w, err := setupTest(tempDir)
	if err != nil {
		t.Error(err)
		return
	}
	defer cleanup(tempDir)

	repoExist := w.RepositoryAlreadyExist()
	assert.Equal(t, false, repoExist)
}

func TestBackupRestoreDirs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "stash-unit-test-")
	if err != nil {
		t.Error(err)
		return
	}

	w, err := setupTest(tempDir)
	if err != nil {
		t.Error(err)
		return
	}
	defer cleanup(tempDir)

	// Initialize Repository
	err = w.InitializeRepository()
	if err != nil {
		t.Error(err)
		return
	}

	backupOpt := BackupOptions{
		BackupPaths: []string{targetPath},
	}
	backupOut, err := w.RunBackup(backupOpt)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(backupOut)

	// delete target then restore
	if err = os.RemoveAll(targetPath); err != nil {
		t.Error(err)
		return
	}
	restoreOpt := RestoreOptions{
		RestorePaths: []string{targetPath},
	}
	restoreOut, err := w.RunRestore(restoreOpt)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(restoreOut)

	// check file
	fileContentByte, err := os.ReadFile(filepath.Join(targetPath, fileName))
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, fileContent, string(fileContentByte))
}

func TestBackupRestoreStdin(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "stash-unit-test-")
	if err != nil {
		t.Error(err)
		return
	}

	w, err := setupTest(tempDir)
	if err != nil {
		t.Error(err)
		return
	}
	defer cleanup(tempDir)

	// Initialize Repository
	err = w.InitializeRepository()
	if err != nil {
		t.Error(err)
		return
	}

	backupOpt := BackupOptions{
		StdinPipeCommands: []Command{stdinPipeCommand},
		StdinFileName:     fileName,
	}
	backupOut, err := w.RunBackup(backupOpt)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("backup output:", backupOut)

	dumpOpt := DumpOptions{
		FileName:           fileName,
		StdoutPipeCommands: []Command{stdoutPipeCommand},
	}
	dumpOut, err := w.Dump(dumpOpt)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("dump output:", dumpOut)
}

func TestBackupRestoreWithScheduling(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "stash-unit-test-")
	if err != nil {
		t.Error(err)
		return
	}

	w, err := setupTest(tempDir)
	if err != nil {
		t.Error(err)
		return
	}
	defer cleanup(tempDir)

	// Initialize Repository
	err = w.InitializeRepository()
	if err != nil {
		t.Error(err)
		return
	}

	w.config.IONice = &ofst.IONiceSettings{
		Class:     pointer.Int32P(2),
		ClassData: pointer.Int32P(3),
	}
	w.config.Nice = &ofst.NiceSettings{
		Adjustment: pointer.Int32P(12),
	}

	backupOpt := BackupOptions{
		BackupPaths: []string{targetPath},
	}
	backupOut, err := w.RunBackup(backupOpt)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(backupOut)

	// delete target then restore
	if err = os.RemoveAll(targetPath); err != nil {
		t.Error(err)
		return
	}
	restoreOpt := RestoreOptions{
		RestorePaths: []string{targetPath},
	}
	restoreOut, err := w.RunRestore(restoreOpt)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(restoreOut)

	// check file
	fileContentByte, err := os.ReadFile(filepath.Join(targetPath, fileName))
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, fileContent, string(fileContentByte))
}

func TestBackupRestoreStdinWithScheduling(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "stash-unit-test-")
	if err != nil {
		t.Error(err)
		return
	}

	w, err := setupTest(tempDir)
	if err != nil {
		t.Error(err)
		return
	}
	defer cleanup(tempDir)

	// Initialize Repository
	err = w.InitializeRepository()
	if err != nil {
		t.Error(err)
		return
	}

	w.config.IONice = &ofst.IONiceSettings{
		Class:     pointer.Int32P(2),
		ClassData: pointer.Int32P(3),
	}
	w.config.Nice = &ofst.NiceSettings{
		Adjustment: pointer.Int32P(12),
	}

	backupOpt := BackupOptions{
		StdinPipeCommands: []Command{stdinPipeCommand},
		StdinFileName:     fileName,
	}
	backupOut, err := w.RunBackup(backupOpt)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("backup output:", backupOut)

	dumpOpt := DumpOptions{
		FileName:           fileName,
		StdoutPipeCommands: []Command{stdoutPipeCommand},
	}
	dumpOut, err := w.Dump(dumpOpt)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("dump output:", dumpOut)
}

func TestBackupRestoreWithArgs(t *testing.T) {
	testCases := []struct {
		name       string
		backupOpt  BackupOptions
		restoreOpt RestoreOptions
	}{
		{
			name: "pass --ignore-inode flag during backup",
			backupOpt: BackupOptions{
				Args: []string{"--ignore-inode"},
			},
		},
		{
			name: "pass --tags during backup and restore",
			backupOpt: BackupOptions{
				Args: []string{"--tag=t1,t2"},
			},
			restoreOpt: RestoreOptions{
				Args: []string{"--tag=t1,t2"},
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "stash-unit-test-")
			if err != nil {
				t.Error(err)
				return
			}

			w, err := setupTest(tempDir)
			if err != nil {
				t.Error(err)
				return
			}
			defer cleanup(tempDir)

			// Initialize Repository
			err = w.InitializeRepository()
			if err != nil {
				t.Error(err)
				return
			}

			// create the source files
			err = os.Remove(filepath.Join(targetPath, fileName))
			if err != nil {
				t.Error(err)
				return
			}
			test.backupOpt.BackupPaths = []string{targetPath}

			_, err = w.RunBackup(test.backupOpt)
			if err != nil {
				t.Error(err)
				return
			}

			// delete target then restore
			if err = os.RemoveAll(targetPath); err != nil {
				t.Error(err)
				return
			}
			test.restoreOpt.RestorePaths = []string{targetPath}

			_, err = w.RunRestore(test.restoreOpt)
			if err != nil {
				t.Error(err)
				return
			}
		})
	}
}

func TestVerifyRepositoryIntegrity(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "stash-unit-test-")
	if err != nil {
		t.Error(err)
		return
	}

	w, err := setupTest(tempDir)
	if err != nil {
		t.Error(err)
		return
	}
	defer cleanup(tempDir)

	// Initialize Repository
	err = w.InitializeRepository()
	if err != nil {
		t.Error(err)
		return
	}

	backupOpt := BackupOptions{
		BackupPaths: []string{targetPath},
	}
	// take two backup
	_, err = w.RunBackup(backupOpt)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = w.RunBackup(backupOpt)
	if err != nil {
		t.Error(err)
		return
	}
	// apply retention policy
	repoStats, err := w.VerifyRepositoryIntegrity()
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, true, *repoStats.Integrity)
}

func getFakeClient(initObjs ...client.Object) (client.WithWatch, error) {
	scheme := runtime.NewScheme()
	if err := coreapi.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := storageapi.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := addonapi.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := core.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(initObjs...).Build(), nil
}

func sampleBackupStorage(transformFuncs ...func(*storageapi.BackupStorage)) *storageapi.BackupStorage {
	bs := &storageapi.BackupStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-backup-storage",
			Namespace: "demo",
		},
		Spec: storageapi.BackupStorageSpec{
			Storage: storageapi.Backend{
				Local: &storageapi.LocalSpec{
					MountPath: localRepoDir,
					SubPath:   "",
				},
			},
			DeletionPolicy: storageapi.DeletionPolicyWipeOut,
		},
	}

	for _, fn := range transformFuncs {
		fn(bs)
	}

	return bs
}
