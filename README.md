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

```go
queue = := mq.New("test_queue");
```

Now you can use it.

--



