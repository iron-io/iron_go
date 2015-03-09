package worker

import (
	"encoding/json"
	"flag"
	"io"
	"os"
)


var (
	TaskDir     string
	payloadFlag string
	TaskId      string
	configFlag  string
)

// call this to parse flags before using the other methods.
func ParseFlags() {
	flag.StringVar(&TaskDir, "d", "", "task dir")
	flag.StringVar(&payloadFlag, "payload", "", "payload file")
	flag.StringVar(&TaskId, "id", "", "task id")
	flag.StringVar(&configFlag, "config", "", "config file")
	flag.Parse()
}

func PayloadReader() (io.ReadCloser, error) {
	return os.Open(payloadFlag)
}

func PayloadFromJSON(v interface{}) error {
	reader, err := PayloadReader()
	if err != nil {
		return err
	}
	defer reader.Close()
	return json.NewDecoder(reader).Decode(v)
}

func IronTaskId() string {
	return idFlag
}
