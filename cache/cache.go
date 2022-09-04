package cache

import (
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

var cacheBucket = []byte("cache")
var cache *bolt.DB

// To intialise the cache db.
func CacheInit() error {
	var err error
	cache, err = bolt.Open("cache.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	return cache.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(cacheBucket)
		return err

	})
}

// To create and update entry in the cache.
func CreateUpdateCacheEntry(name string, value string) error {
	err := cache.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(cacheBucket)

		return b.Put([]byte(name), []byte(value))
	})
	if err != nil {
		return err
	}
	return nil
}

// To get entry from the cache
func GetCacheEntry(name string) (string, error) {
	var value string
	err := cache.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(cacheBucket)
		v := b.Get([]byte(name))
		value = string(v)
		if value != "" {
			return nil
		} else {
			return fmt.Errorf("key Value doesn't exist")
		}
	})
	if err != nil {
		return "", err
	}
	return value, nil
}
