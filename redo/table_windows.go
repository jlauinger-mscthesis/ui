// 28 july 2014

package ui

import (
	"fmt"
	"unsafe"
	"reflect"
)

// #include "winapi_windows.h"
import "C"

type table struct {
	*tablebase
	_hwnd		C.HWND
	noautosize	bool
	colcount		C.int
}

func finishNewTable(b *tablebase, ty reflect.Type) Table {
	t := &table{
		_hwnd:		C.newControl(C.xWC_LISTVIEW,
			C.LVS_REPORT | C.LVS_OWNERDATA | C.LVS_NOSORTHEADER | C.LVS_SHOWSELALWAYS | C.WS_HSCROLL | C.WS_VSCROLL | C.WS_TABSTOP,
			C.WS_EX_CLIENTEDGE),		// WS_EX_CLIENTEDGE without WS_BORDER will show the canonical visual styles border (thanks to MindChild in irc.efnet.net/#winprog)
		tablebase:		b,
	}
	C.setTableSubclass(t._hwnd, unsafe.Pointer(t))
	// LVS_EX_FULLROWSELECT gives us selection across the whole row, not just the leftmost column; this makes the list view work like on other platforms
	// LVS_EX_SUBITEMIMAGES gives us images in subitems, which will be important when both images and checkboxes are added
	C.tableAddExtendedStyles(t._hwnd, C.LVS_EX_FULLROWSELECT | C.LVS_EX_SUBITEMIMAGES)
	for i := 0; i < ty.NumField(); i++ {
		C.tableAppendColumn(t._hwnd, C.int(i), toUTF16(ty.Field(i).Name))
	}
	t.colcount = C.int(ty.NumField())
	return t
}

func (t *table) Unlock() {
	t.unlock()
	// there's a possibility that user actions can happen at this point, before the view is updated
	// alas, this is something we have to deal with, because Unlock() can be called from any thread
	go func() {
		Do(func() {
			t.RLock()
			defer t.RUnlock()
			C.tableUpdate(t._hwnd, C.int(reflect.Indirect(reflect.ValueOf(t.data)).Len()))
		})
	}()
}

//export tableGetCellText
func tableGetCellText(data unsafe.Pointer, row C.int, col C.int, str *C.LPWSTR) {
	t := (*table)(data)
	t.RLock()
	defer t.RUnlock()
	d := reflect.Indirect(reflect.ValueOf(t.data))
	datum := d.Index(int(row)).Field(int(col))
	s := fmt.Sprintf("%v", datum)
	*str = toUTF16(s)
}

// the column autoresize policy is simple:
// on every table.commitResize() call, if the columns have not been resized by the user, autoresize
func (t *table) autoresize() {
	t.RLock()
	defer t.RUnlock()
	if !t.noautosize {
		C.tableAutosizeColumns(t._hwnd, t.colcount)
	}
}

//export tableStopColumnAutosize
func tableStopColumnAutosize(data unsafe.Pointer) {
	t := (*table)(data)
	t.noautosize = true
}

//export tableColumnCount
func tableColumnCount(data unsafe.Pointer) C.int {
	t := (*table)(data)
	return t.colcount
}

func (t *table) hwnd() C.HWND {
	return t._hwnd
}

func (t *table) setParent(p *controlParent) {
	basesetParent(t, p)
}

func (t *table) allocate(x int, y int, width int, height int, d *sizing) []*allocation {
	return baseallocate(t, x, y, width, height, d)
}

const (
	// from C++ Template 05 in http://msdn.microsoft.com/en-us/library/windows/desktop/bb226818%28v=vs.85%29.aspx as this is the best I can do for now
	// there IS a message LVM_APPROXIMATEVIEWRECT that can do calculations, but it doesn't seem to work right when asked to base its calculations on the current width/height on Windows and wine...
	tableWidth = 183
	tableHeight = 50
)

func (t *table) preferredSize(d *sizing) (width, height int) {
	return fromdlgunitsX(tableWidth, d), fromdlgunitsY(tableHeight, d)
}

func (t *table) commitResize(a *allocation, d *sizing) {
	basecommitResize(t, a, d)
	t.RLock()
	defer t.RUnlock()
	t.autoresize()
}

func (t *table) getAuxResizeInfo(d *sizing) {
	basegetAuxResizeInfo(t, d)
}
