// 4 august 2014

package ui

import (
	"fmt"
	"syscall"
	"unsafe"
)

// #include "winapi_windows.h"
import "C"

type container struct {
	*controlSingleHWND
}

type sizing struct {
	sizingbase

	// for size calculations
	baseX           C.int
	baseY           C.int
	internalLeading C.LONG // for Label; see Label.commitResize() for details

	// for the actual resizing
	// possibly the HDWP
}

func makeContainerWindowClass() error {
	var errmsg *C.char

	err := C.makeContainerWindowClass(&errmsg)
	if err != 0 || errmsg != nil {
		return fmt.Errorf("%s: %v", C.GoString(errmsg), syscall.Errno(err))
	}
	return nil
}

func newContainer() *container {
	c := new(container)
	hwnd := C.newContainer(unsafe.Pointer(c))
	if hwnd != c.hwnd {
		panic(fmt.Errorf("inconsistency: hwnd returned by CreateWindowEx() (%p) and hwnd stored in container (%p) differ", hwnd, c.hwnd))
	}
	// don't set preferredSize(); it should never be called
	return c
}

// TODO merge with controlSingleHWND
func (c *container) show() {
	C.ShowWindow(c.hwnd, C.SW_SHOW)
}

// TODO merge with controlSingleHWND
func (c *container) hide() {
	C.ShowWindow(c.hwnd, C.SW_HIDE)
}

//export storeContainerHWND
func storeContainerHWND(data unsafe.Pointer, hwnd C.HWND) {
	c := (*container)(data)
	c.hwnd = hwnd
}

// For Windows, Microsoft just hands you a list of preferred control sizes as part of the MSDN documentation and tells you to roll with it.
// These sizes are given in "dialog units", which are independent of the font in use.
// We need to convert these into standard pixels, which requires we get the device context of the OS window.
// References:
// - http://msdn.microsoft.com/en-us/library/ms645502%28VS.85%29.aspx - the calculation needed
// - http://support.microsoft.com/kb/125681 - to get the base X and Y
// (thanks to http://stackoverflow.com/questions/58620/default-button-size)
// In my tests (see https://github.com/andlabs/windlgunits), the GetTextExtentPoint32() option for getting the base X produces much more accurate results than the tmAveCharWidth option when tested against the sample values given in http://msdn.microsoft.com/en-us/library/windows/desktop/dn742486.aspx#sizingandspacing, but can be off by a pixel in either direction (probably due to rounding errors).

// note on MulDiv():
// div will not be 0 in the usages below
// we also ignore overflow; that isn't likely to happen for our use case anytime soon

func fromdlgunitsX(du int, d *sizing) int {
	return int(C.MulDiv(C.int(du), d.baseX, 4))
}

func fromdlgunitsY(du int, d *sizing) int {
	return int(C.MulDiv(C.int(du), d.baseY, 8))
}

const (
	// TODO figure out how to sort this more nicely
	marginDialogUnits  = 7
	paddingDialogUnits = 4

	// TODO move to group
	groupXMargin       = 6
	groupYMarginTop    = 11 // note this value /includes the groupbox label/
	groupYMarginBottom = 7
)

func (w *window) beginResize() (d *sizing) {
	var baseX, baseY C.int
	var internalLeading C.LONG

	d = new(sizing)

	C.calculateBaseUnits(c.hwnd, &baseX, &baseY, &internalLeading)
	d.baseX = baseX
	d.baseY = baseY
	d.internalLeading = internalLeading

	d.xmargin = fromdlgunitsX(marginDialogUnits, d)
	d.ymargintop = fromdlgunitsY(marginDialogUnits, d)
	d.ymarginbottom = d.ymargintop
	d.xpadding = fromdlgunitsX(paddingDialogUnits, d)
	d.ypadding = fromdlgunitsY(paddingDialogUnits, d)

/*TODO
	if c.isGroup {
		// note that these values apply regardless of whether or not spaced is set
		// this is because Windows groupboxes have the client rect spanning the entire size of the control, not just the active work area
		// the measurements Microsoft give us are for spaced margining; let's just use them
		d.xmargin = fromdlgunitsX(groupXMargin, d)
		d.ymargintop = fromdlgunitsY(groupYMarginTop, d)
		d.ymarginbottom = fromdlgunitsY(groupYMarginBottom, d)

	}
*/

	return d
}
