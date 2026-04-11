package secret

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/zalando/go-keyring"
)

const (
	vaultServiceName = "redis-tui"
	vaultAccountName = "master-vault"
)

// KeyringStore implements the Store interface using the OS native keyring.
type KeyringStore struct {
	mu sync.RWMutex // Protects concurrent map and keyring access
}

func NewKeyringStore() *KeyringStore {
	return &KeyringStore{}
}

// Name returns the name of the secret store provider.
func (s *KeyringStore) Name() string {
	return "OS_Keyring"
}

// Get retrieves a secret from the store.
func (s *KeyringStore) Get(service, user string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, err := s.load()
	if err != nil {
		return nil, err
	}

	users, ok := v.Credentials[service]
	if !ok {
		return nil, ErrNotFound
	}

	pwd, ok := users[user]
	if !ok {
		return nil, ErrNotFound
	}

	out := make([]byte, len(pwd))
	copy(out, pwd)
	return out, nil
}

// Set stores a secret in the store.
func (s *KeyringStore) Set(service, user string, pwd []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, err := s.load()
	if err != nil {
		return err
	}

	if _, ok := v.Credentials[service]; !ok {
		v.Credentials[service] = make(map[string][]byte)
	}

	pwdCopy := make([]byte, len(pwd))
	copy(pwdCopy, pwd)
	v.Credentials[service][user] = pwdCopy
	return s.save(v)
}

// Delete removes a secret from the store.
func (s *KeyringStore) Delete(service, user string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, err := s.load()
	if err != nil {
		return err
	}

	if users, ok := v.Credentials[service]; ok {
		delete(users, user)
		// Clean up map if empty
		if len(users) == 0 {
			delete(v.Credentials, service)
		}
		return s.save(v)
	}

	return ErrNotFound
}

// load fetches and unmarshals the vault from the keyring.
// Assumes the caller holds the appropriate lock.
func (s *KeyringStore) load() (*vault, error) {
	data, err := keyring.Get(vaultServiceName, vaultAccountName)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return &vault{Credentials: make(map[string]map[string][]byte)}, nil
		}
		return nil, ErrUnavailable
	}

	var v vault
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		return nil, errors.New("failed to decode keyring vault")
	}

	if v.Credentials == nil {
		v.Credentials = make(map[string]map[string][]byte)
	}

	return &v, nil
}

// save marshals and writes the vault to the OS keyring.
// Assumes the caller holds the appropriate lock.
func (s *KeyringStore) save(v *vault) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if err := keyring.Set(vaultServiceName, vaultAccountName, string(data)); err != nil {
		return ErrUnavailable
	}
	return nil
}
