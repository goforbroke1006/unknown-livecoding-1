package internal

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/goforbroke1006/unknown-livecoding-1/aggregate"
	"github.com/goforbroke1006/unknown-livecoding-1/domain"
)

type InitStorageFn func() (storage domain.ConnectionsStorage, cancelFn func())

func NewTestSuite(t *testing.T, createFn InitStorageFn) {

	t.Run("connection can be established in 6 seconds", func(t *testing.T) {
		t.Parallel()

		cs, stop := createFn()
		defer stop()

		done := make(chan struct{})
		go func() {
			_ = cs.GetConnection(123)
			done <- struct{}{}
		}()

		select {
		case <-time.After(6 * time.Second):
			t.Error(errors.New("can't open connection in 6 seconds"))
		case <-done:
			// ok
		}
	})

	t.Run("same connection should be open once", func(t *testing.T) {
		t.Parallel()

		const ip = 123

		cs, stop := createFn()
		defer stop()

		var (
			conn1 domain.Connection = nil
			conn2 domain.Connection = nil
			done                    = make(chan struct{})
		)

		go func() {
			conn1 = cs.GetConnection(123)
			done <- struct{}{}
		}()
		select {
		case <-time.After(6 * time.Second):
			t.Error(errors.New("can't open connection in 6 seconds"))
		case <-done:
			// ok
		}

		go func() {
			//start <- struct{}{}
			conn2 = cs.GetConnection(ip)
			done <- struct{}{}
		}()
		//<-start
		select {
		case <-time.After(10 * time.Millisecond):
			t.Fatal(errors.New("can't open same connection immediately"))
		case <-done:
			// ok
		}

		if conn1 == nil {
			t.Fatal("should not be NIL")
		}

		if conn1 != conn2 {
			t.Fatalf("storage return different connections for same ip: %p and %p", conn1, conn2)
		}
	})

	t.Run("OnNewRemoteConnection + GetConnection(immediately)", func(t *testing.T) {
		t.Parallel()

		const ip = 234

		var (
			remoteConn   domain.Connection = nil
			fromStorConn domain.Connection = nil
			done                           = make(chan struct{})
		)

		cs, stop := createFn()
		defer stop()

		remoteConn = aggregate.NewConnection(ip)
		cs.OnNewRemoteConnection(ip, remoteConn)

		go func() {
			fromStorConn = cs.GetConnection(ip)
			done <- struct{}{}
		}()
		select {
		case <-time.After(1 * time.Millisecond):
			t.Error(errors.New("can't get connection immediately"))
		case <-done:
			// ok
		}

		if remoteConn != fromStorConn {
			t.Errorf("storage return different connections for same ip: %p and %p", remoteConn, fromStorConn)
		}
	})

	t.Run("GetConnection + OnNewRemoteConnection (immediately return from GetConnection)", func(t *testing.T) {
		t.Parallel()

		const ip = 234

		var (
			remoteConn   domain.Connection = nil
			fromStorConn domain.Connection = nil
			start                          = make(chan struct{})
			done                           = make(chan struct{})
		)

		cs, stop := createFn()
		defer stop()

		remoteConn = aggregate.NewConnection(ip)

		go func() {
			start <- struct{}{}
			fromStorConn = cs.GetConnection(ip)
			done <- struct{}{}
		}()

		<-start
		cs.OnNewRemoteConnection(ip, remoteConn)

		select {
		case <-time.After(1 * time.Millisecond):
			t.Fatal(errors.New("can't get connection immediately"))
		case <-done:
			// ok
		}

		if fromStorConn == nil {
			t.Fatal(errors.New("should not be NIL"))
		}

		if remoteConn != fromStorConn {
			t.Errorf("storage return different connections for same ip: %p and %p", remoteConn, fromStorConn)
		}
	})

	t.Run("OnNewRemoteConnection + GetConnection + OnNewRemoteConnection+Connection.Close() + GetConnection - old closed, new opened", func(t *testing.T) {
		t.Parallel()

		const ip = 345

		var (
			remoteConn1   domain.Connection = nil
			fromStorConn1 domain.Connection = nil

			remoteConn2   domain.Connection = nil
			fromStorConn2 domain.Connection = nil

			done = make(chan struct{})
		)

		cs, stop := createFn()
		defer stop()

		remoteConn1 = aggregate.NewFakeConnectionOpened(ip)
		cs.OnNewRemoteConnection(ip, remoteConn1)
		go func() {
			fromStorConn1 = cs.GetConnection(ip)
			done <- struct{}{}
		}()
		select {
		case <-time.After(1 * time.Millisecond):
			t.Error(errors.New("can't get connection immediately"))
		case <-done:
			// ok
		}
		if remoteConn1 != fromStorConn1 {
			t.Errorf("storage return different connections for same ip: %p and %p", remoteConn1, fromStorConn1)
		}

		remoteConn2 = aggregate.NewFakeConnectionOpened(ip)
		cs.OnNewRemoteConnection(ip, remoteConn2)
		go func() {
			fromStorConn2 = cs.GetConnection(ip)
			done <- struct{}{}
		}()
		select {
		case <-time.After(1 * time.Millisecond):
			t.Error(errors.New("can't get connection immediately"))
		case <-done:
			// ok
		}
		if remoteConn2 != fromStorConn2 {
			t.Errorf("storage return different connections for same ip: %p and %p", remoteConn2, fromStorConn2)
		}

		if fromStorConn1 == fromStorConn2 {
			t.Fatal("second connection should be new, not same as first connection")
		}

		<-time.After(1500 * time.Millisecond) // wait fake connection is closed
		assert.Equal(t, false, remoteConn1.IsOpen())
	})
}

func NewBenchmarkGetConnection(b *testing.B, createFn InitStorageFn) {
	const ip = 123

	wrongConnCounter := uint64(0)

	for repeatIndex := 0; repeatIndex < b.N; repeatIndex++ {
		func() {
			var (
				remoteConn      domain.Connection = nil
				fromStorageConn domain.Connection = nil
				start                             = make(chan struct{})
				done                              = make(chan struct{})
			)

			cs, stop := createFn()
			defer stop()

			go func() {
				start <- struct{}{}
				fromStorageConn = cs.GetConnection(ip)
				done <- struct{}{}
			}()

			<-start
			remoteConn = aggregate.NewConnection(ip)
			cs.OnNewRemoteConnection(ip, remoteConn)

			<-done
			if remoteConn != fromStorageConn {
				wrongConnCounter++
			}
		}()
	}
	if wrongConnCounter > 0 {
		b.Errorf("for N = %d wrong connection return count %d", b.N, wrongConnCounter)
	}
}
