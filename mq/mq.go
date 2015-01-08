// IronMQ (elastic message queue) client library
package mq

import (
	"errors"
	"time"

	"github.com/iron-io/iron_go/api"
	"github.com/iron-io/iron_go/config"
)

type Queue struct {
	Settings config.Settings
	Name     string
}

type QueueSubscriber struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

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

type PushStatus struct {
	Retried    int    `json:"retried"`
	StatusCode int    `json:"status_code"`
	Status     string `json:"status"`
}

type Subscriber struct {
	Retried    int    `json:"retried"`
	StatusCode int    `json:"status_code"`
	Status     string `json:"status"`
	URL        string `json:"url"`
}

type Alert struct {
	Type      string `json:"type"`
	Direction string `json:direction`
	Trigger   int    `json:trigger`
	Queue     string `queue`
}

func New(queueName string) *Queue {
	return &Queue{Settings: config.Config("iron_mq"), Name: queueName}
}

// ConfigNew uses the specified settings over configuration specified in an iron.json file or
// environment variables to return a Queue object capable of acquiring information about or
// modifying the queue specified by queueName.
func ConfigNew(queueName string, settings *config.Settings) Queue {
	return Queue{Settings: config.ManualConfig("iron_mq", settings), Name: queueName}
}

func ListSettingsQueues(settings config.Settings, page int, perPage int) (queues []Queue, err error) {
	out := []struct {
		Id         string
		Project_id string
		Name       string
	}{}

	q := New("")
	q.Settings = settings
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

func ListProjectQueues(projectId string, token string, page int, perPage int) (queues []Queue, err error) {
	settings := config.Config("iron_mq")
	settings.ProjectId = projectId
	settings.Token = token
	return ListSettingsQueues(settings, page, perPage)
}

func ListQueues(page, perPage int) (queues []Queue, err error) {
	settings := config.Config("iron_mq")
	return ListProjectQueues(settings.ProjectId, settings.Token, page, perPage)
}

func (q Queue) queues(s ...string) *api.URL { return api.Action(q.Settings, "queues", s...) }

// This method is left to support backward compatibility.
// This method is replaced by func ListQueues(page, perPage int) (queues []Queue, err error)
func (q Queue) ListQueues(page, perPage int) (queues []Queue, err error) {
	return ListQueues(page, perPage)
}

func (q Queue) Info() (QueueInfo, error) {
	qi := QueueInfo{}
	err := q.queues(q.Name).Req("GET", nil, &qi)
	return qi, err
}

func (q Queue) Update(qi QueueInfo) (QueueInfo, error) {
	out := QueueInfo{}
	err := q.queues(q.Name).Req("POST", qi, &out)
	return out, err
}

func (q Queue) Delete() (bool, error) {
	err := q.queues(q.Name).Req("DELETE", nil, nil)
	success := err == nil
	return success, err
}

type Subscription struct {
	PushType     string
	Retries      int
	RetriesDelay int
}

// RemoveSubscribers removes subscribers.
func (q Queue) RemoveSubscribers(subscribers ...string) (err error) {
	qi := QueueInfo{Subscribers: make([]QueueSubscriber, len(subscribers))}
	for i, subscriber := range subscribers {
		qi.Subscribers[i].URL = subscriber
	}
	return q.queues(q.Name, "subscribers").Req("DELETE", &qi, nil)
}

// AddSubscribers adds subscribers.
func (q Queue) AddSubscribers(subscribers ...string) (err error) {
	qi := QueueInfo{Subscribers: make([]QueueSubscriber, len(subscribers))}
	for i, subscriber := range subscribers {
		qi.Subscribers[i].URL = subscriber
	}
	return q.queues(q.Name, "subscribers").Req("POST", &qi, nil)
}

func (q Queue) PushString(body string) (id string, err error) {
	ids, err := q.PushStrings(body)
	if err != nil {
		return
	}
	return ids[0], nil
}

// Push adds one or more messages to the end of the queue using IronMQ's defaults:
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

func (q Queue) PushMessage(msg *Message) (id string, err error) {
	ids, err := q.PushMessages(msg)
	if err != nil {
		return
	}
	return ids[0], nil
}

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
		err = errors.New("Couldn't get a single message")
	}

	return
}

// get N messages
func (q Queue) GetN(n int) (msgs []*Message, err error) {
	return q.GetNWithTimeoutAndWait(n, 0, 0)
}

func (q Queue) GetNWithTimeout(n, timeout int) (msgs []*Message, err error) {
	return q.GetNWithTimeoutAndWait(n, timeout, 0)
}

func (q Queue) GetNWithTimeoutAndWait(n, timeout, wait int) (msgs []*Message, err error) {
	out := struct {
		Messages []*Message `json:"messages"`
	}{}

	err = q.queues(q.Name, "messages").
		QueryAdd("n", "%d", n).
		QueryAdd("timeout", "%d", timeout).
		QueryAdd("wait", "%d", wait).
		Req("GET", nil, &out)
	if err != nil {
		return
	}

	for _, msg := range out.Messages {
		msg.q = q
	}

	return out.Messages, nil
}

func (q Queue) Peek() (msg *Message, err error) {
	msgs, err := q.PeekN(1)
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

// peek N messages
func (q Queue) PeekN(n int) (msgs []*Message, err error) {
	msgs, err = q.PeekNWithTimeout(n, 0)

	return
}

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

// Delete all messages in the queue
func (q Queue) Clear() (err error) {
	return q.queues(q.Name, "clear").Req("POST", nil, nil)
}

// Delete message from queue
func (q Queue) DeleteMessage(msgId string) (err error) {
	return q.queues(q.Name, "messages", msgId).Req("DELETE", nil, nil)
}

// Reset timeout of message to keep it reserved
func (q Queue) TouchMessage(msgId string) (err error) {
	return q.queues(q.Name, "messages", msgId, "touch").Req("POST", nil, nil)
}

// Put message back in the queue, message will be available after +delay+ seconds.
func (q Queue) ReleaseMessage(msgId string, delay int64) (err error) {
	in := struct {
		Delay int64 `json:"delay"`
	}{Delay: delay}
	return q.queues(q.Name, "messages", msgId, "release").Req("POST", &in, nil)
}

func (q Queue) MessageSubscribers(msgId string) ([]*Subscriber, error) {
	out := struct {
		Subscribers []*Subscriber `json:"subscribers"`
	}{}
	err := q.queues(q.Name, "messages", msgId, "subscribers").Req("GET", nil, &out)
	return out.Subscribers, err
}

func (q Queue) MessageSubscribersPollN(msgId string, n int) ([]*Subscriber, error) {
	subs, err := q.MessageSubscribers(msgId)
	for {
		time.Sleep(100 * time.Millisecond)
		subs, err = q.MessageSubscribers(msgId)
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

func (q Queue) AddAlerts(alerts ...*Alert) (err error) {
	in := struct {
		Alerts []*Alert `json:"alerts"`
	}{Alerts: alerts}
	return q.queues(q.Name, "alerts").Req("POST", &in, nil)
}

func (q Queue) UpdateAlerts(alerts ...*Alert) (err error) {
	in := struct {
		Alerts []*Alert `json:"alerts"`
	}{Alerts: alerts}
	return q.queues(q.Name, "alerts").Req("PUT", &in, nil)
}

func (q Queue) RemoveAllAlerts() (err error) {
	return q.queues(q.Name, "alerts").Req("DELETE", nil, nil)
}

type AlertInfo struct {
	Id string `json:"id"`
}

func (q Queue) RemoveAlerts(alertIds ...string) (err error) {
	in := struct {
		Alerts []AlertInfo `json:"alerts"`
	}{Alerts: make([]AlertInfo, len(alertIds))}
	for i, alertId := range alertIds {
		(in.Alerts[i]).Id = alertId
	}
	return q.queues(q.Name, "alerts").Req("DELETE", &in, nil)
}

func (q Queue) RemoveAlert(alertId string) (err error) {
	return q.queues(q.Name, "alerts", alertId).Req("DELETE", nil, nil)
}

// Delete message from queue
func (m Message) Delete() (err error) {
	return m.q.DeleteMessage(m.Id)
}

// Reset timeout of message to keep it reserved
func (m Message) Touch() (err error) {
	return m.q.TouchMessage(m.Id)
}

// Put message back in the queue, message will be available after +delay+ seconds.
func (m Message) Release(delay int64) (err error) {
	return m.q.ReleaseMessage(m.Id, delay)
}

func (m Message) Subscribers() (interface{}, error) {
	return m.q.MessageSubscribers(m.Id)
}
