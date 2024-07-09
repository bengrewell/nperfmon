package main

import (
	"fmt"
	wrapper "github.com/bgrewell/nperfmon/pkg"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// Create a channel to receive OS signals
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	// Notify the channel on interrupt or termination signals
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine to handle the signal
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	w, err := wrapper.NewWrapper("127.0.0.1", 5, 30)
	if err != nil {
		panic(err)
	}
	err = w.Start()
	if err != nil {
		panic(err)
	}

	fmt.Println("Press Ctrl+C to exit")
	<-done
	err = w.Stop()
	if err != nil {
		panic(err)
	}
	fmt.Println("Exiting")

}
