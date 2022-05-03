package aggregate

import (
	"time"

	"github.com/goforbroke1006/unknown-livecoding-1/domain"
)

func NewConnection(ip int32) *fakeConnection {
	return &fakeConnection{
		ip: ip,
	}
}

func NewFakeConnectionOpened(ip int32) *fakeConnection {
	return &fakeConnection{
		ip:     ip,
		opened: true,
	}
}

type fakeConnection struct {
	ip     int32
	opened bool
}

var _ domain.Connection = &fakeConnection{}

func (c *fakeConnection) Open() {
	//fmt.Println("opening connection", c.ip)

	const fakeConnectionEstablishingDuration = 5 * time.Second
	time.Sleep(fakeConnectionEstablishingDuration)
	c.opened = true
}

func (c *fakeConnection) Close() {
	const fakeConnectionClosingDuration = 1 * time.Second

	time.Sleep(fakeConnectionClosingDuration)
	c.opened = false
}

func (c fakeConnection) IsOpen() bool {
	return c.opened
}
