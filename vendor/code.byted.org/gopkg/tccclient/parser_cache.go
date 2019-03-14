package tccclient

import (
	"sync"
)

var (
	parserCache *ParserCache
)

func init() {
	parserCache = NewParserCache()
}

type ParserCache struct {
	tccValueCache  map[string]string
	tccResultCache map[string]interface{}
	mu             sync.RWMutex
}

func NewParserCache() *ParserCache {
	return &ParserCache{
		tccValueCache:  make(map[string]string),
		tccResultCache: make(map[string]interface{}),
		mu:             sync.RWMutex{},
	}
}

func (ps *ParserCache) Set(key, value string, result interface{}) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.tccValueCache[key] = value
	ps.tccResultCache[key] = result
}

func (ps *ParserCache) Get(key string) (string, interface{}, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	value, exist := ps.tccValueCache[key]
	return value, ps.tccResultCache[key], exist
}

type TCCParser func(value string, err error, cacheResult interface{}) (interface{}, error)

func (client *Client) GetWithParser(key string, parser TCCParser) (interface{}, error) {
	value, err := client.Get(key)
	cacheValue, cacheResult, exist := parserCache.Get(key)
	if err == nil && exist && cacheValue == value {
		return cacheResult, nil
	}
	result, err := parser(value, err, cacheResult)
	if err == nil {
		parserCache.Set(key, value, result)
	}
	return result, err
}
