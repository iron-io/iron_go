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
)

var (
	escape = url.QueryEscape
)

type Cache struct {
	Config config.Settings
	Name   string
}

func url(c *Cache, action ...string) url.URL {
	cc := c.Config

	parts := append([]string{cc.ApiVersion, "projects", cc.ProjectId, "caches"}, action...)
	for n, part := range parts {
		parts[n] = escape(part)
	}

	u := url.URL{}
	u.Scheme = cc.Protocol
	u.Host = fmt.Sprintf(
		"%s.iron.io:%d/%s/projects/%s",
		escape(cc.Host), cc.Port, strings.Join(parts, "/"))
	return u
}

func query(u *url.URL, values url.Values) *url.URL {
	u.RawQuery = values.Encode()
	return u
}

func (c *Cache) Caches(page, perPage int) ([]Cache, error) {
	query := url.Values{}
	query.Add("page", fmt.Sprintf("%d", page))
	query.Add("per_page", fmt.Sprintf("%d", perPage))

	url := c.url()
	url.RawQuery = query.Encode()

	response, err := c.request("GET", url, nil)
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

	caches = make([]Cache, 0, len(body))
	for _, item := range body {
		caches = append(caches, Cache{
			Host:      c.Host,
			Token:     c.Token,
			ProjectId: item.Project_id,
			Name:      item.Name,
		})
	}

	return
}

func New(domain, token, projectId, name string) *Context {
}

func Caches(domain, token, projectId string) (caches []Cache, err error) {
	c := Cache{
		Domain:    domain,
		Token:     token,
		ProjectId: projectId,
	}

	response, err := c.request("GET", nil)
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

	caches = make([]Cache, 0, len(body))
	for _, item := range body {
		caches = append(caches, Cache{
			Domain:    c.Domain,
			Token:     c.Token,
			ProjectId: item.Project_id,
			Name:      item.Name,
		})
	}

	return
}

type Item struct {
	// The item's data
	Body string `json:"body"`
	// Number of seconds until expiration. Defaults to 7 days, maximum is 30 days.
	ExpiresIn int `json:"expires_in,omitempty"`
	// Caches item only if the key is currently cached.
	Replace bool `json:"replace,omitempty"`
	// Caches item only if the key isn't currently cached.
	Add bool `json:"add,omitempty"`
}

func (i *Item) Gob(value interface{}) error {
	writer := bytes.Buffer{}
	enc := gob.NewEncoder(&writer)
	if err := enc.Encode(value); err != nil {
		return err
	}
	i.Body = writer.String()
	return nil
}

// Set adds an Item to the cache.
func (c Cache) Set(key string, item Item) (err error) {
	body := &bytes.Buffer{}
	encoder := json.NewEncoder(body)
	encoder.Encode(item)
	_, err = c.request("PUT", body, c.Name, "items", key)
	return
}

// Increment increments the corresponding item's value.
func (c Cache) Increment(key string, amount int64) (err error) {
	body := &bytes.Buffer{}
	encoder := json.NewEncoder(body)
	encoder.Encode(map[string]int64{"amount": amount})
	_, err = c.request("POST", body, c.Name, "items", key)
	return
}

// Get gets an item from the cache.
func (c Cache) Get(key string) (value string, err error) {
	//projects/{Project ID}/caches/{Cache Name}/items/{Key}	GET	Get an Item from a Cache

	response, err := c.request("GET", nil, c.Name, "items", key)
	if err != nil {
		return
	}

	body := struct {
		Cache string `json:"cache"`
		Key   string `json:"key"`
		Value string `json:"value"`
	}{}
	err = json.NewDecoder(response.Body).Decode(&body)
	if err != nil {
		return
	}

	return body.Value, err
}

// Delete removes an item from the cache.
func (c Cache) Delete(key string) (err error) {
	_, err = c.request("DELETE", nil, c.Name, "items", key)
	return
}

func (c *Cache) request(method string, body io.Reader, action ...string) (response *http.Response, err error) {
	client := http.Client{}

	request, err := http.NewRequest(method, c.Endpoint(action...), body)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "OAuth "+c.Token)

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
