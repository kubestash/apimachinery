package restic

import (
	"fmt"
	core "k8s.io/api/core/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	storage "kmodules.xyz/objectstore-api/api/v1"
	"kubestash.dev/apimachinery/apis/storage/v1alpha1"
	"os"
	"path/filepath"
)

type backend struct {
	provider       v1alpha1.StorageProvider
	bucket         string
	endpoint       string
	region         string
	insecureTLS    bool
	path           string
	storageAccount string
	envs           map[string]string
}

type Backend struct {
	backend
	storageSecret *core.Secret

	EncryptionSecret *kmapi.ObjectReference
	Directory        string
	BackupStorage    *kmapi.ObjectReference
	MountPath        string
	MaxConnections   int64
	CaCertFile       string
	Repository       string
	RepoRef          string
	SnapRef          string
	Error            error
}

func (b *Backend) createLocalDir() error {
	if b.provider == v1alpha1.ProviderLocal {
		return os.MkdirAll(b.bucket, 0o755)
	}
	return nil
}

func (b *Backend) appendInsecureTLSFlag(args []interface{}) []interface{} {
	if b.insecureTLS {
		return append(args, "--insecure-tls")
	}
	return args
}

func (b *Backend) appendCaCertFlag(args []interface{}) []interface{} {
	if b.CaCertFile != "" {
		return append(args, "--cacert", b.CaCertFile)
	}
	return args
}

func (b *Backend) appendMaxConnectionsFlag(args []interface{}) []interface{} {
	var maxConOption string
	if b.MaxConnections > 0 {
		switch b.provider {
		case storage.ProviderGCS:
			maxConOption = fmt.Sprintf("gs.connections=%d", b.MaxConnections)
		case storage.ProviderAzure:
			maxConOption = fmt.Sprintf("azure.connections=%d", b.MaxConnections)
		case storage.ProviderB2:
			maxConOption = fmt.Sprintf("b2.connections=%d", b.MaxConnections)
		}
	}
	if maxConOption != "" {
		return append(args, "--option", maxConOption)
	}
	return args
}

func (b *Backend) GetCaPath() string {
	return b.CaCertFile
}

func (b *Backend) setupEnv() error {
	b.envs = map[string]string{}
	if b.BackupStorage == nil {
		return fmt.Errorf("missing BackupStorage reference")
	}
	if err := w.setBackupStorageVariables(b); err != nil {
		b.Error = fmt.Errorf("failed to set BackupStorage variables. Reason: %w", err)
		continue
	}
	if b.storageSecret == nil {
		b.Error = fmt.Errorf("missing storage Secret")
		continue
	}

	if value, err := w.getSecretKey(b.storageSecret, RESTIC_PASSWORD, true); err != nil {
		b.Error = err
		continue
	} else {
		b.envs[RESTIC_PASSWORD] = value
	}

	tmpDir, err := os.MkdirTemp(w.Config.ScratchDir, fmt.Sprintf("%s-tmp-", b.Repository))
	if err != nil {
		b.Error = err
		continue
	}
	b.envs[TMPDIR] = tmpDir

	if _, ok := b.storageSecret.Data[CA_CERT_DATA]; ok {
		filePath, err := w.writeSecretKeyToFile(tmpDir, b.storageSecret, CA_CERT_DATA, "ca.crt")
		if err != nil {
			b.Error = err
			continue
		}
		b.CaCertFile = filePath
	}
	switch b.provider {

	case storage.ProviderLocal:
		b.envs[RESTIC_REPOSITORY] = fmt.Sprintf("%s/%s", b.bucket, b.Directory)

	case storage.ProviderS3:
		b.envs[RESTIC_REPOSITORY] = fmt.Sprintf("s3:%s/%s", b.endpoint, filepath.Join(b.bucket, b.path, b.Directory))
		if val, err := w.getSecretKey(b.storageSecret, AWS_ACCESS_KEY_ID, false); err == nil {
			b.envs[AWS_ACCESS_KEY_ID] = val
		} else {
			b.Error = err
			continue
		}

		if val, err := w.getSecretKey(b.storageSecret, AWS_SECRET_ACCESS_KEY, false); err == nil {
			b.envs[AWS_SECRET_ACCESS_KEY] = val
		} else {
			b.Error = err
			continue
		}

		if b.region != "" {
			b.envs[AWS_DEFAULT_REGION] = b.region
		}

	case storage.ProviderGCS:
		b.envs[RESTIC_REPOSITORY] = fmt.Sprintf("gs:%s:/%s", b.bucket, filepath.Join(b.path, b.Directory))
		if val, err := w.getSecretKey(b.storageSecret, GOOGLE_SERVICE_ACCOUNT_JSON_KEY, false); err == nil {
			b.envs[GOOGLE_SERVICE_ACCOUNT_JSON_KEY] = val
		} else {
			b.Error = err
			continue
		}

		if w.isSecretKeyExist(b.storageSecret, GOOGLE_SERVICE_ACCOUNT_JSON_KEY) {
			filePath, err := w.writeSecretKeyToFile(tmpDir, b.storageSecret, GOOGLE_SERVICE_ACCOUNT_JSON_KEY, GOOGLE_SERVICE_ACCOUNT_JSON_KEY)
			if err != nil {

				b.Error = err
				continue
			}
			w.sh.SetEnv(GOOGLE_APPLICATION_CREDENTIALS, filePath)
		}
	}
}
