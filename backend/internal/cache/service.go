// Herlin AI Assistant - Backend Service
// Copyright 2026 Herlin AI. All rights reserved.

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/herlin-ai/herlin-assistant/config"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	client *redis.Client
	ctx    context.Context
}

func NewService(cfg *config.Config) *Service {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	return &Service{
		client: rdb,
		ctx:    context.Background(),
	}
}

func (s *Service) Ping() error {
	return s.client.Ping(s.ctx).Err()
}

func (s *Service) Set(key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return s.client.Set(s.ctx, key, jsonData, expiration).Err()
}

func (s *Service) Get(key string, dest interface{}) error {
	val, err := s.client.Get(s.ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func (s *Service) Delete(key string) error {
	return s.client.Del(s.ctx, key).Err()
}

func (s *Service) Exists(key string) (bool, error) {
	count, err := s.client.Exists(s.ctx, key).Result()
	return count > 0, err
}

func (s *Service) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return s.client.SetNX(s.ctx, key, jsonData, expiration).Result()
}

func (s *Service) Expire(key string, expiration time.Duration) error {
	return s.client.Expire(s.ctx, key, expiration).Err()
}

func (s *Service) TTL(key string) (time.Duration, error) {
	return s.client.TTL(s.ctx, key).Result()
}

func (s *Service) Incr(key string) (int64, error) {
	return s.client.Incr(s.ctx, key).Result()
}

func (s *Service) Decr(key string) (int64, error) {
	return s.client.Decr(s.ctx, key).Result()
}

func (s *Service) HSet(key, field string, value interface{}) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return s.client.HSet(s.ctx, key, field, jsonData).Err()
}

func (s *Service) HGet(key, field string, dest interface{}) error {
	val, err := s.client.HGet(s.ctx, key, field).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func (s *Service) HDel(key string, fields ...string) error {
	return s.client.HDel(s.ctx, key, fields...).Err()
}

func (s *Service) HGetAll(key string) (map[string]string, error) {
	return s.client.HGetAll(s.ctx, key).Result()
}

func (s *Service) LPush(key string, values ...interface{}) error {
	for _, value := range values {
		jsonData, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		if err := s.client.LPush(s.ctx, key, jsonData).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) RPush(key string, values ...interface{}) error {
	for _, value := range values {
		jsonData, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		if err := s.client.RPush(s.ctx, key, jsonData).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) LPop(key string, dest interface{}) error {
	val, err := s.client.LPop(s.ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func (s *Service) RPop(key string, dest interface{}) error {
	val, err := s.client.RPop(s.ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func (s *Service) LRange(key string, start, stop int64) ([]string, error) {
	return s.client.LRange(s.ctx, key, start, stop).Result()
}

func (s *Service) SAdd(key string, members ...interface{}) error {
	for _, member := range members {
		jsonData, err := json.Marshal(member)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		if err := s.client.SAdd(s.ctx, key, jsonData).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) SMembers(key string) ([]string, error) {
	return s.client.SMembers(s.ctx, key).Result()
}

func (s *Service) SIsMember(key string, member interface{}) (bool, error) {
	jsonData, err := json.Marshal(member)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return s.client.SIsMember(s.ctx, key, jsonData).Result()
}

func (s *Service) SRem(key string, members ...interface{}) error {
	for _, member := range members {
		jsonData, err := json.Marshal(member)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		if err := s.client.SRem(s.ctx, key, jsonData).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ZAdd(key string, score float64, member interface{}) error {
	jsonData, err := json.Marshal(member)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return s.client.ZAdd(s.ctx, key, redis.Z{Score: score, Member: jsonData}).Err()
}

func (s *Service) ZRange(key string, start, stop int64) ([]string, error) {
	return s.client.ZRange(s.ctx, key, start, stop).Result()
}

func (s *Service) ZRangeByScore(key string, min, max float64) ([]string, error) {
	return s.client.ZRangeByScore(s.ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", min),
		Max: fmt.Sprintf("%f", max),
	}).Result()
}

func (s *Service) ZRem(key string, members ...interface{}) error {
	for _, member := range members {
		jsonData, err := json.Marshal(member)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		if err := s.client.ZRem(s.ctx, key, jsonData).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) FlushDB() error {
	return s.client.FlushDB(s.ctx).Err()
}

func (s *Service) Close() error {
	return s.client.Close()
}

// Cache helpers for common use cases
func (s *Service) CacheUserSession(userID uint, sessionData interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("session:%d", userID)
	return s.Set(key, sessionData, expiration)
}

func (s *Service) GetUserSession(userID uint, dest interface{}) error {
	key := fmt.Sprintf("session:%d", userID)
	return s.Get(key, dest)
}

func (s *Service) CacheAPIResponse(key string, response interface{}, expiration time.Duration) error {
	cacheKey := fmt.Sprintf("api:%s", key)
	return s.Set(cacheKey, response, expiration)
}

func (s *Service) GetCachedAPIResponse(key string, dest interface{}) error {
	cacheKey := fmt.Sprintf("api:%s", key)
	return s.Get(cacheKey, dest)
}

func (s *Service) InvalidateUserCache(userID uint) error {
	pattern := fmt.Sprintf("user:%d:*", userID)
	iter := s.client.Scan(s.ctx, 0, pattern, 0).Iterator()
	for iter.Next(s.ctx) {
		if err := s.client.Del(s.ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
