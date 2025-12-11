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
	"time"

	addonapi "kubestash.dev/apimachinery/apis/addons/v1alpha1"
	coreapi "kubestash.dev/apimachinery/apis/core/v1alpha1"
	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
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
	stdinPipeCommand  = Command{Name: "echo", Args: []any{"hello"}}
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

	setupOpt := &SetupOptions{
		Backends: []*Backend{
			{
				EncryptionSecret: &kmapi.ObjectReference{
					Name:      es.Name,
					Namespace: es.Namespace,
				},
				BackupStorage: &kmapi.ObjectReference{
					Name:      bs.Name,
					Namespace: bs.Namespace,
				},
				backend: backend{
					provider: storage.ProviderLocal,
					bucket:   localRepoDir,
				},
			},
		},
		ScratchDir:  scratchDir,
		EnableCache: false,
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
	err = w.InitializeRepository(w.Config.Backends[0].Repository)
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
	for _, b := range w.Config.Backends {
		if err = w.InitializeRepository(b.Repository); err != nil {
			t.Error(err)
		}
	}
	for _, b := range w.Config.Backends {
		repoExist := w.repositoryExist(b.Repository)
		assert.Equal(t, true, repoExist)
	}
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

	for _, b := range w.Config.Backends {
		repoExist := w.repositoryExist(b.Repository)
		assert.Equal(t, false, repoExist)
	}
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
	for _, b := range w.Config.Backends {
		if err = w.InitializeRepository(b.Repository); err != nil {
			t.Error(err)
		}
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

	repository := w.Config.Backends[0].Repository
	restoreOut, err := w.RunRestore(repository, restoreOpt)
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
	for _, b := range w.Config.Backends {
		if err = w.InitializeRepository(b.Repository); err != nil {
			t.Error(err)
		}
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
	fmt.Println("backup output:")
	for _, out := range backupOut {
		fmt.Println(out)
	}

	dumpOpt := DumpOptions{
		FileName:           fileName,
		StdoutPipeCommands: []Command{stdoutPipeCommand},
	}

	repository := w.Config.Backends[0].Repository
	dumpOut, err := w.Dump(repository, dumpOpt)
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
	for _, b := range w.Config.Backends {
		if err = w.InitializeRepository(b.Repository); err != nil {
			t.Error(err)
		}
	}
	w.Config.IONice = &ofst.IONiceSettings{
		Class:     ptr.To(int32(2)),
		ClassData: ptr.To(int32(3)),
	}
	w.Config.Nice = &ofst.NiceSettings{
		Adjustment: ptr.To(int32(12)),
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

	restoreOut, err := w.RunRestore(w.Config.Backends[0].Repository, restoreOpt)
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
	for _, b := range w.Config.Backends {
		if err = w.InitializeRepository(b.Repository); err != nil {
			t.Error(err)
		}
	}

	w.Config.IONice = &ofst.IONiceSettings{
		Class:     ptr.To(int32(2)),
		ClassData: ptr.To(int32(3)),
	}
	w.Config.Nice = &ofst.NiceSettings{
		Adjustment: ptr.To(int32(12)),
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
	dumpOut, err := w.Dump(w.Config.Backends[0].Repository, dumpOpt)
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
			for _, b := range w.Config.Backends {
				if err = w.InitializeRepository(b.Repository); err != nil {
					t.Error(err)
				}
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

			_, err = w.RunRestore(w.Config.Backends[0].Repository, test.restoreOpt)
			if err != nil {
				t.Error(err)
				return
			}
		})
	}
}

func TestBackupWithTimeout(t *testing.T) {
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
	for _, b := range w.Config.Backends {
		if err = w.InitializeRepository(b.Repository); err != nil {
			t.Error(err)
		}
	}

	duration := metav1.Duration{Duration: 10 * time.Millisecond}
	w.Config.Timeout = &duration

	backupOpt := BackupOptions{
		StdinPipeCommands: []Command{stdinPipeCommand},
		StdinFileName:     fileName,
	}
	_, err = w.RunBackup(backupOpt)
	assert.Error(t, err, "Timeout error")
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
	for _, b := range w.Config.Backends {
		if err = w.InitializeRepository(b.Repository); err != nil {
			t.Error(err)
		}
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
	repoStats, err := w.VerifyRepositoryIntegrity(w.Config.Backends[0].Repository)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, true, *repoStats.Integrity)
}

func setupTestForMultipleBackends(tempDir string, backendsCount int) (*ResticWrapper, error) {
	setupOpt := &SetupOptions{
		ScratchDir:  filepath.Join(tempDir, "scratch"),
		EnableCache: false,
	}
	if err := os.MkdirAll(setupOpt.ScratchDir, 0o777); err != nil {
		return nil, err
	}

	var initObjs []client.Object
	for idx := range backendsCount {
		localRepoDir = filepath.Join(tempDir, fmt.Sprintf("repo-%d", idx))
		targetPath = filepath.Join(tempDir, fmt.Sprintf("target-%d", idx))
		if err := os.MkdirAll(localRepoDir, 0o777); err != nil {
			return nil, err
		}
		if err := os.MkdirAll(targetPath, 0o777); err != nil {
			return nil, err
		}

		err := os.WriteFile(filepath.Join(targetPath, fileName), []byte(fileContent), os.ModePerm)
		if err != nil {
			return nil, err
		}

		bs := sampleBackupStorage(func(backupStorage *storageapi.BackupStorage) {
			backupStorage.Name = fmt.Sprintf("sample-backup-storage-%d", idx)
		})
		// ss := sampleSecret(&kmapi.ObjectReference{Name: bs.Spec.Storage.GCS.Secret, Namespace: bs.Namespace})
		es := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("sample-encryptionsecret-%d", idx),
				Namespace: bs.Namespace,
			},
			Data: map[string][]byte{
				"RESTIC_PASSWORD": []byte(password),
			},
		}
		initObjs = append(initObjs, bs, es)
		setupOpt.Backends = append(setupOpt.Backends, &Backend{
			Repository: fmt.Sprintf("%s-%s-repository-%d", es.Namespace, storage.ProviderLocal, idx),
			EncryptionSecret: &kmapi.ObjectReference{
				Name:      es.Name,
				Namespace: es.Namespace,
			},
			BackupStorage: &kmapi.ObjectReference{
				Name:      bs.Name,
				Namespace: bs.Namespace,
			},
			backend: backend{
				provider: storage.ProviderLocal,
				bucket:   localRepoDir,
				envs:     map[string]string{},
			},
		})
	}

	var err error
	setupOpt.Client, err = getFakeClient(initObjs...)
	if err != nil {
		return nil, err
	}

	w, err := NewResticWrapper(setupOpt)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func TestMultipleBackedBackupRestoreStdin(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "stash-unit-test-")
	if err != nil {
		t.Error(err)
		return
	}

	w, err := setupTestForMultipleBackends(tempDir, 3)
	if err != nil {
		t.Error(err)
		return
	}
	defer cleanup(tempDir)

	// Initialize Repository
	for _, b := range w.Config.Backends {
		if err = w.InitializeRepository(b.Repository); err != nil {
			t.Error(err)
		}
	}

	w.Config.IONice = &ofst.IONiceSettings{
		Class:     ptr.To(int32(2)),
		ClassData: ptr.To(int32(3)),
	}
	w.Config.Nice = &ofst.NiceSettings{
		Adjustment: ptr.To(int32(12)),
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

	fmt.Println("backup output:")
	for _, out := range backupOut {
		fmt.Println(out)
	}

	dumpOpt := DumpOptions{
		FileName:           fileName,
		StdoutPipeCommands: []Command{stdoutPipeCommand},
	}

	for _, b := range w.Config.Backends {
		klog.Infoln("Dumping backes up data from repository:", b.Repository)
		dumpOut, err := w.Dump(b.Repository, dumpOpt)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println("dump output:", dumpOut)
	}
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
