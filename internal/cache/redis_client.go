package cache

import (
	"context"
	"encoding/json"

	"github.com/madhav-poojari/bharat-digital/internal/models"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func New(addr, pass string) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       0,
	})
	return &Client{rdb: rdb}
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

// MSetJSON writes multiple key->struct values in bulk.
// pairs is map[key]models.CacheValue
func (c *Client) MSetJSON(ctx context.Context, pairs map[string]models.CacheValue) error {
	args := make([]interface{}, 0, len(pairs)*2)
	for k, v := range pairs {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		args = append(args, k, string(b))
	}
	return c.rdb.MSet(ctx, args...).Err()
}

func (c *Client) MGetJSON(ctx context.Context, keys ...string) ([]*models.CacheValue, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	res, err := c.rdb.MGet(ctx, keys...).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	out := make([]*models.CacheValue, 0, len(res))
	for _, r := range res {
		if r == nil {
			out = append(out, nil)
			continue
		}
		s, ok := r.(string)
		if !ok {
			out = append(out, nil)
			continue
		}
		var cv models.CacheValue
		if err := json.Unmarshal([]byte(s), &cv); err != nil {
			out = append(out, nil)
			continue
		}
		out = append(out, &cv)
	}
	return out, nil
}
