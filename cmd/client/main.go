package main

import (
	"log"

	"github.com/cmars/peerless/client"
)

func main() {
	cl := client.New("http://localhost:8080")
	for {
		ok, err := cl.Do()
		if err != nil {
			log.Printf("%+v", err)
		} else {
			if ok {
				log.Println("ok")
			} else {
				log.Println("auth")
			}
		}
	}
}
