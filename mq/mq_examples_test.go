package mq_test

import (
	"fmt"
	"github.com/iron-io/iron_go/mq"
)

var p = fmt.Println

func Example1PushingMessagesToTheQueue() {
	// use the test_queue to push/get messages
	q := mq.New("test_queue")

	q.PushString("Hello, World!")

	// You can also pass multiple messages in a single call.
	q.PushStrings("Message 1", "Message 2")

	// To control parameters like timeout and delay, construct your own message.
	q.PushMessage(&mq.Message{Timeout: 60, Delay: 0, Body: "Hi there"})

	// And finally, all that can be done in bulk as well.
	q.PushMessages(
		&mq.Message{Timeout: 60, Delay: 0, Body: "The first"},
		&mq.Message{Timeout: 60, Delay: 1, Body: "The second"},
		&mq.Message{Timeout: 60, Delay: 2, Body: "The third"},
		&mq.Message{Timeout: 60, Delay: 3, Body: "The fifth"},
	)

	p("all pushed")

	// Output:
	// all pushed
}

func Example2GettingMessagesOffTheQueue() {
	q := mq.New("test_queue")

	// get a single message
	msg, err := q.Get()
	p(err)
	p(msg.Body)

	// get 5 messages
	msgs, err := q.GetN(5)
	p(err)
	p(len(msgs))

	for _, m := range append(msgs, msg) {
		m.Delete()
	}

	// Output:
	// <nil>
	// Hello, World!
	// <nil>
	// 5
}

func Example3DeleteMessagesFromTheQueue() {
	q := mq.New("test_queue")
	msg, err := q.Get()
	p(err)
	msg.Delete()

	// Output:
	// <nil>
}

func Example4ClearQueue() {
	q := mq.New("test_queue")

	info, err := q.Info()

	p(err)
	p("Before Clean(); Name:", info.Name, "Size:", info.Size)

	err = q.Clear()
	p(err)

	info, err = q.Info()

	p(err)
	p("After  Clean(); Name:", info.Name, "Size:", info.Size)

	// Output:
	// <nil>
	// Before Clean(); Name: test_queue Size: 1
	// <nil>
	// <nil>
	// After  Clean(); Name: test_queue Size: 0
}
