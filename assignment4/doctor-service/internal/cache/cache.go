package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment4/doctor-service/internal/model"
	"github.com/redis/go-redis/v9"
)

type CacheClient struct {
	client *redis.Client
	ttl    time.Duration
}

const (
	cacheKeyPrefix = "doctor:"
	cacheListKey   = "doctors:list"
)

func NewCacheClient(client *redis.Client, ttl time.Duration) *CacheClient {
	return &CacheClient{
		client: client,
		ttl:    ttl,
	}
}

func (c *CacheClient) Set(ctx context.Context, doctor *model.Doctor) error {
	data, err := json.Marshal(doctor)
	if err != nil {
		return err
	}
	key := cacheKeyPrefix + doctor.ID
	err = c.client.Set(ctx, key, data, c.ttl).Err()
	if err != nil {
		return model.RedisCacheWriteFailureError
	}
	return nil
}

func (c *CacheClient) SetList(ctx context.Context, doctors []*model.Doctor) error {
	data, err := json.Marshal(doctors)
	if err != nil {
		return err
	}
	err = c.client.Set(ctx, cacheListKey, data, c.ttl).Err()
	if err != nil {
		return model.RedisCacheWriteFailureError
	}
	return nil
}

func (c *CacheClient) Get(ctx context.Context, id string) (*model.Doctor, error) {
	key := cacheKeyPrefix + id
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			return nil, err
		}

		return nil, model.RedisCacheMissError
	}

	var doctor model.Doctor
	err = json.Unmarshal(data, &doctor)
	if err != nil {
		return nil, err
	}
	return &doctor, nil
}

func (c *CacheClient) GetList(ctx context.Context) ([]*model.Doctor, error) {
	key := cacheListKey
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			return nil, err
		}
		return nil, model.RedisCacheMissError
	}
	var doctors []*model.Doctor
	err = json.Unmarshal(data, &doctors)
	if err != nil {
		return nil, err
	}
	return doctors, nil
}

func (c *CacheClient) Delete(ctx context.Context, id string) error {
	key := cacheKeyPrefix + id
	return c.client.Del(ctx, key).Err()
}

func (c *CacheClient) DeleteList(ctx context.Context) error {
	key := cacheListKey
	return c.client.Del(ctx, key).Err()
}

// internal/cache/cache.go

func (c *CacheClient) Allow(ctx context.Context, ip string, limit int) (bool, error) {
	key := "rate_limit:" + ip
	now := time.Now().UnixNano()
	oneMinuteAgo := now - int64(time.Minute)

	pipe := c.client.Pipeline()

	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", oneMinuteAgo))
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, time.Minute)

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count, _ := cmds[2].(*redis.IntCmd).Result()

	return int(count) <= limit, nil
}
