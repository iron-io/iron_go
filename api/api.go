// api provides common functionality for all the iron.io APIs
package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/manveru/go.iron/config"
)

type URL struct {
	URL      url.URL
	Settings config.Settings
}

func Action(cs config.Settings, prefix string, suffix ...string) *URL {
	parts := append([]string{prefix}, suffix...)
	for n, part := range parts {
		parts[n] = url.QueryEscape(part)
	}

	u := &URL{Settings: cs, URL: url.URL{}}
	u.URL.Scheme = cs.Scheme
	u.URL.Host = fmt.Sprintf("%s:%d", url.QueryEscape(cs.Host), cs.Port)
	u.URL.Path = fmt.Sprintf("/%s/projects/%s/%s", cs.ApiVersion, cs.ProjectId, strings.Join(parts, "/"))
	return u
}

func VersionAction(cs config.Settings) *URL {
	u := &URL{Settings: cs, URL: url.URL{Scheme: cs.Scheme}}
	u.URL.Host = fmt.Sprintf("%s:%d", url.QueryEscape(cs.Host), cs.Port)
	u.URL.Path = "/version"
	return u
}

func (u *URL) QueryAdd(key string, format string, value interface{}) *URL {
	query := u.URL.Query()
	query.Add(key, fmt.Sprintf(format, value))
	u.URL.RawQuery = query.Encode()
	return u
}

func (u *URL) Req(method string, in, out interface{}) (err error) {
	var reqBody io.Reader
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(data)
	}
	response, err := u.Request(method, reqBody)
	if err == nil && out != nil {
		return json.NewDecoder(response.Body).Decode(out)
	}
	return
}

var MaxRequestRetries = 5

func (u *URL) Request(method string, body io.Reader) (response *http.Response, err error) {
	client := http.Client{}

	var bodyBytes []byte
	if body == nil {
		bodyBytes = []byte{}
	} else {
		bodyBytes, err = ioutil.ReadAll(body)
		if err != nil {
			return nil, err
		}
	}

	request, err := http.NewRequest(method, u.URL.String(), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "OAuth "+u.Settings.Token)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Accept-Encoding", "gzip/deflate")
	request.Header.Set("User-Agent", u.Settings.UserAgent)

	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	for tries := 0; tries <= MaxRequestRetries; tries++ {
		request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		response, err = client.Do(request)
		if err != nil {
			if err == io.EOF {
				continue
			}
			return
		}

		if response.StatusCode == http.StatusServiceUnavailable {
			delay := (tries + 1) * 10 // smooth out delays from 0-2
			time.Sleep(time.Duration(delay*delay) * time.Millisecond)
		}

		break
	}

	// DumpResponse(response)
	if err = ResponseAsError(response); err != nil {
		return
	}

	return
}

func DumpRequest(req *http.Request) {
	out, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%q\n", out)
}

func DumpResponse(response *http.Response) {
	out, err := httputil.DumpResponse(response, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%q\n", out)
}

var HTTPErrorDescriptions = map[int]string{
	http.StatusUnauthorized:     "The OAuth token is either not provided or invalid",
	http.StatusNotFound:         "The resource, project, or endpoint being requested doesn't exist.",
	http.StatusMethodNotAllowed: "This endpoint doesn't support that particular verb",
	http.StatusNotAcceptable:    "Required fields are missing",
}

func ResponseAsError(response *http.Response) (err error) {
	if response.StatusCode == http.StatusOK {
		return nil
	}

	desc, found := HTTPErrorDescriptions[response.StatusCode]
	if found {
		return Error{Response: response, Error: response.Status + desc}
	}

	out := map[string]interface{}{}
	json.NewDecoder(response.Body).Decode(&out)
	if msg, ok := out["msg"]; ok {
		return errors.New(fmt.Sprint(msg))
	}

	return errors.New(response.Status + ": Unknown API Response")
}

type HTTPError struct {
	Response int
	Error    string
}

func (h HTTPError) Error() string { return h.Error }
