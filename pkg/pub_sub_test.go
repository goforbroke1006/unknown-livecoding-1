package pkg

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPubSubPrimitive_Subscribe(t *testing.T) {
	const topic = "hello"

	t.Run("basic usage", func(t *testing.T) {
		ps := pubSubPrimitive{
			subs: make(map[string][]chan interface{}),
		}
		ch := make(chan interface{}, 2)

		ps.Subscribe(topic, ch)

		ps.Publish(topic, "hello")
		close(ch)

		var results []interface{}
		for msg := range ch {
			results = append(results, msg)
		}
		assert.Equal(t, 1, len(results))
	})

	t.Run("prevent double subscribing on same topic", func(t *testing.T) {
		ps := pubSubPrimitive{
			subs: make(map[string][]chan interface{}),
		}
		ch := make(chan interface{}, 2)

		ps.Subscribe(topic, ch)
		ps.Subscribe(topic, ch)
		ps.Subscribe(topic, ch)
		ps.Subscribe(topic, ch)

		ps.Publish(topic, "hello")
		close(ch)

		var results []interface{}
		for msg := range ch {
			results = append(results, msg)
		}
		assert.Equal(t, 1, len(results))
	})

}

func TestPubSubPrimitive_SubscribeUnsubscribe(t *testing.T) {
	const topic = "hello"

	t.Run("correct count of subscriber after pubSubPrimitive.Unsubscribe called", func(t *testing.T) {
		ps := pubSubPrimitive{
			subs: make(map[string][]chan interface{}),
		}
		assert.Equal(t, 0, len(ps.subs[topic]))

		notifications := make(chan interface{})
		ps.Subscribe(topic, notifications)
		assert.Equal(t, 1, len(ps.subs[topic]))

		ps.Unsubscribe(topic, notifications)
		assert.Equal(t, 0, len(ps.subs[topic]))
	})

	t.Run("after unsubscribe can't receive messages", func(t *testing.T) {
		ps := pubSubPrimitive{
			subs: make(map[string][]chan interface{}),
		}
		ch := make(chan interface{}, 2)
		ps.Subscribe(topic, ch)
		ps.Unsubscribe(topic, ch)
		ps.Publish(topic, "hello")
		close(ch)
		var results []interface{}
		for msg := range ch {
			results = append(results, msg)
		}
		assert.Equal(t, 0, len(results))
	})
}

// BenchmarkPubSubPrimitive_Subscribe check new subscription duration for loaded pub-sub
//
// go test -gcflags=-N -test.bench '^\QBenchmarkPubSubPrimitive_Subscribe\E$' -run ^$ -benchmem -test.benchtime 10000x ./...
//
// 		goos: linux
//		goarch: amd64
//		pkg: github.com/goforbroke1006/unknown-livecoding/pkg
//		cpu: Intel(R) Core(TM) i7-10850H CPU @ 2.70GHz
//		BenchmarkPubSubPrimitive_Subscribe-12              10000               138.7 ns/op            22 B/op          0 allocs/op
//		PASS
//		ok      github.com/goforbroke1006/unknown-livecoding/pkg      0.005s
func BenchmarkPubSubPrimitive_Subscribe(b *testing.B) {
	const topic = "1234"
	ps := pubSubPrimitive{
		subs: make(map[string][]chan interface{}),
	}
	for topicIndex := 1000; topicIndex < 2000; topicIndex++ {
		ps.Subscribe(fmt.Sprintf("%d", topicIndex), nil)
	}
	for i := 0; i < b.N; i++ {
		ps.Subscribe(topic, nil)
	}
}
