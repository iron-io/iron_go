package mq_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/iron-io/iron_go/mq"
	. "github.com/jeffh/go.bdd"
)

func TestUrl(t *testing.T) {
	fmt.Println("Testing URL with spaces")
	mq := mq.New("MyProject - Prod")
	id, err := mq.PushString("hello")
	if err != nil {
		t.Fatal("No good", err)
	}
	fmt.Println("id:", id)
	info, err := mq.Info()
	if err != nil {
		fmt.Println("ERROR:", err)
		t.Fatal("No good", err)
	}
	fmt.Println(info)
}

func TestEverything(t *testing.T) {
	defer PrintSpecReport()

	qname := "queuename3"

	Describe("IronMQ", func() {
		It("Deletes all existing messages", func() {
			c := mq.New(qname)
			c.PushString("hello") // just to ensure queue exists
			Expect(c.Clear(), ToBeNil)

			info, err := c.Info()
			Expect(err, ToBeNil)
			Expect(info.Size, ToEqual, 0x0)
		})

		It("Pushes ands gets a message", func() {
			c := mq.New(qname)
			id1, err := c.PushString("just a little test")
			Expect(err, ToBeNil)
			defer c.DeleteMessage(id1)

			msg, err := c.Get()
			Expect(err, ToBeNil)

			Expect(msg, ToNotBeNil)
			Expect(msg.Id, ToDeepEqual, id1)
			Expect(msg.Body, ToDeepEqual, "just a little test")
		})

		It("Delete messages", func() {
			q := mq.New(qname)

			strings := []string{}
			for n := 0; n < 100; n++ {
				strings = append(strings, fmt.Sprint("test: ", n))
			}

			_, err := q.PushStrings(strings...)
			Expect(err, ToBeNil)

			info, err := q.Info()
			Expect(err, ToBeNil)
			Expect(info.Size, ToEqual, 100)

			msgs, err := q.GetN(100)
			Expect(err, ToBeNil)

			Expect(q.DeleteMessages(msgs), ToBeNil)

			info, err = q.Info()
			Expect(err, ToBeNil)
			Expect(info.Size, ToEqual, 0)
		})

		It("clears the queue", func() {
			q := mq.New(qname)

			strings := []string{}
			for n := 0; n < 100; n++ {
				strings = append(strings, fmt.Sprint("test: ", n))
			}

			_, err := q.PushStrings(strings...)
			Expect(err, ToBeNil)

			info, err := q.Info()
			Expect(err, ToBeNil)
			Expect(info.Size, ToEqual, 100)

			Expect(q.Clear(), ToBeNil)

			info, err = q.Info()
			Expect(err, ToBeNil)
			Expect(info.Size, ToEqual, 0)
		})

		It("Lists all queues", func() {
			c := mq.New(qname)
			queues, err := c.ListQueues(0, 100) // can't check the caches value just yet.
			Expect(err, ToBeNil)
			found := false
			for _, queue := range queues {
				if queue.Name == qname {
					found = true
					break
				}
			}
			Expect(found, ToEqual, true)
		})

		It("releases a message", func() {
			c := mq.New(qname)

			id, err := c.PushString("trying")
			Expect(err, ToBeNil)

			msg, err := c.Get()
			Expect(err, ToBeNil)

			err = msg.Release(3)
			Expect(err, ToBeNil)

			msg, err = c.Get()
			Expect(msg, ToBeNil)

			time.Sleep(3 * time.Second)

			msg, err = c.Get()
			Expect(err, ToBeNil)
			Expect(msg.Id, ToEqual, id)
		})

		It("updates a queue", func() {
			c := mq.New("pushqueue")
			info, err := c.Info()
			qi := mq.QueueInfo{PushType: "multicast"}
			rc, err := c.Update(qi)
			Expect(err, ToBeNil)
			Expect(info.Id, ToEqual, rc.Id)
		})
		It("Adds and removes subscribers", func() {
			queue := mq.New("addSubscribersTest-" + strconv.Itoa(time.Now().Nanosecond()))
			defer queue.Delete()
			qi := mq.QueueInfo{PushType: "multicast"}
			qi, err := queue.Update(qi)
			Expect(qi.PushType, ToEqual, "multicast")
			Expect(err, ToBeNil)
			err = queue.AddSubscribers("http://server1")
			Expect(err, ToBeNil)
			info, err := queue.Info()
			Expect(err, ToBeNil)
			Expect(len(info.Subscribers), ToEqual, 1)
			err = queue.AddSubscribers("http://server2", "http://server3")
			Expect(err, ToBeNil)
			info, err = queue.Info()
			Expect(err, ToBeNil)
			Expect(len(info.Subscribers), ToEqual, 3)
			err = queue.RemoveSubscribers("http://server2")
			Expect(err, ToBeNil)
			info, err = queue.Info()
			Expect(err, ToBeNil)
			Expect(len(info.Subscribers), ToEqual, 2)
			err = queue.RemoveSubscribers("http://server1", "http://server3")
			Expect(err, ToBeNil)
			info, err = queue.Info()
			Expect(err, ToBeNil)
			Expect(len(info.Subscribers), ToEqual, 0)

		})
	})
}

func init() {

}
