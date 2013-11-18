/*
	This code sample demonstrates how to queue a worker from your your existing
	task list.

	http://dev.iron.io/worker/reference/api/
	http://dev.iron.io/worker/reference/api/#queue_a_task
*/
package main

import (
	"bytes"
	"github.com/iron-io/iron_go/api"
	"github.com/iron-io/iron_go/config"
	"log"
)

// payload defines a sample payload document
var payload = `{"tasks":[
{
"code_name" : "Worker-Name",
"timeout" : 20,
"payload" : "{ \"key\" : \"value", \"key\" : \"value\" }"
}]}`

func main() {
	// Create your configuration for iron_worker
	// Find these value in credentials
	config := config.Config("iron_worker")
	config.ProjectId = "your_project_id"
	config.Token = "your_token"

	// Create your endpoint url for tasks
	url := api.ActionEndpoint(config, "tasks")
	log.Printf("Url: %s\n", url.URL.String())

	// Convert the payload to a slice of bytes
	postData := bytes.NewBufferString(payload)

	// Post the request to Iron.io
	resp, err := url.Request("POST", postData)
	if err != nil {
		log.Println(err)
	}

	// Check the status code
	if resp.StatusCode != 200 {
		log.Printf("%v\n", resp)
	}
}
