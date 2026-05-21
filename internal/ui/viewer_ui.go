package ui

import (
	"syscall"
	"unsafe"

	"lan-screen-cast/internal/engine"
)

var viewerEngine *engine.ViewerEngine

func RunViewer(addr string) error {
	var err error
	viewerEngine, err = engine.NewViewerEngine(addr)
	if err != nil {
		return err
	}
	go viewerEngine.AcceptLoop()

	className, _ := syscall.UTF16PtrFromString("LanScreenCastViewer")
	windowName, _ := syscall.UTF16PtrFromString("LAN Screen Cast - Viewer")

	hInst, _, _ := procGetModuleHandleW.Call(0)

	wc := WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		Style:         CS_VREDRAW | CS_HREDRAW,
		LpfnWndProc:   syscall.NewCallback(viewerWndProc),
		HInstance:     syscall.Handle(hInst),
		HCursor:       loadCursor(),
		HbrBackground: syscall.Handle(COLOR_WINDOW + 1),
		LpszClassName: className,
	}
	procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))

	hwnd, _, _ := procCreateWindowEx.Call(
		0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(windowName)),
		WS_OVERLAPPEDWINDOW|WS_VISIBLE,
		CW_USEDEFAULT, CW_USEDEFAULT, 1024, 768,
		0, 0, hInst, 0,
	)

	procSetTimer.Call(hwnd, 1, 33, 0) // 30 FPS

	var msg [28]byte
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg[0])), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg[0])))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg[0])))
	}

	return nil
}

func viewerWndProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_PAINT:
		var ps PAINTSTRUCT
		procBeginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))
		viewerPaint(syscall.Handle(ps.Hdc))
		procEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))
		return 0

	case WM_TIMER:
		if viewerEngine != nil && viewerEngine.FrameBuffer().HasUpdate() {
			viewerEngine.FrameBuffer().ClearUpdate()
			procInvalidateRect.Call(uintptr(hwnd), 0, 0)
		}
		return 0

	case WM_SIZE:
		procInvalidateRect.Call(uintptr(hwnd), 0, 0)
		return 0

	case WM_CLOSE:
		procDestroyWindow.Call(uintptr(hwnd))
		return 0

	case WM_DESTROY:
		if viewerEngine != nil {
			viewerEngine.Stop()
		}
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProc.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return ret
}

func viewerPaint(hdc syscall.Handle) {
	if viewerEngine == nil {
		return
	}
	w, h, pixels := viewerEngine.FrameBuffer().Snapshot()
	if len(pixels) == 0 {
		return
	}

	var rc RECT
	procGetClientRect.Call(0, uintptr(unsafe.Pointer(&rc))) // hwnd not needed for paint DC
	cw := int(rc.Right - rc.Left)
	ch := int(rc.Bottom - rc.Top)
	if cw == 0 || ch == 0 {
		return
	}

	bi := BITMAPINFOHEADER{
		BiSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
		BiWidth:       int32(w),
		BiHeight:      -int32(h),
		BiPlanes:      1,
		BiBitCount:    32,
		BiCompression: BI_RGB,
	}

	procStretchDIBits.Call(
		uintptr(hdc),
		0, 0, uintptr(cw), uintptr(ch),
		0, 0, uintptr(w), uintptr(h),
		uintptr(unsafe.Pointer(&pixels[0])),
		uintptr(unsafe.Pointer(&bi)),
		DIB_RGB_COLORS, SRCCOPY,
	)
}
