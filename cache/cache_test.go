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

		It("Gets meta-information about an item", func() {
			err := c.Put("forever", &cache.Item{Value: "and ever", Expiration: 0})
			Expect(err, ToBeNil)
			value, err := c.GetMeta("forever")
			Expect(err, ToBeNil)
			Expect(value, ToDeepEqual,
				map[string]interface{}{
					"key":     "forever",
					"cas":     5.766174005002437e+18,
					"cache":   "cachename",
					"value":   "and ever",
					"expires": "9999-01-01T01:00:00+01:00",
					"flags":   0,
				},
			)
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
