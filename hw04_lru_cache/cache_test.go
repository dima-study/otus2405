package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(3)

		lru, ok := c.(*lruCache)
		require.True(t, ok, "Cache is lruCache")

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)
		validateQueue(
			t,
			lru.queue,
			[]interface{}{lruItemValue{"aaa", 100}},
			"aaa=100",
		)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)
		validateQueue(
			t,
			lru.queue,
			[]interface{}{
				lruItemValue{"bbb", 200},
				lruItemValue{"aaa", 100},
			},
			"bbb=200, aaa=100",
		)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)
		validateQueue(
			t,
			lru.queue,
			[]interface{}{
				lruItemValue{"aaa", 100},
				lruItemValue{"bbb", 200},
			},
			"aaa=100, bbb=200",
		)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)
		validateQueue(
			t,
			lru.queue,
			[]interface{}{
				lruItemValue{"bbb", 200},
				lruItemValue{"aaa", 100},
			},
			"bbb=200, aaa=100",
		)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)
		validateQueue(
			t,
			lru.queue,
			[]interface{}{
				lruItemValue{"aaa", 300},
				lruItemValue{"bbb", 200},
			},
			"aaa=300, bbb=200",
		)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)
		validateQueue(
			t,
			lru.queue,
			[]interface{}{
				lruItemValue{"aaa", 300},
				lruItemValue{"bbb", 200},
			},
			"aaa=300, bbb=200",
		)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)

		wasInCache = c.Set("ccc", 400)
		require.False(t, wasInCache)
		validateQueue(
			t,
			lru.queue,
			[]interface{}{
				lruItemValue{"ccc", 400},
				lruItemValue{"aaa", 300},
				lruItemValue{"bbb", 200},
			},
			"ccc=400, aaa=300, bbb=200",
		)

		wasInCache = c.Set("ddd", 500)
		require.False(t, wasInCache)
		validateQueue(
			t,
			lru.queue,
			[]interface{}{
				lruItemValue{"ddd", 500},
				lruItemValue{"ccc", 400},
				lruItemValue{"aaa", 300},
			},
			"ddd=500, ccc=400, aaa=300",
		)

		_, ok = lru.items["bbb"]
		require.False(t, ok, "bbb key removed")
	})

	t.Run("purge logic", func(t *testing.T) {
		c := NewCache(3)

		lru, ok := c.(*lruCache)
		require.True(t, ok, "Cache is lruCache")

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)
		validateQueue(
			t,
			lru.queue,
			[]interface{}{lruItemValue{"aaa", 100}},
			"aaa=100",
		)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)
		validateQueue(
			t,
			lru.queue,
			[]interface{}{
				lruItemValue{"bbb", 200},
				lruItemValue{"aaa", 100},
			},
			"bbb=200, aaa=100",
		)

		c.Clear()

		lru, ok = c.(*lruCache)
		require.True(t, ok, "Cache is lruCache")
		validateQueue(
			t,
			lru.queue,
			[]interface{}{},
			"empty",
		)
		require.Equal(t, 0, lru.queue.Len(), "queue len=0")

		keys := len(lru.items)
		require.Equal(t, 0, keys, "items len=0")
	})
}

func TestCacheMultithreading(t *testing.T) {
	t.Skip() // Remove me if task with asterisk completed.

	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}

func validateQueue(t *testing.T, list List, expected interface{}, msg string) {
	got := make([]interface{}, 0, list.Len())
	for front := list.Front(); front != nil; front = front.Next {
		got = append(got, front.Value)
	}

	require.Equal(t, expected, got, msg)
}
