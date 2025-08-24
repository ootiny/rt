package _rt_package_name_

import (
	"time"

	"github.com/ykytech/study-blocks/main/libs/rt/ristretto"
)

type IMemCache interface {
	Get(key string) []byte
	Delete(key string) bool
	DeleteAll() bool
	Set(key string, value []byte, expirationSecond uint32) bool
	Close() error
}

// LocalMemCache 表示内存缓存系统
type LocalMemCache struct {
	cache *ristretto.Cache[string, []byte]
}

// NewMemoryCache 创建一个新的内存缓存
func NewLocalMemCache(maxBytes int) (IMemCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, []byte]{
		NumCounters: int64(maxBytes / 10),
		MaxCost:     int64(maxBytes),
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	} else {
		return &LocalMemCache{cache: cache}, nil
	}
}

// Close 关闭内存缓存
func (p *LocalMemCache) Close() error {
	p.cache.Close()
	return nil
}

// Set 设置一个缓存条目
func (p *LocalMemCache) Set(key string, value []byte, expirationSecond uint32) bool {
	if expirationSecond > 0 {
		return p.cache.SetWithTTL(key, value, int64(len(key)+len(value)), time.Duration(expirationSecond)*time.Second)
	} else {
		return p.cache.Set(key, value, int64(len(key)+len(value)))
	}
}

// Delete 删除一个缓存条目
func (p *LocalMemCache) Delete(key string) bool {
	p.cache.Del(key)
	return true
}

// DeleteAll
func (p *LocalMemCache) DeleteAll() bool {
	p.cache.Clear()
	return true
}

// Get 获取一个缓存条目
func (p *LocalMemCache) Get(key string) []byte {
	if buf, ok := p.cache.Get(key); !ok {
		return nil
	} else {
		return buf
	}
}
