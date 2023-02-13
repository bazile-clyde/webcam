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
	_, err := webcam.Open("/dev/video0")
	handleError(err)

	_, err = webcam.Open("/dev/video11")
	handleError(err)

	fmt.Print("finished")
}
