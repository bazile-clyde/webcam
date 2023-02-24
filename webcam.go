// Library for working with webcams and other video capturing devices.
// It depends entirely on v4l2 framework, thus will compile and work
// only on Linux machine
package webcam

import (
	"errors"
	"fmt"
	"github.com/blackjack/webcam/ioctl"
	"os"
	"reflect"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Webcam object
type Webcam struct {
	fd             uintptr
	bufcount       uint32
	buffers        [][]byte
	buffersCapture [][]byte
	buffersOutput  [][]byte
	streaming      bool
}

type ControlID uint32

type Control struct {
	Name string
	Min  int32
	Max  int32
}

// Open a webcam with a given path
// Checks if device is a v4l2 device and if it is
// capable to stream video
func Open(path string) (*Webcam, error) {

	handle, err := unix.Open(path, unix.O_RDWR|unix.O_NONBLOCK, 0666)
	fd := uintptr(handle)

	if fd < 0 || err != nil {
		return nil, err
	}

	supportsVideoCapture, supportsVideoStreaming, err := checkCapabilities(fd)

	if err != nil {
		return nil, err
	}

	if !supportsVideoCapture {
		return nil, errors.New("Not a video capture device")
	}

	if !supportsVideoStreaming {
		return nil, errors.New("Device does not support the streaming I/O method")
	}

	w := new(Webcam)
	w.fd = uintptr(fd)
	w.bufcount = 256
	return w, nil
}

func Open_v2(path string) (*Webcam, error) {

	handle, err := unix.Open(path, unix.O_RDWR|unix.O_NONBLOCK, 0666)
	fd := uintptr(handle)

	if fd < 0 || err != nil {
		return nil, err
	}

	hasM2M, err := checkCapabilities_v2(fd)

	if err != nil {
		return nil, err
	}

	if !hasM2M {
		return nil, errors.New("not a video mem-to-mem device")
	}

	w := new(Webcam)
	w.fd = fd
	w.bufcount = 256
	return w, nil
}

// Returns image formats supported by the device alongside with
// their text description
// Not that this function is somewhat experimental. Frames are not ordered in
// any meaning, also duplicates can occur so it's up to developer to clean it up.
// See http://linuxtv.org/downloads/v4l-dvb-apis/vidioc-enum-framesizes.html
// for more information
func (w *Webcam) GetSupportedFormats() map[PixelFormat]string {

	result := make(map[PixelFormat]string)
	var err error
	var code uint32
	var desc string
	var index uint32

	for index = 0; err == nil; index++ {
		code, desc, err = getPixelFormat(w.fd, index)

		if err != nil {
			break
		}

		result[PixelFormat(code)] = desc
	}

	return result
}

// Returns supported frame sizes for a given image format
func (w *Webcam) GetSupportedFrameSizes(f PixelFormat) []FrameSize {
	result := make([]FrameSize, 0)

	var index uint32
	var err error

	for index = 0; err == nil; index++ {
		s, err := getFrameSize(w.fd, index, uint32(f))

		if err != nil {
			break
		}

		result = append(result, s)
	}

	return result
}

// Sets desired image format and frame size
// Note, that device driver can change that values.
// Resulting values are returned by a function
// alongside with an error if any
func (w *Webcam) SetImageFormat(f PixelFormat, width, height uint32) (PixelFormat, uint32, uint32, error) {

	code := uint32(f)
	cw := width
	ch := height

	err := setImageFormat(w.fd, &code, &width, &height)

	if err != nil {
		return 0, 0, 0, err
	} else {
		return PixelFormat(code), cw, ch, nil
	}
}

func (w *Webcam) SetImageFormat_v2(f PixelFormat, width, height uint32) (PixelFormat, uint32, uint32, error) {

	code := uint32(f)
	cw := width
	ch := height

	err := setImageFormat_v2(w.fd, &code, &width, &height)

	if err != nil {
		return 0, 0, 0, err
	} else {
		return PixelFormat(code), cw, ch, nil
	}
}

// Set the number of frames to be buffered.
// Not allowed if streaming is already on.
func (w *Webcam) SetBufferCount(count uint32) error {
	if w.streaming {
		return errors.New("Cannot set buffer count when streaming")
	}
	w.bufcount = count
	return nil
}

// Get a map of available controls.
func (w *Webcam) GetControls() map[ControlID]Control {
	cmap := make(map[ControlID]Control)
	for _, c := range queryControls(w.fd) {
		cmap[ControlID(c.id)] = Control{c.name, c.min, c.max}
	}
	return cmap
}

// Get the value of a control.
func (w *Webcam) GetControl(id ControlID) (int32, error) {
	return getControl(w.fd, uint32(id))
}

// Set a control.
func (w *Webcam) SetControl(id ControlID, value int32) error {
	return setControl(w.fd, uint32(id), value)
}

// Get the framerate.
func (w *Webcam) GetFramerate() (float32, error) {
	return getFramerate(w.fd)
}

// Set FPS
func (w *Webcam) SetFramerate(fps float32) error {
	return setFramerate(w.fd, 1000, uint32(1000*(fps)))
}

func (w *Webcam) requestAndMapQueryBuffer(_type uint32, buffers [][]byte) error {
	if err := mmapRequestBuffers_v2(w.fd, _type, &w.bufcount); err != nil {
		return errors.New(fmt.Sprintf("Failed to map buffers: %v : %v", err.Error(), _type))
	}

	buffers = make([][]byte, w.bufcount, w.bufcount)
	for index, _ := range buffers {
		var length uint32
		buffer, err := mmapQueryBuffer_v2(w.fd, _type, uint32(index), &length)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to map memory: %v : %v", err.Error(), _type))
		}

		buffers[index] = buffer
	}
	return nil
}

func (w *Webcam) enqueueBuffer(_type uint32, buffers [][]byte) error {
	for index, _ := range buffers {
		if err := mmapEnqueueBuffer_v2(w.fd, &buffers[index]); err != nil {
			return errors.New(fmt.Sprintf("Failed to enqueue buffer: %s : %d", err.Error(), _type))
		}
	}
	return nil
}

func (w *Webcam) StartStreaming_v2() error {
	if w.streaming {
		return errors.New("Already streaming")
	}

	if err := w.requestAndMapQueryBuffer(V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE, w.buffersOutput); err != nil {
		return err
	}

	if err := w.requestAndMapQueryBuffer(V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE, w.buffersCapture); err != nil {
		return err
	}

	if err := w.enqueueBuffer(V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE, w.buffersOutput); err != nil {
		return err
	}

	if err := w.enqueueBuffer(V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE, w.buffersCapture); err != nil {
		return err
	}

	if err := startStreaming_v2(w.fd, V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE); err != nil {
		return err
	}

	if err := startStreaming_v2(w.fd, V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE); err != nil {
		return err
	}

	w.streaming = true
	return nil
}

// Start streaming process
func (w *Webcam) StartStreaming() error {
	if w.streaming {
		return errors.New("Already streaming")
	}

	err := mmapRequestBuffers(w.fd, &w.bufcount)

	if err != nil {
		return errors.New("Failed to map request buffers: " + string(err.Error()))
	}

	w.buffers = make([][]byte, w.bufcount, w.bufcount)
	for index, _ := range w.buffers {
		var length uint32

		buffer, err := mmapQueryBuffer(w.fd, uint32(index), &length)

		if err != nil {
			return errors.New("Failed to map memory: " + string(err.Error()))
		}

		w.buffers[index] = buffer
	}

	for index, _ := range w.buffers {

		err := mmapEnqueueBuffer(w.fd, uint32(index))

		if err != nil {
			return errors.New("Failed to enqueue buffer: " + string(err.Error()))
		}

	}

	err = startStreaming(w.fd)

	if err != nil {
		return errors.New("Failed to start streaming: " + string(err.Error()))
	}
	w.streaming = true

	return nil
}

// Read a single frame from the webcam
// If frame cannot be read at the moment
// function will return empty slice
func (w *Webcam) ReadFrame() ([]byte, error) {
	result, index, err := w.GetFrame()
	if err == nil {
		w.ReleaseFrame(index)
	}
	return result, err
}

func readFrameFromSource(src *Webcam) ([]byte, error) {
	timeout := uint32(5) // 5 seconds
	for {
		err := src.WaitForFrame(timeout)

		switch err.(type) {
		case nil:
		case *Timeout:
			fmt.Fprint(os.Stderr, err.Error())
			continue
		default:
			panic(err.Error())
		}

		frame, err := src.ReadFrame()
		if err != nil {
			return nil, err
		}

		if len(frame) != 0 {
			return frame, nil
		}
	}
}

func (w *Webcam) ReadFrame_v2(src *Webcam) ([]byte, error) {
	buf := &v4l2_buffer{}
	buf.memory = V4L2_MEMORY_MMAP
	buf.length = 1

	planes := [1]v4l2_plane{{}}                                                          // must have a pointer that refers to the newly created object to avoid GC.
	NativeByteOrder.PutUint64(buf.union[:], uint64(uintptr(unsafe.Pointer(&planes[0])))) // for 32-bit arch use PutUint32

	buf._type = V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE
	if err := ioctl.Ioctl(w.fd, VIDIOC_DQBUF, uintptr(unsafe.Pointer(buf))); err != nil {
		return nil, errors.New(fmt.Sprintf("cannot dequeue output buffer: %s", err.Error()))
	}

	frame, err := readFrameFromSource(src)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot get frame from source: %s", err.Error()))
	}

	// NativeByteOrder.PutUint32(*(*[]byte)(unsafe.Pointer(&planes[0].bytesused)), uint32(uintptr(unsafe.Pointer(&frame)))) // for 32-bit arch use PutUint32
	if err := mmapEnqueueBuffer_v2(w.fd, &w.buffersOutput[0]); err != nil {
		return nil, errors.New(fmt.Sprintf("cannot enqueue output buffer: %s", err.Error()))
	}

	buf._type = V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE
	if err := ioctl.Ioctl(w.fd, VIDIOC_DQBUF, uintptr(unsafe.Pointer(buf))); err != nil {
		return nil, errors.New(fmt.Sprintf("cannot dequeue capture buffer: %s", err.Error()))
	}

	// len := planes[0].bytesused
	if err = mmapEnqueueBuffer_v2(w.fd, &w.buffersCapture[0]); err != nil {
		return nil, errors.New(fmt.Sprintf("cannot enqueue capture buffer: %s", err.Error()))
	}

	return frame, nil
}

// Get a single frame from the webcam and return the frame and
// the buffer index. To return the buffer, ReleaseFrame must be called.
// If frame cannot be read at the moment
// function will return empty slice
func (w *Webcam) GetFrame() ([]byte, uint32, error) {
	var index uint32
	var length uint32

	err := mmapDequeueBuffer(w.fd, &index, &length)

	if err != nil {
		return nil, 0, err
	}

	return w.buffers[int(index)][:length], index, nil

}

// Release the frame buffer that was obtained via GetFrame
func (w *Webcam) ReleaseFrame(index uint32) error {
	return mmapEnqueueBuffer(w.fd, index)
}

// Wait until frame could be read
func (w *Webcam) WaitForFrame(timeout uint32) error {

	count, err := waitForFrame(w.fd, timeout)

	if count < 0 || err != nil {
		return err
	} else if count == 0 {
		return new(Timeout)
	} else {
		return nil
	}
}

func (w *Webcam) StopStreaming() error {
	if !w.streaming {
		return errors.New("Request to stop streaming when not streaming")
	}
	w.streaming = false
	if err := closeBuffers(w.buffersOutput); err != nil {
		return err
	}
	if err := closeBuffers(w.buffersCapture); err != nil {
		return err
	}
	if err := closeBuffers(w.buffers); err != nil {
		return err
	}

	return stopStreaming(w.fd)
}

func closeBuffers(buffers [][]byte) error {
	for _, buffer := range buffers {
		err := mmapReleaseBuffer(buffer)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close the device
func (w *Webcam) Close() error {
	if w.streaming {
		w.StopStreaming()
	}

	err := unix.Close(int(w.fd))

	return err
}

// Sets automatic white balance correction
func (w *Webcam) SetAutoWhiteBalance(val bool) error {
	v := int32(0)
	if val {
		v = 1
	}
	return setControl(w.fd, V4L2_CID_AUTO_WHITE_BALANCE, v)
}

func gobytes(p unsafe.Pointer, n int) []byte {

	h := reflect.SliceHeader{uintptr(p), n, n}
	s := *(*[]byte)(unsafe.Pointer(&h))

	return s
}
