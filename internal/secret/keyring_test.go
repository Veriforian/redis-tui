package secret

import (
	"bytes"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestKeyringStore_Lifecycle(t *testing.T) {
	keyring.MockInit()

	store := NewKeyringStore()

	t.Run("Get not found", func(t *testing.T) {
		_, err := store.Get("svc", "user")
		if err != ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Set and Get", func(t *testing.T) {
		err := store.Set("redis-local", "admin", []byte("secret123"))
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		res, err := store.Get("redis-local", "admin")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if !bytes.Equal(res, []byte("secret123")) {
			t.Errorf("expected 'secret123', got %s", string(res))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := store.Delete("redis-local", "admin")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = store.Get("redis-local", "admin")
		if err != ErrNotFound {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})
}
