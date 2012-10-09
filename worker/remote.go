package worker

import (
	"encoding/json"
	"flag"
	"os"
)

var (
	dFlag       = flag.String("d", "", "root")
	payloadFlag = flag.String("payload", "", "payload file")
	idFlag      = flag.String("id", "", "task id")
)

func init() {
	flag.Parse()
}

func PayloadReader() (io.Reader, error) {
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
