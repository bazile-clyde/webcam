package main

import "C"
import (
	"errors"
	"fmt"
	"github.com/blackjack/webcam"
	"os"
)

const V4L2_PIX_FMT_YUYV = 0x56595559

func panicOnError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// Programming a V4L2 device consists of these steps:
//   - Opening the device
//   - Changing device properties, selecting a video and audio input, video standard, picture brightness a. o.
//   - Negotiating a data format
//   - Negotiating an input/output method
//   - The actual input/output loop
//   - Closing the device
//
// In practice most steps are optional and can be executed out of order.
func main() {
	// frames are sent to output devices like graphics cards and read from capture devices like cameras.
	source, err := webcam.Open("/dev/video0")
	panicOnError(err)
	defer source.Close()

	err = source.SetBufferCount(1)
	panicOnError(err)

	var format webcam.PixelFormat = V4L2_PIX_FMT_YUYV
	formatDesc := source.GetSupportedFormats()
	if _, ok := formatDesc[format]; !ok {
		panicOnError(errors.New(fmt.Sprintf("cannot support pixel format %v", format)))
	}

	fmt.Println("Available formats:")
	for _, s := range formatDesc {
		fmt.Fprintln(os.Stderr, s)
	}

	frames := source.GetSupportedFrameSizes(format)
	size := frames[len(frames)-1]

	f, w, h, err := source.SetImageFormat(format, size.MaxWidth, size.MaxHeight)
	panicOnError(err)

	_, err = fmt.Fprintf(os.Stderr, "Resulting image format: %s %dx%d\n", formatDesc[f], w, h)
	panicOnError(err)

	err = source.StartStreaming()
	panicOnError(err)

	timeout := uint32(5) // 5 seconds
	err = source.WaitForFrame(timeout)
	panicOnError(err)

	_, err = source.ReadFrame()
	panicOnError(err)

	// "output” refers to the raw frames being encoded
	output, err := webcam.Open_v2("/dev/video11")
	panicOnError(err)
	defer output.Close()

	fmt.Println("finished")
}
