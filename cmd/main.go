package main

import (
	"fmt"
	"time"

	"github.com/goforbroke1006/unknown-livecoding-1/aggregate"
	"github.com/goforbroke1006/unknown-livecoding-1/domain"
	"github.com/goforbroke1006/unknown-livecoding-1/internal"
)

func main() {
	storage := internal.NewConnectionStorageOnChan(1024)
	go storage.Run()
	defer storage.Shutdown()

	start := time.Now()

	go func() {
		for {
			for ip := int32(200); ip >= 100; ip-- {
				<-time.After(10 * time.Millisecond)
				storage.OnNewRemoteConnection(ip, aggregate.NewConnection(ip))
			}
		}
	}()

	var opened []domain.Connection
	for ip := int32(100); ip <= 200; ip++ {
		conn := storage.GetConnection(ip)
		opened = append(opened, conn)
	}

	fmt.Printf("all connection (count %d) are opened in %.4f seconds!!!\n", len(opened), time.Since(start).Seconds())
	<-time.After(10 * time.Second)
	fmt.Println("bye...")
}
