package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/argon2"
)

const (
	saltSize  = 16
	nonceSize = 12
	keySize   = 32
)

var ErrInvalidPassword = errors.New("invalid master password or corrupted vault")

type FileStore struct {
	path string
	key  []byte
	salt []byte
	mu   sync.RWMutex
}

func Name() string {
	return "Encrypted_File"
}

// Get retrieves a secret from the store.
func (s *FileStore) Get(service, user string) ([]byte, error) {
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
func (s *FileStore) Set(service, user string, pwd []byte) error {
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
func (s *FileStore) Delete(service, user string) error {
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

// NewFileStore initializes the file store. It requires the absolute path to the
// vault file and the user's master password.
func NewFileStore(path string, password []byte) (*FileStore, error) {
	if len(password) == 0 {
		return nil, errors.New("master password cannot be empty")
	}

	// Ensure dir exists, with strict permissions
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create vault directory: %w", err)
	}

	store := &FileStore{
		path: path,
		key:  make([]byte, keySize),
		salt: make([]byte, saltSize),
	}

	// Attempt to read the salt from the existing file.
	// If it fails, we'll generate a new salt and write it to the file.
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read vault file: %w", err)
		}

		if _, err := io.ReadFull(rand.Reader, store.salt); err != nil {
			return nil, fmt.Errorf("failed to generate salt: %w", err)
		}
		store.deriveKey(password)
		return store, nil
	}

	if len(data) < saltSize+nonceSize {
		return nil, errors.New("vault file is corrupted")
	}

	copy(store.salt, data[:saltSize])
	store.deriveKey(password)

	// Fail fast. Verify pass is correct by attempting decryption.
	if _, err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

// deriveKey uses Argon2id to derive a 32-byte key
func (s *FileStore) deriveKey(password []byte) {
	derived := argon2.IDKey(password, s.salt, 1, 64*1024, 4, keySize)
	copy(s.key, derived)
	// Clear the derived key slice
	clear(derived)
}

func (s *FileStore) load() (*vault, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &vault{Credentials: make(map[string]map[string][]byte)}, nil
		}
		return nil, ErrUnavailable
	}

	if len(data) < saltSize+nonceSize {
		return nil, errors.New("vault file is corrupted")
	}

	salt := data[:saltSize]
	nonce := data[saltSize : saltSize+nonceSize]
	ciphertext := data[saltSize+nonceSize:]

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Open auths and decrypts the vault. If the master password is incorrect,
	// the ciphertext will be nil and an error will be returned.
	plaintext, err := gcm.Open(nil, nonce, ciphertext, salt)
	if err != nil {
		return nil, err
	}
	defer clear(plaintext)

	var v vault
	if err := json.Unmarshal(plaintext, &v); err != nil {
		return nil, errors.New("failed to decode vault json")
	}

	if v.Credentials == nil {
		v.Credentials = make(map[string]map[string][]byte)
	}

	return &v, nil
}

func (s *FileStore) save(v *vault) error {
	plaintext, err := json.Marshal(v)
	if err != nil {
		return err
	}
	defer clear(plaintext)

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, s.salt)

	finalPayload := make([]byte, 0, len(s.salt)+len(nonce)+len(ciphertext))
	finalPayload = append(finalPayload, s.salt...)
	finalPayload = append(finalPayload, nonce...)
	finalPayload = append(finalPayload, ciphertext...)

	// Write to tmp file, then rename for atomic updates.
	tmpFile := s.path + ".tmp"
	if err := os.WriteFile(tmpFile, finalPayload, 0o600); err != nil {
		return fmt.Errorf("failed to write tmp file: %w", err)
	}
	if err := os.Rename(tmpFile, s.path); err != nil {
		return fmt.Errorf("failed to commit vault file: %w", err)
	}

	return nil
}
