package capture

import "time"

type Frame struct {
	Width     int
	Height    int
	Pixels    []byte // RGBA, row-major, stride = Width * 4
	Timestamp time.Time
}

type Capturer interface {
	Capture() (*Frame, error)
	Release()
}
