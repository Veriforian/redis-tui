package redis

import (
	"fmt"
	"time"

	"github.com/davidbudnick/redis-tui/internal/types"
	"github.com/redis/go-redis/v9"
)

// ExportKeys exports keys matching a pattern to a map
func (c *Client) ExportKeys(pattern string) (map[string]interface{}, error) {
	allKeys, err := c.scanAll(pattern, 100)
	if err != nil {
		return nil, err
	}

	export := make(map[string]interface{})
	chunkSize := 100

	for i := 0; i < len(allKeys); i += chunkSize {
		end := min(i+chunkSize, len(allKeys))
		chunk := allKeys[i:end]

		// Pipeline TYPE + TTL for each key in chunk
		pipe := c.pipeline()
		typeCmds := make([]*redis.StatusCmd, len(chunk))
		ttlCmds := make([]*redis.DurationCmd, len(chunk))
		for j, key := range chunk {
			typeCmds[j] = pipe.Type(c.ctx, key)
			ttlCmds[j] = pipe.TTL(c.ctx, key)
		}
		_, _ = pipe.Exec(c.ctx)

		// Group keys by type for value fetching
		type keyMeta struct {
			key     string
			keyType string
			ttl     time.Duration
		}
		metas := make([]keyMeta, 0, len(chunk))
		for j, key := range chunk {
			kt := typeCmds[j].Val()
			ttl := ttlCmds[j].Val()
			if ttl < 0 {
				ttl = 0
			}
			metas = append(metas, keyMeta{key: key, keyType: kt, ttl: ttl})
		}

		// Pipeline value fetches grouped by type
		pipe = c.pipeline()
		type valueFetch struct {
			meta keyMeta
			cmd  interface{}
		}
		fetches := make([]valueFetch, 0, len(metas))
		for _, m := range metas {
			var cmd interface{}
			switch m.keyType {
			case "string":
				cmd = pipe.Get(c.ctx, m.key)
			case "list":
				cmd = pipe.LRange(c.ctx, m.key, 0, -1)
			case "set":
				cmd = pipe.SMembers(c.ctx, m.key)
			case "zset":
				cmd = pipe.ZRangeWithScores(c.ctx, m.key, 0, -1)
			case "hash":
				cmd = pipe.HGetAll(c.ctx, m.key)
			case "stream":
				cmd = pipe.XRange(c.ctx, m.key, "-", "+")
			default:
				continue
			}
			fetches = append(fetches, valueFetch{meta: m, cmd: cmd})
		}
		_, _ = pipe.Exec(c.ctx)

		// Collect results
		for _, f := range fetches {
			keyData := map[string]interface{}{
				"type": f.meta.keyType,
				"ttl":  f.meta.ttl.Seconds(),
			}

			switch f.meta.keyType {
			case "string":
				if cmd, ok := f.cmd.(*redis.StringCmd); ok && cmd.Err() == nil {
					keyData["value"] = cmd.Val()
				} else {
					continue
				}
			case "list":
				if cmd, ok := f.cmd.(*redis.StringSliceCmd); ok && cmd.Err() == nil {
					keyData["value"] = cmd.Val()
				} else {
					continue
				}
			case "set":
				if cmd, ok := f.cmd.(*redis.StringSliceCmd); ok && cmd.Err() == nil {
					keyData["value"] = cmd.Val()
				} else {
					continue
				}
			case "zset":
				if cmd, ok := f.cmd.(*redis.ZSliceCmd); ok && cmd.Err() == nil {
					members := make([]map[string]interface{}, len(cmd.Val()))
					for k, z := range cmd.Val() {
						members[k] = map[string]interface{}{"member": z.Member, "score": z.Score}
					}
					keyData["value"] = members
				} else {
					continue
				}
			case "hash":
				if cmd, ok := f.cmd.(*redis.MapStringStringCmd); ok && cmd.Err() == nil {
					keyData["value"] = cmd.Val()
				} else {
					continue
				}
			case "stream":
				if cmd, ok := f.cmd.(*redis.XMessageSliceCmd); ok && cmd.Err() == nil {
					entries := make([]map[string]interface{}, len(cmd.Val()))
					for k, e := range cmd.Val() {
						entries[k] = map[string]interface{}{"id": e.ID, "fields": e.Values}
					}
					keyData["value"] = entries
				} else {
					continue
				}
			}

			export[f.meta.key] = keyData
		}
	}

	return export, nil
}

// ImportKeys imports keys from a map
func (c *Client) ImportKeys(data map[string]interface{}) (int, error) {
	count := 0

	for key, keyDataRaw := range data {
		keyData, ok := keyDataRaw.(map[string]interface{})
		if !ok {
			continue
		}

		keyType, _ := keyData["type"].(string)
		ttlSecs, _ := keyData["ttl"].(float64)
		ttl := time.Duration(ttlSecs) * time.Second

		switch keyType {
		case "string":
			if val, ok := keyData["value"].(string); ok {
				_ = c.SetString(key, val, ttl)
				count++
			}
		case "list":
			if vals, ok := keyData["value"].([]interface{}); ok {
				strs := make([]string, 0, len(vals))
				for _, v := range vals {
					if s, ok := v.(string); ok {
						strs = append(strs, s)
					}
				}
				if len(strs) > 0 {
					_ = c.RPush(key, strs...)
				}
				if ttl > 0 {
					_ = c.SetTTL(key, ttl)
				}
				count++
			}
		case "set":
			if vals, ok := keyData["value"].([]interface{}); ok {
				strs := make([]string, 0, len(vals))
				for _, v := range vals {
					if s, ok := v.(string); ok {
						strs = append(strs, s)
					}
				}
				if len(strs) > 0 {
					_ = c.SAdd(key, strs...)
				}
				if ttl > 0 {
					_ = c.SetTTL(key, ttl)
				}
				count++
			}
		case "zset":
			if vals, ok := keyData["value"].([]interface{}); ok {
				members := make([]redis.Z, 0, len(vals))
				for _, v := range vals {
					if m, ok := v.(map[string]interface{}); ok {
						member, _ := m["member"].(string)
						score, _ := m["score"].(float64)
						members = append(members, redis.Z{Score: score, Member: member})
					}
				}
				if len(members) > 0 {
					_ = c.ZAddBatch(key, members...)
				}
				if ttl > 0 {
					_ = c.SetTTL(key, ttl)
				}
				count++
			}
		case "hash":
			if vals, ok := keyData["value"].(map[string]interface{}); ok {
				fields := make(map[string]string, len(vals))
				for field, val := range vals {
					if s, ok := val.(string); ok {
						fields[field] = s
					}
				}
				if len(fields) > 0 {
					_ = c.HSetMap(key, fields)
				}
				if ttl > 0 {
					_ = c.SetTTL(key, ttl)
				}
				count++
			}
		}
	}

	return count, nil
}

// CompareKeys compares two keys and returns their values
func (c *Client) CompareKeys(key1, key2 string) (types.RedisValue, types.RedisValue, error) {
	val1, err := c.GetValue(key1)
	if err != nil {
		return types.RedisValue{}, types.RedisValue{}, fmt.Errorf("error getting key1: %w", err)
	}

	val2, err := c.GetValue(key2)
	if err != nil {
		return val1, types.RedisValue{}, fmt.Errorf("error getting key2: %w", err)
	}

	return val1, val2, nil
}
