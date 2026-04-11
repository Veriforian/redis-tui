package secret

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestFileStore_Lifecycle(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.enc")
	masterPassword := []byte("super_secure_master_password")

	t.Run("Create new vault", func(t *testing.T) {
		store, err := NewFileStore(vaultPath, masterPassword)
		if err != nil {
			t.Fatalf("failed to create FileStore: %v", err)
		}
		if store == nil {
			t.Fatal("store is nil")
		}
		if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
			t.Fatal("vault file was not created on disk")
		}
	})

	t.Run("Set and Get secret", func(t *testing.T) {
		store, _ := NewFileStore(vaultPath, masterPassword)

		err := store.Set("redis-prod", "user1", []byte("my_redis_pass"))
		if err != nil {
			t.Fatalf("failed to set secret: %v", err)
		}

		store2, err := NewFileStore(vaultPath, masterPassword)
		if err != nil {
			t.Fatalf("failed to open existing vault: %v", err)
		}

		retrieved, err := store2.Get("redis-prod", "user1")
		if err != nil {
			t.Fatalf("failed to get secret: %v", err)
		}
		if !bytes.Equal(retrieved, []byte("my_redis_pass")) {
			t.Errorf("expected 'my_redis_pass', got %s", string(retrieved))
		}
	})

	t.Run("Get missing secret", func(t *testing.T) {
		store, _ := NewFileStore(vaultPath, masterPassword)
		_, err := store.Get("redis-prod", "unknown_user")
		if err != ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Invalid master password", func(t *testing.T) {
		_, err := NewFileStore(vaultPath, []byte("wrong_password"))
		if err == nil {
			t.Fatal("expected error when using wrong master password")
		}
	})

	t.Run("Delete secret", func(t *testing.T) {
		store, _ := NewFileStore(vaultPath, masterPassword)

		err := store.Delete("redis-prod", "user1")
		if err != nil {
			t.Fatalf("failed to delete secret: %v", err)
		}

		_, err = store.Get("redis-prod", "user1")
		if err != ErrNotFound {
			t.Errorf("expected ErrNotFound after deletion, got %v", err)
		}
	})
}

func TestFileStore_CorruptedFile(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.enc")

	// Write garbage data
	if err := os.WriteFile(vaultPath, []byte("too_short"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := NewFileStore(vaultPath, []byte("pass"))
	if err == nil || err.Error() != "vault file is corrupted" {
		t.Errorf("expected corruption error, got %v", err)
	}
}
