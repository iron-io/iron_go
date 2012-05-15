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
		It("Lists all caches", func() {
			c := cache.New("cachename")
			_, err := c.ListCaches(0, 100) // can't check the caches value just yet.
			Expect(err, ToBeNil)
		})

		It("Puts a value into the cache", func() {
			c := cache.New("cachename")
			err := c.Set(&cache.Item{
				Key:        "keyname",
				Value:      []byte("value"),
				Expiration: 2 * time.Second,
			})
			Expect(err, ToBeNil)
		})

		It("Gets a value from the cache", func() {
			c := cache.New("cachename")
			value, err := c.Get("keyname")
			Expect(err, ToBeNil)
			Expect(value, ToEqual, "value")
		})
	})
}
