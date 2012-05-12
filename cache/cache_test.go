package cache_test

import (
	"github.com/manveru/go.iron/cache"
	. "github.com/sdegutis/go.bdd"
	"testing"
)

func TestEverything(t *testing.T) {}

func init() {
	defer PrintSpecReport()

	Describe("IronCache", func() {
		token := "asdf"
		projectId := "asdf"

		It("Lists all caches", func() {
			caches, err := cache.Caches(token, projectId)
			Expect(err, ToBeNil)
			Expect(len(caches), ToEqual, 1)
			Expect(caches, ToDeepEqual, 42)
		})

		It("Puts a value into the cache", func() {})
	})
}
