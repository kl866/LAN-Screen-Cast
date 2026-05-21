package ui

import "syscall"

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassEx      = user32.NewProc("RegisterClassExW")
	procCreateWindowEx       = user32.NewProc("CreateWindowExW")
	procDefWindowProc        = user32.NewProc("DefWindowProcW")
	procGetMessage           = user32.NewProc("GetMessageW")
	procTranslateMessage     = user32.NewProc("TranslateMessage")
	procDispatchMessage      = user32.NewProc("DispatchMessageW")
	procPostQuitMessage      = user32.NewProc("PostQuitMessage")
	procSetTimer             = user32.NewProc("SetTimer")
	procInvalidateRect       = user32.NewProc("InvalidateRect")
	procBeginPaint           = user32.NewProc("BeginPaint")
	procEndPaint             = user32.NewProc("EndPaint")
	procGetClientRect        = user32.NewProc("GetClientRect")
	procStretchDIBits        = gdi32.NewProc("StretchDIBits")
	procDestroyWindow        = user32.NewProc("DestroyWindow")
	procLoadCursorW          = user32.NewProc("LoadCursorW")
	procGetModuleHandleW     = kernel32.NewProc("GetModuleHandleW")
	procSetWindowTextW       = user32.NewProc("SetWindowTextW")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procEnableWindow         = user32.NewProc("EnableWindow")
	procSendMessageW         = user32.NewProc("SendMessageW")
)

const (
	CS_VREDRAW           = 1
	CS_HREDRAW           = 2
	WS_OVERLAPPEDWINDOW  = 0x00CF0000
	WS_VISIBLE           = 0x10000000
	WS_CHILD             = 0x40000000
	WS_BORDER            = 0x00800000
	WS_TABSTOP           = 0x00010000
	CW_USEDEFAULT        = 0x80000000
	WM_PAINT             = 15
	WM_TIMER             = 275
	WM_DESTROY           = 2
	WM_CLOSE             = 16
	WM_SIZE              = 5
	WM_COMMAND           = 273
	DIB_RGB_COLORS       = 0
	BI_RGB               = 0
	SRCCOPY              = 0x00CC0020
	BS_PUSHBUTTON        = 0x00000000
	SS_LEFT              = 0x00000000
	ES_LEFT              = 0x0000
	BN_CLICKED           = 0
	COLOR_WINDOW         = 5
	IDC_START            = 103
	IDC_STOP             = 104
	IDC_STATUS           = 105
	IDC_IP               = 101
	IDC_PORT             = 102
	GWLP_USERDATA        = -21
)

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     syscall.Handle
	HIcon         syscall.Handle
	HCursor       syscall.Handle
	HbrBackground syscall.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       syscall.Handle
}

type PAINTSTRUCT struct {
	Hdc         syscall.Handle
	FErase      int32
	RcPaint     [4]int32
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

type RECT struct {
	Left, Top, Right, Bottom int32
}

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

func loadCursor() syscall.Handle {
	c, _, _ := procLoadCursorW.Call(0, 32512) // IDC_ARROW
	return syscall.Handle(c)
}
