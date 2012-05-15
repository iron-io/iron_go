package cache_test

import (
	"fmt"
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
			Expect(string(value), ToEqual, "value")
		})
	})
}

func ExampleStoringData() {
	// For configuration info, see http://dev.iron.io/articles/configuration
	// test_cache is the default cache name
	c := cache.New("test_cache")

	// All values are stored as []byte.
	c.Set(&cache.Item{Key: "item 1", Value: []byte("IronCache")})

	// They are retrieved as []byte as well.
	value, err := c.Get("item 1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", value)

	// We can store things using JSON
	cache.JSON.Set(c, &cache.Item{Key: "item 2", Object: map[string]string{"Hello": "IronCache"}})

	// And get them as JSON again.
	obj := map[string]string{}
	cache.JSON.Get(c, "item 2", &obj)
	fmt.Printf("%#v\n", obj)

	// Output:
	// []byte{0x49, 0x72, 0x6f, 0x6e, 0x43, 0x61, 0x63, 0x68, 0x65}
	// map[string]string{"Hello":"IronCache"}
}
