package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var queueKey = "urls:queue"
var urlKey = "url:%s"

type URLTracker struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Result    string    `json:"result,omitempty"`
	Error     string    `json:"error,omitempty"`
}

type RedisClient struct {
	client     *redis.Client
	expiration time.Duration
}

func NewRedisClient(addr string) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisClient{client: client}
}

func (r *RedisClient) StoreURL(ctx context.Context, tracker *URLTracker) error {
	data, err := json.Marshal(tracker)
	if err != nil {
		return err
	}

	key := fmt.Sprintf(urlKey, tracker.ID)
	err = r.client.Set(ctx, key, data, r.expiration).Err()
	if err != nil {
		return err
	}

	if err := r.client.LPush(ctx, queueKey, tracker.ID).Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisClient) GetURL(ctx context.Context, id string) (*URLTracker, error) {
	key := fmt.Sprintf(urlKey, id)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var tracker URLTracker
	err = json.Unmarshal([]byte(data), &tracker)
	if err != nil {
		return nil, err
	}

	return &tracker, nil
}

func (r *RedisClient) UpdateURL(ctx context.Context, t *URLTracker) error {
	tracker, err := r.GetURL(ctx, t.ID)
	if err != nil {
		return err
	}

	tracker.Status = t.Status
	tracker.UpdatedAt = time.Now()
	tracker.Result = t.Result
	tracker.Error = t.Error

	data, err := json.Marshal(tracker)
	if err != nil {
		return err
	}

	key := fmt.Sprintf(urlKey, t.ID)
	return r.client.Set(ctx, key, data, r.expiration).Err()
}

func (r *RedisClient) GetAllURLs(ctx context.Context) ([]*URLTracker, error) {
	var results []*URLTracker

	keys, err := r.client.Keys(ctx, "url:*").Result()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		data, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var tracker URLTracker
		if err := json.Unmarshal([]byte(data), &tracker); err != nil {
			continue
		}

		results = append(results, &tracker)
	}

	return results, nil
}

func (r *RedisClient) DequeueURL(ctx context.Context) (*URLTracker, error) {
	id, err := r.client.RPop(ctx, "urls:queue").Result()
	if err != nil {
		if err == redis.Nil {
			//
			return nil, nil
		}
		return nil, err
	}

	tracker, err := r.GetURL(ctx, id)
	if err != nil {
		return nil, err
	}

	return tracker, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
