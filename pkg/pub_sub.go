package pkg

import "sync"

type PubSub interface {
	Subscribe(topic string, ch chan interface{})
	Publish(topic string, msg interface{}) bool
	Unsubscribe(topic string, ch chan interface{})
}

func NewPubSub() *pubSubPrimitive {
	return &pubSubPrimitive{
		subs: make(map[string][]chan interface{}),
	}
}

type pubSubPrimitive struct {
	subs   map[string][]chan interface{}
	subsMx sync.RWMutex
}

var _ PubSub = &pubSubPrimitive{}

func (ps *pubSubPrimitive) Subscribe(topic string, ch chan interface{}) {
	ps.subsMx.Lock()
	defer ps.subsMx.Unlock()

	// spend O(N) to remove subscription duplicates
	for existingChIndex, existingCh := range ps.subs[topic] {
		if ch == existingCh {
			ps.subs[topic] = append(ps.subs[topic][0:existingChIndex], ps.subs[topic][existingChIndex+1:]...)
		}
	}

	ps.subs[topic] = append(ps.subs[topic], ch)
}

func (ps *pubSubPrimitive) Publish(topic string, msg interface{}) bool {
	ps.subsMx.RLock()
	defer ps.subsMx.RUnlock()

	for _, ch := range ps.subs[topic] {
		ch <- msg
	}

	return len(ps.subs[topic]) > 0
}

func (ps *pubSubPrimitive) Unsubscribe(topic string, ch chan interface{}) {
	ps.subsMx.Lock()
	defer ps.subsMx.Unlock()

	for existingChIndex, existingCh := range ps.subs[topic] {
		if ch == existingCh {
			ps.subs[topic] = append(ps.subs[topic][0:existingChIndex], ps.subs[topic][existingChIndex+1:]...)
		}
	}
}
