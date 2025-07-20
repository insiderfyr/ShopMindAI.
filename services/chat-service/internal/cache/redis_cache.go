package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	// Cache configuration
	defaultTTL        = 5 * time.Minute
	maxTTL           = 24 * time.Hour
	lockTTL          = 30 * time.Second
	stampedeFactor   = 0.8 // Probabilistic early expiration
	hotKeyThreshold  = 100 // Requests per minute to consider hot
)

// CacheManager handles all caching operations with advanced features
type CacheManager struct {
	client       *redis.ClusterClient
	logger       *logrus.Logger
	
	// Hot key tracking
	hotKeys      map[string]*hotKeyStats
	hotKeysMu    sync.RWMutex
	
	// Metrics
	hits         int64
	misses       int64
	errors       int64
	metricsMu    sync.RWMutex
}

type hotKeyStats struct {
	count       int64
	lastAccess  time.Time
	ttlBoost    time.Duration
}

// CacheOptions configures cache behavior
type CacheOptions struct {
	TTL             time.Duration
	Lock            bool
	StampedeProtect bool
	Compress        bool
}

// NewCacheManager creates a new cache manager
func NewCacheManager(client *redis.ClusterClient, logger *logrus.Logger) *CacheManager {
	cm := &CacheManager{
		client:  client,
		logger:  logger,
		hotKeys: make(map[string]*hotKeyStats),
	}
	
	// Start hot key cleanup
	go cm.cleanupHotKeys()
	
	return cm
}

// Get retrieves a value from cache with stampede protection
func (cm *CacheManager) Get(ctx context.Context, key string, dest interface{}, opts *CacheOptions) error {
	// Track hot keys
	cm.trackHotKey(key)
	
	// Try to get from cache
	val, err := cm.client.Get(ctx, key).Result()
	if err == redis.Nil {
		cm.incrementMisses()
		return ErrCacheMiss
	}
	if err != nil {
		cm.incrementErrors()
		return fmt.Errorf("redis get: %w", err)
	}
	
	// Check for probabilistic early expiration (stampede protection)
	if opts != nil && opts.StampedeProtect {
		ttl, _ := cm.client.TTL(ctx, key).Result()
		if cm.shouldRefreshEarly(ttl) {
			cm.incrementMisses() // Count as miss to trigger refresh
			return ErrCacheMiss
		}
	}
	
	// Deserialize
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		cm.incrementErrors()
		return fmt.Errorf("unmarshal: %w", err)
	}
	
	cm.incrementHits()
	return nil
}

// Set stores a value in cache with options
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, opts *CacheOptions) error {
	// Serialize
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	
	// Determine TTL
	ttl := cm.calculateTTL(key, opts)
	
	// Set with pipeline for performance
	pipe := cm.client.Pipeline()
	pipe.Set(ctx, key, data, ttl)
	
	// Add to hot key set if applicable
	if cm.isHotKey(key) {
		pipe.ZAdd(ctx, "hot_keys", redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: key,
		})
	}
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		cm.incrementErrors()
		return fmt.Errorf("redis set: %w", err)
	}
	
	return nil
}

// GetOrSet implements read-through cache with locking
func (cm *CacheManager) GetOrSet(ctx context.Context, key string, dest interface{}, 
	loader func() (interface{}, error), opts *CacheOptions) error {
	
	// Try cache first
	err := cm.Get(ctx, key, dest, opts)
	if err == nil {
		return nil
	}
	
	// Use distributed lock to prevent stampede
	if opts != nil && opts.Lock {
		lockKey := fmt.Sprintf("lock:%s", key)
		locked, err := cm.acquireLock(ctx, lockKey, lockTTL)
		if err != nil {
			return fmt.Errorf("acquire lock: %w", err)
		}
		
		if !locked {
			// Someone else is loading, wait and retry
			time.Sleep(100 * time.Millisecond)
			return cm.Get(ctx, key, dest, opts)
		}
		
		defer cm.releaseLock(ctx, lockKey)
	}
	
	// Double-check cache after acquiring lock
	err = cm.Get(ctx, key, dest, nil)
	if err == nil {
		return nil
	}
	
	// Load data
	data, err := loader()
	if err != nil {
		return fmt.Errorf("loader: %w", err)
	}
	
	// Store in cache
	if err := cm.Set(ctx, key, data, opts); err != nil {
		cm.logger.WithError(err).Warn("Failed to cache loaded data")
	}
	
	// Copy to destination
	dataBytes, _ := json.Marshal(data)
	return json.Unmarshal(dataBytes, dest)
}

// Delete removes a key from cache
func (cm *CacheManager) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	
	pipe := cm.client.Pipeline()
	for _, key := range keys {
		pipe.Del(ctx, key)
	}
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		cm.incrementErrors()
		return fmt.Errorf("redis delete: %w", err)
	}
	
	return nil
}

// InvalidatePattern removes all keys matching a pattern
func (cm *CacheManager) InvalidatePattern(ctx context.Context, pattern string) error {
	// Use SCAN for memory efficiency
	iter := cm.client.Scan(ctx, 0, pattern, 100).Iterator()
	
	batch := make([]string, 0, 100)
	for iter.Next(ctx) {
		batch = append(batch, iter.Val())
		
		if len(batch) >= 100 {
			if err := cm.Delete(ctx, batch...); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	
	if len(batch) > 0 {
		if err := cm.Delete(ctx, batch...); err != nil {
			return err
		}
	}
	
	return iter.Err()
}

// Distributed locking
func (cm *CacheManager) acquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return cm.client.SetNX(ctx, key, "1", ttl).Result()
}

func (cm *CacheManager) releaseLock(ctx context.Context, key string) {
	cm.client.Del(ctx, key)
}

// Hot key management
func (cm *CacheManager) trackHotKey(key string) {
	cm.hotKeysMu.Lock()
	defer cm.hotKeysMu.Unlock()
	
	stats, exists := cm.hotKeys[key]
	if !exists {
		stats = &hotKeyStats{}
		cm.hotKeys[key] = stats
	}
	
	stats.count++
	stats.lastAccess = time.Now()
	
	// Increase TTL boost for frequently accessed keys
	if stats.count > hotKeyThreshold {
		stats.ttlBoost = time.Duration(math.Min(
			float64(stats.count/hotKeyThreshold)*float64(time.Hour),
			float64(maxTTL),
		))
	}
}

func (cm *CacheManager) isHotKey(key string) bool {
	cm.hotKeysMu.RLock()
	defer cm.hotKeysMu.RUnlock()
	
	stats, exists := cm.hotKeys[key]
	return exists && stats.count > hotKeyThreshold
}

func (cm *CacheManager) calculateTTL(key string, opts *CacheOptions) time.Duration {
	baseTTL := defaultTTL
	if opts != nil && opts.TTL > 0 {
		baseTTL = opts.TTL
	}
	
	// Add TTL boost for hot keys
	cm.hotKeysMu.RLock()
	stats, exists := cm.hotKeys[key]
	cm.hotKeysMu.RUnlock()
	
	if exists && stats.ttlBoost > 0 {
		return baseTTL + stats.ttlBoost
	}
	
	return baseTTL
}

// Probabilistic early expiration
func (cm *CacheManager) shouldRefreshEarly(ttl time.Duration) bool {
	if ttl <= 0 {
		return true
	}
	
	// Calculate probability of early refresh
	remainingRatio := float64(ttl) / float64(defaultTTL)
	if remainingRatio > stampedeFactor {
		return false
	}
	
	// Exponentially increasing probability as TTL approaches 0
	probability := math.Pow(1-remainingRatio/stampedeFactor, 3)
	return rand.Float64() < probability
}

// Cleanup hot keys periodically
func (cm *CacheManager) cleanupHotKeys() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		cm.hotKeysMu.Lock()
		now := time.Now()
		
		for key, stats := range cm.hotKeys {
			// Remove keys not accessed in last hour
			if now.Sub(stats.lastAccess) > time.Hour {
				delete(cm.hotKeys, key)
			} else if now.Sub(stats.lastAccess) > 10*time.Minute {
				// Decay count for keys not recently accessed
				stats.count = stats.count / 2
			}
		}
		
		cm.hotKeysMu.Unlock()
	}
}

// Metrics
func (cm *CacheManager) incrementHits() {
	cm.metricsMu.Lock()
	cm.hits++
	cm.metricsMu.Unlock()
}

func (cm *CacheManager) incrementMisses() {
	cm.metricsMu.Lock()
	cm.misses++
	cm.metricsMu.Unlock()
}

func (cm *CacheManager) incrementErrors() {
	cm.metricsMu.Lock()
	cm.errors++
	cm.metricsMu.Unlock()
}

// GetMetrics returns cache metrics
func (cm *CacheManager) GetMetrics() CacheMetrics {
	cm.metricsMu.RLock()
	defer cm.metricsMu.RUnlock()
	
	total := cm.hits + cm.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(cm.hits) / float64(total)
	}
	
	return CacheMetrics{
		Hits:    cm.hits,
		Misses:  cm.misses,
		Errors:  cm.errors,
		HitRate: hitRate,
	}
}

// CacheMetrics holds cache performance metrics
type CacheMetrics struct {
	Hits    int64
	Misses  int64
	Errors  int64
	HitRate float64
}

var ErrCacheMiss = fmt.Errorf("cache miss")