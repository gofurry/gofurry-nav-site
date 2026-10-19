package uptime

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	fiberredis "github.com/gofiber/storage/redis/v3"
	"github.com/redis/go-redis/v9"
	bolt "go.etcd.io/bbolt"
)

var (
	entriesBucket = []byte("uptime_entries")
	typeField     = []byte{0, 't'}
	expiresField  = []byte{0, 'e'}
)

const (
	typeHash = "hash"
	typeSet  = "set"
	typeZSet = "zset"
)

// BoltStore exposes a deliberately narrow Redis-compatible client over Bbolt.
// Fiber uptime currently accepts a concrete Fiber Redis storage even though it
// only uses hashes, sets, sorted sets, expiry, pipelines, and three fixed Lua
// transactions. Implementing that exact surface keeps uptime self-contained
// without running or connecting to Redis.
type BoltStore struct {
	db      *bolt.DB
	client  *boltRedisClient
	storage *fiberredis.Storage
	mu      sync.Mutex
	closed  bool
}

func OpenBoltStore(path string) (*BoltStore, error) {
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, err
	}
	if err = db.Update(func(tx *bolt.Tx) error {
		_, createErr := tx.CreateBucketIfNotExists(entriesBucket)
		return createErr
	}); err != nil {
		_ = db.Close()
		return nil, err
	}
	client := &boltRedisClient{db: db}
	return &BoltStore{
		db:      db,
		client:  client,
		storage: fiberredis.NewFromConnection(client),
	}, nil
}

func (s *BoltStore) FiberStorage() *fiberredis.Storage { return s.storage }

func (s *BoltStore) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	closed := s.closed
	s.mu.Unlock()
	if closed {
		return errors.New("uptime Bbolt storage is closed")
	}
	return s.db.View(func(tx *bolt.Tx) error {
		if tx.Bucket(entriesBucket) == nil {
			return errors.New("uptime Bbolt bucket is unavailable")
		}
		return nil
	})
}

func (s *BoltStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if s.storage != nil {
		_ = s.storage.Close()
	}
	return s.db.Close()
}

type boltRedisClient struct {
	redis.UniversalClient
	db *bolt.DB
}

func (c *boltRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	if err := contextError(ctx); err != nil {
		cmd.SetErr(err)
		return cmd
	}
	err := c.db.View(func(tx *bolt.Tx) error {
		if tx.Bucket(entriesBucket) == nil {
			return errors.New("uptime Bbolt bucket is unavailable")
		}
		return nil
	})
	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal("PONG")
	}
	return cmd
}

func (c *boltRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	var added int64
	err := c.update(ctx, func(root *bolt.Bucket) error {
		entry, err := entryBucket(root, key, typeSet, true)
		if err != nil {
			return err
		}
		for _, member := range members {
			field := dataField(stringValue(member))
			if entry.Get(field) == nil {
				added++
				if err = entry.Put(field, []byte{1}); err != nil {
					return err
				}
			}
		}
		return nil
	})
	setIntResult(cmd, added, err)
	return cmd
}

func (c *boltRedisClient) SRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	var removed int64
	err := c.update(ctx, func(root *bolt.Bucket) error {
		entry, err := entryBucket(root, key, typeSet, false)
		if err != nil || entry == nil {
			return err
		}
		for _, member := range members {
			field := dataField(stringValue(member))
			if entry.Get(field) != nil {
				removed++
				if err = entry.Delete(field); err != nil {
					return err
				}
			}
		}
		return nil
	})
	setIntResult(cmd, removed, err)
	return cmd
}

func (c *boltRedisClient) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	cmd := redis.NewStringSliceCmd(ctx)
	var values []string
	err := c.update(ctx, func(root *bolt.Bucket) error {
		entry, err := entryBucket(root, key, typeSet, false)
		if err != nil || entry == nil {
			return err
		}
		return entry.ForEach(func(field, _ []byte) error {
			if isDataField(field) {
				values = append(values, string(field[1:]))
			}
			return nil
		})
	})
	if err != nil {
		cmd.SetErr(err)
	} else {
		sort.Strings(values)
		cmd.SetVal(values)
	}
	return cmd
}

func (c *boltRedisClient) SCard(ctx context.Context, key string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	var count int64
	err := c.update(ctx, func(root *bolt.Bucket) error {
		entry, err := entryBucket(root, key, typeSet, false)
		if err != nil || entry == nil {
			return err
		}
		return entry.ForEach(func(field, _ []byte) error {
			if isDataField(field) {
				count++
			}
			return nil
		})
	})
	setIntResult(cmd, count, err)
	return cmd
}

func (c *boltRedisClient) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	var added int64
	err := c.update(ctx, func(root *bolt.Bucket) error {
		return hset(root, key, values, &added)
	})
	setIntResult(cmd, added, err)
	return cmd
}

func (c *boltRedisClient) HSetNX(ctx context.Context, key, field string, value interface{}) *redis.BoolCmd {
	cmd := redis.NewBoolCmd(ctx)
	set := false
	err := c.update(ctx, func(root *bolt.Bucket) error {
		entry, err := entryBucket(root, key, typeHash, true)
		if err != nil {
			return err
		}
		name := dataField(field)
		if entry.Get(name) != nil {
			return nil
		}
		set = true
		return entry.Put(name, []byte(stringValue(value)))
	})
	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(set)
	}
	return cmd
}

func (c *boltRedisClient) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx)
	var value []byte
	err := c.update(ctx, func(root *bolt.Bucket) error {
		entry, err := entryBucket(root, key, typeHash, false)
		if err != nil {
			return err
		}
		if entry == nil || entry.Get(dataField(field)) == nil {
			return redis.Nil
		}
		value = append(value, entry.Get(dataField(field))...)
		return nil
	})
	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(string(value))
	}
	return cmd
}

func (c *boltRedisClient) HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd {
	cmd := redis.NewMapStringStringCmd(ctx)
	values := make(map[string]string)
	err := c.update(ctx, func(root *bolt.Bucket) error {
		entry, err := entryBucket(root, key, typeHash, false)
		if err != nil || entry == nil {
			return err
		}
		return entry.ForEach(func(field, value []byte) error {
			if isDataField(field) {
				values[string(field[1:])] = string(value)
			}
			return nil
		})
	})
	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(values)
	}
	return cmd
}

func (c *boltRedisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	var added int64
	err := c.update(ctx, func(root *bolt.Bucket) error {
		entry, err := entryBucket(root, key, typeZSet, true)
		if err != nil {
			return err
		}
		for _, member := range members {
			field := dataField(stringValue(member.Member))
			if entry.Get(field) == nil {
				added++
			}
			if err = entry.Put(field, encodeFloat(member.Score)); err != nil {
				return err
			}
		}
		return nil
	})
	setIntResult(cmd, added, err)
	return cmd
}

func (c *boltRedisClient) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	cmd := redis.NewStringSliceCmd(ctx)
	var members []scoredMember
	err := c.update(ctx, func(root *bolt.Bucket) error {
		entry, err := entryBucket(root, key, typeZSet, false)
		if err != nil || entry == nil {
			return err
		}
		min, minExclusive, err := scoreBoundary(opt.Min)
		if err != nil {
			return err
		}
		max, maxExclusive, err := scoreBoundary(opt.Max)
		if err != nil {
			return err
		}
		return entry.ForEach(func(field, raw []byte) error {
			if !isDataField(field) {
				return nil
			}
			score := decodeFloat(raw)
			if score < min || score > max || minExclusive && score == min || maxExclusive && score == max {
				return nil
			}
			members = append(members, scoredMember{member: string(field[1:]), score: score})
			return nil
		})
	})
	if err != nil {
		cmd.SetErr(err)
		return cmd
	}
	sort.Slice(members, func(i, j int) bool {
		if members[i].score == members[j].score {
			return members[i].member < members[j].member
		}
		return members[i].score < members[j].score
	})
	values := make([]string, len(members))
	for i := range members {
		values[i] = members[i].member
	}
	cmd.SetVal(values)
	return cmd
}

func (c *boltRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	var removed int64
	err := c.update(ctx, func(root *bolt.Bucket) error {
		entry, err := entryBucket(root, key, typeZSet, false)
		if err != nil || entry == nil {
			return err
		}
		for _, member := range members {
			field := dataField(stringValue(member))
			if entry.Get(field) != nil {
				removed++
				if err = entry.Delete(field); err != nil {
					return err
				}
			}
		}
		return nil
	})
	setIntResult(cmd, removed, err)
	return cmd
}

func (c *boltRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	var deleted int64
	err := c.update(ctx, func(root *bolt.Bucket) error {
		for _, key := range keys {
			if root.Bucket([]byte(key)) != nil {
				deleted++
				if err := root.DeleteBucket([]byte(key)); err != nil {
					return err
				}
			}
		}
		return nil
	})
	setIntResult(cmd, deleted, err)
	return cmd
}

func (c *boltRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	cmd := redis.NewBoolCmd(ctx)
	found := false
	err := c.update(ctx, func(root *bolt.Bucket) error {
		entry := root.Bucket([]byte(key))
		if entry == nil || expired(root, key, entry) {
			return nil
		}
		found = true
		return entry.Put(expiresField, encodeInt64(time.Now().Add(expiration).UnixNano()))
	})
	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(found)
	}
	return cmd
}

func (c *boltRedisClient) Eval(ctx context.Context, _ string, keys []string, args ...interface{}) *redis.Cmd {
	cmd := redis.NewCmd(ctx)
	var result int64
	err := c.update(ctx, func(root *bolt.Bucket) error {
		var err error
		switch {
		case len(keys) == 1 && len(args) == 2:
			result, err = evalSetMax(root, keys[0], stringValue(args[0]), stringValue(args[1]))
		case len(keys) == 2 && len(args) == 7:
			result, err = evalWriteDaily(root, keys, args)
		case len(keys) == 2 && len(args) == 2:
			result, err = evalCleanupInstance(root, keys, args)
		default:
			err = errors.New("unsupported uptime Redis script")
		}
		return err
	})
	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(result)
	}
	return cmd
}

func (c *boltRedisClient) Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	pipe := &boltPipeline{client: c}
	if err := fn(pipe); err != nil {
		return pipe.commands, err
	}
	for _, command := range pipe.commands {
		if err := command.Err(); err != nil {
			return pipe.commands, err
		}
	}
	return pipe.commands, nil
}

func (c *boltRedisClient) update(ctx context.Context, fn func(*bolt.Bucket) error) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	return c.db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists(entriesBucket)
		if err != nil {
			return err
		}
		return fn(root)
	})
}

type boltPipeline struct {
	redis.Pipeliner
	client   *boltRedisClient
	commands []redis.Cmder
}

func (p *boltPipeline) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	cmd := p.client.Expire(ctx, key, expiration)
	p.commands = append(p.commands, cmd)
	return cmd
}

func (p *boltPipeline) HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd {
	cmd := p.client.HGetAll(ctx, key)
	p.commands = append(p.commands, cmd)
	return cmd
}

func (p *boltPipeline) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	cmd := p.client.ZRangeByScore(ctx, key, opt)
	p.commands = append(p.commands, cmd)
	return cmd
}

func (p *boltPipeline) SCard(ctx context.Context, key string) *redis.IntCmd {
	cmd := p.client.SCard(ctx, key)
	p.commands = append(p.commands, cmd)
	return cmd
}

func entryBucket(root *bolt.Bucket, key, kind string, create bool) (*bolt.Bucket, error) {
	entry := root.Bucket([]byte(key))
	if entry != nil && expired(root, key, entry) {
		entry = nil
	}
	if entry == nil {
		if !create {
			return nil, nil
		}
		var err error
		entry, err = root.CreateBucket([]byte(key))
		if err != nil {
			return nil, err
		}
		if err = entry.Put(typeField, []byte(kind)); err != nil {
			return nil, err
		}
		return entry, nil
	}
	if existing := string(entry.Get(typeField)); existing != kind {
		return nil, fmt.Errorf("uptime Bbolt key %q has type %q, want %q", key, existing, kind)
	}
	return entry, nil
}

func expired(root *bolt.Bucket, key string, entry *bolt.Bucket) bool {
	raw := entry.Get(expiresField)
	if len(raw) != 8 || decodeInt64(raw) > time.Now().UnixNano() {
		return false
	}
	_ = root.DeleteBucket([]byte(key))
	return true
}

func hset(root *bolt.Bucket, key string, values []interface{}, added *int64) error {
	if len(values)%2 != 0 {
		return errors.New("uptime Bbolt HSET requires field/value pairs")
	}
	entry, err := entryBucket(root, key, typeHash, true)
	if err != nil {
		return err
	}
	for i := 0; i < len(values); i += 2 {
		field := dataField(stringValue(values[i]))
		if entry.Get(field) == nil && added != nil {
			*added++
		}
		if err = entry.Put(field, []byte(stringValue(values[i+1]))); err != nil {
			return err
		}
	}
	return nil
}

func evalSetMax(root *bolt.Bucket, key, field, rawValue string) (int64, error) {
	value, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		return 0, err
	}
	entry, err := entryBucket(root, key, typeHash, true)
	if err != nil {
		return 0, err
	}
	name := dataField(field)
	if currentRaw := entry.Get(name); currentRaw != nil {
		current, parseErr := strconv.ParseInt(string(currentRaw), 10, 64)
		if parseErr != nil {
			return 0, parseErr
		}
		if current >= value {
			return 0, nil
		}
	}
	return 1, entry.Put(name, []byte(rawValue))
}

func evalWriteDaily(root *bolt.Bucket, keys []string, args []interface{}) (int64, error) {
	daily, err := entryBucket(root, keys[0], typeHash, true)
	if err != nil {
		return 0, err
	}
	finalized := daily.Get(dataField("finalized"))
	score, err := strconv.ParseFloat(stringValue(args[5]), 64)
	if err != nil {
		return 0, err
	}
	if err = zadd(root, keys[1], stringValue(args[1]), score); err != nil {
		return 0, err
	}
	result := int64(0)
	if len(finalized) == 0 || string(finalized) == "0" {
		var added int64
		if err = hset(root, keys[0], []interface{}{
			"service_id", args[0], "day", args[1], "up_slots", args[2],
			"expected_slots", args[3], "finalized", args[4],
		}, &added); err != nil {
			return 0, err
		}
		result = 1
	}
	ttlSeconds, err := strconv.ParseInt(stringValue(args[6]), 10, 64)
	if err != nil {
		return 0, err
	}
	if ttlSeconds > 0 {
		expiresAt := encodeInt64(time.Now().Add(time.Duration(ttlSeconds) * time.Second).UnixNano())
		for _, key := range keys {
			if entry := root.Bucket([]byte(key)); entry != nil {
				if err = entry.Put(expiresField, expiresAt); err != nil {
					return 0, err
				}
			}
		}
	}
	return result, nil
}

func evalCleanupInstance(root *bolt.Bucket, keys []string, args []interface{}) (int64, error) {
	index, err := entryBucket(root, keys[0], typeZSet, false)
	if err != nil || index == nil {
		return 0, err
	}
	member := stringValue(args[0])
	rawScore := index.Get(dataField(member))
	if rawScore == nil {
		return 0, nil
	}
	cutoff, err := strconv.ParseFloat(stringValue(args[1]), 64)
	if err != nil {
		return 0, err
	}
	if decodeFloat(rawScore) > cutoff {
		return 0, nil
	}
	if root.Bucket([]byte(keys[1])) != nil {
		if err = root.DeleteBucket([]byte(keys[1])); err != nil {
			return 0, err
		}
	}
	if err = index.Delete(dataField(member)); err != nil {
		return 0, err
	}
	return 1, nil
}

func zadd(root *bolt.Bucket, key, member string, score float64) error {
	entry, err := entryBucket(root, key, typeZSet, true)
	if err != nil {
		return err
	}
	return entry.Put(dataField(member), encodeFloat(score))
}

type scoredMember struct {
	member string
	score  float64
}

func scoreBoundary(value string) (float64, bool, error) {
	exclusive := strings.HasPrefix(value, "(")
	value = strings.TrimPrefix(value, "(")
	switch value {
	case "-inf":
		return math.Inf(-1), exclusive, nil
	case "+inf":
		return math.Inf(1), exclusive, nil
	default:
		score, err := strconv.ParseFloat(value, 64)
		return score, exclusive, err
	}
}

func dataField(value string) []byte        { return append([]byte{1}, []byte(value)...) }
func isDataField(value []byte) bool        { return len(value) > 0 && value[0] == 1 }
func stringValue(value interface{}) string { return fmt.Sprint(value) }

func encodeInt64(value int64) []byte {
	result := make([]byte, 8)
	binary.BigEndian.PutUint64(result, uint64(value))
	return result
}

func decodeInt64(value []byte) int64   { return int64(binary.BigEndian.Uint64(value)) }
func encodeFloat(value float64) []byte { return encodeInt64(int64(math.Float64bits(value))) }
func decodeFloat(value []byte) float64 { return math.Float64frombits(uint64(decodeInt64(value))) }

func contextError(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func setIntResult(cmd *redis.IntCmd, value int64, err error) {
	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(value)
	}
}
