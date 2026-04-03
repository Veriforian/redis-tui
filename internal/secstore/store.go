package secstore

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/99designs/keyring"
)

const (
	serviceName   = "redis-tui"
	localKeyFile  = "master.key"
	localDataFile = "secrets.enc"
)

type Store struct {
	kr         keyring.Keyring
	configPath string
}

func NewStore(configPath string, userSecret string) (*Store, error) {
	kr, err := keyring.Open(keyring.Config{
		ServiceName: serviceName,
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.SecretServiceBackend,
			keyring.PassBackend,
		},
	})

	if err == nil {
		testKey := "redis-tui-test"
		_ = kr.Set(keyring.Item{Key: testKey, Data: []byte("ping")})

		if checkErr := kr.Remove(testKey); checkErr == nil {
			return &Store{kr: kr, configPath: configPath}, nil
		}
	}

	var encryptionKey string

	if userSecret != "" {
		salt := []byte("redis-tui-salt")
		keyBytes, _ := pbkdf2.Key(sha256.New, userSecret, salt, 100000, 32)
		encryptionKey = base64.StdEncoding.EncodeToString(keyBytes)
	} else {
		key, err := ensureLocalKey(filepath.Join(configPath, localKeyFile))
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
		FileDir: configPath,
		FilePasswordFunc: func(s string) (string, error) {
			return encryptionKey, nil
		},
	})
	if err != nil {
		return nil, err
	}

	return &Store{kr: krFile, configPath: configPath}, nil
}

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

func (s *Store) Save(connectionID string, password string) error {
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

func (s *Store) Load(connectionID string) (string, error) {
	item, err := s.kr.Get(connectionID)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", nil // Normal, connection has no password
		}
		return "", err
	}
	return string(item.Data), nil
}
