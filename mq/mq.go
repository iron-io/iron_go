// Package mq contains the IronMQ (elastic message queue) Go client library
package mq

import (
	"errors"
	"time"

	"github.com/iron-io/iron_go/api"
	"github.com/iron-io/iron_go/config"
)

var (
	// ErrNoMessages is returned when an attempt is made to get or peek a message from the queue,
	// but no more messages are present.
	ErrNoMessages = errors.New("mq: Couldn't get a single message")
)

// Queue represents an IronMQ message queue
type Queue struct {
	Settings config.Settings
	Name     string
}

// QueueSubscriber represents a HTTP endpoint subscriber for an IronMQ message queue
type QueueSubscriber struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

// QueueInfo represents information about an IronMQ message queue
type QueueInfo struct {
	Id            string            `json:"id,omitempty"`
	Name          string            `json:"name,omitempty"`
	PushType      string            `json:"push_type,omitempty"`
	Reserved      int               `json:"reserved,omitempty"`
	RetriesDelay  int               `json:"retries,omitempty"`
	Retries       int               `json:"retries_delay,omitempty"`
	Size          int               `json:"size,omitempty"`
	Subscribers   []QueueSubscriber `json:"subscribers,omitempty"`
	Alerts        []Alert           `json:"alerts,omitempty"`
	TotalMessages int               `json:"total_messages,omitempty"`
	ErrorQueue    string            `json:"error_queue,omitempty"`
}

// Message represents a message which can be sent and retrieved from an IronMQ message queue
type Message struct {
	Id   string `json:"id,omitempty"`
	Body string `json:"body"`
	// Timeout is the amount of time in seconds allowed for processing the
	// message.
	Timeout int64 `json:"timeout,omitempty"`
	// Delay is the amount of time in seconds to wait before adding the message
	// to the queue.
	Delay int64 `json:"delay,omitempty"`
	q     Queue
}

// PushStatus represents the status of a pushed message to a subscriber
type PushStatus struct {
	Retried    int    `json:"retried"`
	StatusCode int    `json:"status_code"`
	Status     string `json:"status"`
}

// Subscriber represents an HTTP endpoint which retrieves information from an IronMQ message
// queue, and the status of the subscriber
type Subscriber struct {
	Retried    int    `json:"retried"`
	StatusCode int    `json:"status_code"`
	Status     string `json:"status"`
	URL        string `json:"url"`
}

// Alert represents an alert for a pull queue
type Alert struct {
	Type      string `json:"type"`
	Direction string `json:direction`
	Trigger   int    `json:trigger`
	Queue     string `queue`
}

// New returns a new Queue struct containing the specified name, and using the default settings
// for an IronMQ message queue
func New(queueName string) *Queue {
	return &Queue{Settings: config.Config("iron_mq"), Name: queueName}
}

// ListQueues returns a slice of Queue structs, specifying the page and number of queues per
// page to return from the API request
func ListQueues(page, perPage int) (queues []Queue, err error) {
	out := []struct {
		Id         string
		Project_id string
		Name       string
	}{}

	q := New("")
	err = q.queues().
		QueryAdd("page", "%d", page).
		QueryAdd("per_page", "%d", perPage).
		Req("GET", nil, &out)
	if err != nil {
		return
	}

	queues = make([]Queue, 0, len(out))
	for _, item := range out {
		queues = append(queues, Queue{
			Settings: q.Settings,
			Name:     item.Name,
		})
	}

	return
}

func (q Queue) queues(s ...string) *api.URL { return api.Action(q.Settings, "queues", s...) }

// ListQueues method is left to support backward compatibility.
// This method is replaced by func ListQueues(page, perPage int) (queues []Queue, err error)
func (q Queue) ListQueues(page, perPage int) (queues []Queue, err error) {
	return ListQueues(page, perPage)
}

// Info returns a QueueInfo struct containing information about the specified queue
func (q Queue) Info() (QueueInfo, error) {
	qi := QueueInfo{}
	err := q.queues(q.Name).Req("GET", nil, &qi)
	return qi, err
}

// Update modifies the queue information using the input QueueInfo struct
func (q Queue) Update(qi QueueInfo) (QueueInfo, error) {
	out := QueueInfo{}
	err := q.queues(q.Name).Req("POST", qi, &out)
	return out, err
}

// Delete deletes the current IronMQ message queue, returning both true and nil error
// on success, or false and an error on failure
func (q Queue) Delete() (bool, error) {
	err := q.queues(q.Name).Req("DELETE", nil, nil)
	success := err == nil
	return success, err
}

// Subscription represents a HTTP endpoint subscription to an IronMQ message queue
type Subscription struct {
	PushType     string
	Retries      int
	RetriesDelay int
}

// Subscribe adds a subscription to an IronMQ message queue, with one or more subscribers
// as specified by the variadic string argument
func (q Queue) Subscribe(subscription Subscription, subscribers ...string) (err error) {
	in := QueueInfo{
		PushType:     subscription.PushType,
		Retries:      subscription.Retries,
		RetriesDelay: subscription.RetriesDelay,
		Subscribers:  make([]QueueSubscriber, len(subscribers)),
	}
	for i, subscriber := range subscribers {
		in.Subscribers[i].URL = subscriber
	}
	return q.queues(q.Name).Req("POST", &in, nil)
}

// PushString pushes a simple string to an IronMQ message queue, and returns the ID
// of the message
func (q Queue) PushString(body string) (id string, err error) {
	ids, err := q.PushStrings(body)
	if err != nil {
		return
	}
	return ids[0], nil
}

// PushStrings adds one or more messages to the end of the queue using IronMQ's defaults:
//	timeout - 60 seconds
//	delay - none
//
// Identical to PushMessages with Message{Timeout: 60, Delay: 0}
func (q Queue) PushStrings(bodies ...string) (ids []string, err error) {
	msgs := make([]*Message, 0, len(bodies))
	for _, body := range bodies {
		msgs = append(msgs, &Message{Body: body})
	}

	return q.PushMessages(msgs...)
}

// PushMessage pushes a Message struct to an IronMQ message queue, and returns the ID
// of the message
func (q Queue) PushMessage(msg *Message) (id string, err error) {
	ids, err := q.PushMessages(msg)
	if err != nil {
		return
	}
	return ids[0], nil
}

// PushMessages adds one or more messages to the end of an IronMQ message queue
func (q Queue) PushMessages(msgs ...*Message) (ids []string, err error) {
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
func (q Queue) Get() (msg *Message, err error) {
	msgs, err := q.GetN(1)
	if err != nil {
		return
	}

	if len(msgs) > 0 {
		msg = msgs[0]
	} else {
		err = ErrNoMessages
	}

	return
}

// GetN reserves N messages on the queue using the default timeout
func (q Queue) GetN(n int) (msgs []*Message, err error) {
	msgs, err = q.GetNWithTimeout(n, 0)

	return
}

// GetNWithTimeout reserves N messages on the queue using the specified timeout
func (q Queue) GetNWithTimeout(n, timeout int) (msgs []*Message, err error) {
	out := struct {
		Messages []*Message `json:"messages"`
	}{}

	err = q.queues(q.Name, "messages").
		QueryAdd("n", "%d", n).
		QueryAdd("timeout", "%d", timeout).
		Req("GET", nil, &out)
	if err != nil {
		return
	}

	for _, msg := range out.Messages {
		msg.q = q
	}

	return out.Messages, nil
}

// Peek looks at the next message on the queue, without reserving it
func (q Queue) Peek() (msg *Message, err error) {
	msgs, err := q.PeekN(1)
	if err != nil {
		return
	}

	if len(msgs) > 0 {
		msg = msgs[0]
	} else {
		err = ErrNoMessages
	}

	return
}

// PeekN looks at the next N messages on the queue using the default timeout
func (q Queue) PeekN(n int) (msgs []*Message, err error) {
	msgs, err = q.PeekNWithTimeout(n, 0)

	return
}

// PeekNWithTimeout looks at the next N messages on the queue using the default timeout
func (q Queue) PeekNWithTimeout(n, timeout int) (msgs []*Message, err error) {
	out := struct {
		Messages []*Message `json:"messages"`
	}{}

	err = q.queues(q.Name, "messages", "peek").
		QueryAdd("n", "%d", n).
		QueryAdd("timeout", "%d", timeout).
		Req("GET", nil, &out)
	if err != nil {
		return
	}

	for _, msg := range out.Messages {
		msg.q = q
	}

	return out.Messages, nil
}

// Clear deletes all messages in the queue
func (q Queue) Clear() (err error) {
	return q.queues(q.Name, "clear").Req("POST", nil, nil)
}

// DeleteMessage removes a message with the specified ID from queue
func (q Queue) DeleteMessage(msgID string) (err error) {
	return q.queues(q.Name, "messages", msgID).Req("DELETE", nil, nil)
}

// TouchMessage resets the timeout of message with the specified ID, to keep it reserved
func (q Queue) TouchMessage(msgID string) (err error) {
	return q.queues(q.Name, "messages", msgID, "touch").Req("POST", nil, nil)
}

// ReleaseMessage puts a message back in the queue, and then makes the message
// available again after +delay+ seconds.
func (q Queue) ReleaseMessage(msgID string, delay int64) (err error) {
	in := struct {
		Delay int64 `json:"delay"`
	}{Delay: delay}
	return q.queues(q.Name, "messages", msgID, "release").Req("POST", &in, nil)
}

// MessageSubscribers returns a slice of Subscriber structs attached to the message
// with the specified ID
func (q Queue) MessageSubscribers(msgID string) ([]*Subscriber, error) {
	out := struct {
		Subscribers []*Subscriber `json:"subscribers"`
	}{}
	err := q.queues(q.Name, "messages", msgID, "subscribers").Req("GET", nil, &out)
	return out.Subscribers, err
}

// MessageSubscribersPollN returns a slice of Subscriber structs attached to the message
// with the specified ID, while also polling them to ensure that at least N are ready
func (q Queue) MessageSubscribersPollN(msgID string, n int) ([]*Subscriber, error) {
	subs, err := q.MessageSubscribers(msgID)
	for {
		time.Sleep(100 * time.Millisecond)
		subs, err = q.MessageSubscribers(msgID)
		if err != nil {
			return subs, err
		}
		if len(subs) >= n && actualPushStatus(subs) {
			return subs, nil
		}
	}
	return subs, err
}

func actualPushStatus(subs []*Subscriber) bool {
	for _, sub := range subs {
		if sub.Status == "queued" {
			return false
		}
	}

	return true
}

// AddAlerts adds a variadic number of Alert structs to a pull queue
func (q Queue) AddAlerts(alerts ...*Alert) (err error) {
	in := struct {
		Alerts []*Alert `json:"alerts"`
	}{Alerts: alerts}
	return q.queues(q.Name, "alerts").Req("POST", &in, nil)
}

// UpdateAlerts replaces a variadic number of Alert structs on a pull queue
func (q Queue) UpdateAlerts(alerts ...*Alert) (err error) {
	in := struct {
		Alerts []*Alert `json:"alerts"`
	}{Alerts: alerts}
	return q.queues(q.Name, "alerts").Req("PUT", &in, nil)
}

// RemoveAllAlerts removes all alerts from an IronMQ message queue
func (q Queue) RemoveAllAlerts() (err error) {
	return q.queues(q.Name, "alerts").Req("DELETE", nil, nil)
}

// AlertInfo represents information about an alert
type AlertInfo struct {
	Id string `json:"id"`
}

// RemoveAlerts removes a variadic number of alerts with the specified alert IDs
// from an IronMQ message queue
func (q Queue) RemoveAlerts(alertIds ...string) (err error) {
	in := struct {
		Alerts []AlertInfo `json:"alerts"`
	}{Alerts: make([]AlertInfo, len(alertIds))}
	for i, alertID := range alertIds {
		(in.Alerts[i]).Id = alertID
	}
	return q.queues(q.Name, "alerts").Req("DELETE", &in, nil)
}

// RemoveAlert removes a single alert with the specified alert ID from an IronMQ
// message queue
func (q Queue) RemoveAlert(alertID string) (err error) {
	return q.queues(q.Name, "alerts", alertID).Req("DELETE", nil, nil)
}

// Delete removes a message from queue
func (m Message) Delete() (err error) {
	return m.q.DeleteMessage(m.Id)
}

// Touch reset the timeout of message to keep it reserved
func (m Message) Touch() (err error) {
	return m.q.TouchMessage(m.Id)
}

// Release puts a message back in the queue, and will make the message available
// again after +delay+ seconds.
func (m Message) Release(delay int64) (err error) {
	return m.q.ReleaseMessage(m.Id, delay)
}

// Subscribers returns the subscribers attached to this message
func (m Message) Subscribers() (interface{}, error) {
	return m.q.MessageSubscribers(m.Id)
}
