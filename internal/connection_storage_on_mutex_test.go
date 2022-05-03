package internal

import (
	"testing"

	"github.com/goforbroke1006/unknown-livecoding-1/domain"
)

func TestConnectionStorageOnMutex_GetConnection(t *testing.T) {
	const size = 1024

	NewTestSuite(t, func() (storage domain.ConnectionsStorage, cancelFn func()) {
		storage = NewConnectionStorageOnMutex(size)

		return storage, func() {}
	})
}

// BenchmarkConnectionStorageOnMutex_GetConnection checks efficiency open connection callback running
//
// go test -gcflags=-N -test.bench '^\QBenchmarkConnectionStorageOnMutex_GetConnection\E$' -run ^$ -benchmem -test.benchtime 10000x ./...
//
//goos: linux
//goarch: amd64
//pkg: github.com/goforbroke1006/unknown-livecoding/internal
//cpu: Intel(R) Core(TM) i5-6300HQ CPU @ 2.30GHz
//BenchmarkConnectionStorageOnMutex_GetConnection-4          10000             24601 ns/op           51135 B/op         28 allocs/op
//
//goos: linux
//goarch: amd64
//pkg: github.com/goforbroke1006/unknown-livecoding/internal
//cpu: Intel(R) Core(TM) i5-6300HQ CPU @ 2.30GHz
//BenchmarkConnectionStorageOnMutex_GetConnection-4          10000             25093 ns/op           51133 B/op         28 allocs/op
//
func BenchmarkConnectionStorageOnMutex_GetConnection(b *testing.B) {
	const size = 1024

	NewBenchmarkGetConnection(b, func() (storage domain.ConnectionsStorage, cancelFn func()) {
		storage = NewConnectionStorageOnMutex(size)

		return storage, func() {}
	})
}
