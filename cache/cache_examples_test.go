package cache_test

import (
	"github.com/manveru/go.iron/cache"
)

func SetExample() {
	c := cache.New("test_cache")
	c.Set("number_item", 42)
	// Output: hi
}

/*
require 'iron_cache'

# For configuration info, see http://dev.iron.io/articles/configuration
cache = IronCache::Client.new()

# Set the default cache name
cache.cache_name = "test_cache"

# Numbers will get stored as numbers
cache.items.put("number_item", 42)

# Strings get stored as strings
cache.items.put("string_item", "Hello, IronCache")

# Objects and dicts get JSON-encoded and stored as strings
complex_item = {"test" => "this is a dict", "args" => ["apples", "oranges"] }
cache.items.put("complex_item", complex_item)
*/

/*
require 'iron_cache'

# For configuration info, see http://dev.iron.io/articles/configuration
cache = IronCache::Client.new()

# Set the default cache name
cache.cache_name = "test_cache"

# Numbers can be incremented
cache.items.increment("number_item", -10)

# Everything else throws a 400 error
cache.items.increment("number_item", "a lot")
cache.items.increment("string_item", 10)
cache.items.increment("complex_item", 10)
*/

/*
require 'iron_cache'

# For configuration info, see http://dev.iron.io/articles/configuration
cache = IronCache::Client.new()

# Set the default cache name
cache.cache_name = "test_cache"

# Numbers can be decremented
cache.items.increment("number_item", -10)

# Everything else throws a 400 error
cache.items.increment("number_item", "negative a lot")
cache.items.increment("string_item", -10)
cache.items.increment("complex_item", -10)
*/

/*
require 'iron_cache'

# For configuration info, see http://dev.iron.io/articles/configuration
cache = IronCache::Client.new()

# Set the default cache name
cache.cache_name = "test_cache"

# Numbers will get stored as numbers
p cache.items.get("number_item").value

# Strings get stored as strings
p cache.items.get("string_item").value

# Objects and dicts get JSON-encoded and stored as strings
p cache.items.get("complex_item").value
*/

/*
require 'iron_cache'

# For configuration info, see http://dev.iron.io/articles/configuration
cache = IronCache::Client.new()

# Set the default cache name
cache.cache_name = "test_cache"

# Immediately delete an item
cache.items.delete("string_item")
*/
