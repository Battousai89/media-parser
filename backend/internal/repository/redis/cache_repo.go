package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/media-parser/backend/internal/config"
	"github.com/media-parser/backend/internal/repository"
	"github.com/redis/go-redis/v9"
)

type URLCacheEntry = repository.URLCacheEntry

type CacheRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCacheRepository(cfg *config.RedisConfig) *CacheRepository {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &CacheRepository{
		client: client,
		ttl:    time.Duration(cfg.CacheTTL) * time.Hour,
	}
}

func (r *CacheRepository) Set(ctx context.Context, key string, value string) error {
	return r.client.Set(ctx, key, value, r.ttl).Err()
}

func (r *CacheRepository) SetNX(ctx context.Context, key string, value string) (bool, error) {
	return r.client.SetNX(ctx, key, value, r.ttl).Result()
}

func (r *CacheRepository) SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *CacheRepository) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *CacheRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *CacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (r *CacheRepository) URLCacheKey(url string) string {
	return fmt.Sprintf("url_cache:%s", url)
}

func (r *CacheRepository) IsURLCached(ctx context.Context, url string) (bool, error) {
	return r.Exists(ctx, r.URLCacheKey(url))
}

func (r *CacheRepository) GetURLCache(ctx context.Context, url string) (*URLCacheEntry, error) {
	key := r.URLCacheKey(url)
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var entry URLCacheEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (r *CacheRepository) SetURLCache(ctx context.Context, url, hash string, ttl time.Duration) error {
	key := r.URLCacheKey(url)
	entry := URLCacheEntry{
		Hash:      hash,
		ParsedAt:  time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	if ttl == 0 {
		ttl = r.ttl
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *CacheRepository) DeleteURLCache(ctx context.Context, url string) error {
	key := r.URLCacheKey(url)
	return r.client.Del(ctx, key).Err()
}

func (r *CacheRepository) MarkURLAsCached(ctx context.Context, url string) error {
	return r.Set(ctx, r.URLCacheKey(url), "1")
}

func (r *CacheRepository) GetURLMetadata(ctx context.Context, url string) (map[string]string, error) {
	key := fmt.Sprintf("url_meta:%s", url)
	return r.client.HGetAll(ctx, key).Result()
}

func (r *CacheRepository) SetURLMetadata(ctx context.Context, url string, metadata map[string]interface{}) error {
	key := fmt.Sprintf("url_meta:%s", url)
	return r.client.HMSet(ctx, key, metadata).Err()
}

func (r *CacheRepository) Close() error {
	return r.client.Close()
}

func (r *CacheRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *CacheRepository) Client() *redis.Client {
	return r.client
}
