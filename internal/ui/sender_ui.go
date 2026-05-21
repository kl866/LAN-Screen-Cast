package ui

import (
	"fmt"
	"syscall"
	"unsafe"

	"lan-screen-cast/internal/engine"
)

var (
	senderEngine  *engine.SenderEngine
	senderRunning bool
	hwndStartBtn  syscall.Handle
	hwndStopBtn   syscall.Handle
	hwndStatusTxt syscall.Handle
	hwndIPEdit    syscall.Handle
	hwndPortEdit  syscall.Handle
)

func RunSender() error {
	className, _ := syscall.UTF16PtrFromString("LanScreenCastSender")
	windowName, _ := syscall.UTF16PtrFromString("LAN Screen Cast - Sender")

	hInst, _, _ := procGetModuleHandleW.Call(0)

	wc := WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		Style:         CS_VREDRAW | CS_HREDRAW,
		LpfnWndProc:   syscall.NewCallback(senderWndProc),
		HInstance:     syscall.Handle(hInst),
		HCursor:       loadCursor(),
		HbrBackground: syscall.Handle(COLOR_WINDOW + 1),
		LpszClassName: className,
	}
	procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))

	hwnd, _, _ := procCreateWindowEx.Call(
		0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(windowName)),
		WS_OVERLAPPEDWINDOW|WS_VISIBLE,
		CW_USEDEFAULT, CW_USEDEFAULT, 340, 240,
		0, 0, hInst, 0,
	)

	createSenderControls(syscall.Handle(hwnd), syscall.Handle(hInst))

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

func createSenderControls(hwnd, hInst syscall.Handle) {
	staticClass, _ := syscall.UTF16PtrFromString("STATIC")
	editClass, _ := syscall.UTF16PtrFromString("EDIT")
	btnClass, _ := syscall.UTF16PtrFromString("BUTTON")

	// IP label
	ipLabel, _ := syscall.UTF16PtrFromString("Target IP:")
	procCreateWindowEx.Call(0, uintptr(unsafe.Pointer(staticClass)),
		uintptr(unsafe.Pointer(ipLabel)),
		WS_CHILD|WS_VISIBLE|SS_LEFT, 10, 15, 80, 20,
		uintptr(hwnd), 0, uintptr(hInst), 0)

	// IP edit
	ipDefault, _ := syscall.UTF16PtrFromString("192.168.1.100")
	ret, _, _ := procCreateWindowEx.Call(0, uintptr(unsafe.Pointer(editClass)),
		uintptr(unsafe.Pointer(ipDefault)),
		WS_CHILD|WS_VISIBLE|WS_BORDER|ES_LEFT|WS_TABSTOP,
		95, 12, 200, 22, uintptr(hwnd), IDC_IP, uintptr(hInst), 0)
	hwndIPEdit = syscall.Handle(ret)

	// Port label
	portLabel, _ := syscall.UTF16PtrFromString("Port:")
	procCreateWindowEx.Call(0, uintptr(unsafe.Pointer(staticClass)),
		uintptr(unsafe.Pointer(portLabel)),
		WS_CHILD|WS_VISIBLE|SS_LEFT, 10, 50, 80, 20,
		uintptr(hwnd), 0, uintptr(hInst), 0)

	// Port edit
	portDefault, _ := syscall.UTF16PtrFromString("9527")
	ret2, _, _ := procCreateWindowEx.Call(0, uintptr(unsafe.Pointer(editClass)),
		uintptr(unsafe.Pointer(portDefault)),
		WS_CHILD|WS_VISIBLE|WS_BORDER|ES_LEFT|WS_TABSTOP,
		95, 47, 200, 22, uintptr(hwnd), IDC_PORT, uintptr(hInst), 0)
	hwndPortEdit = syscall.Handle(ret2)

	// Start button
	startText, _ := syscall.UTF16PtrFromString("Start Sharing")
	ret3, _, _ := procCreateWindowEx.Call(0, uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(startText)),
		WS_CHILD|WS_VISIBLE|BS_PUSHBUTTON|WS_TABSTOP,
		10, 85, 130, 30, uintptr(hwnd), IDC_START, uintptr(hInst), 0)
	hwndStartBtn = syscall.Handle(ret3)

	// Stop button (initially disabled)
	stopText, _ := syscall.UTF16PtrFromString("Stop Sharing")
	ret4, _, _ := procCreateWindowEx.Call(0, uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(stopText)),
		WS_CHILD|WS_VISIBLE|BS_PUSHBUTTON|WS_TABSTOP,
		160, 85, 130, 30, uintptr(hwnd), IDC_STOP, uintptr(hInst), 0)
	hwndStopBtn = syscall.Handle(ret4)
	procEnableWindow.Call(uintptr(hwndStopBtn), 0)

	// Status text
	statusText, _ := syscall.UTF16PtrFromString("Ready")
	ret5, _, _ := procCreateWindowEx.Call(0, uintptr(unsafe.Pointer(staticClass)),
		uintptr(unsafe.Pointer(statusText)),
		WS_CHILD|WS_VISIBLE|SS_LEFT, 10, 135, 300, 40,
		uintptr(hwnd), IDC_STATUS, uintptr(hInst), 0)
	hwndStatusTxt = syscall.Handle(ret5)
}

func senderWndProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_COMMAND:
		ctrlID := uintptr(wParam & 0xFFFF)
		notify := (wParam >> 16) & 0xFFFF
		switch ctrlID {
		case IDC_START:
			if notify == BN_CLICKED {
				startSharing()
			}
		case IDC_STOP:
			if notify == BN_CLICKED {
				stopSharing()
			}
		}
		return 0

	case WM_CLOSE:
		if senderEngine != nil {
			senderEngine.Stop()
		}
		procDestroyWindow.Call(uintptr(hwnd))
		return 0

	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProc.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return ret
}

func getWindowText(hwnd syscall.Handle) string {
	length, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
	if length == 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(length+1))
	return syscall.UTF16ToString(buf)
}

func setStatus(text string) {
	t, _ := syscall.UTF16PtrFromString(text)
	procSetWindowTextW.Call(uintptr(hwndStatusTxt), uintptr(unsafe.Pointer(t)))
}

func startSharing() {
	if senderRunning {
		return
	}
	ip := getWindowText(hwndIPEdit)
	port := getWindowText(hwndPortEdit)
	addr := fmt.Sprintf("%s:%s", ip, port)

	setStatus("Connecting...")
	eng, err := engine.NewSenderEngine(addr)
	if err != nil {
		setStatus(fmt.Sprintf("Error: %v", err))
		return
	}
	senderEngine = eng
	senderRunning = true

	procEnableWindow.Call(uintptr(hwndStartBtn), 0)
	procEnableWindow.Call(uintptr(hwndStopBtn), 1)

	go func() {
		setStatus("Waiting in queue...")
		if err := eng.Run(); err != nil {
			setStatus(fmt.Sprintf("Error: %v", err))
		}
		senderRunning = false
		procEnableWindow.Call(uintptr(hwndStartBtn), 1)
		procEnableWindow.Call(uintptr(hwndStopBtn), 0)
		setStatus("Stopped")
	}()
}

func stopSharing() {
	if senderEngine != nil {
		senderEngine.Stop()
		senderEngine = nil
	}
	senderRunning = false
	procEnableWindow.Call(uintptr(hwndStartBtn), 1)
	procEnableWindow.Call(uintptr(hwndStopBtn), 0)
	setStatus("Stopped")
}
