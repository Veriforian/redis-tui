package testutil

import (
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestFullMockRedisClient_ListOps(t *testing.T) {
	tests := []struct {
		name    string
		setErr  func(*FullMockRedisClient)
		call    func(*FullMockRedisClient) error
		logName string
	}{
		{"RPush success", nil, func(m *FullMockRedisClient) error { return m.RPush("list", "a", "b") }, "RPush"},
		{"RPush error", func(m *FullMockRedisClient) { m.RPushError = errTest }, func(m *FullMockRedisClient) error { return m.RPush("list", "a") }, "RPush"},
		{"LSet success", nil, func(m *FullMockRedisClient) error { return m.LSet("list", 0, "val") }, "LSet"},
		{"LSet error", func(m *FullMockRedisClient) { m.LSetError = errTest }, func(m *FullMockRedisClient) error { return m.LSet("list", 0, "val") }, "LSet"},
		{"LRem success", nil, func(m *FullMockRedisClient) error { return m.LRem("list", 1, "val") }, "LRem"},
		{"LRem error", func(m *FullMockRedisClient) { m.LRemError = errTest }, func(m *FullMockRedisClient) error { return m.LRem("list", 1, "val") }, "LRem"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewFullMockRedisClient()
			if tt.setErr != nil {
				tt.setErr(m)
			}
			err := tt.call(m)
			if tt.setErr != nil {
				if !errors.Is(err, errTest) {
					t.Errorf("expected errTest, got %v", err)
				}
			} else {
				AssertNoError(t, err, tt.logName)
			}
			AssertEqual(t, m.Calls[0], tt.logName, "call name")
		})
	}
}

func TestFullMockRedisClient_SetOps(t *testing.T) {
	tests := []struct {
		name    string
		setErr  func(*FullMockRedisClient)
		call    func(*FullMockRedisClient) error
		logName string
	}{
		{"SAdd success", nil, func(m *FullMockRedisClient) error { return m.SAdd("set", "a", "b") }, "SAdd"},
		{"SAdd error", func(m *FullMockRedisClient) { m.SAddError = errTest }, func(m *FullMockRedisClient) error { return m.SAdd("set", "a") }, "SAdd"},
		{"SRem success", nil, func(m *FullMockRedisClient) error { return m.SRem("set", "a") }, "SRem"},
		{"SRem error", func(m *FullMockRedisClient) { m.SRemError = errTest }, func(m *FullMockRedisClient) error { return m.SRem("set", "a") }, "SRem"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewFullMockRedisClient()
			if tt.setErr != nil {
				tt.setErr(m)
			}
			err := tt.call(m)
			if tt.setErr != nil {
				if !errors.Is(err, errTest) {
					t.Errorf("expected errTest, got %v", err)
				}
			} else {
				AssertNoError(t, err, tt.logName)
			}
			AssertEqual(t, m.Calls[0], tt.logName, "call name")
		})
	}
}

func TestFullMockRedisClient_ZSetOps(t *testing.T) {
	t.Run("ZAdd success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		err := m.ZAdd("zset", 1.5, "member")
		AssertNoError(t, err, "ZAdd")
		AssertEqual(t, m.Calls[0], "ZAdd", "call name")
	})
	t.Run("ZAdd error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.ZAddError = errTest
		err := m.ZAdd("zset", 1.5, "member")
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
	t.Run("ZRem success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		err := m.ZRem("zset", "member")
		AssertNoError(t, err, "ZRem")
		AssertEqual(t, m.Calls[0], "ZRem", "call name")
	})
	t.Run("ZRem error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.ZRemError = errTest
		err := m.ZRem("zset", "member")
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_HashOps(t *testing.T) {
	t.Run("HSet success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		err := m.HSet("hash", "field", "value")
		AssertNoError(t, err, "HSet")
		AssertEqual(t, m.Calls[0], "HSet", "call name")
	})
	t.Run("HSet error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.HSetError = errTest
		err := m.HSet("hash", "field", "value")
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
	t.Run("HDel success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		err := m.HDel("hash", "field1", "field2")
		AssertNoError(t, err, "HDel")
		AssertEqual(t, m.Calls[0], "HDel", "call name")
	})
	t.Run("HDel error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.HDelError = errTest
		err := m.HDel("hash", "field1")
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_StreamOps(t *testing.T) {
	t.Run("XAdd success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.XAddResult = "1234-0"
		id, err := m.XAdd("stream", map[string]any{"k": "v"})
		AssertNoError(t, err, "XAdd")
		AssertEqual(t, id, "1234-0", "XAdd result")
		AssertEqual(t, m.Calls[0], "XAdd", "call name")
	})
	t.Run("XAdd error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.XAddError = errTest
		_, err := m.XAdd("stream", map[string]any{"k": "v"})
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
	t.Run("XDel success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		err := m.XDel("stream", "1234-0")
		AssertNoError(t, err, "XDel")
		AssertEqual(t, m.Calls[0], "XDel", "call name")
	})
	t.Run("XDel error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.XDelError = errTest
		err := m.XDel("stream", "1234-0")
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_PFAdd(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		err := m.PFAdd("hll", "a", "b", "c")
		AssertNoError(t, err, "PFAdd")
		AssertEqual(t, m.Calls[0], "PFAdd", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.PFAddError = errTest
		err := m.PFAdd("hll", "a")
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_PFCount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.PFCountResult = 42
		got, err := m.PFCount("hll")
		AssertNoError(t, err, "PFCount")
		AssertEqual(t, got, int64(42), "PFCount result")
		AssertEqual(t, m.Calls[0], "PFCount", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.PFCountError = errTest
		_, err := m.PFCount("hll")
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_SetBit(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		err := m.SetBit("bitmap", 7, 1)
		AssertNoError(t, err, "SetBit")
		AssertEqual(t, m.Calls[0], "SetBit", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.SetBitError = errTest
		err := m.SetBit("bitmap", 7, 1)
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_GetBit(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		got, err := m.GetBit("bitmap", 7)
		AssertNoError(t, err, "GetBit")
		AssertEqual(t, got, int64(0), "GetBit result")
		AssertEqual(t, m.Calls[0], "GetBit", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.GetBitError = errTest
		_, err := m.GetBit("bitmap", 7)
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_BitCount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.BitCountResult = 5
		got, err := m.BitCount("bitmap")
		AssertNoError(t, err, "BitCount")
		AssertEqual(t, got, int64(5), "BitCount result")
		AssertEqual(t, m.Calls[0], "BitCount", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.BitCountError = errTest
		_, err := m.BitCount("bitmap")
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_GeoAdd(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		loc := &redis.GeoLocation{Name: "place", Longitude: 1.0, Latitude: 2.0}
		err := m.GeoAdd("geo", loc)
		AssertNoError(t, err, "GeoAdd")
		AssertEqual(t, m.Calls[0], "GeoAdd", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.GeoAddError = errTest
		err := m.GeoAdd("geo", &redis.GeoLocation{})
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_GeoPos(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		pos := &redis.GeoPos{Longitude: 1.0, Latitude: 2.0}
		m.GeoPosResult = []*redis.GeoPos{pos}
		got, err := m.GeoPos("geo", "place")
		AssertNoError(t, err, "GeoPos")
		AssertSliceLen(t, got, 1, "GeoPos result")
		AssertEqual(t, got[0].Longitude, 1.0, "longitude")
		AssertEqual(t, got[0].Latitude, 2.0, "latitude")
		AssertEqual(t, m.Calls[0], "GeoPos", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.GeoPosError = errTest
		_, err := m.GeoPos("geo", "place")
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_JSONGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.JSONGetResult = `{"name":"test"}`
		got, err := m.JSONGet("key")
		AssertNoError(t, err, "JSONGet")
		AssertEqual(t, got, `{"name":"test"}`, "JSONGet result")
		AssertEqual(t, m.Calls[0], "JSONGet", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.JSONGetError = errTest
		_, err := m.JSONGet("key")
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_JSONGetPath(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.JSONGetResult = `"test"`
		got, err := m.JSONGetPath("key", "$.name")
		AssertNoError(t, err, "JSONGetPath")
		AssertEqual(t, got, `"test"`, "JSONGetPath result")
		AssertEqual(t, m.Calls[0], "JSONGetPath", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.JSONGetError = errTest
		_, err := m.JSONGetPath("key", "$.name")
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

func TestFullMockRedisClient_JSONSet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewFullMockRedisClient()
		err := m.JSONSet("key", `{"a":1}`)
		AssertNoError(t, err, "JSONSet")
		AssertEqual(t, m.Calls[0], "JSONSet", "call name")
	})

	t.Run("error", func(t *testing.T) {
		m := NewFullMockRedisClient()
		m.JSONSetError = errTest
		err := m.JSONSet("key", `{"a":1}`)
		if !errors.Is(err, errTest) {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}
