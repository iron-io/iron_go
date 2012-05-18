package mq

import (
	"errors"
	"fmt"

	"github.com/manveru/go.iron/api"
	"github.com/manveru/go.iron/config"
)

type Queue struct {
	Settings config.Settings
	Name     string
}

type Message struct {
	Id   string `json:"id,omitempty"`
	Body string `json:"body"`
	// Timeout is the amount of time in seconds allowed for processing the
	// message.
	Timeout int64 `json:"timeout,omitempty"`
	// Delay is the amount of time in seconds to wait before adding the message
	// to the queue.
	Delay int64 `json:"delay,omitempty"`
	q     *Queue
}

func New(queueName string) *Queue {
	return &Queue{Settings: config.Config("iron_mq"), Name: queueName}
}

func (q *Queue) queues(s ...string) *api.URL { return api.Action(q.Settings, "queues", s...) }

func (q *Queue) ListQueues(page, perPage int) (queues []*Queue, err error) {
	out := []struct {
		Id         string
		Project_id string
		Name       string
	}{}

	err = q.queues().
		QueryAdd("page", "%d", page).
		QueryAdd("per_page", "%d", perPage).
		Req("GET", nil, &out)
	if err != nil {
		return
	}

	queues = make([]*Queue, 0, len(out))
	for _, item := range out {
		queues = append(queues, &Queue{
			Settings: q.Settings,
			Name:     item.Name,
		})
	}

	return
}

type QueueInfo struct {
	TotalMessages int    `json:"total_messages"`
	Name          string `json:"queuename"`
	Size          int    `json:"size"`
}

func (q *Queue) Info() (QueueInfo, error) {
	qi := QueueInfo{}
	err := q.queues(q.Name).Req("GET", nil, &qi)
	return qi, err
}

// Push adds a message to the end of the queue using IronMQ's defaults:
//	timeout - 60 seconds
//	delay - none
func (q *Queue) Push(msg string) (id string, err error) {
	return q.PushMessage(&Message{Body: msg})
}

func (q *Queue) PushMessage(msg *Message) (id string, err error) {
	ids, err := q.PushMessages(msg)
	if err != nil {
		return
	}
	return ids[0], nil
}

func (q *Queue) PushMessages(msgs ...*Message) (ids []string, err error) {
	in := struct {
		Messages []*Message `json:"messages"`
	}{Messages: msgs}

	out := struct {
		IDs []string `json:"ids"`
		Msg string   `json:"msg"`
	}{}

	err = q.queues(q.Name, "messages").Req("POST", &in, &out)
	return out.IDs, err
}

// Get reserves a message from the queue.
// The message will not be deleted, but will be reserved until the timeout
// expires. If the timeout expires before the message is deleted, the message
// will be placed back onto the queue.
// As a result, be sure to Delete a message after you're done with it.
func (q *Queue) Get() (msg *Message, err error) {
	msgs, err := q.GetN(1)
	if err != nil {
		return
	}

	if len(msgs) > 0 {
		msg = msgs[0]
	} else {
		err = errors.New("Couldn't get a single message")
	}

	return
}

func (q *Queue) GetN(n int) (msgs []*Message, err error) {
	out := struct {
		Messages []*Message `json:"messages"`
	}{}

	err = q.queues(q.Name, "messages").
		QueryAdd("n", "%d", n).
		Req("GET", nil, &out)
	if err != nil {
		return
	}

	for _, msg := range out.Messages {
		fmt.Println(msg)
		msg.q = q
	}

	return out.Messages, nil
}

func (q *Queue) DeleteMsg(msgId string) (err error) {
	return q.queues(q.Name, "messages", msgId).Req("DELETE", nil, nil)
}
