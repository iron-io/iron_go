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

type URL url.URL

func Action(cs config.Settings, prefix string, suffix ...string) *URL {
	parts := append([]string{prefix}, suffix...)
	for n, part := range parts {
		parts[n] = url.QueryEscape(part)
	}

	u := &URL{}
	u.Scheme = cs.Protocol
	u.Host = fmt.Sprintf("%s:%d", url.QueryEscape(cs.Host), cs.Port)
	u.Path = fmt.Sprintf("/%s/projects/%s/%s", cs.ApiVersion, cs.ProjectId, strings.Join(parts, "/"))
	return u
}

func (u *URL) QueryAdd(key string, format string, value interface{}) *URL {
	query := u.Query()
	query.Add(key, fmt.Sprintf(format, value))
	u.RawQuery = query.Encode()
	return u
}

func (u *URL) Req(s config.Settings, method string, in, out interface{}) (err error) {
	var reqBody io.Reader
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(data)
	}
	response, err := Request(s, method, u, reqBody)
	if err == nil && out != nil {
		return json.NewDecoder(response.Body).Decode(out)
	}
	return
}

func (u *URL) String() string    { return (*url.URL)(u).String() }
func (u *URL) Query() url.Values { return (*url.URL)(u).Query() }

func Request(s config.Settings, method string, url fmt.Stringer, body io.Reader) (response *http.Response, err error) {
	client := http.Client{}

	request, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "OAuth "+s.Token)

	if body == nil {
		request.Header.Set("Accept", "application/json")
		request.Header.Set("Accept-Encoding", "gzip/deflate")
	} else {
		request.Header.Set("Content-Type", "application/json")
	}

	DumpRequest(request)
	if response, err = client.Do(request); err != nil {
		return
	}
	DumpResponse(response)
	if err = ResToErr(response); err != nil {
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

func ResToErr(response *http.Response) (err error) {
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

func JSONRequest() {}
