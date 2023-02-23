package webcam

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"unsafe"

	"github.com/blackjack/webcam/ioctl"
	"golang.org/x/sys/unix"
)

type controlType int

const (
	c_int controlType = iota
	c_bool
	c_menu
)

type control struct {
	id     uint32
	name   string
	c_type controlType
	min    int32
	max    int32
}

const (
	V4L2_CAP_VIDEO_CAPTURE             uint32 = 0x00000001
	V4L2_CAP_STREAMING                 uint32 = 0x04000000
	V4L2_CAP_VIDEO_M2M                        = 0x00008000
	V4L2_CAP_VIDEO_M2M_MPLANE                 = 0x00004000
	V4L2_BUF_TYPE_VIDEO_CAPTURE        uint32 = 1
	V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE        = 9
	V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE         = 10
	V4L2_MEMORY_MMAP                   uint32 = 1
	V4L2_FIELD_ANY                     uint32 = 0
)

const (
	V4L2_FRMSIZE_TYPE_DISCRETE   uint32 = 1
	V4L2_FRMSIZE_TYPE_CONTINUOUS uint32 = 2
	V4L2_FRMSIZE_TYPE_STEPWISE   uint32 = 3
)

const (
	V4L2_CID_BASE               uint32 = 0x00980900
	V4L2_CID_AUTO_WHITE_BALANCE uint32 = V4L2_CID_BASE + 12
	V4L2_CID_PRIVATE_BASE       uint32 = 0x08000000
)

const (
	V4L2_CTRL_TYPE_INTEGER      uint32 = 1
	V4L2_CTRL_TYPE_BOOLEAN      uint32 = 2
	V4L2_CTRL_TYPE_MENU         uint32 = 3
	V4L2_CTRL_TYPE_BUTTON       uint32 = 4
	V4L2_CTRL_TYPE_INTEGER64    uint32 = 5
	V4L2_CTRL_TYPE_CTRL_CLASS   uint32 = 6
	V4L2_CTRL_TYPE_STRING       uint32 = 7
	V4L2_CTRL_TYPE_BITMASK      uint32 = 8
	V4L2_CTRL_TYPE_INTEGER_MENU uint32 = 9

	V4L2_CTRL_COMPOUND_TYPES uint32 = 0x0100
	V4L2_CTRL_TYPE_U8        uint32 = 0x0100
	V4L2_CTRL_TYPE_U16       uint32 = 0x0101
	V4L2_CTRL_TYPE_U32       uint32 = 0x0102
)

const (
	V4L2_CTRL_FLAG_DISABLED  uint32 = 0x00000001
	V4L2_CTRL_FLAG_NEXT_CTRL uint32 = 0x80000000
)

var (
	VIDIOC_QUERYCAP  = ioctl.IoR(uintptr('V'), 0, unsafe.Sizeof(v4l2_capability{}))
	VIDIOC_ENUM_FMT  = ioctl.IoRW(uintptr('V'), 2, unsafe.Sizeof(v4l2_fmtdesc{}))
	VIDIOC_S_FMT     = ioctl.IoRW(uintptr('V'), 5, unsafe.Sizeof(v4l2_format{}))
	VIDIOC_G_FMT     = ioctl.IoRW(uintptr('V'), 4, unsafe.Sizeof(v4l2_format{}))
	VIDIOC_REQBUFS   = ioctl.IoRW(uintptr('V'), 8, unsafe.Sizeof(v4l2_requestbuffers{}))
	VIDIOC_QUERYBUF  = ioctl.IoRW(uintptr('V'), 9, unsafe.Sizeof(v4l2_buffer{}))
	VIDIOC_QBUF      = ioctl.IoRW(uintptr('V'), 15, unsafe.Sizeof(v4l2_buffer{}))
	VIDIOC_DQBUF     = ioctl.IoRW(uintptr('V'), 17, unsafe.Sizeof(v4l2_buffer{}))
	VIDIOC_G_PARM    = ioctl.IoRW(uintptr('V'), 21, unsafe.Sizeof(v4l2_streamparm{}))
	VIDIOC_S_PARM    = ioctl.IoRW(uintptr('V'), 22, unsafe.Sizeof(v4l2_streamparm{}))
	VIDIOC_G_CTRL    = ioctl.IoRW(uintptr('V'), 27, unsafe.Sizeof(v4l2_control{}))
	VIDIOC_S_CTRL    = ioctl.IoRW(uintptr('V'), 28, unsafe.Sizeof(v4l2_control{}))
	VIDIOC_QUERYCTRL = ioctl.IoRW(uintptr('V'), 36, unsafe.Sizeof(v4l2_queryctrl{}))
	//sizeof int32
	VIDIOC_STREAMON        = ioctl.IoW(uintptr('V'), 18, 4)
	VIDIOC_STREAMOFF       = ioctl.IoW(uintptr('V'), 19, 4)
	VIDIOC_ENUM_FRAMESIZES = ioctl.IoRW(uintptr('V'), 74, unsafe.Sizeof(v4l2_frmsizeenum{}))
	__p                    = unsafe.Pointer(uintptr(0))
	NativeByteOrder        = getNativeByteOrder()
)

type v4l2_capability struct {
	driver       [16]uint8
	card         [32]uint8
	bus_info     [32]uint8
	version      uint32
	capabilities uint32
	device_caps  uint32
	reserved     [3]uint32
}

type v4l2_fmtdesc struct {
	index       uint32
	_type       uint32
	flags       uint32
	description [32]uint8
	pixelformat uint32
	reserved    [4]uint32
}

type v4l2_frmsizeenum struct {
	index        uint32
	pixel_format uint32
	_type        uint32
	union        [24]uint8
	reserved     [2]uint32
}

type v4l2_frmsize_discrete struct {
	Width  uint32
	Height uint32
}

type v4l2_frmsize_stepwise struct {
	Min_width   uint32
	Max_width   uint32
	Step_width  uint32
	Min_height  uint32
	Max_height  uint32
	Step_height uint32
}

// Hack to make go compiler properly align union
type v4l2_format_aligned_union struct {
	data [200 - unsafe.Sizeof(__p)]byte
	_    unsafe.Pointer
}

type v4l2_format struct {
	_type uint32
	union v4l2_format_aligned_union
}

type v4l2_pix_format struct {
	Width        uint32
	Height       uint32
	Pixelformat  uint32
	Field        uint32
	Bytesperline uint32
	Sizeimage    uint32
	Colorspace   uint32
	Priv         uint32
	Flags        uint32
	Ycbcr_enc    uint32
	Quantization uint32
	Xfer_func    uint32
}

const VIDEO_MAX_PLANES = 8

type v4l2_pix_format_mplane struct {
	Width        uint32
	Height       uint32
	Pixelformat  uint32
	Field        uint32
	ColorSpace   uint32
	PlaneFmt     [VIDEO_MAX_PLANES]v4l2_plane_pix_format
	NumPlanes    uint8
	Flags        uint8
	Ycbcr_enc    uint32
	Quantization uint32
	Xfer_func    uint32
	Reserved     [7]uint8
}

type v4l2_plane_pix_format struct {
	SizeImage    uint32
	BytesPerLine uint32
	Reserved     [6]uint16
}

type v4l2_requestbuffers struct {
	count    uint32
	_type    uint32
	memory   uint32
	reserved [2]uint32
}

type v4l2_buffer struct {
	index     uint32
	_type     uint32
	bytesused uint32
	flags     uint32
	field     uint32
	timestamp unix.Timeval
	timecode  v4l2_timecode
	sequence  uint32
	memory    uint32
	union     [unsafe.Sizeof(__p)]uint8
	length    uint32
	reserved2 uint32
	reserved  uint32
}

type v4l2_timecode struct {
	_type    uint32
	flags    uint32
	frames   uint8
	seconds  uint8
	minutes  uint8
	hours    uint8
	userbits [4]uint8
}

type v4l2_queryctrl struct {
	id            uint32
	_type         uint32
	name          [32]uint8
	minimum       int32
	maximum       int32
	step          int32
	default_value int32
	flags         uint32
	reserved      [2]uint32
}

type v4l2_control struct {
	id    uint32
	value int32
}

type v4l2_fract struct {
	numerator   uint32
	denominator uint32
}

type v4l2_streamparm_union struct {
	capability     uint32
	output_mode    uint32
	time_per_frame v4l2_fract
	extended_mode  uint32
	buffers        uint32
	reserved       [4]uint32
	data           [200 - (10 * unsafe.Sizeof(uint32(0)))]byte
}

type v4l2_streamparm struct {
	_type uint32
	union v4l2_streamparm_union
}

type union [unsafe.Sizeof(__p)]uint8 // varies for 32-bit and 64-bit
type v4l2_plane struct {
	bytesused   uint32
	length      uint32
	m           union
	data_offset uint32
	reserved    [11]uint32
}

func checkCapabilities_v2(fd uintptr) (bool, error) {
	caps := &v4l2_capability{}
	err := ioctl.Ioctl(fd, VIDIOC_QUERYCAP, uintptr(unsafe.Pointer(caps)))
	if err != nil {
		return false, err
	}

	supportsM2M := (caps.capabilities & V4L2_CAP_VIDEO_M2M) != 0
	supportsM2MMPlane := (caps.capabilities & V4L2_CAP_VIDEO_M2M_MPLANE) != 0
	return supportsM2M || supportsM2MMPlane, nil

}

func checkCapabilities(fd uintptr) (supportsVideoCapture bool, supportsVideoStreaming bool, err error) {

	caps := &v4l2_capability{}

	err = ioctl.Ioctl(fd, VIDIOC_QUERYCAP, uintptr(unsafe.Pointer(caps)))

	if err != nil {
		return
	}

	supportsVideoCapture = (caps.capabilities & V4L2_CAP_VIDEO_CAPTURE) != 0
	supportsVideoStreaming = (caps.capabilities & V4L2_CAP_STREAMING) != 0
	return

}

func getPixelFormat(fd uintptr, index uint32) (code uint32, description string, err error) {

	fmtdesc := &v4l2_fmtdesc{}

	fmtdesc.index = index
	fmtdesc._type = V4L2_BUF_TYPE_VIDEO_CAPTURE

	err = ioctl.Ioctl(fd, VIDIOC_ENUM_FMT, uintptr(unsafe.Pointer(fmtdesc)))

	if err != nil {
		return
	}

	code = fmtdesc.pixelformat
	description = CToGoString(fmtdesc.description[:])

	return
}

func getFrameSize(fd uintptr, index uint32, code uint32) (frameSize FrameSize, err error) {

	frmsizeenum := &v4l2_frmsizeenum{}
	frmsizeenum.index = index
	frmsizeenum.pixel_format = code

	err = ioctl.Ioctl(fd, VIDIOC_ENUM_FRAMESIZES, uintptr(unsafe.Pointer(frmsizeenum)))

	if err != nil {
		return
	}

	switch frmsizeenum._type {

	case V4L2_FRMSIZE_TYPE_DISCRETE:
		discrete := &v4l2_frmsize_discrete{}
		err = binary.Read(bytes.NewBuffer(frmsizeenum.union[:]), NativeByteOrder, discrete)

		if err != nil {
			return
		}

		frameSize.MinWidth = discrete.Width
		frameSize.MaxWidth = discrete.Width
		frameSize.StepWidth = 0
		frameSize.MinHeight = discrete.Height
		frameSize.MaxHeight = discrete.Height
		frameSize.StepHeight = 0

	case V4L2_FRMSIZE_TYPE_CONTINUOUS:

	case V4L2_FRMSIZE_TYPE_STEPWISE:
		stepwise := &v4l2_frmsize_stepwise{}
		err = binary.Read(bytes.NewBuffer(frmsizeenum.union[:]), NativeByteOrder, stepwise)

		if err != nil {
			return
		}

		frameSize.MinWidth = stepwise.Min_width
		frameSize.MaxWidth = stepwise.Max_width
		frameSize.StepWidth = stepwise.Step_width
		frameSize.MinHeight = stepwise.Min_height
		frameSize.MaxHeight = stepwise.Max_height
		frameSize.StepHeight = stepwise.Step_height
	}

	return
}

func setImageFormat_v2(fd uintptr, formatcode *uint32, width *uint32, height *uint32) (err error) {

	format := &v4l2_format{
		_type: V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE,
	}

	pix_mp := v4l2_pix_format_mplane{
		Width:       *width,
		Height:      *height,
		Pixelformat: *formatcode,
	}

	pixbytes := &bytes.Buffer{}
	if err = binary.Write(pixbytes, NativeByteOrder, pix_mp); err != nil {
		return
	}

	copy(format.union.data[:], pixbytes.Bytes())
	if err = ioctl.Ioctl(fd, VIDIOC_G_FMT, uintptr(unsafe.Pointer(format))); err != nil {
		return
	}

	if err = ioctl.Ioctl(fd, VIDIOC_S_FMT, uintptr(unsafe.Pointer(format))); err != nil {
		return
	}

	format._type = V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE
	if err = ioctl.Ioctl(fd, VIDIOC_G_FMT, uintptr(unsafe.Pointer(format))); err != nil {
		return
	}

	pixReverse := &v4l2_pix_format{}
	if err = binary.Read(bytes.NewBuffer(format.union.data[:]), NativeByteOrder, pixReverse); err != nil {
		return
	}

	*width = pixReverse.Width
	*height = pixReverse.Height
	*formatcode = pixReverse.Pixelformat

	return
}

func setImageFormat(fd uintptr, formatcode *uint32, width *uint32, height *uint32) (err error) {

	format := &v4l2_format{
		_type: V4L2_BUF_TYPE_VIDEO_CAPTURE,
	}

	pix := v4l2_pix_format{
		Width:       *width,
		Height:      *height,
		Pixelformat: *formatcode,
		Field:       V4L2_FIELD_ANY,
	}

	pixbytes := &bytes.Buffer{}
	err = binary.Write(pixbytes, NativeByteOrder, pix)

	if err != nil {
		return
	}

	copy(format.union.data[:], pixbytes.Bytes())

	err = ioctl.Ioctl(fd, VIDIOC_S_FMT, uintptr(unsafe.Pointer(format)))

	if err != nil {
		return
	}

	pixReverse := &v4l2_pix_format{}
	err = binary.Read(bytes.NewBuffer(format.union.data[:]), NativeByteOrder, pixReverse)

	if err != nil {
		return
	}

	*width = pixReverse.Width
	*height = pixReverse.Height
	*formatcode = pixReverse.Pixelformat

	return

}

func mmapRequestBuffers_v2(fd uintptr, _type uint32, buf_count *uint32) (err error) {

	req := &v4l2_requestbuffers{}
	req.count = *buf_count
	req._type = _type
	req.memory = V4L2_MEMORY_MMAP

	err = ioctl.Ioctl(fd, VIDIOC_REQBUFS, uintptr(unsafe.Pointer(req)))

	if err != nil {
		return
	}

	*buf_count = req.count

	return

}

func mmapRequestBuffers(fd uintptr, buf_count *uint32) (err error) {

	req := &v4l2_requestbuffers{}
	req.count = *buf_count
	req._type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	req.memory = V4L2_MEMORY_MMAP

	err = ioctl.Ioctl(fd, VIDIOC_REQBUFS, uintptr(unsafe.Pointer(req)))

	if err != nil {
		return
	}

	*buf_count = req.count

	return

}

func mmapQueryBuffer_v2(fd uintptr, _type uint32, index uint32, length *uint32) (buffer []byte, err error) {

	req := &v4l2_buffer{}

	req._type = _type
	req.index = index

	if req.reserved != 0 || req.reserved2 != 0 {
		panic("The reserved and reserved2 fields must be set to 0")
	}

	if unsafe.Sizeof(__p) != 8 {
		panic(fmt.Sprintf("not on 64-bit arch: size of pointer is %d bytes", unsafe.Sizeof(__p)))
	}

	planes := [1]v4l2_plane{{}} // must have a pointer that refers to the newly created object to avoid GC.
	// for 32-bit arch use PutUint32
	NativeByteOrder.PutUint64(req.union[:], uint64(uintptr(unsafe.Pointer(&planes[0]))))
	req.length = 1 // number of elements in req.m.planes

	fmt.Println("BEFORE")
	fmt.Println("Planes Bytes:")
	fmt.Println(hex.Dump(req.union[:]))
	fmt.Println("Planes[0]:")
	fmt.Printf(hex.Dump(*(*[]byte)(unsafe.Pointer(&planes[0]))))
	fmt.Println("DONE")

	if err = ioctl.Ioctl(fd, VIDIOC_QUERYBUF, uintptr(unsafe.Pointer(req))); err != nil {
		err = errors.New(fmt.Sprintf("cannot query the status of the buffer: %v", err.Error()))
		return
	}

	fmt.Println("AFTER")
	fmt.Println("Planes Bytes:")
	fmt.Println(hex.Dump(req.union[:]))
	fmt.Println("Planes[0]:")
	fmt.Printf(hex.Dump(*(*[]byte)(unsafe.Pointer(&planes[0]))))
	fmt.Println("DONE")

	plane := &v4l2_plane{}
	if err = binary.Read(bytes.NewBuffer(*(*[]byte)(unsafe.Pointer(&planes[0]))), NativeByteOrder, plane); err != nil {
		err = errors.New(fmt.Sprintf("cannot read plane: %v", err.Error()))
		return
	}

	fmt.Println("Bytes:")
	fmt.Println(hex.Dump(req.union[:]))
	fmt.Printf("Bytes used: %d\n", plane.bytesused)
	fmt.Printf("Length set: %d\n", plane.length)
	// buffer, err = unix.Mmap(int(fd), int64(offset), int(*length), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	// if err != nil {
	// 	err = errors.New(fmt.Sprintf("cannot map file into memory: %v", err.Error()))
	// }
	return
}

func mmapQueryBuffer(fd uintptr, index uint32, length *uint32) (buffer []byte, err error) {

	req := &v4l2_buffer{}

	req._type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	req.memory = V4L2_MEMORY_MMAP
	req.index = index

	err = ioctl.Ioctl(fd, VIDIOC_QUERYBUF, uintptr(unsafe.Pointer(req)))

	if err != nil {
		return
	}

	var offset uint32
	err = binary.Read(bytes.NewBuffer(req.union[:]), NativeByteOrder, &offset)

	if err != nil {
		return
	}

	*length = req.length

	buffer, err = unix.Mmap(int(fd), int64(offset), int(req.length), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	return
}

func mmapDequeueBuffer(fd uintptr, index *uint32, length *uint32) (err error) {

	buffer := &v4l2_buffer{}

	buffer._type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	buffer.memory = V4L2_MEMORY_MMAP

	err = ioctl.Ioctl(fd, VIDIOC_DQBUF, uintptr(unsafe.Pointer(buffer)))

	if err != nil {
		return
	}

	*index = buffer.index
	*length = buffer.bytesused

	return

}

func mmapEnqueueBuffer_v2(fd uintptr, _type uint32, index uint32) (err error) {

	buffer := &v4l2_buffer{}

	buffer._type = _type
	buffer.memory = V4L2_MEMORY_MMAP
	buffer.index = index

	err = ioctl.Ioctl(fd, VIDIOC_QBUF, uintptr(unsafe.Pointer(buffer)))
	return
}

func mmapEnqueueBuffer(fd uintptr, index uint32) (err error) {

	buffer := &v4l2_buffer{}

	buffer._type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	buffer.memory = V4L2_MEMORY_MMAP
	buffer.index = index

	err = ioctl.Ioctl(fd, VIDIOC_QBUF, uintptr(unsafe.Pointer(buffer)))
	return

}

func mmapReleaseBuffer(buffer []byte) (err error) {
	err = unix.Munmap(buffer)
	return
}

func startStreaming_v2(fd uintptr, uintPointer uint32) (err error) {
	err = ioctl.Ioctl(fd, VIDIOC_STREAMON, uintptr(unsafe.Pointer(&uintPointer)))
	return

}

func startStreaming(fd uintptr) (err error) {

	var uintPointer uint32 = V4L2_BUF_TYPE_VIDEO_CAPTURE
	err = ioctl.Ioctl(fd, VIDIOC_STREAMON, uintptr(unsafe.Pointer(&uintPointer)))
	return

}

func stopStreaming(fd uintptr) (err error) {

	var uintPointer uint32 = V4L2_BUF_TYPE_VIDEO_CAPTURE
	err = ioctl.Ioctl(fd, VIDIOC_STREAMOFF, uintptr(unsafe.Pointer(&uintPointer)))
	return

}

func waitForFrame(fd uintptr, timeout uint32) (count int, err error) {

	for {
		fds := &unix.FdSet{}
		fds.Set(int(fd))

		var oneSecInNsec int64 = 1e9
		timeoutNsec := int64(timeout) * oneSecInNsec
		nativeTimeVal := unix.NsecToTimeval(timeoutNsec)
		tv := &nativeTimeVal

		count, err = unix.Select(int(fd+1), fds, nil, nil, tv)

		if count < 0 && err == unix.EINTR {
			continue
		}
		return
	}

}

func getControl(fd uintptr, id uint32) (int32, error) {
	ctrl := &v4l2_control{}
	ctrl.id = id
	err := ioctl.Ioctl(fd, VIDIOC_G_CTRL, uintptr(unsafe.Pointer(ctrl)))
	return ctrl.value, err
}

func setControl(fd uintptr, id uint32, val int32) error {
	ctrl := &v4l2_control{}
	ctrl.id = id
	ctrl.value = val
	return ioctl.Ioctl(fd, VIDIOC_S_CTRL, uintptr(unsafe.Pointer(ctrl)))
}

func getFramerate(fd uintptr) (float32, error) {
	param := &v4l2_streamparm{}
	param._type = V4L2_BUF_TYPE_VIDEO_CAPTURE

	err := ioctl.Ioctl(fd, VIDIOC_G_PARM, uintptr(unsafe.Pointer(param)))
	if err != nil {
		return 0, err
	}
	tf := param.union.time_per_frame
	if tf.denominator == 0 || tf.numerator == 0 {
		return 0, fmt.Errorf("Invalid framerate (%d/%d)", tf.denominator, tf.numerator)
	}
	return float32(tf.denominator) / float32(tf.numerator), nil
}

func setFramerate(fd uintptr, num, denom uint32) error {
	param := &v4l2_streamparm{}
	param._type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	param.union.time_per_frame.numerator = num
	param.union.time_per_frame.denominator = denom
	return ioctl.Ioctl(fd, VIDIOC_S_PARM, uintptr(unsafe.Pointer(param)))
}

func queryControls(fd uintptr) []control {
	controls := []control{}
	var err error
	// Don't use V42L_CID_BASE since it is the same as brightness.
	var id uint32
	for err == nil {
		id |= V4L2_CTRL_FLAG_NEXT_CTRL
		query := &v4l2_queryctrl{}
		query.id = id
		err = ioctl.Ioctl(fd, VIDIOC_QUERYCTRL, uintptr(unsafe.Pointer(query)))
		id = query.id
		if err == nil {
			if (query.flags & V4L2_CTRL_FLAG_DISABLED) != 0 {
				continue
			}
			var c control
			switch query._type {
			default:
				continue
			case V4L2_CTRL_TYPE_INTEGER, V4L2_CTRL_TYPE_INTEGER64:
				c.c_type = c_int
			case V4L2_CTRL_TYPE_BOOLEAN:
				c.c_type = c_bool
			case V4L2_CTRL_TYPE_MENU:
				c.c_type = c_menu
			}
			c.id = id
			c.name = CToGoString(query.name[:])
			c.min = query.minimum
			c.max = query.maximum
			controls = append(controls, c)
		}
	}
	return controls
}

func getNativeByteOrder() binary.ByteOrder {
	var i int32 = 0x01020304
	u := unsafe.Pointer(&i)
	pb := (*byte)(u)
	b := *pb
	if b == 0x04 {
		return binary.LittleEndian
	} else {
		return binary.BigEndian
	}
}

func CToGoString(c []byte) string {
	n := -1
	for i, b := range c {
		if b == 0 {
			break
		}
		n = i
	}
	return string(c[:n+1])
}
