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

func (c *Cache) action(suffix ...string) *api.URL {
	return api.Action(c.Settings, "caches", suffix...)
}

func (c *Cache) ListCaches(page, perPage int) (caches []*Cache, err error) {
	u := c.action().
		QueryAdd("page", "%d", page).
		QueryAdd("per_page", "%d", perPage)

	response, err := api.Request(c.Settings, "GET", u, nil)
	if err != nil {
		return
	}

	body := []struct {
		Project_id string
		Name       string
	}{}
	err = json.NewDecoder(response.Body).Decode(&body)
	if err != nil {
		return
	}

	caches = make([]*Cache, 0, len(body))
	for _, item := range body {
		caches = append(caches, &Cache{
			Settings: c.Settings,
			Name:     item.Name,
		})
	}

	return
}

// Set adds an Item to the cache.
func (c *Cache) Set(item *Item) (err error) {
	body := &bytes.Buffer{}
	encoder := json.NewEncoder(body)
	encoder.Encode(struct {
		Body      []byte `json:"body"`
		ExpiresIn int    `json:"expires_in,omitempty"`
		Replace   bool   `json:"replace,omitempty"`
		Add       bool   `json:"add,omitempty"`
	}{
		Body:      item.Value,
		ExpiresIn: int(item.Expiration.Seconds()),
		Replace:   item.Replace,
		Add:       item.Add,
	})

	_, err = api.Request(c.Settings, "PUT", c.action(c.Name, "items", item.Key), body)
	return
}

// Increment increments the corresponding item's value.
func (c Cache) Increment(key string, amount int64) (err error) {
	body := &bytes.Buffer{}
	encoder := json.NewEncoder(body)
	encoder.Encode(map[string]int64{"amount": amount})
	_, err = api.Request(c.Settings, "POST", c.action(c.Name, "items", key), body)
	return
}

// Get gets an item from the cache.
func (c Cache) Get(key string) (value []byte, err error) {
	//projects/{Project ID}/caches/{Cache Name}/items/{Key}	GET	Get an Item from a Cache

	response, err := api.Request(c.Settings, "GET", c.action(c.Name, "items", key), nil)
	if err != nil {
		return
	}

	body := struct {
		Cache string `json:"cache"`
		Key   string `json:"key"`
		Value []byte `json:"value"`
	}{}
	err = json.NewDecoder(response.Body).Decode(&body)
	if err != nil {
		return
	}

	return body.Value, err
}

// Delete removes an item from the cache.
func (c Cache) Delete(key string) (err error) {
	_, err = api.Request(c.Settings, "DELETE", c.action(c.Name, "items", key), nil)
	return
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
