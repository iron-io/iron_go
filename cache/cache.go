// client for the IronCache REST API
package cache

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/manveru/go.iron/config"
	"time"
)

var (
	escape = url.QueryEscape
	JSON   = Codec{Marshal: json.Marshal, Unmarshal: json.Unmarshal}
	Gob    = Codec{Marshal: gobMarshal, Unmarshal: gobUnmarshal}
)

type cacheURL url.URL

type stringer interface {
	String() string
}

type Cache struct {
	Config config.Settings
	Name   string
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
	return &Cache{Config: config.Config("iron_cache"), Name: cacheName}
}

func (c *Cache) action(suffix ...string) *cacheURL {
	cc := c.Config

	parts := append([]string{"caches"}, suffix...)
	for n, part := range parts {
		parts[n] = escape(part)
	}

	u := &cacheURL{}
	u.Scheme = cc.Protocol
	u.Host = fmt.Sprintf("%s:%d", escape(cc.Host), cc.Port)
	u.Path = fmt.Sprintf("/%s/projects/%s/%s", cc.ApiVersion, cc.ProjectId, strings.Join(parts, "/"))
	return u
}

func (c *cacheURL) QueryAdd(key string, format string, value interface{}) *cacheURL {
	query := c.Query()
	query.Add(key, fmt.Sprintf(format, value))
	c.RawQuery = query.Encode()
	return c
}

func (c *cacheURL) String() string    { return (*url.URL)(c).String() }
func (c *cacheURL) Query() url.Values { return (*url.URL)(c).Query() }

func (c *Cache) ListCaches(page, perPage int) (caches []*Cache, err error) {
	u := c.action().QueryAdd("page", "%d", page).QueryAdd("per_page", "%d", perPage)

	response, err := c.request("GET", u, nil)
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
			Config: c.Config,
			Name:   item.Name,
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

	_, err = c.request("PUT", c.action(c.Name, "items", item.Key), body)
	return
}

// Increment increments the corresponding item's value.
func (c Cache) Increment(key string, amount int64) (err error) {
	body := &bytes.Buffer{}
	encoder := json.NewEncoder(body)
	encoder.Encode(map[string]int64{"amount": amount})
	_, err = c.request("POST", c.action(c.Name, "items", key), body)
	return
}

// Get gets an item from the cache.
func (c Cache) Get(key string) (value []byte, err error) {
	//projects/{Project ID}/caches/{Cache Name}/items/{Key}	GET	Get an Item from a Cache

	response, err := c.request("GET", c.action(c.Name, "items", key), nil)
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
	_, err = c.request("DELETE", c.action(c.Name, "items", key), nil)
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

func (c *Cache) request(method string, endpoint stringer, body io.Reader) (response *http.Response, err error) {
	client := http.Client{}

	request, err := http.NewRequest(method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "OAuth "+c.Config.Token)

	if body == nil {
		request.Header.Set("Accept", "application/json")
		request.Header.Set("Accept-Encoding", "gzip/deflate")
	} else {
		request.Header.Set("Content-Type", "application/json")
	}

	// dumpRequest(request)
	if response, err = client.Do(request); err != nil {
		return
	}
	//dumpResponse(response)
	if err = resToErr(response); err != nil {
		return
	}

	return
}

func dumpRequest(req *http.Request) {
	out, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%q\n", out)
}

func dumpResponse(response *http.Response) {
	out, err := httputil.DumpResponse(response, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%q\n", out)
}

func resToErr(response *http.Response) (err error) {
	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return errors.New("Invalid authentication: The OAuth token is either not provided or invalid")
	case http.StatusNotFound:
		return errors.New("Invalid endpoint: The resource, project, or endpoint being requested doesn't exist.")
	case http.StatusMethodNotAllowed:
		return errors.New("Invalid HTTP method: This endpoint doesn't support that particular verb")
	case http.StatusNotAcceptable:
		return errors.New("Invalid request: Required fields are missing")
	default:
		return errors.New("Unknown API Response: " + response.Status)
	}

	panic("There is no way you'll encounter this")
}
