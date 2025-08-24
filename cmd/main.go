package main

import (
	"log"

	"github.com/lukemassa/doorbell/internal/doorbell"
)

func main() {
	controller, err := doorbell.New("tcp://localhost:1883")
	if err != nil {
		log.Fatal(err)
	}
	err = controller.Run()
	if err != nil {
		log.Fatal(err)
	}
}
