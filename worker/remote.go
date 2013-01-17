package worker

import (
	"encoding/json"
	"flag"
	"io"
	"os"
)


var (
	taskDirFlag string
	payloadFlag string
	idFlag      string
)

// call this to parse flags before using the other methods.
func ParseFlags() {
	flag.StringVar(&taskDirFlag, "d", "", "task dir")
	flag.StringVar(&payloadFlag, "payload", "", "payload file")
	flag.StringVar(&idFlag, "id", "", "task id")
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
