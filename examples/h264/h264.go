package main

import (
	"fmt"
	"github.com/blackjack/webcam"
)

func handleError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	cam, err := webcam.Open("/dev/video0")
	handleError(err)
	defer cam.Close()

	encoder, err := webcam.Open_v2("/dev/video11")
	handleError(err)
	defer encoder.Close()

	fmt.Print("finished")
}
