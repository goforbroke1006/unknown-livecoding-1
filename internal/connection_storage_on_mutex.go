package internal

import (
	"context"
	"fmt"
	"sync"

	"github.com/goforbroke1006/unknown-livecoding-1/aggregate"
	"github.com/goforbroke1006/unknown-livecoding-1/domain"
	"github.com/goforbroke1006/unknown-livecoding-1/pkg"
)

func NewConnectionStorageOnMutex(initSize int) *connectionStorageOnMutex {
	return &connectionStorageOnMutex{
		cache:        make(map[int32]domain.Connection, initSize),
		remoteConnPS: pkg.NewPubSub(),
		stopInit:     make(chan struct{}),
		stopDone:     make(chan struct{}),
	}
}

var _ domain.ConnectionsStorage = &connectionStorageOnMutex{}

type connectionStorageOnMutex struct {
	cache   map[int32]domain.Connection
	cacheMx sync.RWMutex

	remoteConnPS pkg.PubSub

	stopInit chan struct{}
	stopDone chan struct{}
}

func (c *connectionStorageOnMutex) Run() {
	// do not implement me
}

func (c *connectionStorageOnMutex) GetConnection(ipAddress int32) (result domain.Connection) {

	topic := fmt.Sprintf("%d", ipAddress)
	notifyConnCh := make(chan interface{})
	c.remoteConnPS.Subscribe(topic, notifyConnCh)
	defer c.remoteConnPS.Unsubscribe(topic, notifyConnCh)

	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) { // try open new connection
		newConn := aggregate.NewConnection(ipAddress)
		newConn.Open()
		select {
		case <-ctx.Done():
			go func() { newConn.Close() }()
		default:
			c.remoteConnPS.Publish(topic, remoteConnChunk{
				remotePeer: ipAddress,
				conn:       newConn,
				needUpdate: true,
			})
		}
	}(ctx)

	go func(ctx context.Context) { // try to find existing
		c.cacheMx.RLock()
		existentConn, found := c.cache[ipAddress]
		c.cacheMx.RUnlock()

		if !found {
			return
		}

		select {
		case <-ctx.Done():
			// do nothing
		default:
			c.remoteConnPS.Publish(topic, remoteConnChunk{
				remotePeer: ipAddress,
				conn:       existentConn,
				needUpdate: false,
			})
		}
	}(ctx)

	// wait for connection
	select {
	case chunk := <-notifyConnCh:
		cancel()

		// check remote connection exists but was not notified because has no subscribers in right time
		// or another subscription had updated cache already
		c.cacheMx.RLock()
		remoteConn, remoteConnFound := c.cache[ipAddress]
		c.cacheMx.RUnlock()

		if remoteConnFound {
			result = remoteConn
		} else {
			result = chunk.(remoteConnChunk).conn

			if chunk.(remoteConnChunk).needUpdate {
				c.cacheMx.Lock()
				c.cache[ipAddress] = chunk.(remoteConnChunk).conn
				c.cacheMx.Unlock()
			}
		}

		break
	}

	return result
}

func (c *connectionStorageOnMutex) OnNewRemoteConnection(remotePeer int32, conn domain.Connection) {
	topic := fmt.Sprintf("%d", remotePeer)
	if published := c.remoteConnPS.Publish(topic, remoteConnChunk{
		remotePeer: remotePeer,
		conn:       conn,
		needUpdate: true,
	}); !published {
		c.cacheMx.Lock()
		old, found := c.cache[remotePeer]
		if found {
			go func() {
				old.Close()
			}()
		}
		c.cache[remotePeer] = conn
		c.cacheMx.Unlock()
	}
}

func (c *connectionStorageOnMutex) Shutdown() {
	c.closeAllConnections()
}

func (c *connectionStorageOnMutex) closeAllConnections() {
	c.cacheMx.Lock()
	defer c.cacheMx.Unlock()

	for ip, conn := range c.cache {
		go func(conn domain.Connection) { conn.Close() }(conn)
		delete(c.cache, ip)
	}
}

type remoteConnChunk struct {
	remotePeer int32
	conn       domain.Connection
	needUpdate bool
}
