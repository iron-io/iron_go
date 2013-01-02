package mq_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/iron-io/iron_go/mq"
	. "github.com/jeffh/go.bdd"
)

func TestEverything(t *testing.T) {}

func init() {
	defer PrintSpecReport()

	Describe("IronMQ", func() {
		It("Deletes all existing messages", func() {
			c := mq.New("queuename")
			Expect(c.Clear(), ToBeNil)

			info, err := c.Info()
			Expect(err, ToBeNil)
			Expect(info.Size, ToEqual, 0x0)
		})

		It("Pushes ands gets a message", func() {
			c := mq.New("queuename")
			id1, err := c.PushString("just a little test")
			Expect(err, ToBeNil)
			defer c.DeleteMessage(id1)

			msg, err := c.Get()
			Expect(err, ToBeNil)

			Expect(msg, ToNotBeNil)
			Expect(msg.Id, ToDeepEqual, id1)
			Expect(msg.Body, ToDeepEqual, "just a little test")
		})

		It("clears the queue", func() {
			q := mq.New("queuename")

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
			c := mq.New("queuename")
			queues, err := c.ListQueues(0, 100) // can't check the caches value just yet.
			Expect(err, ToBeNil)
			found := false
			for _, queue := range queues {
				if queue.Name == "queuename" {
					found = true
					break
				}
			}
			Expect(found, ToEqual, true)
		})

		It("releases a message", func() {
			c := mq.New("queuename")

			id, err := c.PushString("trying")
			Expect(err, ToBeNil)

			msg, err := c.Get()
			Expect(err, ToBeNil)

			err = msg.Release(3)
			Expect(err, ToBeNil)

			msg, err = c.Get()
			Expect(msg, ToEqual, nil)

			time.Sleep(3)

			msg, err = c.Get()
			Expect(err, ToBeNil)
			Expect(msg.Id, ToEqual, id)
		})
	})
}
