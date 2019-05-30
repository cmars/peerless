package main

import (
	"log"

	"github.com/cmars/peerless/client"
)

func main() {
	cl := client.New("http://localhost:8080")
	// Ensure we have a token
	var ok bool
	var err error
	for !ok {
		ok, err = cl.Do()
		if err != nil {
			panic(err)
		}
	}
	// Make 100 copies and set them loose
	for i := 0; i < 100; i++ {
		go func() {
			newCl := cl.Clone()
			for {
				_, err := newCl.Do()
				if err != nil {
					panic(err)
				}
			}
		}()
	}
	log.Println("clients all started")
	<-(chan struct{})(nil)
}
