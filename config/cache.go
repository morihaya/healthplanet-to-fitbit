package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Cache struct {
	ProcessedDates map[string]bool `json:"processed_dates"`
	mu             sync.RWMutex
}

func LoadCache() (*Cache, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, "cache.json")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Cache{
				ProcessedDates: make(map[string]bool),
			}, nil
		}
		return nil, err
	}
	defer f.Close()

	var c Cache
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil, err
	}
	if c.ProcessedDates == nil {
		c.ProcessedDates = make(map[string]bool)
	}

	return &c, nil
}

func SaveCache(c *Cache) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	path := filepath.Join(dir, "cache.json")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(c)
}

func (c *Cache) Add(date string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ProcessedDates[date] = true
}

func (c *Cache) Has(date string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ProcessedDates[date]
}
