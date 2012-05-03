// A library to talk with IronWorker (iron.io)
package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"time"
)

type Worker struct {
	ProjectId, Token, UserAgent string
	ApiVersion                  int
	BaseURL                     *url.URL
}

func New(projectId, token string) *Worker {
	baseURL, err := url.ParseRequestURI("https://worker-aws-us-east-1.iron.io:443/")
	// baseURL, err := url.ParseRequestURI("http://localhost:7001/")
	if err != nil {
		panic(err)
	}

	return &Worker{
		Token:      token,
		ProjectId:  projectId,
		BaseURL:    baseURL,
		ApiVersion: 2,
		UserAgent:  "go.iron/worker 0.1",
	}
}

func dumpRequest(req *http.Request) {
	out, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%q\n", out)
}

func dumpResponse(res *http.Response) {
	out, err := httputil.DumpResponse(res, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%q\n", out)
}

func (w *Worker) request(method, action string, body io.Reader) (res *http.Response, err error) {
	client := http.Client{}
	uri := fmt.Sprintf("%s%d/projects/%s/%s", w.BaseURL, w.ApiVersion, w.ProjectId, action)
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip/deflate")
	req.Header.Set("Authorization", "OAuth "+w.Token)

	switch method {
	case "GET", "DELETE":
	default:
		req.Header.Set("Content-Type", "application/json")
	}

	// dumpRequest(req)

	res, err = client.Do(req)
	if res.StatusCode != httpOk {
		return res, resToErr(res)
	}

	// dumpResponse(res)

	return res, err
}

func (w *Worker) getJSON(action string, data interface{}) (err error) {
	res, err := w.request("GET", action, nil)
	if err != nil {
		return
	}

	err = json.NewDecoder(res.Body).Decode(data)

	return
}

func resToErr(res *http.Response) (err *APIError) {
	switch res.StatusCode {
	case httpUnauthorized:
		return &APIError{Response: res,
			Msg: "Invalid authentication: The OAuth token is either not provided or invalid"}
	case httpNotFound:
		return &APIError{Response: res,
			Msg: "Invalid endpoint: The resource, project, or endpoint being requested doesn't exist."}
	case httpMethodNotAllowed:
		return &APIError{Response: res,
			Msg: "Invalid HTTP method: This endpoint doesn't support that particular verb"}
	case httpNotAcceptable:
		return &APIError{Response: res,
			Msg: "Invalid request: Required fields are missing"}
	default:
		body := make([]byte, 0, res.ContentLength)
		res.Body.Read(body)
		msg := fmt.Sprintf("Unknown API Response %s: %q", res.Status, string(body))
		return &APIError{Msg: msg, Response: res}
	}

	panic("There is no way you'll encounter this")
}

func responseWithBody(res *http.Response, data interface{}) (err error) {
	return
}

const (
	httpOk               = 200
	httpUnauthorized     = 401
	httpNotFound         = 404
	httpMethodNotAllowed = 405
	httpNotAcceptable    = 406
)

type APIError struct {
	Msg      string
	Response *http.Response
}

func (a APIError) Error() string {
	return a.Msg
}

func (w *Worker) post(action string, obj interface{}) (res *http.Response, err error) {
	body := &bytes.Buffer{}
	encoder := json.NewEncoder(body)
	m := make(map[string]interface{}, 1)
	m[action] = obj
	encoder.Encode(m)

	return w.request("POST", action, body)
}

var GoCodeRunner = []byte(`#!/bin/sh
root() {
  while [ $# -gt 0 ]; do
    if [ "$1" = "-d" ]; then
      printf "%s\n" "$2"
      break
    fi
  done
}
cd "$(root "$@")"
chmod +x worker
./worker "$@"
`)

// NewGoCodePackage creates an IronWorker code package from a Go package.
// The codeName is the name the code package will have have once it is uploaded.
//
// The packageArgs are equivalent to the args you would give to `go build`.
//
// If the packageArgs are a list of .go files, they will be treated as a list of
// source files specifying a single Go package.
//
// Otherwise it's treated as a Go package name that will be searched in GOPATH,
// and compiled.
//
// Please note that if you don't use `package main` for your Go package, this
// function will not work as expected.
func NewGoCodePackage(codeName string, packageArgs ...string) (code Code, err error) {
	tempDir, err := ioutil.TempDir("", "iron-go-build-"+codeName)
	if err != nil {
		return
	}

	defer os.RemoveAll(tempDir)

	os.Setenv("GOOS", "linux")
	os.Setenv("GOARCH", "amd64")
	args := []string{"build", "-x", "-v", "-o", tempDir + "/worker"}
	cmd := exec.Command("go", append(args, packageArgs...)...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Fprintln(os.Stderr, "vvvvvvvvvvvvvvvvvvvv go build output vvvvvvvvvvvvvvvvvvvv")
		fmt.Fprintln(os.Stderr, string(output))
		fmt.Fprintln(os.Stderr, "^^^^^^^^^^^^^^^^^^^^ go build output ^^^^^^^^^^^^^^^^^^^^")
		return
	}

	fd, err := os.Open(tempDir + "/worker")
	if err != nil {
		return
	}
	defer fd.Close()
	workerExe, err := ioutil.ReadAll(fd)
	if err != nil {
		return
	}

	code = Code{
		Name:     codeName,
		Runtime:  "sh",
		FileName: "__runner__.sh",
		Source: CodeSource{
			"worker":        workerExe,
			"__runner__.sh": GoCodeRunner,
		},
	}

	return
}

// WaitForTask returns a channel that will receive the completed task and is closed afterwards.
// If an error occured during the wait, the channel will be closed.
func (w *Worker) WaitForTask(taskId string) chan TaskInfo {
	out := make(chan TaskInfo)
	go func() {
		for {
			info, err := w.TaskInfo(taskId)
			if err != nil {
				close(out)
				return
			}

			if info.Status == "queued" || info.Status == "running" {
				time.Sleep(5)
			} else {
				out <- info
				return
			}
		}
	}()

	return out
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}
