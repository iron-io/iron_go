// A library to talk with IronWorker (iron.io)
package worker

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/iron-io/go.iron/api"
	"github.com/iron-io/go.iron/config"
)

type Worker struct {
	Settings config.Settings
}

func New() *Worker {
	return &Worker{Settings: config.Config("iron_worker")}
}

func (w *Worker) codes(s ...string) *api.URL     { return api.Action(w.Settings, "codes", s...) }
func (w *Worker) tasks(s ...string) *api.URL     { return api.Action(w.Settings, "tasks", s...) }
func (w *Worker) schedules(s ...string) *api.URL { return api.Action(w.Settings, "schedules", s...) }

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
		defer close(out)
		for {
			info, err := w.TaskInfo(taskId)
			if err != nil {
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

func (w *Worker) WaitForTaskLog(taskId string) chan []byte {
	out := make(chan []byte)
	go func() {
		defer close(out)
		for {
			log, err := w.TaskLog(taskId)
			if err != nil {
				println(err.Error())
				return
			}
			out <- log
			return
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
