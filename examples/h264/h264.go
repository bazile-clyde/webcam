package main

import "C"
import (
	"errors"
	"fmt"
	"github.com/blackjack/webcam"
	"os"
)

const V4L2_PIX_FMT_YUYV = 0x56595559
const V4L2_PIX_FMT_MJPG = 0x47504A4D

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
// Frames are sent to output devices like graphics cards and read from capture devices like cameras.
func main() {
	source, err := webcam.Open("/dev/video0")
	panicOnError(err)
	defer source.Close()

	err = source.SetBufferCount(1)
	panicOnError(err)

	var format webcam.PixelFormat = V4L2_PIX_FMT_MJPG
	formatDesc := source.GetSupportedFormats()
	if _, ok := formatDesc[format]; !ok {
		panicOnError(errors.New(fmt.Sprintf("cannot support pixel format %v", format)))
	}

	fmt.Println("Available formats:")
	for k, s := range formatDesc {
		fmt.Fprintln(os.Stderr, s, k)
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

	// frame, err := source.ReadFrame()
	_, err = source.ReadFrame()
	panicOnError(err)

	// img, err := jpeg.Decode(bytes.NewReader(frame))
	// panicOnError(err)

	// file, err := os.Create("img.jpg")
	// panicOnError(err)

	// defer file.Close()
	// panicOnError(jpeg.Encode(file, img, nil))

	codec, err := webcam.Open_v2("/dev/video11")
	panicOnError(err)
	defer codec.Close()

	err = codec.SetBufferCount(1)
	panicOnError(err)

	f, w, h, err = codec.SetImageFormat_v2(format, size.MaxWidth, size.MaxHeight)
	panicOnError(err)

	_, err = fmt.Fprintf(os.Stderr, "Resulting image format: %s %dx%d\n", formatDesc[f], w, h)
	panicOnError(err)

	fmt.Println("finished")
}
