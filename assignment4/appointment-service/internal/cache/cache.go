package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment4/appointment-service/internal/model"
	"github.com/redis/go-redis/v9"
)

type CacheClient struct {
	client *redis.Client
	ttl    time.Duration
}

const (
	cacheKeyPrefix = "appointment:"
	cacheListKey   = "appointments:list"
)

func NewCacheClient(client *redis.Client, ttl time.Duration) *CacheClient {
	return &CacheClient{
		client: client,
		ttl:    ttl,
	}
}

func (c *CacheClient) Set(ctx context.Context, appointment *model.Appointment) error {
	data, err := json.Marshal(appointment)
	if err != nil {
		return err
	}
	key := cacheKeyPrefix + appointment.ID
	err = c.client.Set(ctx, key, data, c.ttl).Err()
	if err != nil {
		return model.RedisCacheWriteFailureError
	}
	return nil
}

func (c *CacheClient) SetList(ctx context.Context, appointments []*model.Appointment) error {
	data, err := json.Marshal(appointments)
	if err != nil {
		return err
	}
	err = c.client.Set(ctx, cacheListKey, data, c.ttl).Err()
	if err != nil {
		return model.RedisCacheWriteFailureError
	}
	return nil
}

func (c *CacheClient) Get(ctx context.Context, id string) (*model.Appointment, error) {
	key := cacheKeyPrefix + id
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			return nil, err
		}

		return nil, model.RedisCacheMissError
	}

	var appointment model.Appointment
	err = json.Unmarshal(data, &appointment)
	if err != nil {
		return nil, err
	}
	return &appointment, nil
}

func (c *CacheClient) GetList(ctx context.Context) ([]*model.Appointment, error) {
	key := cacheListKey
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			return nil, err
		}
		return nil, model.RedisCacheMissError
	}
	var appointments []*model.Appointment
	err = json.Unmarshal(data, &appointments)
	if err != nil {
		return nil, err
	}
	return appointments, nil
}

func (c *CacheClient) Delete(ctx context.Context, id string) error {
	key := cacheKeyPrefix + id
	return c.client.Del(ctx, key).Err()
}

func (c *CacheClient) DeleteList(ctx context.Context) error {
	key := cacheListKey
	return c.client.Del(ctx, key).Err()
}

func (c *CacheClient) Update(ctx context.Context, appointment *model.Appointment) error {
	key := cacheKeyPrefix + appointment.ID
	data, err := json.Marshal(appointment)
	if err != nil {
		return err
	}

	pipeline := c.client.Pipeline()

	pipeline.Del(ctx, cacheListKey)
	pipeline.Set(ctx, key, data, c.ttl)

	_, err = pipeline.Exec(ctx)
	if err != nil {
		return model.RedisCacheWriteFailureError
	}
	return nil
}

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
