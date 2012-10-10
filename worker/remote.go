package worker

import (
	"encoding/json"
	"flag"
	"io"
	"os"
)

var (
	envFlag     = flag.String("e", "", "environment")
	taskDirFlag = flag.String("d", "", "task dir")
	payloadFlag = flag.String("payload", "", "payload file")
	idFlag      = flag.String("id", "", "task id")
)

func init() {
	flag.Parse()
}

func PayloadReader() (io.ReadCloser, error) {
	return os.Open(*payloadFlag)
}

func PayloadFromJSON(v interface{}) error {
	fd, err := os.Open(*payloadFlag)
	if err != nil {
		return err
	}
	return json.NewDecoder(fd).Decode(v)
}

func IronTaskId() string {
	return *idFlag
}

func IronEnvironment() string {
	return *envFlag
}
