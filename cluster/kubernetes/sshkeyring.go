package kubernetes

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/weaveworks/flux/ssh"
)

const (
	// The private key file must have these permissions, or ssh will refuse to
	// use it
	privateKeyFileMode = os.FileMode(0400)
)

// SSHKeyRingConfig is used to configure the keyring with key generation
// options and the parameters of its backing kubernetes secret resource.
// SecretVolumeMountPath must be mounted RW for regenerate() to work, and to
// set the privateKeyFileMode on the identity secret file.
type SSHKeyRingConfig struct {
	SecretAPI             v1.SecretInterface
	SecretName            string
	SecretVolumeMountPath string // e.g. "/etc/fluxd/ssh"
	SecretDataKey         string // e.g. "identity"
	KeyBits               ssh.OptionalValue
	KeyType               ssh.OptionalValue
}

type sshKeyRing struct {
	sync.RWMutex
	SSHKeyRingConfig
	publicKey              ssh.PublicKey
	expectedPrivateKeyPath string
	realPrivateKeyPath     string
}

// NewSSHKeyRing constructs an sshKeyRing backed by a kubernetes secret
// resource. The keyring is initialised with the key that was previously stored
// in the secret (either by regenerate() or an administrator), or a freshly
// generated key if none was found.
func NewSSHKeyRing(config SSHKeyRingConfig) (*sshKeyRing, error) {
	skr := &sshKeyRing{SSHKeyRingConfig: config}
	skr.expectedPrivateKeyPath = filepath.Join(skr.SecretVolumeMountPath, skr.SecretDataKey)

	fileInfo, err := os.Stat(skr.expectedPrivateKeyPath)
	switch {
	case os.IsNotExist(err):
		if err := skr.Regenerate(); err != nil {
			return nil, err
		}
		skr.publicKey, skr.realPrivateKeyPath = skr.KeyPair()
	case err != nil:
		return nil, err
	case fileInfo.Mode() != privateKeyFileMode:
		if err := os.Chmod(skr.expectedPrivateKeyPath, privateKeyFileMode); err != nil {
			return nil, err
		}
		fallthrough
	default:
		publicKey, err := ssh.ExtractPublicKey(skr.expectedPrivateKeyPath)
		if err != nil {
			return nil, err
		}
		skr.realPrivateKeyPath = skr.expectedPrivateKeyPath
		skr.publicKey = publicKey
	}

	return skr, nil
}

// KeyPair returns the current public key and the path to its corresponding
// private key. The private key file is guaranteed to exist for the lifetime of
// the process, however as the returned pair can be discarded from the keyring
// at any time by use of the regenerate() method it is inadvisable to cache the
// results for long periods; instead request the key pair from the ring
// immediately prior to each use.
func (skr *sshKeyRing) KeyPair() (publicKey ssh.PublicKey, privateKeyPath string) {
	skr.RLock()
	defer skr.RUnlock()
	return skr.publicKey, skr.expectedPrivateKeyPath
}

// regenerate creates a new keypair in the configured SecretVolumeMountPath and
// updates the kubernetes secret resource with the private key so that it will
// be available to the keyring after restart. If this operation is successful
// the keyPair() method will return the new pair; if it fails for any reason,
// keyPair() will continue to return the existing pair.
//
// BUG(awh) Updating the kubernetes secret should be done via an ephemeral
// external executable invoked with coredumps disabled and using
// syscall.Mlockall(MCL_FUTURE) in conjunction with an appropriate ulimit to
// ensure the private key isn't unintentionally written to persistent storage.
func (skr *sshKeyRing) Regenerate() error {
	privateKeyPath, privateKey, publicKey, err := ssh.KeyGen(skr.KeyBits, skr.KeyType, skr.SecretVolumeMountPath)
	if err != nil {
		return err
	}

	// Prepare a symlink pointing at the new key, to be moved later.
	tmpSymlinkPath := filepath.Join(filepath.Dir(privateKeyPath), "tmp-identity")
	if err = os.Symlink(privateKeyPath, tmpSymlinkPath); err != nil {
		return err
	}
	if err = os.Chmod(tmpSymlinkPath, privateKeyFileMode); err != nil {
		return err
	}

	patch := map[string]map[string]string{
		"data": map[string]string{
			"identity": base64.StdEncoding.EncodeToString(privateKey),
		},
	}

	jsonPatch, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	_, err = skr.SecretAPI.Patch(skr.SecretName, types.StrategicMergePatchType, jsonPatch)
	if err != nil {
		return err
	}

	// The secret is updated, and Kubernetes will eventually make sure
	// it's mounted and that `identity` points at it. In the meantime,
	// change the symlink to point to our copy of it.
	if err = os.Rename(tmpSymlinkPath, skr.expectedPrivateKeyPath); err != nil {
		os.Remove(tmpSymlinkPath)
		return err
	}

	skr.Lock()
	skr.realPrivateKeyPath = privateKeyPath
	skr.publicKey = publicKey
	skr.Unlock()

	return nil
}
