package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

func (c *CacheClient) RetryAfter(ctx context.Context, ip string) (time.Duration, error) {
	key := "rate_limit:" + ip
	oldest, err := c.client.ZRange(ctx, key, 0, 0).Result()
	if err != nil {
		return 0, err
	}
	if len(oldest) == 0 {
		return time.Minute, nil
	}

	oldestTs, err := strconv.ParseInt(oldest[0], 10, 64)
	if err != nil {
		return 0, err
	}

	retryAfter := time.Duration(oldestTs + int64(time.Minute) - time.Now().UnixNano())
	if retryAfter < time.Second {
		retryAfter = time.Second
	}
	if retryAfter > time.Minute {
		retryAfter = time.Minute
	}

	return retryAfter, nil
}
