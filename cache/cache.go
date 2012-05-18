// client for the IronCache REST API
package cache

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"time"

	"github.com/manveru/go.iron/api"
	"github.com/manveru/go.iron/config"
)

var (
	JSON = Codec{Marshal: json.Marshal, Unmarshal: json.Unmarshal}
	Gob  = Codec{Marshal: gobMarshal, Unmarshal: gobUnmarshal}
)

type Cache struct {
	Settings config.Settings
	Name     string
}

type Item struct {
	// Key is the Item's key
	Key string
	// Value is the Item's value
	Value []byte
	// Object is the Item's value for use with a Codec.
	Object interface{}
	// Number of seconds until expiration. The zero value defaults to 7 days,
	// maximum is 30 days.
	Expiration time.Duration
	// Caches item only if the key is currently cached.
	Replace bool
	// Caches item only if the key isn't currently cached.
	Add bool
}

// New returns a struct ready to make requests with.
// The cacheName argument is used as namespace.
func New(cacheName string) *Cache {
	return &Cache{Settings: config.Config("iron_cache"), Name: cacheName}
}

func (c *Cache) caches(suffix ...string) *api.URL {
	return api.Action(c.Settings, "caches", suffix...)
}

func (c *Cache) ListCaches(page, perPage int) (caches []*Cache, err error) {
	out := []struct {
		Project_id string
		Name       string
	}{}

	err = c.caches().
		QueryAdd("page", "%d", page).
		QueryAdd("per_page", "%d", perPage).
		Req("GET", nil, &out)
	if err != nil {
		return
	}

	caches = make([]*Cache, 0, len(out))
	for _, item := range out {
		caches = append(caches, &Cache{
			Settings: c.Settings,
			Name:     item.Name,
		})
	}

	return
}

// Set adds an Item to the cache.
func (c Cache) Set(item *Item) (err error) {
	in := struct {
		Body      []byte `json:"body"`
		ExpiresIn int    `json:"expires_in,omitempty"`
		Replace   bool   `json:"replace,omitempty"`
		Add       bool   `json:"add,omitempty"`
	}{
		Body:      item.Value,
		ExpiresIn: int(item.Expiration.Seconds()),
		Replace:   item.Replace,
		Add:       item.Add,
	}

	return c.caches(c.Name, "items", item.Key).Req("PUT", &in, nil)
}

// Increment increments the corresponding item's value.
func (c Cache) Increment(key string, amount int64) (err error) {
	in := map[string]int64{"amount": amount}
	return c.caches(c.Name, "items", key).Req("POST", &in, nil)
}

// Get gets an item from the cache.
func (c Cache) Get(key string) (value []byte, err error) {
	out := struct {
		Cache string `json:"cache"`
		Key   string `json:"key"`
		Value []byte `json:"value"`
	}{}
	if err = c.caches(c.Name, "items", key).Req("GET", nil, &out); err == nil {
		value = out.Value
	}
	return
}

// Delete removes an item from the cache.
func (c Cache) Delete(key string) (err error) {
	return c.caches(c.Name, "items", key).Req("DELETE", nil, nil)
}

type Codec struct {
	Marshal   func(interface{}) ([]byte, error)
	Unmarshal func([]byte, interface{}) error
}

func (cd Codec) Set(c *Cache, item *Item) (err error) {
	if item.Value, err = cd.Marshal(item.Object); err != nil {
		return
	}

	return c.Set(item)
}

func (cd Codec) Get(c *Cache, key string, v interface{}) (item *Item, err error) {
	str, err := c.Get(key)
	if err != nil {
		return
	}

	bts := []byte(str)

	err = cd.Unmarshal(bts, v)
	if err != nil {
		return
	}

	return &Item{Key: key, Value: bts, Object: v}, nil
}

func gobMarshal(v interface{}) ([]byte, error) {
	writer := bytes.Buffer{}
	enc := gob.NewEncoder(&writer)
	err := enc.Encode(v)
	return writer.Bytes(), err
}

func gobUnmarshal(marshalled []byte, v interface{}) error {
	reader := bytes.NewBuffer(marshalled)
	dec := gob.NewDecoder(reader)
	return dec.Decode(v)
}
