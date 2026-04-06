package testutil

import (
	"errors"
	"testing"
	"time"
)

func TestFullMockRedisClient_SetString(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		err := m.SetString("key", "value", 0)
		AssertNoError(t, err, "SetString")
		AssertEqual(t, m.Calls[0], "SetString", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.SetStringError = errTest
		err := m.SetString("key", "value", 0)
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_SetTTL(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		err := m.SetTTL("key", 10*time.Second)
		AssertNoError(t, err, "SetTTL")
		AssertEqual(t, m.Calls[0], "SetTTL", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.SetTTLError = errTest
		err := m.SetTTL("key", 10*time.Second)
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_Rename(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		err := m.Rename("old", "new")
		AssertNoError(t, err, "Rename")
		AssertEqual(t, m.Calls[0], "Rename", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.RenameError = errTest
		err := m.Rename("old", "new")
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_Copy(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		err := m.Copy("src", "dst", false)
		AssertNoError(t, err, "Copy")
		AssertEqual(t, m.Calls[0], "Copy", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.CopyError = errTest
		err := m.Copy("src", "dst", true)
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_DeleteKeys(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		n, err := m.DeleteKeys("a", "b")
		AssertNoError(t, err, "DeleteKeys")
		AssertEqual(t, n, int64(0), "DeleteKeys result")
		AssertEqual(t, m.Calls[0], "DeleteKeys", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.DeleteKeysError = errTest
		_, err := m.DeleteKeys("a")
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_BulkDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.BulkDeleteResult = 5
		n, err := m.BulkDelete("user:*")
		AssertNoError(t, err, "BulkDelete")
		AssertEqual(t, n, 5, "BulkDelete result")
		AssertEqual(t, m.Calls[0], "BulkDelete", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.BulkDeleteError = errTest
		_, err := m.BulkDelete("user:*")
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_BatchSetTTL(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.BatchTTLResult = 10
		n, err := m.BatchSetTTL("user:*", 60*time.Second)
		AssertNoError(t, err, "BatchSetTTL")
		AssertEqual(t, n, 10, "BatchSetTTL result")
		AssertEqual(t, m.Calls[0], "BatchSetTTL", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.BatchSetTTLError = errTest
		_, err := m.BatchSetTTL("user:*", 60*time.Second)
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_SetIncludeTypes(t *testing.T) {
	m := NewFullMockRedisClient()
	m.SetIncludeTypes(true)
	AssertEqual(t, m.Calls[0], "SetIncludeTypes", "call name")
}
