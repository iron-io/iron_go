package mq_test

import (
	"testing"

	"github.com/manveru/go.iron/mq"
	. "github.com/sdegutis/go.bdd"
)

func TestEverything(t *testing.T) {}

func init() {
	defer PrintSpecReport()

	Describe("IronMQ", func() {
		It("Lists all queues", func() {
			c := mq.New("queuename")
			_, err := c.ListQueues(0, 100) // can't check the caches value just yet.
			Expect(err, ToBeNil)
		})

		It("Deletes all existing messages", func() {
			c := mq.New("queuename")

			for {
				info, err := c.Info()
				Expect(err, ToBeNil)
				if info.Size == 0 {
					break
				}

				msgs, err := c.GetN(info.Size)
				Expect(err, ToBeNil)

				for _, msg := range msgs {
					err = c.DeleteMsg(msg.Id)
					Expect(err, ToBeNil)
				}
			}
		})

		It("Pushes ands gets a message", func() {
			c := mq.New("queuename")
			id1, err := c.PushString("just a little test")
			Expect(err, ToBeNil)
			defer c.DeleteMsg(id1)

			msg, err := c.Get()
			Expect(err, ToBeNil)

			Expect(msg, ToNotBeNil)
			Expect(msg.Id, ToDeepEqual, id1)
			Expect(msg.Body, ToDeepEqual, "just a little test")
		})
	})
}
