package cache_test

import (
	"github.com/manveru/go.iron/cache"
	. "github.com/sdegutis/go.bdd"
	"os"
	"testing"
)

func TestEverything(t *testing.T) {}

func init() {
	defer PrintSpecReport()

	Describe("IronCache", func() {
		domain := "cache-aws-us-east-1"
		token := os.Getenv("IRON_TOKEN")
		projectId := os.Getenv("IRON_PROJECT")

		It("Lists all caches", func() {
			caches, err := cache.Caches(domain, token, projectId)
			Expect(err, ToBeNil)
			Expect(len(caches), ToEqual, 0)
			Expect(caches, ToDeepEqual, []cache.Cache{})
		})

		It("Puts a value into the cache", func() {
			/*
				c := cache.Cache{
					Token: token, ProjectId: projectId, Name: "cachename", Domain: domain,
				}
				err := c.Set("keyname", cache.Item{Body: "value", ExpiresIn: 1, Replace: true})
				Expect(err, ToBeNil)
			*/
		})

		It("Gets a value from the cache", func() {
			c := cache.Cache{
				Token: token, ProjectId: projectId, Name: "cachename", Domain: domain,
			}
			value, err := c.Get("keyname")
			Expect(err, ToBeNil)
			Expect(value, ToBeNil)
		})
	})
}
