package mq_test

import (
	"fmt"
	"github.com/iron-io/iron_go/mq"
)

var p = fmt.Println

func assert(b bool, msg ...interface{}) {
	if !b {
		panic(fmt.Sprintln(msg...))
	}
}

func Example1PushingMessagesToTheQueue() {
	// use a queue named "test_queue" to push/get messages
	q := mq.New("test_queue")

	total := 0

	id, err := q.PushString("Hello, World!")
	assert(err == nil, err)
	assert(len(id) > 1, len(id))
	total++

	// You can also pass multiple messages in a single call.
	ids, err := q.PushStrings("Message 1", "Message 2")
	assert(err == nil, err)
	assert(len(ids) == 2, len(ids))
	total += len(ids)

	// To control parameters like timeout and delay, construct your own message.
	id, err = q.PushMessage(&mq.Message{Timeout: 60, Delay: 0, Body: "Hi there"})
	assert(err == nil, err)
	assert(len(id) > 10, len(id))
	total++

	// And finally, all that can be done in bulk as well.
	ids, err = q.PushMessages(
		&mq.Message{Timeout: 60, Delay: 0, Body: "The first"},
		&mq.Message{Timeout: 60, Delay: 1, Body: "The second"},
		&mq.Message{Timeout: 60, Delay: 2, Body: "The third"},
		&mq.Message{Timeout: 60, Delay: 3, Body: "The fifth"},
	)
	assert(err == nil, err)
	assert(len(ids) == 4, len(ids))
	total += len(ids)

	p("pushed a total of", total, "messages")

	// Output:
	// pushed a total of 8 messages
}

func Example2GettingMessagesOffTheQueue() {
	q := mq.New("test_queue")

	// get a single message
	msg, err := q.Get()
	assert(err == nil, err)
	fmt.Sprintf("The message says: %q\n", msg.Body)

	// when we're done handling a message, we have to delete it, or it
	// will be put back into the queue after a timeout.

	// get 5 messages
	msgs, err := q.GetN(5)
	assert(err == nil, err)
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
	err = msg.Delete()
	p(err)

	// Output:
	// <nil>
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
