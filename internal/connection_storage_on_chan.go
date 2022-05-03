package internal

import (
	"context"
	"fmt"

	"github.com/goforbroke1006/unknown-livecoding-1/aggregate"
	"github.com/goforbroke1006/unknown-livecoding-1/domain"
	"github.com/goforbroke1006/unknown-livecoding-1/pkg"
)

type operationKind string

const (
	operationKindRead  = operationKind("read")
	operationKindWrite = operationKind("write")
)

type operation struct {
	kind operationKind
	addr int32
	conn domain.Connection
}

// NewConnectionStorage create storage with initial size of cache
func NewConnectionStorageOnChan(initSize int) *connectionStorageOnChan {
	return &connectionStorageOnChan{
		cache: make(map[int32]domain.Connection, initSize),

		readConnPS:       pkg.NewPubSub(),
		operations:       make(chan operation, initSize),
		operationsAnswer: make(chan struct{}),

		stopInit: make(chan struct{}),
		stopDone: make(chan struct{}),
	}
}

type connectionStorageOnChan struct {
	cache map[int32]domain.Connection

	readConnPS       pkg.PubSub
	operations       chan operation
	operationsAnswer chan struct{}

	stopInit chan struct{}
	stopDone chan struct{}
}

var _ domain.ConnectionsStorage = &connectionStorageOnChan{}

// GetConnection try to get connection in 3 parallel ways
// 1. Expect for new remote connection
// 2. Expect reading existing connection
// 3. Await new opened connection (step skips if remote connection from step 1 appears)
func (c connectionStorageOnChan) GetConnection(ipAddress int32) (result domain.Connection) {

	topic := fmt.Sprintf("%d", ipAddress)
	const getConnOptions = 3 // get existing, open new, catch from remote
	notifyConnCh := make(chan interface{}, getConnOptions)
	c.readConnPS.Subscribe(topic, notifyConnCh)
	defer c.readConnPS.Unsubscribe(topic, notifyConnCh)

	ctx, cancel := context.WithCancel(context.Background())

	go c.reading(ipAddress)

	go func(ctx context.Context) { // try open new connection
		newConn := aggregate.NewConnection(ipAddress)
		newConn.Open()
		select {
		case <-ctx.Done():
			go func() { newConn.Close() }()
		default:
			c.writing(ipAddress, newConn)
		}
	}(ctx)

	// wait for connection
	select {
	case chunk := <-notifyConnCh:
		cancel()
		result = chunk.(connState).conn
		break
	}

	return result
}

// OnNewRemoteConnection store new connection from remote peer
func (c connectionStorageOnChan) OnNewRemoteConnection(remotePeer int32, conn domain.Connection) {
	c.writing(remotePeer, conn)
}

// Run process next type of operations
// * stopping storage pipelines and clear it
// * reading from cache
// * writing to cache
func (c *connectionStorageOnChan) Run() {
RunLoop:
	for {
		select {
		case <-c.stopInit:
			break RunLoop
		case chunk := <-c.operations:
			c.processOperation(chunk)
		}
	}

	c.closeAllConnections()

	c.stopDone <- struct{}{}
}

func (c *connectionStorageOnChan) processOperation(chunk operation) {
	topic := fmt.Sprintf("%d", chunk.addr)
	var connection domain.Connection

	switch chunk.kind {
	case operationKindWrite:
		if old, found := c.cache[chunk.addr]; found {
			go func() { old.Close() }()
		}
		c.cache[chunk.addr] = chunk.conn
		connection = chunk.conn

	case operationKindRead:
		if conn, found := c.cache[chunk.addr]; found {
			connection = conn
		}
	}

	if connection != nil {
		state := connState{
			ip:   chunk.addr,
			conn: connection,
		}
		c.readConnPS.Publish(topic, state)
	}

	c.operationsAnswer <- struct{}{}
}

// Shutdown break loop inside connectionStorage.Run() method and return to the 'owner' goroutine
func (c connectionStorageOnChan) Shutdown() {
	c.stopInit <- struct{}{}
	<-c.stopDone
}

// reading func send request for reading
// You have to subscribe on connectionStorage.readConnPS(fmt.Sprint("%d", ip), yourChannel) to read data
func (c connectionStorageOnChan) reading(ip int32) {
	c.operations <- operation{kind: operationKindRead, addr: ip}
	<-c.operationsAnswer
}

// writing func send request for writing connection to cache
func (c connectionStorageOnChan) writing(ip int32, conn domain.Connection) {
	c.operations <- operation{kind: operationKindWrite, addr: ip, conn: conn}
	<-c.operationsAnswer
}

// closeAllConnections delete all keeping connection.
// Should be run after connectionStorage.Run loop break to prevent race on connectionStorage.cache
func (c connectionStorageOnChan) closeAllConnections() {
	for ip, conn := range c.cache {
		go func() { conn.Close() }()
		delete(c.cache, ip)
	}
}

// connState transport data about connection's changes
type connState struct {
	ip   int32
	conn domain.Connection
}
