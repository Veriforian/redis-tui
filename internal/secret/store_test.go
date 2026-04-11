package secret

import (
	"bytes"
	"errors"
	"testing"

	"github.com/davidbudnick/redis-tui/internal/service"
)

// mockStore implements service.StoreService for testing ChainStore routing.
type mockStore struct {
	name   string
	getRes []byte
	getErr error
	setErr error
	delErr error

	// Track calls
	getCalled bool
	setCalled bool
	delCalled bool
}

func (m *mockStore) Name() string { return m.name }
func (m *mockStore) Get(service, user string) ([]byte, error) {
	m.getCalled = true
	return m.getRes, m.getErr
}

func (m *mockStore) Set(service, user string, pwd []byte) error {
	m.setCalled = true
	return m.setErr
}

func (m *mockStore) Delete(service, user string) error {
	m.delCalled = true
	return m.delErr
}

func TestChainStore_Get(t *testing.T) {
	tests := []struct {
		name          string
		stores        []*mockStore
		expectedBytes []byte
		expectedErr   error
	}{
		{
			name: "first store succeeds",
			stores: []*mockStore{
				{name: "store1", getRes: []byte("pass1"), getErr: nil},
				{name: "store2", getRes: []byte("pass2"), getErr: nil},
			},
			expectedBytes: []byte("pass1"),
			expectedErr:   nil,
		},
		{
			name: "fallback to second store on ErrNotFound",
			stores: []*mockStore{
				{name: "store1", getErr: ErrNotFound},
				{name: "store2", getRes: []byte("pass2"), getErr: nil},
			},
			expectedBytes: []byte("pass2"),
			expectedErr:   nil,
		},
		{
			name: "fallback on ErrUnavailable",
			stores: []*mockStore{
				{name: "store1", getErr: ErrUnavailable},
				{name: "store2", getRes: []byte("pass2"), getErr: nil},
			},
			expectedBytes: []byte("pass2"),
			expectedErr:   nil,
		},
		{
			name: "fails immediately on critical error",
			stores: []*mockStore{
				{name: "store1", getErr: errors.New("critical DB failure")},
				{name: "store2", getRes: []byte("pass2"), getErr: nil}, // Should not be reached
			},
			expectedBytes: nil,
			expectedErr:   errors.New("store1 provider failed: critical DB failure"),
		},
		{
			name: "all stores fail normally",
			stores: []*mockStore{
				{name: "store1", getErr: ErrUnavailable},
				{name: "store2", getErr: ErrNotFound},
			},
			expectedBytes: nil,
			expectedErr:   ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert mockStore to interfaces
			interfaces := make([]service.StoreService, len(tt.stores))
			for i, s := range tt.stores {
				interfaces[i] = s
			}

			chain := NewChainStore(interfaces...)
			res, err := chain.Get("svc", "usr")

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					if !errors.Is(err, tt.expectedErr) {
						t.Errorf("expected error %v, got %v", tt.expectedErr, err)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !bytes.Equal(res, tt.expectedBytes) {
					t.Errorf("expected %s, got %s", tt.expectedBytes, res)
				}
			}
		})
	}
}

func TestChainStore_Set(t *testing.T) {
	t.Run("sets on first available store", func(t *testing.T) {
		s1 := &mockStore{name: "s1", setErr: ErrUnavailable}
		s2 := &mockStore{name: "s2", setErr: nil}
		chain := NewChainStore(s1, s2)

		err := chain.Set("svc", "usr", []byte("pass"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !s1.setCalled || !s2.setCalled {
			t.Errorf("expected both stores to be called, s1: %v, s2: %v", s1.setCalled, s2.setCalled)
		}
	})

	t.Run("fails if all unavailable", func(t *testing.T) {
		s1 := &mockStore{name: "s1", setErr: ErrUnavailable}
		chain := NewChainStore(s1)

		err := chain.Set("svc", "usr", []byte("pass"))
		if !errors.Is(err, ErrUnavailable) {
			t.Errorf("expected ErrUnavailable, got %v", err)
		}
	})
}

func TestChainStore_Delete(t *testing.T) {
	t.Run("delete on first available store", func(t *testing.T) {
		d1 := &mockStore{name: "d1", delErr: ErrUnavailable}
		d2 := &mockStore{name: "d2", delErr: nil}
		chain := NewChainStore(d1, d2)

		err := chain.Delete("svc", "usr")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !d1.delCalled || !d2.delCalled {
			t.Errorf("expected both stores to be called, s1: %v, s2: %v", d1.delCalled, d2.delCalled)
		}
	})

	t.Run("fails if all unavailable", func(t *testing.T) {
		d1 := &mockStore{name: "d1", delErr: ErrUnavailable}
		chain := NewChainStore(d1)

		err := chain.Delete("svc", "usr")
		if !errors.Is(err, ErrUnavailable) {
			t.Errorf("expected ErrUnavailable, got %v", err)
		}
	})
}
