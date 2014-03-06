Iron.io Go Client Library
-------------

# IronMQ

[IronMQ](http://www.iron.io/products/mq) is an elastic message queue for managing data and event flow within cloud applications and between systems.

The [full API documentation is here](http://dev.iron.io/mq/reference/api/) and this client tries to stick to the API as
much as possible so if you see an option in the API docs, you can use it in the methods below. 

You can find [Go docs here](http://go.pkgdoc.org/github.com/iron-io/iron_go).

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
queues, err := q.ListQueues(0, 100);
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

As of now the library doesn't support deleting of queues.

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
