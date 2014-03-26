Iron.io Go Client Library
-------------

# IronMQ

[IronMQ](http://www.iron.io/products/mq) is an elastic message queue for managing data and event flow within cloud applications and between systems.

The [full API documentation is here](http://dev.iron.io/mq/reference/api/) and this client tries to stick to the API as
much as possible so if you see an option in the API docs, you can use it in the methods below. 

You can find [Go docs here](http://godoc.org/github.com/iron-io/iron_go).

## Getting Started

### Get credentials

To start using iron_go, you need to sign up and get an oauth token.

1. Go to http://iron.io/ and sign up.
2. Create new project at http://hud.iron.io/dashboard
3. Download the iron.json file from "Credentials" block of project

--

### Configure

1\. Reference the library:

```go
import "github.com/iron-io/iron_go/mq"
```

2\. [Setup your Iron.io credentials](http://dev.iron.io/mq/reference/configuration/)

3\. Create an IronMQ client object:

```go
queue := mq.New("test_queue");
```

## The Basics

### Get Queues List

```go
queues, err := mq.ListQueues(0, 100);
for _, element := range queues {
	fmt.Println(element.Name);
}
```

--

### Get a Queue Object

You can have as many queues as you want, each with their own unique set of messages.

```go
queue := mq.New("test_queue");
```

Now you can use it.

--

### Post a Message on a Queue

Messages are placed on the queue in a FIFO arrangement.
If a queue does not exist, it will be created upon the first posting of a message.

```go
id, err := q.PushString("Hello, World!")
```

--

### Retrieve Queue Information

```go
info, err := q.Info()
fmt.Println(info.Name);
```

--

### Get a Message off a Queue

```go
msg, err := q.Get()
fmt.Printf("The message says: %q\n", msg.Body)
```

--

### Delete a Message from a Queue

```go
msg, _ := q.Get()
// perform some actions with a message here
msg.Delete()
```

Be sure to delete a message from the queue when you're done with it.

--

## Queues

### Retrieve Queue Information

```go
info, err := q.Info()
fmt.Println(info.Name);
fmt.Println(info.Size);
```

QueueInfo struct consists of the following fields:

```go
type QueueInfo struct {
	Id            string            `json:"id,omitempty"`
	Name          string            `json:"name,omitempty"`
	PushType      string            `json:"push_type,omitempty"`
	Reserved      int               `json:"reserved,omitempty"`
	RetriesDelay  int               `json:"retries,omitempty"`
	Retries       int               `json:"retries_delay,omitempty"`
	Size          int               `json:"size,omitempty"`
	Subscribers   []QueueSubscriber `json:"subscribers,omitempty"`
	TotalMessages int               `json:"total_messages,omitempty"`
	ErrorQueue    string            `json:"error_queue,omitempty"`
}
```

--

### Delete a Message Queue

```go
deleted, err := q.Delete()
if(deleted) { 
  fmt.Println("Successfully deleted")
} else {
  fmt.Println("Cannot delete, because of error: ", err)
}
```

--

### Post Messages to a Queue

**Single message:**

```go
id, err := q.PushString("Hello, World!")
// To control parameters like timeout and delay, construct your own message.
id, err := q.PushMessage(&mq.Message{Timeout: 60, Delay: 0, Body: "Hi there"})
```

**Multiple messages:**

You can also pass multiple messages in a single call.

```go
ids, err := q.PushStrings("Message 1", "Message 2")
```

To control parameters like timeout and delay, construct your own message.

```go
ids, err = q.PushMessages(
	&mq.Message{Timeout: 60, Delay: 0,  Body: "The first"},
	&mq.Message{Timeout: 60, Delay: 10, Body: "The second"},
	&mq.Message{Timeout: 60, Delay: 10, Body: "The third"},
	&mq.Message{Timeout: 60, Delay: 0,  Body: "The fifth"},
)
```

**Parameters:**

* `Timeout`: After timeout (in seconds), item will be placed back onto queue.
You must delete the message from the queue to ensure it does not go back onto the queue.
 Default is 60 seconds. Minimum is 30 seconds. Maximum is 86,400 seconds (24 hours).

* `Delay`: The item will not be available on the queue until this many seconds have passed.
Default is 0 seconds. Maximum is 604,800 seconds (7 days).

--

### Get Messages from a Queue

```go
msg, err := q.Get()
fmt.Printf("The message says: %q\n", msg.Body)
```

When you pop/get a message from the queue, it is no longer on the queue but it still exists within the system.
You have to explicitly delete the message or else it will go back onto the queue after the `timeout`.
The default `timeout` is 60 seconds. Minimal `timeout` is 30 seconds.

You also can get several messages at a time:

```go
// get 5 messages
msgs, err := q.GetN(5)
```

And with timeout param:

```go
messages, err := q.GetNWithTimeout(4, 600)
```

### Touch a Message on a Queue

Touching a reserved message extends its timeout by the duration specified when the message was created, which is 60 seconds by default.

```go
msg, _ := q.Get()
err := msg.Touch()
```

There is another way to touch a message without getting it:

```go
err := q.TouchMessage("5987586196292186572")
```

--

### Release Message

```go
msg, _ := q.Get()
delay  := 30
err := msg.release(delay)
```

Or another way to release a message without creation of message object:

```go
delay := 30
err := q.ReleaseMessage("5987586196292186572", delay)
```

**Optional parameters:**

* `delay`: The item will not be available on the queue until this many seconds have passed.
Default is 0 seconds. Maximum is 604,800 seconds (7 days).

--

### Delete a Message from a Queue

```go
msg, _ := q.Get()
// perform some actions with a message here
err := msg.Delete()
```

Or

```go
err := q.DeleteMessage("5987586196292186572")
```

Be sure to delete a message from the queue when you're done with it.

--

### Peek Messages from a Queue

Peeking at a queue returns the next messages on the queue, but it does not reserve them.

```go
message, err := q.Peek()
```

There is a way to get several messages not reserving them:

```go
messages, err := q.PeekN(50)
for _, m := range messages {
  fmt.Println(m.Body)
}
```

And with timeout param:

```go
messages, err := q.PeekNWithTimeout(4, 600)
```

--

### Clear a Queue

```go
err := q.Clear()
```

### Add an Alert to a Queue

[Check out our Blog Post on Queue Alerts](http://blog.iron.io).

Alerts have now been incorporated into IronMQ. This feature lets developers control actions based on the activity within a queue. With alerts, actions can be triggered when the number of messages in a queue reach a certain threshold. These actions can include things like auto-scaling, failure detection, load-monitoring, and system health.

You may add up to 5 alerts per queue.

**Required parameters:**
* `type`: required - "fixed" or "progressive". In case of alert's type set to "fixed", alert will be triggered when queue size pass value set by trigger parameter. When type set to "progressive", alert will be triggered when queue size pass any of values, calculated by trigger * N where N >= 1. For example, if trigger set to 10, alert will be triggered at queue sizes 10, 20, 30, etc.
* `direction`: required - "asc" or "desc". Set direction in which queue size must be changed when pass trigger value. If direction set to "asc" queue size must growing to trigger alert. When direction is "desc" queue size must decreasing to trigger alert.
* `trigger`: required. It will be used to calculate actual values of queue size when alert must be triggered. See type field description. Trigger must be integer value greater than 0.
* `queue`: required. Name of queue which will be used to post alert messages.

```go
err := q.AddAlerts(
  &mq.Alert{Queue: "new_milestone_queue", Trigger: 10, Direction: "asc",  Type: "progressive"},
  &mq.Alert{Queue: "low_level_queue",     Trigger: 5,  Direction: "desc", Type: "fixed" })
```

#### Update alerts in a queue
```go
err := q.AddAlerts(
  &mq.Alert{Queue: "milestone_queue", Trigger: 100, Direction: "asc",  Type: "progressive"})
```

#### Remove alerts from a queue

You can delete an alert from a queue by id:

```go
err := q.RemoveAlert("532fdf593663ed6afa06ed16")
```

Or delete several alerts by ids:

```go
err := q.RemoveAlerts("532f59663ed6afed16483052", "559663ed6af6483399b3400a")
```

Also you can delete all alerts

```go
err := q.RemoveAllAlerts()
```

Please, remember, that passing zero of alerts while update process will lead to deleating of all previously added alerts.

```go
q.AddAlerts(
  &mq.Alert{Queue: "alert1", Trigger: 10, Direction: "asc", Type: "progressive"},
  &mq.Alert{Queue: "alert2", Trigger: 5,  Direction: "desc", Type: "fixed" })
info, _ := q.Info() // 2

q.UpdateAlerts()
info, _ = q.Info()  // 0
```

--

## Push Queues

IronMQ push queues allow you to setup a queue that will push to an endpoint, rather than having to poll the endpoint. 
[Here's the announcement for an overview](http://blog.iron.io/2013/01/ironmq-push-queues-reliable-message.html). 

### Update a Message Queue

```go
queueInfo := mq.QueueInfo{
	//...
} 
info, err := q.Update(queueInfo);
```

QueueInfo struct consists of following fields:

```go
type QueueInfo struct {
	PushType      string            `json:"push_type,omitempty"`
	RetriesDelay  int               `json:"retries,omitempty"`
	Retries       int               `json:"retries_delay,omitempty"`
	Subscribers   []QueueSubscriber `json:"subscribers,omitempty"`
	// and some other fields not related to push queues
}
```

**The following parameters are all related to Push Queues:**

* `push_type`: Either `multicast` to push to all subscribers or `unicast` to push to one and only one subscriber. Default is `multicast`.
* `retries`: How many times to retry on failure. Default is 3. Maximum is 100.
* `retries_delay`: Delay between each retry in seconds. Default is 60.
* `subscribers`: An array of `QueueSubscriber` 
This set of subscribers will replace the existing subscribers.
To add or remove subscribers, see the add subscribers endpoint or the remove subscribers endpoint.

QueueSubscriber has the following structure:

```go
type QueueSubscriber struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}
```

--

### Set Subscribers on a Queue

Subscribers can be any HTTP endpoint. `push_type` is one of:

* `multicast`: will push to all endpoints/subscribers
* `unicast`: will push to one and only one endpoint/subscriber

```go
subscription := mq.Subscription {
	PushType: "multicast",
	Retries:  3,
	RetriesDelay: 60,
}
err := q.Subscribe(
	subscription, 
	"http://mysterious-brook-1807.herokuapp.com/ironmq_push_3", 
	"http://mysterious-brook-1807.herokuapp.com/ironmq_push_4")
```

--

<!--- 

TODO: IMPLEMENT IT

### Add/Remove Subscribers on a Queue

```ruby
queue.add_subscriber({:url => "http://nowhere.com"})

queue.add_subscribers([
  {:url => 'http://first.endpoint.xx/process'},
  {:url => 'http://second.endpoint.xx/process'}
])


queue.remove_subscriber({url: "http://nowhere.com"})

queue.remove_subscribers([
  {:url => 'http://first.endpoint.xx/process'},
  {:url => 'http://second.endpoint.xx/process'}
])
```

- -

-->

### Get Message Push Status

After pushing a message:

```go
subscribers, err := message.Subscribers()
```

Returns an array of subscribers with status.

--

### Revert Queue Back to Pull Queue

If you want to revert you queue just update `push_type` to `'pull'`.

```ruby
q.Update(mq.QueueInfo{
	PushType: "pull",	
});
```

--

## Further Links

* [IronMQ Overview](http://dev.iron.io/mq/)
* [IronMQ REST/HTTP API](http://dev.iron.io/mq/reference/api/)
* [Push Queues](http://dev.iron.io/mq/reference/push_queues/)
* [Other Client Libraries](http://dev.iron.io/mq/libraries/)
* [Live Chat, Support & Fun](http://get.iron.io/chat)

-------------
© 2011 - 2014 Iron.io Inc. All Rights Reserved.
