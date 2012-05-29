package cache_test

import (
	"testing"
	"time"

	"github.com/manveru/go.iron/cache"
	. "github.com/sdegutis/go.bdd"
)

func TestEverything(t *testing.T) {}

func init() {
	defer PrintSpecReport()

	Describe("IronCache", func() {
		c := cache.New("cachename")

		It("Lists all caches", func() {
			_, err := c.ListCaches(0, 100) // can't check the caches value just yet.
			Expect(err, ToBeNil)
		})

		It("Puts a value into the cache", func() {
			err := c.Put("keyname", &cache.Item{
				Value:      "value",
				Expiration: 2 * time.Second,
			})
			Expect(err, ToBeNil)
		})

		It("Gets a value from the cache", func() {
			value, err := c.Get("keyname")
			Expect(err, ToBeNil)
			Expect(value, ToEqual, "value")
		})

		It("Sets numeric items", func() {
			err := c.Set("number", 42)
			Expect(err, ToBeNil)
			value, err := c.Get("number")
			Expect(err, ToBeNil)
			Expect(value.(float64), ToEqual, 42.0)
		})
	})
}
