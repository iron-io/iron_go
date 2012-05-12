package cache

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Cache struct {
	Domain    string
	Token     string
	ProjectId string
	Name      string
}

func (c Cache) Endpoint(action string) string {
	// BaseURL: "http://staging-dev.iron.io.s3-website-us-east-1.amazonaws.com:433",
	return "https://" + c.Domain + ".iron.io/1/projects/" + c.ProjectId + "/caches/" + action
}

func (c Cache) request(method, action string, body interface{}) (response *http.Response, err error) {
	client := http.Client{}

	request, err := http.NewRequest("GET", c.Endpoint(action), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Accept-Encoding", "gzip/deflate")
	request.Header.Set("Authorization", "OAuth "+c.Token)

	dumpRequest(request)
	if response, err = client.Do(request); err != nil {
		return
	}
	if err = resToErr(response); err != nil {
		return
	}

	dumpResponse(response)
	return
}

func Caches(token, projectId string) (caches []Cache, err error) {
	c := Cache{
		Domain:    "cache-aws-us-east-1",
		Token:     "asdf",
		ProjectId: "qwert",
	}

	response, err := c.request("GET", "caches", nil)
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

// Put adds an Item to the cache.
func (c Cache) Put(key string, item *Item) (err error) {
	body := &bytes.Buffer{}
	encoder := json.NewEncoder(body)
	encoder.Encode(item)
	_, err = c.request("PUT", "items/"+key, body)
	return
}

// Increment increments the corresponding item's value.
func (c Cache) Increment(key string, amount int64) (err error) {
	body := &bytes.Buffer{}
	encoder := json.NewEncoder(body)
	encoder.Encode(map[string]int64{"amount": amount})
	_, err = c.request("POST", "items/"+url.QueryEscape(key), body)
	return
}

// Get gets an item from the cache.
func (c Cache) Get(key string) (value string, err error) {
	//projects/{Project ID}/caches/{Cache Name}/items/{Key}	GET	Get an Item from a Cache

	response, err := c.request("GET", "items/"+url.QueryEscape(key), nil)
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

	return body.Value
}

// Delete removes an item from the cache.
func (c Cache) Delete(key string) (err error) {
	_, err = c.request("DELETE", "items/"+url.QueryEscape(key), nil)
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
