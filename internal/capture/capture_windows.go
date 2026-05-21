package capture

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32                = syscall.NewLazyDLL("user32.dll")
	gdi32                 = syscall.NewLazyDLL("gdi32.dll")
	procGetDC             = user32.NewProc("GetDC")
	procReleaseDC         = user32.NewProc("ReleaseDC")
	procGetSystemMetrics  = user32.NewProc("GetSystemMetrics")
	procCreateCompatibleDC  = gdi32.NewProc("CreateCompatibleDC")
	procDeleteDC            = gdi32.NewProc("DeleteDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procDeleteObject          = gdi32.NewProc("DeleteObject")
	procSelectObject          = gdi32.NewProc("SelectObject")
	procBitBlt                = gdi32.NewProc("BitBlt")
	procGetDIBits             = gdi32.NewProc("GetDIBits")
)

const (
	SM_CXSCREEN    = 0
	SM_CYSCREEN    = 1
	SRCCOPY        = 0x00CC0020
	DIB_RGB_COLORS = 0
	BI_RGB         = 0
)

type GdiCapturer struct {
	hdcScreen syscall.Handle
	hdcMem    syscall.Handle
	hbmMem    syscall.Handle
	hbmOld    syscall.Handle
	width     int
	height    int
	released  bool
}

func NewGdiCapturer() (*GdiCapturer, error) {
	hdcScreen, _, _ := procGetDC.Call(0)
	if hdcScreen == 0 {
		return nil, fmt.Errorf("GetDC failed")
	}

	w, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	h, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)
	width := int(w)
	height := int(h)

	hdcMem, _, _ := procCreateCompatibleDC.Call(hdcScreen)
	if hdcMem == 0 {
		procReleaseDC.Call(0, hdcScreen)
		return nil, fmt.Errorf("CreateCompatibleDC failed")
	}

	hbmMem, _, _ := procCreateCompatibleBitmap.Call(hdcScreen, uintptr(width), uintptr(height))
	if hbmMem == 0 {
		procDeleteDC.Call(hdcMem)
		procReleaseDC.Call(0, hdcScreen)
		return nil, fmt.Errorf("CreateCompatibleBitmap failed")
	}

	hbmOld, _, _ := procSelectObject.Call(hdcMem, hbmMem)

	return &GdiCapturer{
		hdcScreen: syscall.Handle(hdcScreen),
		hdcMem:    syscall.Handle(hdcMem),
		hbmMem:    syscall.Handle(hbmMem),
		hbmOld:    syscall.Handle(hbmOld),
		width:     width,
		height:    height,
	}, nil
}

func (c *GdiCapturer) Capture() (*Frame, error) {
	if c.released {
		return nil, fmt.Errorf("capturer released")
	}

	procBitBlt.Call(
		uintptr(c.hdcMem), 0, 0, uintptr(c.width), uintptr(c.height),
		uintptr(c.hdcScreen), 0, 0, SRCCOPY,
	)

	bufSize := c.width * c.height * 4
	pixels := make([]byte, bufSize)

	// BITMAPINFO — 44 bytes
	bi := make([]byte, 44)
	*(*uint32)(unsafe.Pointer(&bi[0])) = 44          // biSize
	*(*int32)(unsafe.Pointer(&bi[4])) = int32(c.width) // biWidth
	*(*int32)(unsafe.Pointer(&bi[8])) = -int32(c.height) // biHeight (negative = top-down)
	*(*uint16)(unsafe.Pointer(&bi[12])) = 1           // biPlanes
	*(*uint16)(unsafe.Pointer(&bi[14])) = 32          // biBitCount
	*(*uint32)(unsafe.Pointer(&bi[16])) = BI_RGB      // biCompression

	procGetDIBits.Call(
		uintptr(c.hdcMem), uintptr(c.hbmMem),
		0, uintptr(c.height),
		uintptr(unsafe.Pointer(&pixels[0])),
		uintptr(unsafe.Pointer(&bi[0])),
		DIB_RGB_COLORS,
	)

	return &Frame{
		Width:     c.width,
		Height:    c.height,
		Pixels:    pixels,
		Timestamp: time.Now(),
	}, nil
}

func (c *GdiCapturer) Release() {
	if c.released {
		return
	}
	c.released = true
	procSelectObject.Call(uintptr(c.hdcMem), uintptr(c.hbmOld))
	procDeleteObject.Call(uintptr(c.hbmMem))
	procDeleteDC.Call(uintptr(c.hdcMem))
	procReleaseDC.Call(0, uintptr(c.hdcScreen))
}
