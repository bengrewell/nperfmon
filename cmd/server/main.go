package main

import (
	"fmt"
	"github.com/BGrewell/go-iperf"
	"time"
)

func main() {

	// Run an iperf server
	s := iperf.NewServer()
	err := s.Start()
	if err != nil {
		panic(err)
	}

	for s.Running {
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Server finished")
}
