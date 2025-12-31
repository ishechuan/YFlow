package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"yflow/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis客户端封装
type RedisClient struct {
	client *redis.Client
	config *config.RedisConfig
}

// NewRedisClient 创建Redis客户端实例
func NewRedisClient(cfg *config.RedisConfig) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &RedisClient{
		client: client,
		config: cfg,
	}
}

// Ping 测试Redis连接
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close 关闭Redis连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// GetClient 获取底层Redis客户端（用于监控等场景）
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// GetKey 获取带前缀的键名
func (r *RedisClient) GetKey(key string) string {
	return r.config.Prefix + key
}

// Set 设置键值对
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, r.GetKey(key), value, expiration).Err()
}

// Get 获取键值
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, r.GetKey(key)).Result()
}

// GetBytes 获取二进制数据
func (r *RedisClient) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return r.client.Get(ctx, r.GetKey(key)).Bytes()
}

// Delete 删除键
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, r.GetKey(key)).Err()
}

// DeleteByPattern 根据模式删除键
func (r *RedisClient) DeleteByPattern(ctx context.Context, pattern string) error {
	keys, err := r.client.Keys(ctx, r.GetKey(pattern)).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Exists 检查键是否存在
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, r.GetKey(key)).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// SetJSON 存储JSON数据
func (r *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// 将对象序列化为JSON
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %w", err)
	}

	// 存储JSON字符串
	return r.client.Set(ctx, r.GetKey(key), jsonData, expiration).Err()
}

// GetJSON 获取JSON数据
func (r *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	// 获取JSON字符串
	jsonStr, err := r.client.Get(ctx, r.GetKey(key)).Result()
	if err != nil {
		return err
	}

	// 反序列化JSON
	return json.Unmarshal([]byte(jsonStr), dest)
}

// HSet 设置哈希表字段
func (r *RedisClient) HSet(ctx context.Context, key, field string, value interface{}) error {
	return r.client.HSet(ctx, r.GetKey(key), field, value).Err()
}

// HGet 获取哈希表字段
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, r.GetKey(key), field).Result()
}

// HGetAll 获取哈希表所有字段
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, r.GetKey(key)).Result()
}

// HDel 删除哈希表字段
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, r.GetKey(key), fields...).Err()
}
