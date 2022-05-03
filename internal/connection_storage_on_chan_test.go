package internal

import (
	"testing"

	"github.com/goforbroke1006/unknown-livecoding-1/aggregate"
	"github.com/goforbroke1006/unknown-livecoding-1/domain"
)

func TestConnectionStorageOnChan_GetConnection(t *testing.T) {
	const size = 1024

	NewTestSuite(t, func() (storage domain.ConnectionsStorage, cancelFn func()) {
		storage = NewConnectionStorageOnChan(size)
		go storage.Run()
		return storage, func() {
			storage.Shutdown()
		}
	})
}

// BenchmarkConnectionStorageOnChan_GetConnection checks efficiency open connection callback running
//
// go test -gcflags=-N -test.bench '^\QBenchmarkConnectionStorageOnChan_GetConnection\E$' -run ^$ -benchmem -test.benchtime 10000x ./...
//
//goos: linux
//goarch: amd64
//pkg: github.com/goforbroke1006/unknown-livecoding/internal
//cpu: Intel(R) Core(TM) i5-6300HQ CPU @ 2.30GHz
//BenchmarkConnectionStorageOnChan_GetConnection-4           10000             55317 ns/op           93066 B/op         39 allocs/op
//
//goos: linux
//goarch: amd64
//pkg: github.com/goforbroke1006/unknown-livecoding/internal
//cpu: Intel(R) Core(TM) i5-6300HQ CPU @ 2.30GHz
//BenchmarkConnectionStorageOnChan_GetConnection-4           10000             56766 ns/op           93071 B/op         39 allocs/op
//
func BenchmarkConnectionStorageOnChan_GetConnection(b *testing.B) {
	const size = 1024

	NewBenchmarkGetConnection(b, func() (storage domain.ConnectionsStorage, cancelFn func()) {
		storage = NewConnectionStorageOnChan(size)
		go storage.Run()
		return storage, func() {
			storage.Shutdown()
		}
	})
}

// go test -gcflags=-N -test.bench '^\QBenchmarkConnectionStorage_Shutdown\E$' -run ^$ -benchmem -test.benchtime 10000x
func BenchmarkConnectionStorageOnChan_Shutdown(b *testing.B) {
	const size = 5
	const startPeer = int32(101)

	for repeatIndex := 0; repeatIndex < b.N; repeatIndex++ {
		cs := NewConnectionStorageOnChan(size)
		go cs.Run()

		b.StopTimer()
		for peer := startPeer; peer < startPeer+size; peer++ {
			cs.OnNewRemoteConnection(peer, aggregate.NewConnection(peer))
		}
		b.StartTimer()

		cs.Shutdown()
	}
}
