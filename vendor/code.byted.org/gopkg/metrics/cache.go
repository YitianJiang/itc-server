package metrics

import "sync"

type cachedTags struct {
	tb []byte
}

func (e *cachedTags) Bytes() []byte {
	if e == nil {
		return nil
	}
	return e.tb
}

type tagscache struct {
	mu sync.RWMutex
	m  map[string]*cachedTags
}

func (c *tagscache) get(key []byte) *cachedTags {
	c.mu.RLock()
	ret := c.m[ss(key)]
	c.mu.RUnlock()
	return ret
}

func (c *tagscache) set(key []byte, tt *cachedTags) {
	c.mu.Lock()
	if c.m == nil {
		c.m = make(map[string]*cachedTags)
	}
	c.m[string(key)] = tt
	c.mu.Unlock()
}

func (c *tagscache) Get(tags []T, extTagBytes []byte) *cachedTags {
	k := make([]byte, 0, 500)

	// XXX: we dont sort the tags to improve performance
	// for v2 api, the tags list should be stable all the time
	// for v1 api which use map to store tags, we sort it in Map2Tags
	k = appendTags(k, tags)
	if e := c.get(k); e != nil {
		return e
	}
	b := make([]byte, 0, len(k)+1+len(extTagBytes))
	b = append(b, k...)
	if len(extTagBytes) > 0 {
		b = append(b, '|')
		b = append(b, extTagBytes...)
	}
	e := &cachedTags{b}
	c.set(k, e)
	return e
}
