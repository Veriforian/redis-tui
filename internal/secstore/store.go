package secstore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/99designs/keyring"
)

const (
	serviceName   = "redis-tui"
	localKeyFile  = "master.key"
	masterKeyName = "master-key"
	vaultFileName = "secrets.enc"
)

// Store stores secrets in memory and on disk.
type Store struct {
	mu        sync.RWMutex
	kr        keyring.Keyring
	configDir string
	masterKey []byte
	secrets   map[string]string
	vaultPath string
}

func (s *Store) Close() error {
	return nil
}

// NewStore creates a new Store instance.
func NewStore(configDir string, userSecret string) (*Store, error) {
	kr, err := keyring.Open(keyring.Config{
		ServiceName: serviceName,
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.SecretServiceBackend,
			keyring.PassBackend,
			keyring.FileBackend,
		},
	})

	if err == nil {
		var masterKey []byte
		item, err := kr.Get(masterKeyName)
		if err != nil {
			if errors.Is(err, keyring.ErrKeyNotFound) {
				masterKey = make([]byte, 32)
				if _, err := io.ReadFull(rand.Reader, masterKey); err != nil {
					return nil, err
				}

				if err := kr.Set(keyring.Item{
					Key:  masterKeyName,
					Data: masterKey,
				}); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}

		return &Store{
			mu:        sync.RWMutex{},
			kr:        kr,
			configDir: configDir,
			vaultPath: filepath.Join(configDir, vaultFileName),
			secrets:   make(map[string]string),
			masterKey: item.Data,
		}, nil
	}

	var encryptionKey string

	if userSecret != "" {
		salt := []byte("redis-tui-salt")
		keyBytes, _ := pbkdf2.Key(sha256.New, userSecret, salt, 100000, 32)
		encryptionKey = base64.StdEncoding.EncodeToString(keyBytes)
	} else {
		key, err := ensureLocalKey(filepath.Join(configDir, localKeyFile))
		if err != nil {
			return nil, err
		}
		encryptionKey = base64.StdEncoding.EncodeToString(key)
	}

	krFile, err := keyring.Open(keyring.Config{
		ServiceName: serviceName,
		AllowedBackends: []keyring.BackendType{
			keyring.FileBackend,
		},
		FileDir: configDir,
		FilePasswordFunc: func(s string) (string, error) {
			return encryptionKey, nil
		},
	})
	if err != nil {
		return nil, err
	}

	return &Store{kr: krFile, configDir: configDir}, nil
}

// ensureLocalKey ensures a local key exists at the given path.
func ensureLocalKey(path string) ([]byte, error) {
	if key, err := os.ReadFile(path); err == nil {
		return key, nil
	}

	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, key, 0o600); err != nil {
		return nil, err
	}

	return key, nil
}

// Save saves a secret, commiting it to disk. Falls back to local keyring if not using masterkey.
func (s *Store) Save(connectionID string, password string) error {
	if s.masterKey != nil {
		s.Set(connectionID, password)
		err := s.Commit()
		if err != nil {
			return err
		}

		return nil
	}

	// Fallback to local keyring
	if password == "" {
		err := s.kr.Remove(connectionID)
		if err != nil {
			if errors.Is(err, keyring.ErrKeyNotFound) {
				return nil // Normal, nothing saved to remove or save
			}

			return err
		}
	}
	return s.kr.Set(keyring.Item{
		Key:  connectionID,
		Data: []byte(password),
	})
}

// Load retrieves a secret from memory instantly, falls back to keyring if not using masterkey.
func (s *Store) Load(connectionID string) (string, error) {
	if s.masterKey != nil {
		item, err := s.Get(connectionID)
		if err != nil {
			return "", err
		}

		return item, nil
	} else {

		item, err := s.kr.Get(connectionID)
		if err != nil {
			if errors.Is(err, keyring.ErrKeyNotFound) {
				return "", nil // Normal, connection has no password
			}
			return "", err
		}
		return string(item.Data), nil
	}
}

// Set updates a secret in memory. You must call Commit() to save to disk.
func (s *Store) Set(connectionID string, password string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if password == "" {
		delete(s.secrets, connectionID)
	} else {
		s.secrets[connectionID] = password
	}
}

// Get retrieves a secret from memory instantly, falls back to keyring if not using masterkey.
func (s *Store) Get(connectionID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.secrets[connectionID], nil
}

// Commit encrypts the current state of the secrets map and writes it to disk.
func (s *Store) Commit() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.secrets) == 0 {
		_ = os.Remove(s.vaultPath)
		return nil
	}

	plaintext, err := json.Marshal(s.secrets)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)

	return os.WriteFile(s.vaultPath, ciphertext, 0o600)
}

// LoadVault loads the encrypted secrets from disk.
func (s *Store) LoadVault() error {
	data, err := os.ReadFile(s.vaultPath)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonceSize := aesgcm.NonceSize()
	if len(data) < nonceSize {
		return errors.New("invalid ciphertext")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	return json.Unmarshal(plaintext, &s.secrets)
}
