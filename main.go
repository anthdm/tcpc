package main

import (
	"fmt"
	"log"
	"time"

	"github.com/anthdm/tcpc/tcpc"
)

func main() {
	channelLocal, err := tcpc.New[string](":3000", ":4000")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		time.Sleep(5 * time.Second)
		channelLocal.Sendchan <- "GG"
	}()

	channelRemote, err := tcpc.New[string](":4000", ":3000")
	if err != nil {
		log.Fatal(err)
	}

	msg := <-channelRemote.Recvchan

	fmt.Println("received from channel over the wire: ", msg)
}
