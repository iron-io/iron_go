// api provides common functionality for all the iron.io APIs
package api

import (
	"bytes"
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
	u.URL.Scheme = cs.Protocol
	u.URL.Host = fmt.Sprintf("%s:%d", url.QueryEscape(cs.Host), cs.Port)
	u.URL.Path = fmt.Sprintf("/%s/projects/%s/%s", cs.ApiVersion, cs.ProjectId, strings.Join(parts, "/"))
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

func (u *URL) Request(method string, body io.Reader) (response *http.Response, err error) {
	client := http.Client{}

	request, err := http.NewRequest(method, u.URL.String(), body)
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

	// DumpRequest(request)
	if response, err = client.Do(request); err != nil {
		return
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

func ResponseAsError(response *http.Response) (err error) {
	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return errors.New(response.Status + ": The OAuth token is either not provided or invalid")
	case http.StatusNotFound:
		return errors.New(response.Status + ": The resource, project, or endpoint being requested doesn't exist.")
	case http.StatusMethodNotAllowed:
		return errors.New(response.Status + ": This endpoint doesn't support that particular verb")
	case http.StatusNotAcceptable:
		return errors.New(response.Status + ": Required fields are missing")
	default:
		out := map[string]interface{}{}
		json.NewDecoder(response.Body).Decode(&out)
		if msg, ok := out["msg"]; ok {
			return errors.New(fmt.Sprint(msg))
		} else {
			return errors.New(response.Status + ": Unknown API Response")
		}
	}

	panic("There is no way you'll encounter this")
}
