package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	_ "golang.org/x/image/webp"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")

	procCreateWindowExW      = user32.NewProc("CreateWindowExW")
	procDefWindowProcW       = user32.NewProc("DefWindowProcW")
	procRegisterClassExW     = user32.NewProc("RegisterClassExW")
	procShowWindow           = user32.NewProc("ShowWindow")
	procUpdateWindow         = user32.NewProc("UpdateWindow")
	procGetMessageW          = user32.NewProc("GetMessageW")
	procTranslateMessage     = user32.NewProc("TranslateMessage")
	procDispatchMessageW     = user32.NewProc("DispatchMessageW")
	procPostQuitMessage      = user32.NewProc("PostQuitMessage")
	procSendMessageW         = user32.NewProc("SendMessageW")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procMessageBoxW          = user32.NewProc("MessageBoxW")
	procGetModuleHandleW     = kernel32.NewProc("GetModuleHandleW")
	procCreateFontW          = gdi32.NewProc("CreateFontW")
	procGetOpenFileNameW     = comdlg32.NewProc("GetOpenFileNameW")
	procSetCursorPos         = user32.NewProc("SetCursorPos")
	procGetCursorPos         = user32.NewProc("GetCursorPos")
	procGetAsyncKeyState     = user32.NewProc("GetAsyncKeyState")
	procMouseEvent           = user32.NewProc("mouse_event")
	procLoadImageW           = user32.NewProc("LoadImageW")
	procSetWindowPos         = user32.NewProc("SetWindowPos")
	procInvalidateRect       = user32.NewProc("InvalidateRect")
	procGetClientRect        = user32.NewProc("GetClientRect")
	procClientToScreen       = user32.NewProc("ClientToScreen")
	procDeleteObject         = gdi32.NewProc("DeleteObject")
	procCreatePen            = gdi32.NewProc("CreatePen")
	procCreateSolidBrush     = gdi32.NewProc("CreateSolidBrush")
	procSelectObject         = gdi32.NewProc("SelectObject")
	procRectangle            = gdi32.NewProc("Rectangle")
	procBeginPaint           = user32.NewProc("BeginPaint")
	procEndPaint             = user32.NewProc("EndPaint")
	procFillRect             = user32.NewProc("FillRect")
	procDestroyWindow        = user32.NewProc("DestroyWindow")
	procGetSystemMetrics     = user32.NewProc("GetSystemMetrics")
	procSetTimer             = user32.NewProc("SetTimer")
	procKillTimer            = user32.NewProc("KillTimer")
	procSetTextColor         = gdi32.NewProc("SetTextColor")
	procTextOut              = gdi32.NewProc("TextOutW")
	procCreateCompatibleDC   = gdi32.NewProc("CreateCompatibleDC")
	procGetStockObject       = gdi32.NewProc("GetStockObject")
	procDeleteDC             = gdi32.NewProc("DeleteDC")
	procBitBlt               = gdi32.NewProc("BitBlt")
	procGetDC                = user32.NewProc("GetDC")
	procReleaseDC            = user32.NewProc("ReleaseDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procGetDesktopWindow     = user32.NewProc("GetDesktopWindow")
	procStretchBlt           = gdi32.NewProc("StretchBlt")
	procGetObject            = gdi32.NewProc("GetObjectW")
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	procScreenToClient       = user32.NewProc("ScreenToClient")
	procSetForegroundWindow  = user32.NewProc("SetForegroundWindow")
	procSetBkMode            = gdi32.NewProc("SetBkMode")
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_BORDER           = 0x00800000
	WS_TABSTOP          = 0x00010000
	WS_POPUP            = 0x80000000
	ES_AUTOHSCROLL      = 0x0080
	ES_CENTER           = 0x0001
	BS_PUSHBUTTON       = 0x00000000
	BS_GROUPBOX         = 0x00000007
	SS_BITMAP           = 0x0000000E
	SS_CENTERIMAGE      = 0x00000800
	WM_DESTROY          = 0x0002
	WM_COMMAND          = 0x0111
	WM_SETFONT          = 0x0030
	WM_SETTEXT          = 0x000C
	WM_GETTEXT          = 0x000D
	WM_GETTEXTLENGTH    = 0x000E
	WM_PAINT            = 0x000F
	WM_ERASEBKGND       = 0x0014
	WM_KEYDOWN          = 0x0100
	WM_TIMER            = 0x0113
	WM_USER             = 0x0400
	WM_APP              = 0x8000
	SW_SHOW             = 5
	SW_HIDE             = 0
	MB_OK               = 0x00000000
	MB_ICONWARNING      = 0x00000030
	MB_ICONERROR        = 0x00000010
	MB_ICONINFORMATION  = 0x00000040
	MOUSEEVENTF_LEFTDOWN = 0x0002
	MOUSEEVENTF_LEFTUP   = 0x0004
	VK_OEM_PLUS         = 0xBB
	VK_OEM_MINUS        = 0xBD
	VK_NUMPAD_ADD       = 0x6B
	VK_NUMPAD_SUBTRACT  = 0x6D
	VK_ESCAPE           = 0x1B
	VK_RETURN           = 0x0D
	VK_LBUTTON          = 0x01
	CW_USEDEFAULT       = 0x80000000
	OFN_FILEMUSTEXIST   = 0x00001000
	OFN_HIDEREADONLY    = 0x00000004
	OFN_EXPLORER        = 0x00080000
	PBM_SETRANGE32      = 0x0401
	PBM_SETPOS          = 0x0402
	CS_HREDRAW          = 0x0002
	CS_VREDRAW          = 0x0001
	COLOR_WINDOW        = 5
	STM_SETIMAGE        = 0x0172
	IMAGE_BITMAP        = 0
	LR_LOADFROMFILE     = 0x00000010
	LR_CREATEDIBSECTION = 0x00002000
	PS_SOLID            = 0
	DEFAULT_CHARSET     = 1
	OUTLINE_FONTTYPE    = 0
	CLIP_DEFAULT_PRECIS = 0
	DEFAULT_QUALITY     = 0
	DEFAULT_PITCH       = 0
	FF_DONTCARE         = 0
	WS_EX_TOPMOST       = 0x00000008
	WS_EX_LAYERED       = 0x00080000
	LWA_COLORKEY        = 0x00000001
	LWA_ALPHA           = 0x00000002
	SRCCOPY             = 0x00CC0020
	TRANSPARENT         = 1

	TRANSPARENT_COLOR = 0x00FF00FF
)

type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type OPENFILENAMEW struct {
	LStructSize       uint32
	HwndOwner         uintptr
	HInstance         uintptr
	LpstrFilter       *uint16
	LpstrCustomFilter *uint16
	NMaxCustFilter    uint32
	NFilterIndex      uint32
	LpstrFile         *uint16
	NMaxFile          uint32
	LpstrFileTitle    *uint16
	NMaxFileTitle     uint32
	LpstrInitialDir   *uint16
	LpstrTitle        *uint16
	Flags             uint32
	NFileOffset       uint16
	NFileExtension    uint16
	LpstrDefExt       *uint16
	LCustData         uintptr
	LpfnHook          uintptr
	LpTemplateName    *uint16
	PvReserved        uintptr
	DwReserved        uint32
	FlagsEx           uint32
}

type MSG struct {
	Hwnd    uintptr
	Message uint32
	_       [4]byte
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	PtX     int32
	PtY     int32
}

type POINT struct{ X, Y int32 }
type RECT struct{ Left, Top, Right, Bottom int32 }

type ProcessResult struct {
	paths      []Path
	imgW       int
	imgH       int
	previewBMP string
	err        error
}

var (
	hMainWnd uintptr
	hFont    uintptr

	hEditThreshold uintptr
	hEditBlur      uintptr
	hEditMinLen    uintptr
	hEditSimplify  uintptr
	hEditSpeed     uintptr
	hEditLinePause uintptr
	hEditStatus    uintptr
	hProgress      uintptr
	hPreview       uintptr

	imagePath  string
	paths      []Path
	imgW, imgH int
	calibrated bool
	canvasTL   Pt
	canvasBR   Pt

	drawSpeed float64 = 0.01
	linePause float64 = 0.1
	speedStep float64 = 0.005

	maxImageSize = 800

	timerDraw    uintptr = 1
	timerProcess uintptr = 2
	timerSel     uintptr = 3

	drawState      int = 0
	drawPathsCopy  []Path
	drawCurSpeed   float64
	drawCurPause   float64

	selState   int = 0
	selStartPt POINT
	selEndPt   POINT

	hSelWnd       uintptr
	selWndClass   = "DrawSelWnd_v5"
	selRegistered bool
	previewBMP    string

	// 预加载的预览位图句柄
	hPreviewBMP uintptr

	processResultChan chan ProcessResult
)

type Pt struct{ X, Y int }
type Path []Pt

func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func getEditText(h uintptr) string {
	n, _, _ := procGetWindowTextLengthW.Call(h)
	if n == 0 {
		return ""
	}
	buf := make([]uint16, n+1)
	procGetWindowTextW.Call(h, uintptr(unsafe.Pointer(&buf[0])), n+1)
	return syscall.UTF16ToString(buf)
}

func setEditText(h uintptr, text string) {
	procSendMessageW.Call(h, WM_SETTEXT, 0, uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func msgBox(title, msg string, style uintptr) int {
	ret, _, _ := procMessageBoxW.Call(0, uintptr(unsafe.Pointer(utf16Ptr(msg))), uintptr(unsafe.Pointer(utf16Ptr(title))), style)
	return int(ret)
}

func setStatus(msg string) {
	if hEditStatus != 0 {
		setEditText(hEditStatus, msg)
	}
}

func setProgress(cur, total int) {
	if total <= 0 {
		return
	}
	procSendMessageW.Call(hProgress, PBM_SETRANGE32, 0, uintptr(total))
	procSendMessageW.Call(hProgress, PBM_SETPOS, 0, uintptr(cur))
}

func parseInt(s string, def int) int {
	n := 0
	for _, c := range strings.TrimSpace(s) {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			return def
		}
	}
	if n == 0 {
		return def
	}
	return n
}

func parseFloat(s string, def float64) float64 {
	n, dot, div := 0, false, 1.0
	for _, c := range strings.TrimSpace(s) {
		if c >= '0' && c <= '9' {
			if dot {
				div *= 10
			}
			n = n*10 + int(c-'0')
		} else if c == '.' {
			dot = true
		}
	}
	if v := float64(n) / div; v > 0 {
		return v
	}
	return def
}

func openFileDlg(title, filter string) string {
	buf := make([]uint16, 1024)
	ofn := OPENFILENAMEW{
		LStructSize: uint32(unsafe.Sizeof(OPENFILENAMEW{})),
		HwndOwner:   hMainWnd,
		LpstrFilter: utf16Ptr(filter),
		LpstrFile:   &buf[0],
		NMaxFile:    1024,
		LpstrTitle:  utf16Ptr(title),
		Flags:       OFN_FILEMUSTEXIST | OFN_HIDEREADONLY | OFN_EXPLORER,
	}
	if ret, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn))); ret == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

func setCursorPos(x, y int)    { procSetCursorPos.Call(uintptr(x), uintptr(y)) }
func getCursorPos() (int, int) { var p POINT; procGetCursorPos.Call(uintptr(unsafe.Pointer(&p))); return int(p.X), int(p.Y) }
func mouseDown()               { procMouseEvent.Call(MOUSEEVENTF_LEFTDOWN, 0, 0, 0, 0) }
func mouseUp()                 { procMouseEvent.Call(MOUSEEVENTF_LEFTUP, 0, 0, 0, 0) }
func isKeyDown(vk uint16) bool { ret, _, _ := procGetAsyncKeyState.Call(uintptr(vk)); return ret&0x8000 != 0 }

type ImageProcessor struct {
	W, H  int
	Gray  [][]float64
	Edges [][]float64
}

func LoadImage(path string) (*ImageProcessor, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	header := make([]byte, 512)
	n, _ := f.Read(header)
	f.Seek(0, 0)

	var img image.Image
	var decodeErr error

	if n >= 2 && header[0] == 0xFF && header[1] == 0xD8 {
		img, decodeErr = jpeg.Decode(f)
	} else if n >= 8 && string(header[0:8]) == "\x89PNG\r\n\x1a\n" {
		img, decodeErr = png.Decode(f)
	} else if n >= 6 && (string(header[0:6]) == "GIF87a" || string(header[0:6]) == "GIF89a") {
		img, decodeErr = gif.Decode(f)
	} else if n >= 2 && header[0] == 'B' && header[1] == 'M' {
		img, decodeErr = decodeBMP(f)
	} else if n >= 12 && string(header[0:4]) == "RIFF" && string(header[8:12]) == "WEBP" {
		img, decodeErr = webpDecode(f)
	} else {
		img, _, decodeErr = image.Decode(f)
	}

	if decodeErr != nil {
		return nil, fmt.Errorf("解码失败: %v", decodeErr)
	}

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	if w > maxImageSize || h > maxImageSize {
		scale := math.Min(float64(maxImageSize)/float64(w), float64(maxImageSize)/float64(h))
		newW, newH := int(float64(w)*scale), int(float64(h)*scale)
		dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
		for y := 0; y < newH; y++ {
			for x := 0; x < newW; x++ {
				srcX := int(float64(x) / float64(newW) * float64(w))
				srcY := int(float64(y) / float64(newH) * float64(h))
				if srcX >= w {
					srcX = w - 1
				}
				if srcY >= h {
					srcY = h - 1
				}
				dst.Set(x, y, img.At(srcX+b.Min.X, srcY+b.Min.Y))
			}
		}
		img = dst
		w, h = newW, newH
	}

	ip := &ImageProcessor{W: w, H: h, Gray: make([][]float64, h), Edges: make([][]float64, h)}
	for y := 0; y < h; y++ {
		ip.Gray[y] = make([]float64, w)
		ip.Edges[y] = make([]float64, w)
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			ip.Gray[y][x] = float64(r>>8)*0.299 + float64(g>>8)*0.587 + float64(b>>8)*0.114
		}
	}
	return ip, nil
}

func (ip *ImageProcessor) Blur(k int) {
	if k < 1 {
		return
	}
	if k%2 == 0 {
		k++
	}
	r, sig := k/2, float64(k/2)/3.0
	kern := make([]float64, k)
	for i := range kern {
		x := float64(i - r)
		kern[i] = math.Exp(-x*x/(2*sig*sig)) / (sig * math.Sqrt(2*math.Pi))
	}
	sum := 0.0
	for _, v := range kern {
		sum += v
	}
	for i := range kern {
		kern[i] /= sum
	}
	tmp := make([][]float64, ip.H)
	for y := 0; y < ip.H; y++ {
		tmp[y] = make([]float64, ip.W)
		for x := 0; x < ip.W; x++ {
			for j := 0; j < k; j++ {
				xx := x + j - r
				if xx < 0 {
					xx = 0
				} else if xx >= ip.W {
					xx = ip.W - 1
				}
				tmp[y][x] += ip.Gray[y][xx] * kern[j]
			}
		}
	}
	for y := 0; y < ip.H; y++ {
		for x := 0; x < ip.W; x++ {
			ip.Gray[y][x] = 0
			for j := 0; j < k; j++ {
				yy := y + j - r
				if yy < 0 {
					yy = 0
				} else if yy >= ip.H {
					yy = ip.H - 1
				}
				ip.Gray[y][x] += tmp[yy][x] * kern[j]
			}
		}
	}
}

func (ip *ImageProcessor) Sobel() {
	for y := 1; y < ip.H-1; y++ {
		for x := 1; x < ip.W-1; x++ {
			gx := -ip.Gray[y-1][x-1] + ip.Gray[y-1][x+1] - 2*ip.Gray[y][x-1] + 2*ip.Gray[y][x+1] - ip.Gray[y+1][x-1] + ip.Gray[y+1][x+1]
			gy := -ip.Gray[y-1][x-1] - 2*ip.Gray[y-1][x] - ip.Gray[y-1][x+1] + ip.Gray[y+1][x-1] + 2*ip.Gray[y+1][x] + ip.Gray[y+1][x+1]
			ip.Edges[y][x] = math.Sqrt(gx*gx + gy*gy)
		}
	}
}

func (ip *ImageProcessor) Threshold(t float64) {
	for y := 0; y < ip.H; y++ {
		for x := 0; x < ip.W; x++ {
			if ip.Edges[y][x] >= t {
				ip.Edges[y][x] = 255
			} else {
				ip.Edges[y][x] = 0
			}
		}
	}
}

func (ip *ImageProcessor) Trace(minLen int) []Path {
	vis := make([][]bool, ip.H)
	for y := range vis {
		vis[y] = make([]bool, ip.W)
	}
	dx := []int{-1, -1, -1, 0, 0, 1, 1, 1}
	dy := []int{-1, 0, 1, -1, 1, -1, 0, 1}
	var result []Path
	for y := 1; y < ip.H-1; y++ {
		for x := 1; x < ip.W-1; x++ {
			if ip.Edges[y][x] > 0 && !vis[y][x] {
				var p Path
				cx, cy := x, y
				for cx >= 0 && cx < ip.W && cy >= 0 && cy < ip.H && ip.Edges[cy][cx] > 0 && !vis[cy][cx] {
					p = append(p, Pt{cx, cy})
					vis[cy][cx] = true
					found := false
					for d := 0; d < 8; d++ {
						if nx, ny := cx+dx[d], cy+dy[d]; nx >= 0 && nx < ip.W && ny >= 0 && ny < ip.H && ip.Edges[ny][nx] > 0 && !vis[ny][nx] {
							cx, cy = nx, ny
							found = true
							break
						}
					}
					if !found {
						break
					}
				}
				if len(p) >= minLen {
					result = append(result, p)
				}
			}
		}
	}
	return result
}

func SimplifyPath(path Path, eps float64) Path {
	if len(path) <= 2 {
		return path
	}
	mx, mi := 0.0, 0
	p0, pn := path[0], path[len(path)-1]
	for i := 1; i < len(path)-1; i++ {
		if d := ptLineDist(path[i], p0, pn); d > mx {
			mx, mi = d, i
		}
	}
	if mx > eps {
		l, r := SimplifyPath(path[:mi+1], eps), SimplifyPath(path[mi:], eps)
		return append(l[:len(l)-1], r...)
	}
	return Path{p0, pn}
}

func ptLineDist(p, a, b Pt) float64 {
	dx, dy := float64(b.X-a.X), float64(b.Y-a.Y)
	if ls := dx*dx + dy*dy; ls > 0 {
		t := (float64(p.X-a.X)*dx + float64(p.Y-a.Y)*dy) / ls
		if t < 0 {
			t = 0
		} else if t > 1 {
			t = 1
		}
		return math.Sqrt((float64(p.X)-(float64(a.X)+t*dx))*(float64(p.X)-(float64(a.X)+t*dx)) + (float64(p.Y)-(float64(a.Y)+t*dy))*(float64(p.Y)-(float64(a.Y)+t*dy)))
	}
	return math.Sqrt(float64((p.X-a.X)*(p.X-a.X) + (p.Y-a.Y)*(p.Y-a.Y)))
}

func SavePreviewBMP(paths []Path, w, h, pw, ph int, outPath string) error {
	img := image.NewRGBA(image.Rect(0, 0, pw, ph))
	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 255}
	for y := 0; y < ph; y++ {
		for x := 0; x < pw; x++ {
			img.Set(x, y, white)
		}
	}
	if len(paths) > 0 && w > 0 && h > 0 {
		sc := math.Min(float64(pw)/float64(w), float64(ph)/float64(h)) * 0.9
		ox := (pw - int(float64(w)*sc)) / 2
		oy := (ph - int(float64(h)*sc)) / 2
		for _, path := range paths {
			for i := 1; i < len(path); i++ {
				x0 := int(float64(path[i-1].X)*sc) + ox
				y0 := int(float64(path[i-1].Y)*sc) + oy
				x1 := int(float64(path[i].X)*sc) + ox
				y1 := int(float64(path[i].Y)*sc) + oy
				drawLineColor(img, x0, y0, x1, y1, black)
			}
		}
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	bmpW, bmpH := img.Rect.Dx(), img.Rect.Dy()
	rowSize := (bmpW*3 + 3) & ^3
	fileSize := 54 + rowSize*bmpH
	header := make([]byte, fileSize)

	header[0] = 'B'
	header[1] = 'M'
	le(header, 2, uint32(fileSize))
	le(header, 10, uint32(54))
	le(header, 14, uint32(40))
	le(header, 18, int32(bmpW))
	le(header, 22, int32(bmpH))
	le(header, 26, uint16(1))
	le(header, 28, uint16(24))
	le(header, 34, uint32(rowSize*bmpH))

	pixelOffset := 54
	for y := 0; y < bmpH; y++ {
		bmpRow := bmpH - 1 - y
		for x := 0; x < bmpW; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			b8 := byte(b >> 8)
			g8 := byte(g >> 8)
			r8 := byte(r >> 8)
			idx := pixelOffset + bmpRow*rowSize + x*3
			header[idx] = b8
			header[idx+1] = g8
			header[idx+2] = r8
		}
	}

	f.Write(header)
	return nil
}

func le(b []byte, off int, v interface{}) {
	switch val := v.(type) {
	case uint32:
		b[off] = byte(val)
		b[off+1] = byte(val >> 8)
		b[off+2] = byte(val >> 16)
		b[off+3] = byte(val >> 24)
	case int32:
		b[off] = byte(val)
		b[off+1] = byte(val >> 8)
		b[off+2] = byte(val >> 16)
		b[off+3] = byte(val >> 24)
	case uint16:
		b[off] = byte(val)
		b[off+1] = byte(val >> 8)
	}
}

func drawLineColor(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	dx, dy := absi(x1-x0), absi(y1-y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	e := dx - dy
	for {
		if x0 >= 0 && x0 < img.Rect.Dx() && y0 >= 0 && y0 < img.Rect.Dy() {
			img.Set(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * e
		if e2 > -dy {
			e -= dy
			x0 += sx
		}
		if e2 < dx {
			e += dx
			y0 += sy
		}
	}
}

func absi(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func ProcessImage(imgPath string, blur, thresh, minLen int, simplify float64) ([]Path, int, int, error) {
	ip, err := LoadImage(imgPath)
	if err != nil {
		return nil, 0, 0, err
	}
	ip.Blur(blur)
	ip.Sobel()
	ip.Threshold(float64(thresh))
	p := ip.Trace(minLen)
	for i := range p {
		p[i] = SimplifyPath(p[i], simplify)
	}
	return p, ip.W, ip.H, nil
}

func doProcessAsync(path, previewPath string, blur, thresh, minLen int, simplify float64) {
	p, w, h, err := ProcessImage(path, blur, thresh, minLen, simplify)

	result := ProcessResult{
		paths:      p,
		imgW:       w,
		imgH:       h,
		previewBMP: previewPath,
		err:        err,
	}

	if err == nil && len(p) > 0 {
		if saveErr := SavePreviewBMP(p, w, h, 400, 300, previewPath); saveErr != nil {
			result.err = saveErr
		}
	}

	select {
	case processResultChan <- result:
	default:
	}
}

func doDrawing() {
	cw := canvasBR.X - canvasTL.X
	ch := canvasBR.Y - canvasTL.Y
	sx := float64(cw) / float64(imgW)
	sy := float64(ch) / float64(imgH)
	sc := sx
	if sy < sx {
		sc = sy
	}
	offsetX := canvasTL.X + (cw-int(float64(imgW)*sc))/2
	offsetY := canvasTL.Y + (ch-int(float64(imgH)*sc))/2

	total := len(drawPathsCopy)
	setStatus(fmt.Sprintf("开始绘画，共 %d 条路径", total))
	defer func() { mouseUp(); drawState = 0 }()

	for i, path := range drawPathsCopy {
		if drawState != 1 {
			return
		}
		setProgress(i+1, total)
		if len(path) == 0 {
			continue
		}
		setCursorPos(int(float64(path[0].X)*sc)+offsetX, int(float64(path[0].Y)*sc)+offsetY)
		time.Sleep(10 * time.Millisecond)
		mouseDown()
		time.Sleep(5 * time.Millisecond)
		for j := 1; j < len(path); j++ {
			if drawState != 1 {
				mouseUp()
				return
			}
			setCursorPos(int(float64(path[j].X)*sc)+offsetX, int(float64(path[j].Y)*sc)+offsetY)
			if drawCurSpeed > 0.002 {
				time.Sleep(time.Duration(drawCurSpeed * 100 * float64(time.Millisecond)))
			}
		}
		mouseUp()
		if drawCurPause > 0.01 {
			time.Sleep(time.Duration(drawCurPause * 100 * float64(time.Millisecond)))
		}
	}
	setStatus("绘画完成！")
	setProgress(total, total)
}

func doStart() {
	if len(paths) == 0 {
		msgBox("提示", "没有可绘制的路径，请先解析图片", MB_OK|MB_ICONWARNING)
		return
	}
	if !calibrated {
		msgBox("提示", "请先校准画布位置", MB_OK|MB_ICONWARNING)
		return
	}
	if drawState == 1 {
		return
	}

	drawPathsCopy = make([]Path, len(paths))
	copy(drawPathsCopy, paths)
	drawCurSpeed = parseFloat(getEditText(hEditSpeed), 0.01)
	drawCurPause = parseFloat(getEditText(hEditLinePause), 0.1)
	drawState = 1

	go doDrawing()
}

// ============================================================
// 选区窗口 - 简化版，只绘制边框，不加载图片
// ============================================================

func selWndProc(hwnd uintptr, msg uint32, wp, lp uintptr) uintptr {
	switch msg {
	case WM_PAINT:
		var ps struct {
			hdc        uintptr
			fErase     int32
			rcPaint    RECT
			fRestore   int32
			fIncUpdate int32
			rgbReserved [32]byte
		}
		procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		hdc := ps.hdc

		// 获取屏幕尺寸
		sxR, _, _ := procGetSystemMetrics.Call(0)
		syR, _, _ := procGetSystemMetrics.Call(1)

		// 填充透明色
		brush, _, _ := procCreateSolidBrush.Call(TRANSPARENT_COLOR)
		var rect RECT
		rect.Right = int32(sxR)
		rect.Bottom = int32(syR)
		procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rect)), brush)
		procDeleteObject.Call(brush)

		// 只在拖拽时绘制选区边框
		if selState == 1 {
			x1 := int(selStartPt.X)
			y1 := int(selStartPt.Y)
			x2 := int(selEndPt.X)
			y2 := int(selEndPt.Y)

			if x1 > x2 { x1, x2 = x2, x1 }
			if y1 > y2 { y1, y2 = y2, y1 }

			selW := x2 - x1
			selH := y2 - y1

			// 绘制预览图片（如果已加载）
			if hPreviewBMP != 0 && selW > 0 && selH > 0 {
				memDC, _, _ := procCreateCompatibleDC.Call(hdc)
				oldBmp, _, _ := procSelectObject.Call(memDC, hPreviewBMP)

				var bmpInfo [6]uintptr
				ret, _, _ := procGetObject.Call(hPreviewBMP, 24, uintptr(unsafe.Pointer(&bmpInfo[0])))
				if ret != 0 {
					bmpW := int(bmpInfo[1])
					bmpH := int(bmpInfo[2])
					if bmpW > 0 && bmpH > 0 {
						procStretchBlt.Call(hdc, uintptr(x1), uintptr(y1), uintptr(selW), uintptr(selH),
							memDC, 0, 0, uintptr(bmpW), uintptr(bmpH), SRCCOPY)
					}
				}

				procSelectObject.Call(memDC, oldBmp)
				procDeleteDC.Call(memDC)
			}

			// 绘制边框
			pen, _, _ := procCreatePen.Call(PS_SOLID, 3, 0x00FFFFFF)
			oldPen, _, _ := procSelectObject.Call(hdc, pen)
			nullBrush, _, _ := procGetStockObject.Call(5)
			oldBrush, _, _ := procSelectObject.Call(hdc, nullBrush)
			procRectangle.Call(hdc, uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2))
			procSelectObject.Call(hdc, oldPen)
			procSelectObject.Call(hdc, oldBrush)
			procDeleteObject.Call(pen)

			// 尺寸文字
			sizeText := fmt.Sprintf("%d x %d", selW, selH)
			procSetBkMode.Call(hdc, TRANSPARENT)
			procSetTextColor.Call(hdc, 0x00000000)
			procTextOut.Call(hdc, uintptr(x1+5), uintptr(y1+5), uintptr(unsafe.Pointer(utf16Ptr(sizeText))))
		}

		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case WM_ERASEBKGND:
		return 1

	case WM_KEYDOWN:
		if wp == VK_ESCAPE {
			selState = -1
			procKillTimer.Call(hMainWnd, timerSel)
			closeSelWindow()
			setStatus("画布校准已取消")
			return 0
		}
		if wp == VK_RETURN && selState == 1 {
			procKillTimer.Call(hMainWnd, timerSel)
			closeSelWindow()
			finishSelection()
			return 0
		}
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wp, lp)
	return ret
}

func openSelWindow() {
	if !selRegistered {
		wc := WNDCLASSEXW{
			CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
			Style:         0, // 不用 CS_HREDRAW|CS_VREDRAW，减少重绘
			LpfnWndProc:   syscall.NewCallback(selWndProc),
			HInstance:     0,
			HbrBackground: 0,
			LpszClassName: utf16Ptr(selWndClass),
		}
		procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
		selRegistered = true
	}

	// 预加载预览位图
	if previewBMP != "" && hPreviewBMP == 0 {
		bmpPtr, _, _ := procLoadImageW.Call(0, uintptr(unsafe.Pointer(utf16Ptr(previewBMP))), IMAGE_BITMAP, 0, 0, LR_LOADFROMFILE|LR_CREATEDIBSECTION)
		if bmpPtr != 0 {
			hPreviewBMP = bmpPtr
		}
	}

	sxR, _, _ := procGetSystemMetrics.Call(0)
	syR, _, _ := procGetSystemMetrics.Call(1)

	hSelWnd, _, _ = procCreateWindowExW.Call(
		uintptr(WS_EX_LAYERED|WS_EX_TOPMOST),
		uintptr(unsafe.Pointer(utf16Ptr(selWndClass))),
		uintptr(unsafe.Pointer(utf16Ptr("框选绘画区域"))),
		uintptr(WS_POPUP|WS_VISIBLE),
		0, 0, uintptr(sxR), uintptr(syR),
		0, 0, 0, 0,
	)

	if hSelWnd == 0 {
		setStatus("创建选区窗口失败")
		return
	}

	procSetLayeredWindowAttributes.Call(hSelWnd, TRANSPARENT_COLOR, 0, LWA_COLORKEY)
	procShowWindow.Call(hSelWnd, SW_SHOW)
	procSetForegroundWindow.Call(hSelWnd)
}

func closeSelWindow() {
	if hSelWnd != 0 {
		procDestroyWindow.Call(hSelWnd)
		hSelWnd = 0
	}
	// 不删除 hPreviewBMP，保留给下次使用
}

func onSelTimer() {
	switch selState {
	case 0:
		if isKeyDown(VK_ESCAPE) {
			selState = -1
			procKillTimer.Call(hMainWnd, timerSel)
			closeSelWindow()
			setStatus("画布校准已取消")
			return
		}
		if ret, _, _ := procGetAsyncKeyState.Call(uintptr(VK_LBUTTON)); ret&0x8000 != 0 {
			x, y := getCursorPos()
			selStartPt = POINT{X: int32(x), Y: int32(y)}
			selEndPt = selStartPt
			selState = 1
		}

	case 1:
		if isKeyDown(VK_ESCAPE) {
			selState = -1
			procKillTimer.Call(hMainWnd, timerSel)
			closeSelWindow()
			setStatus("画布校准已取消")
			return
		}
		x, y := getCursorPos()
		// 只有坐标变化时才重绘
		if selEndPt.X != int32(x) || selEndPt.Y != int32(y) {
			selEndPt = POINT{X: int32(x), Y: int32(y)}
			if hSelWnd != 0 {
				procInvalidateRect.Call(hSelWnd, 0, 1) // bErase=TRUE
			}
		}
	}
}

func finishSelection() {
	tl := selStartPt
	br := selEndPt
	if tl.X > br.X {
		tl.X, br.X = br.X, tl.X
	}
	if tl.Y > br.Y {
		tl.Y, br.Y = br.Y, tl.Y
	}

	if br.X-tl.X < 10 || br.Y-tl.Y < 10 {
		setStatus("选区太小，请重新框选")
		return
	}

	canvasTL = Pt{int(tl.X), int(tl.Y)}
	canvasBR = Pt{int(br.X), int(br.Y)}
	calibrated = true

	setStatus(fmt.Sprintf("画布已校准: (%d,%d)-(%d,%d) | 大小: %dx%d",
		canvasTL.X, canvasTL.Y, canvasBR.X, canvasBR.Y,
		canvasBR.X-canvasTL.X, canvasBR.Y-canvasTL.Y))
}

func handleProcessResult(result ProcessResult) {
	if result.err != nil {
		setStatus(fmt.Sprintf("解析失败: %v", result.err))
		return
	}

	paths = result.paths
	imgW, imgH = result.imgW, result.imgH
	previewBMP = result.previewBMP

	// 删除旧的预览位图
	if hPreviewBMP != 0 {
		procDeleteObject.Call(hPreviewBMP)
		hPreviewBMP = 0
	}

	// 加载新的预览位图
	if previewBMP != "" {
		bmpPtr, _, _ := procLoadImageW.Call(0, uintptr(unsafe.Pointer(utf16Ptr(previewBMP))), IMAGE_BITMAP, 0, 0, LR_LOADFROMFILE|LR_CREATEDIBSECTION)
		if bmpPtr != 0 {
			hPreviewBMP = bmpPtr
		}
		// 更新主窗口预览
		bmpPtr2, _, _ := procLoadImageW.Call(0, uintptr(unsafe.Pointer(utf16Ptr(previewBMP))), IMAGE_BITMAP, 0, 0, LR_LOADFROMFILE|LR_CREATEDIBSECTION)
		if bmpPtr2 != 0 {
			procSendMessageW.Call(hPreview, STM_SETIMAGE, IMAGE_BITMAP, bmpPtr2)
		}
	}

	setStatus(fmt.Sprintf("解析完成，共 %d 条路径 | 图片: %dx%d", len(paths), imgW, imgH))
	setProgress(0, 1)
}

func wndProc(hwnd uintptr, msg uint32, wp, lp uintptr) uintptr {
	switch msg {
	case WM_TIMER:
		switch wp {
		case timerSel:
			onSelTimer()
		case timerProcess:
			select {
			case result := <-processResultChan:
				handleProcessResult(result)
				procKillTimer.Call(hwnd, timerProcess)
			default:
			}
		}

	case WM_COMMAND:
		switch int(wp & 0xFFFF) {
		case 1:
			filter := "Image" + string(rune(0)) + "*.png;*.jpg;*.jpeg;*.bmp;*.gif;*.webp" + string(rune(0)) + "All" + string(rune(0)) + "*.*" + string(rune(0))
			if path := openFileDlg("选择图片", filter); path != "" {
				imagePath = path
				setStatus("已选择: " + path)

				blur := parseInt(getEditText(hEditBlur), 5)
				if blur%2 == 0 {
					blur++
				}
				thresh := parseInt(getEditText(hEditThreshold), 100)
				minLen := parseInt(getEditText(hEditMinLen), 30)
				simplify := parseFloat(getEditText(hEditSimplify), 2.0)

				previewPath := os.TempDir() + "\\draw_preview.bmp"

				setStatus("正在解析图片...")

				go doProcessAsync(imagePath, previewPath, blur, thresh, minLen, simplify)
				procSetTimer.Call(hwnd, timerProcess, 100, 0)
			}
		case 2:
			if imagePath == "" {
				msgBox("提示", "请先选择图片", MB_OK|MB_ICONWARNING)
				return 0
			}
			blur := parseInt(getEditText(hEditBlur), 5)
			if blur%2 == 0 {
				blur++
			}
			thresh := parseInt(getEditText(hEditThreshold), 100)
			minLen := parseInt(getEditText(hEditMinLen), 30)
			simplify := parseFloat(getEditText(hEditSimplify), 2.0)

			previewPath := os.TempDir() + "\\draw_preview.bmp"

			setStatus("正在解析图片...")

			go doProcessAsync(imagePath, previewPath, blur, thresh, minLen, simplify)
			procSetTimer.Call(hwnd, timerProcess, 100, 0)

		case 3:
			if len(paths) == 0 {
				msgBox("提示", "请先解析图片", MB_OK|MB_ICONWARNING)
				return 0
			}
			selState = 0
			openSelWindow()
			procSetTimer.Call(hwnd, timerSel, 50, 0)
			setStatus("拖动鼠标框选绘画区域 | 按 Enter 确认 | 按 Esc 取消")
		case 4:
			doStart()
		case 5:
			if drawState == 1 {
				drawState = 0
				mouseUp()
				setStatus("绘画已停止")
			}
		}

	case WM_DESTROY:
		procKillTimer.Call(hwnd, timerSel)
		procKillTimer.Call(hwnd, timerProcess)
		if hPreviewBMP != 0 {
			procDeleteObject.Call(hPreviewBMP)
			hPreviewBMP = 0
		}
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wp, lp)
	return ret
}

func createWnd(exStyle uint32, className, title string, style uint32, x, y, w, h int, parent, menu, inst uintptr) uintptr {
	ret, _, _ := procCreateWindowExW.Call(uintptr(exStyle), uintptr(unsafe.Pointer(utf16Ptr(className))), uintptr(unsafe.Pointer(utf16Ptr(title))), uintptr(style), uintptr(x), uintptr(y), uintptr(w), uintptr(h), parent, menu, inst, 0)
	return ret
}

func createBtn(text string, x, y, w, h int, parent uintptr, id int) uintptr {
	return createWnd(0, "BUTTON", text, BS_PUSHBUTTON|WS_TABSTOP|WS_VISIBLE|WS_CHILD, x, y, w, h, parent, uintptr(id), 0)
}

func createEdit(text string, x, y, w, h int, parent uintptr, id int) uintptr {
	return createWnd(0, "EDIT", text, WS_VISIBLE|WS_CHILD|WS_BORDER|ES_AUTOHSCROLL|ES_CENTER, x, y, w, h, parent, uintptr(id), 0)
}

func createLabel(text string, x, y, w, h int, parent uintptr) uintptr {
	return createWnd(0, "STATIC", text, WS_VISIBLE|WS_CHILD, x, y, w, h, parent, 0, 0)
}

func createGroup(text string, x, y, w, h int, parent uintptr) uintptr {
	return createWnd(0, "BUTTON", text, BS_GROUPBOX|WS_VISIBLE|WS_CHILD, x, y, w, h, parent, 0, 0)
}

func setFont(h uintptr) {
	if hFont != 0 {
		procSendMessageW.Call(h, WM_SETFONT, uintptr(hFont), 1)
	}
}

func WinMain(hInst uintptr) int {
	processResultChan = make(chan ProcessResult, 1)

	className := utf16Ptr("DrawAssistant_v12")
	hFont, _, _ = procCreateFontW.Call(14, 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 0, DEFAULT_CHARSET, OUTLINE_FONTTYPE, CLIP_DEFAULT_PRECIS, DEFAULT_QUALITY, DEFAULT_PITCH|FF_DONTCARE, 0)

	wc := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		Style:         CS_HREDRAW | CS_VREDRAW,
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     hInst,
		HbrBackground: COLOR_WINDOW + 1,
		LpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	hMainWnd = createWnd(0, "DrawAssistant_v12", "你画我猜 - 辅助绘画工具", WS_OVERLAPPEDWINDOW, CW_USEDEFAULT, CW_USEDEFAULT, 800, 560, 0, 0, hInst)

	createGroup("参数设置", 10, 10, 240, 290, hMainWnd)
	createLabel("边缘阈值 (0-255):", 22, 38, 110, 20, hMainWnd)
	hEditThreshold = createEdit("100", 140, 36, 95, 24, hMainWnd, 101)
	createLabel("模糊程度 (奇数):", 22, 68, 110, 20, hMainWnd)
	hEditBlur = createEdit("5", 140, 66, 95, 24, hMainWnd, 102)
	createLabel("最小轮廓长度:", 22, 98, 110, 20, hMainWnd)
	hEditMinLen = createEdit("30", 140, 96, 95, 24, hMainWnd, 103)
	createLabel("简化精度:", 22, 128, 110, 20, hMainWnd)
	hEditSimplify = createEdit("2.0", 140, 126, 95, 24, hMainWnd, 104)
	createLabel("绘画速度 (秒/点):", 22, 158, 110, 20, hMainWnd)
	hEditSpeed = createEdit("0.01", 140, 156, 95, 24, hMainWnd, 105)
	createLabel("线条间隔 (秒):", 22, 188, 110, 20, hMainWnd)
	hEditLinePause = createEdit("0.1", 140, 186, 95, 24, hMainWnd, 106)

	createGroup("操作", 260, 10, 520, 90, hMainWnd)
	createBtn("选择图片", 275, 38, 110, 35, hMainWnd, 1)
	createBtn("解析图片", 395, 38, 110, 35, hMainWnd, 2)
	createBtn("框选画布", 515, 38, 110, 35, hMainWnd, 3)
	createBtn("开始绘画", 635, 38, 110, 35, hMainWnd, 4)
	createBtn("停止", 755, 38, 70, 35, hMainWnd, 5)

	createGroup("预览", 260, 110, 520, 330, hMainWnd)
	hPreview = createWnd(0, "STATIC", "", SS_BITMAP|SS_CENTERIMAGE|WS_VISIBLE|WS_CHILD|WS_BORDER, 275, 135, 490, 290, hMainWnd, 0, 0)

	hProgress = createWnd(0, "msctls_progress32", "", WS_VISIBLE|WS_CHILD, 275, 450, 490, 22, hMainWnd, 0, 0)

	createGroup("状态", 10, 465, 240, 50, hMainWnd)
	hEditStatus = createEdit("就绪 - 请选择图片", 22, 485, 215, 24, hMainWnd, 301)
	createLabel("快捷键: +加速 | -减速 | ESC停止", 260, 480, 520, 20, hMainWnd)

	for _, h := range []uintptr{hMainWnd, hEditThreshold, hEditBlur, hEditMinLen, hEditSimplify, hEditSpeed, hEditLinePause, hEditStatus, hProgress, hPreview} {
		setFont(h)
	}

	procShowWindow.Call(hMainWnd, SW_SHOW)
	procUpdateWindow.Call(hMainWnd)

	var msg MSG
	for {
		if ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0); ret == 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
	return int(msg.WParam)
}

func main() {
	image.RegisterFormat("png", "\x89PNG\r\n\x1a\n", png.Decode, png.DecodeConfig)
	image.RegisterFormat("jpeg", "\xff\xd8", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("gif", "GIF8", gif.Decode, gif.DecodeConfig)
	image.RegisterFormat("bmp", "BM", decodeBMP, decodeBMPConfig)
	hInst, _, _ := procGetModuleHandleW.Call(0)
	os.Exit(WinMain(hInst))
}

func decodeBMP(r io.Reader) (image.Image, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(data) < 54 {
		return nil, fmt.Errorf("BMP file too small")
	}
	if data[0] != 'B' || data[1] != 'M' {
		return nil, fmt.Errorf("not a BMP file")
	}

	offset := binary.LittleEndian.Uint32(data[10:14])
	width := int32(binary.LittleEndian.Uint32(data[18:22]))
	height := int32(binary.LittleEndian.Uint32(data[22:26]))
	bpp := binary.LittleEndian.Uint16(data[28:30])

	if bpp != 24 && bpp != 32 {
		return nil, fmt.Errorf("unsupported BMP bit depth: %d", bpp)
	}

	topDown := false
	if height < 0 {
		height = -height
		topDown = true
	}

	w, h := int(width), int(height)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rowSize := ((w*int(bpp) + 31) / 32) * 4

	for y := 0; y < h; y++ {
		bmpY := y
		if !topDown {
			bmpY = h - 1 - y
		}
		rowOffset := int(offset) + bmpY*rowSize
		for x := 0; x < w; x++ {
			pixelOffset := rowOffset + x*(int(bpp)/8)
			if pixelOffset+2 >= len(data) {
				continue
			}
			b := data[pixelOffset]
			g := data[pixelOffset+1]
			rv := data[pixelOffset+2]
			a := byte(255)
			if bpp == 32 && pixelOffset+3 < len(data) {
				a = data[pixelOffset+3]
			}
			img.SetRGBA(x, y, color.RGBA{rv, g, b, a})
		}
	}
	return img, nil
}

func decodeBMPConfig(r io.Reader) (image.Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return image.Config{}, err
	}
	if len(data) < 26 {
		return image.Config{}, fmt.Errorf("BMP file too small")
	}
	width := binary.LittleEndian.Uint32(data[18:22])
	height := binary.LittleEndian.Uint32(data[22:26])
	if height > 0x7FFFFFFF {
		height = uint32(-int32(height))
	}
	return image.Config{Width: int(width), Height: int(height)}, nil
}

func webpDecode(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	return img, err
}
