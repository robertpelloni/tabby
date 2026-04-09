// Package ui provides Go bindings for the BTK native UI toolkit.
//
// It exposes a Go-friendly API for creating native desktop applications
// with windows, tabs, menus, toolbars, terminal widgets, and more.
//
// This package uses CGo to bridge Go code with the BTK C++ framework
// via a flat C API defined in bridge.h.
//
// Usage:
//
//	app := ui.NewApp()
//	defer app.Destroy()
//
//	win := app.NewWindow()
//	win.SetTitle("Tabby")
//	win.SetSize(1200, 800)
//	win.Show()
//
//	app.Run()
package ui

/*
#cgo CXXFLAGS: -std=c++17 -I${SRCDIR}
#cgo windows LDFLAGS: -lgdi32 -luser32 -lkernel32 -lshell32 -lole32 -lcomdlg32
#cgo linux LDFLAGS: -lX11 -lXext -lGL -lpthread -ldl
#cgo darwin LDFLAGS: -framework Cocoa -framework Carbon -framework CoreFoundation -framework IOKit

#include "bridge.h"
#include <stdlib.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

// Orientation for splitters
type Orientation int

const (
	Horizontal Orientation = 0
	Vertical   Orientation = 1
)

// EchoMode for line edits
type EchoMode int

const (
	EchoNormal       EchoMode = 0
	EchoPassword      EchoMode = 1
	EchoNoEcho       EchoMode = 2
	EchoPasswordEdit EchoMode = 3
)

// Callback types
type CloseCallback func()
type SizeCallback func(width, height int)
type DataCallback func(data string)
type MenuCallback func()

// ---- Application ----

// App represents the BTK application instance
type App struct {
	handle C.TabbyApp
}

// NewApp creates a new BTK application instance
func NewApp() *App {
	// CGo requires C argv conversion
	argc := C.int(1)
	argv := [1]*C.char{C.CString("tabby")}
	defer C.free(unsafe.Pointer(argv[0]))

	app := &App{
		handle: C.TabbyApp_Create(argc, (*C.char)(unsafe.Pointer(&argv[0]))),
	}
	return app
}

// Run starts the application event loop. Blocks until exit.
func (a *App) Run() int {
	return int(C.TabbyApp_Run(a.handle))
}

// Quit requests the application to exit
func (a *App) Quit() {
	C.TabbyApp_Quit(a.handle)
}

// Destroy cleans up the application
func (a *App) Destroy() {
	C.TabbyApp_Destroy(a.handle)
}

// SetStyle sets the application style
func (a *App) SetStyle(style string) {
	cs := C.CString(style)
	defer C.free(unsafe.Pointer(cs))
	C.TabbyApp_SetStyle(a.handle, cs)
}

// ProcessEvents processes pending events
func (a *App) ProcessEvents() {
	C.TabbyApp_ProcessEvents(a.handle)
}

// SetDarkMode enables or disables dark mode
func (a *App) SetDarkMode(dark bool) {
	v := 0
	if dark {
		v = 1
	}
	C.TabbyApp_SetDarkMode(a.handle, C.int(v))
}

// SetFont sets the default application font
func (a *App) SetFont(family string, size int) {
	cs := C.CString(family)
	defer C.free(unsafe.Pointer(cs))
	C.TabbyApp_SetFont(a.handle, cs, C.int(size))
}

// SettingsPath returns the platform-specific settings directory
func (a *App) SettingsPath() string {
	cs := C.TabbyApp_SettingsPath()
	defer C.Tabby_FreeString(cs)
	return C.GoString(cs)
}

// ---- Window ----

// Window represents a main application window
type Window struct {
	handle      C.TabbyWindow
	closeCb     CloseCallback
	resizeCb    SizeCallback
}

// NewWindow creates a new main window
func (a *App) NewWindow() *Window {
	return &Window{
		handle: C.TabbyWindow_Create(a.handle),
	}
}

// Destroy destroys the window
func (w *Window) Destroy() {
	C.TabbyWindow_Destroy(w.handle)
}

// Show shows the window
func (w *Window) Show() {
	C.TabbyWindow_Show(w.handle)
}

// Hide hides the window
func (w *Window) Hide() {
	C.TabbyWindow_Hide(w.handle)
}

// SetTitle sets the window title
func (w *Window) SetTitle(title string) {
	cs := C.CString(title)
	defer C.free(unsafe.Pointer(cs))
	C.TabbyWindow_SetTitle(w.handle, cs)
}

// SetSize sets the window size
func (w *Window) SetSize(width, height int) {
	C.TabbyWindow_SetSize(w.handle, C.int(width), C.int(height))
}

// SetMinimumSize sets the minimum window size
func (w *Window) SetMinimumSize(width, height int) {
	C.TabbyWindow_SetMinimumSize(w.handle, C.int(width), C.int(height))
}

// GetSize returns the current window size
func (w *Window) GetSize() (int, int) {
	var width, height C.int
	C.TabbyWindow_GetSize(w.handle, &width, &height)
	return int(width), int(height)
}

// SetPosition sets the window position
func (w *Window) SetPosition(x, y int) {
	C.TabbyWindow_SetPosition(w.handle, C.int(x), C.int(y))
}

// Maximize maximizes the window
func (w *Window) Maximize() {
	C.TabbyWindow_Maximize(w.handle)
}

// Minimize minimizes the window
func (w *Window) Minimize() {
	C.TabbyWindow_Minimize(w.handle)
}

// Restore restores the window
func (w *Window) Restore() {
	C.TabbyWindow_Restore(w.handle)
}

// SetFullScreen toggles full screen mode
func (w *Window) SetFullScreen(full bool) {
	v := 0
	if full {
		v = 1
	}
	C.TabbyWindow_SetFullScreen(w.handle, C.int(v))
}

// Close closes the window
func (w *Window) Close() {
	C.TabbyWindow_Close(w.handle)
}

// OnClose sets the callback for when the window is about to close
func (w *Window) OnClose(cb CloseCallback) {
	w.closeCb = cb
	// Store Go callback in a map to prevent GC
	storeWindowCallback(w)
	C.TabbyWindow_OnClose(w.handle,
		(C.TabbyCloseCallback)(unsafe.Pointer(C.tabbyCloseCallback)),
		unsafe.Pointer(w))
}

// OnResize sets the callback for when the window is resized
func (w *Window) OnResize(cb SizeCallback) {
	w.resizeCb = cb
	storeWindowCallback(w)
	C.TabbyWindow_OnResize(w.handle,
		(C.TabbySizeCallback)(unsafe.Pointer(C.tabbySizeCallback)),
		unsafe.Pointer(w))
}

// SetCentralWidget sets the central widget of the window
func (w *Window) SetCentralWidget(widget Widget) {
	C.TabbyWindow_SetCentralWidget(w.handle, widget.handle())
}

// MenuBar returns the window's menu bar (creates if needed)
func (w *Window) MenuBar() *MenuBar {
	return &MenuBar{handle: C.TabbyWindow_MenuBar(w.handle), window: w}
}

// StatusBar returns the window's status bar
func (w *Window) StatusBar() *StatusBar {
	return &StatusBar{handle: C.TabbyWindow_StatusBar(w.handle)}
}

// AddToolBar adds a toolbar to the window
func (w *Window) AddToolBar(title string) *ToolBar {
	cs := C.CString(title)
	defer C.free(unsafe.Pointer(cs))
	return &ToolBar{handle: C.TabbyWindow_AddToolBar(w.handle, cs)}
}

// ---- MenuBar ----

// MenuBar represents a menu bar
type MenuBar struct {
	handle C.TabbyMenuBar
	window *Window
}

// AddMenu adds a menu to the menu bar
func (mb *MenuBar) AddMenu(title string) *Menu {
	cs := C.CString(title)
	defer C.free(unsafe.Pointer(cs))
	return &Menu{handle: C.TabbyMenuBar_AddMenu(mb.handle, cs)}
}

// ---- Menu ----

// Menu represents a dropdown menu
type Menu struct {
	handle C.TabbyMenu
}

// AddAction adds an action to the menu
func (m *Menu) AddAction(text string) *Action {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	return &Action{handle: C.TabbyMenu_AddAction(m.handle, cs)}
}

// AddSeparator adds a separator to the menu
func (m *Menu) AddSeparator() {
	C.TabbyMenu_AddSeparator(m.handle)
}

// ---- Action ----

// Action represents a menu/toolbar action
type Action struct {
	handle C.TabbyAction
	cb     MenuCallback
}

// SetShortcut sets the keyboard shortcut
func (a *Action) SetShortcut(shortcut string) {
	cs := C.CString(shortcut)
	defer C.free(unsafe.Pointer(cs))
	C.TabbyAction_SetShortcut(a.handle, cs)
}

// SetCheckable makes the action checkable
func (a *Action) SetCheckable(checkable bool) {
	v := 0
	if checkable {
		v = 1
	}
	C.TabbyAction_SetCheckable(a.handle, C.int(v))
}

// SetChecked sets the checked state
func (a *Action) SetChecked(checked bool) {
	v := 0
	if checked {
		v = 1
	}
	C.TabbyAction_SetChecked(a.handle, C.int(v))
}

// IsChecked returns whether the action is checked
func (a *Action) IsChecked() bool {
	return C.TabbyAction_IsChecked(a.handle) != 0
}

// SetEnabled sets whether the action is enabled
func (a *Action) SetEnabled(enabled bool) {
	v := 0
	if enabled {
		v = 1
	}
	C.TabbyAction_SetEnabled(a.handle, C.int(v))
}

// SetToolTip sets the tooltip text
func (a *Action) SetToolTip(tip string) {
	cs := C.CString(tip)
	defer C.free(unsafe.Pointer(cs))
	C.TabbyAction_SetToolTip(a.handle, cs)
}

// OnTriggered sets the callback for when the action is triggered
func (a *Action) OnTriggered(cb MenuCallback) {
	a.cb = cb
	storeActionCallback(a)
	C.TabbyAction_OnTriggered(a.handle,
		(C.TabbyMenuCallback)(unsafe.Pointer(C.tabbyMenuCallback)),
		unsafe.Pointer(a))
}

// ---- ToolBar ----

// ToolBar represents a toolbar
type ToolBar struct {
	handle C.TabbyToolBar
}

// AddAction adds an action to the toolbar
func (t *ToolBar) AddAction(action *Action) {
	C.TabbyToolBar_AddAction(t.handle, action.handle)
}

// AddSeparator adds a separator
func (t *ToolBar) AddSeparator() {
	C.TabbyToolBar_AddSeparator(t.handle)
}

// ---- StatusBar ----

// StatusBar represents a status bar
type StatusBar struct {
	handle C.TabbyStatusBar
}

// ShowMessage shows a temporary message in the status bar
func (s *StatusBar) ShowMessage(message string, timeoutMs int) {
	cs := C.CString(message)
	defer C.free(unsafe.Pointer(cs))
	C.TabbyStatusBar_ShowMessage(s.handle, cs, C.int(timeoutMs))
}

// ---- TabWidget ----

// TabWidget represents a tab widget
type TabWidget struct {
	handle C.TabbyTabWidget
}

// NewTabWidget creates a new tab widget
func NewTabWidget() *TabWidget {
	return &TabWidget{handle: C.TabbyTabWidget_Create()}
}

// Destroy destroys the tab widget
func (t *TabWidget) Destroy() {
	C.TabbyTabWidget_Destroy(t.handle)
}

// AddTab adds a widget as a new tab
func (t *TabWidget) AddTab(widget Widget, label string) int {
	cs := C.CString(label)
	defer C.free(unsafe.Pointer(cs))
	return int(C.TabbyTabWidget_AddTab(t.handle, widget.handle(), cs))
}

// RemoveTab removes a tab by index
func (t *TabWidget) RemoveTab(index int) {
	C.TabbyTabWidget_RemoveTab(t.handle, C.int(index))
}

// SetTabText sets the text of a tab
func (t *TabWidget) SetTabText(index int, text string) {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	C.TabbyTabWidget_SetTabText(t.handle, C.int(index), cs)
}

// SetCurrentIndex switches to a tab by index
func (t *TabWidget) SetCurrentIndex(index int) {
	C.TabbyTabWidget_SetCurrentIndex(t.handle, C.int(index))
}

// CurrentIndex returns the current tab index
func (t *TabWidget) CurrentIndex() int {
	return int(C.TabbyTabWidget_CurrentIndex(t.handle))
}

// Count returns the number of tabs
func (t *TabWidget) Count() int {
	return int(C.TabbyTabWidget_Count(t.handle))
}

// SetTabsClosable enables close buttons on tabs
func (t *TabWidget) SetTabsClosable(closable bool) {
	v := 0
	if closable {
		v = 1
	}
	C.TabbyTabWidget_SetTabsClosable(t.handle, C.int(v))
}

// SetMovable enables tab reordering
func (t *TabWidget) SetMovable(movable bool) {
	v := 0
	if movable {
		v = 1
	}
	C.TabbyTabWidget_SetMovable(t.handle, C.int(v))
}

// ---- Splitter ----

// Splitter represents a splitter widget
type Splitter struct {
	handle C.TabbySplitter
}

// NewSplitter creates a new splitter
func NewSplitter(orientation Orientation) *Splitter {
	return &Splitter{handle: C.TabbySplitter_Create(C.int(orientation))}
}

// Destroy destroys the splitter
func (s *Splitter) Destroy() {
	C.TabbySplitter_Destroy(s.handle)
}

// AddWidget adds a widget to the splitter
func (s *Splitter) AddWidget(widget Widget) {
	C.TabbySplitter_AddWidget(s.handle, widget.handle())
}

// SetSizes sets the sizes of the splitter's children
func (s *Splitter) SetSizes(sizes []int) {
	csizes := make([]C.int, len(sizes))
	for i, v := range sizes {
		csizes[i] = C.int(v)
	}
	C.TabbySplitter_SetSizes(s.handle, (*C.int)(unsafe.Pointer(&csizes[0])), C.int(len(sizes)))
}

// GetSizes returns the sizes of the splitter's children
func (s *Splitter) GetSizes() []int {
	var count C.int
	var sizes [32]C.int
	C.TabbySplitter_GetSizes(s.handle, (*C.int)(unsafe.Pointer(&sizes[0])), &count)
	result := make([]int, int(count))
	for i := 0; i < int(count); i++ {
		result[i] = int(sizes[i])
	}
	return result
}

// ---- Terminal ----

// Terminal represents a terminal emulator widget
type Terminal struct {
	handle  C.TabbyTerminalWidget
	inputCb DataCallback
	sizeCb  SizeCallback
	titleCb DataCallback
}

// NewTerminal creates a new terminal widget
func NewTerminal() *Terminal {
	return &Terminal{handle: C.TabbyTerminalWidget_Create()}
}

// Destroy destroys the terminal
func (t *Terminal) Destroy() {
	C.TabbyTerminalWidget_Destroy(t.handle)
}

// SetSize sets the terminal size in columns/rows
func (t *Terminal) SetSize(cols, rows int) {
	C.TabbyTerminalWidget_SetSize(t.handle, C.int(cols), C.int(rows))
}

// GetSize returns the terminal size in columns/rows
func (t *Terminal) GetSize() (int, int) {
	var cols, rows C.int
	C.TabbyTerminalWidget_GetSize(t.handle, &cols, &rows)
	return int(cols), int(rows)
}

// Write writes data to the terminal display
func (t *Terminal) Write(data []byte) {
	if len(data) > 0 {
		C.TabbyTerminalWidget_Write(t.handle, (*C.char)(unsafe.Pointer(&data[0])), C.int(len(data)))
	}
}

// WriteString writes a string to the terminal display
func (t *Terminal) WriteString(data string) {
	if len(data) > 0 {
		C.TabbyTerminalWidget_Write(t.handle, C.CString(data), C.int(len(data)))
	}
}

// SetColorScheme sets the terminal color scheme (JSON)
func (t *Terminal) SetColorScheme(json string) {
	cs := C.CString(json)
	defer C.free(unsafe.Pointer(cs))
	C.TabbyTerminalWidget_SetColorScheme(t.handle, cs)
}

// SetFont sets the terminal font
func (t *Terminal) SetFont(family string, size int) {
	cs := C.CString(family)
	defer C.free(unsafe.Pointer(cs))
	C.TabbyTerminalWidget_SetFont(t.handle, cs, C.int(size))
}

// SetScrollback sets the scrollback buffer size
func (t *Terminal) SetScrollback(lines int) {
	C.TabbyTerminalWidget_SetScrollback(t.handle, C.int(lines))
}

// Clear clears the terminal
func (t *Terminal) Clear() {
	C.TabbyTerminalWidget_Clear(t.handle)
}

// SetFocus gives keyboard focus to the terminal
func (t *Terminal) SetFocus() {
	C.TabbyTerminalWidget_SetFocus(t.handle)
}

// HasFocus returns whether the terminal has focus
func (t *Terminal) HasFocus() bool {
	return C.TabbyTerminalWidget_HasFocus(t.handle) != 0
}

// ---- Basic Widgets ----

// Label represents a text label
type Label struct {
	handle C.TabbyLabel
}

// NewLabel creates a new label
func NewLabel(text string) *Label {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	return &Label{handle: C.TabbyLabel_Create(cs)}
}

// SetText sets the label text
func (l *Label) SetText(text string) {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	C.TabbyLabel_SetText(l.handle, cs)
}

// LineEdit represents a single-line text input
type LineEdit struct {
	handle C.TabbyLineEdit
}

// NewLineEdit creates a new line edit
func NewLineEdit() *LineEdit {
	return &LineEdit{handle: C.TabbyLineEdit_Create()}
}

// SetText sets the text
func (l *LineEdit) SetText(text string) {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	C.TabbyLineEdit_SetText(l.handle, cs)
}

// Text returns the current text
func (l *LineEdit) Text() string {
	cs := C.TabbyLineEdit_Text(l.handle)
	defer C.Tabby_FreeString(cs)
	return C.GoString(cs)
}

// SetPlaceholder sets placeholder text
func (l *LineEdit) SetPlaceholder(text string) {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	C.TabbyLineEdit_SetPlaceholder(l.handle, cs)
}

// SetEchoMode sets the echo mode
func (l *LineEdit) SetEchoMode(mode EchoMode) {
	C.TabbyLineEdit_SetEchoMode(l.handle, C.int(mode))
}

// Button represents a push button
type Button struct {
	handle C.TabbyButton
	cb     MenuCallback
}

// NewButton creates a new button
func NewButton(text string) *Button {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	return &Button{handle: C.TabbyButton_Create(cs)}
}

// SetText sets the button text
func (b *Button) SetText(text string) {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	C.TabbyButton_SetText(b.handle, cs)
}

// SetCheckable makes the button checkable
func (b *Button) SetCheckable(checkable bool) {
	v := 0
	if checkable {
		v = 1
	}
	C.TabbyButton_SetCheckable(b.handle, C.int(v))
}

// SetChecked sets the checked state
func (b *Button) SetChecked(checked bool) {
	v := 0
	if checked {
		v = 1
	}
	C.TabbyButton_SetChecked(b.handle, C.int(v))
}

// OnClicked sets the click callback
func (b *Button) OnClicked(cb MenuCallback) {
	b.cb = cb
	storeButtonCallback(b)
	C.TabbyButton_OnClicked(b.handle,
		(C.TabbyMenuCallback)(unsafe.Pointer(C.tabbyMenuCallback)),
		unsafe.Pointer(b))
}

// ComboBox represents a dropdown combo box
type ComboBox struct {
	handle C.TabbyComboBox
}

// NewComboBox creates a new combo box
func NewComboBox() *ComboBox {
	return &ComboBox{handle: C.TabbyComboBox_Create()}
}

// AddItem adds an item
func (c *ComboBox) AddItem(text string) {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	C.TabbyComboBox_AddItem(c.handle, cs)
}

// Clear removes all items
func (c *ComboBox) Clear() {
	C.TabbyComboBox_Clear(c.handle)
}

// CurrentIndex returns the selected index
func (c *ComboBox) CurrentIndex() int {
	return int(C.TabbyComboBox_CurrentIndex(c.handle))
}

// SetCurrentIndex sets the selected index
func (c *ComboBox) SetCurrentIndex(index int) {
	C.TabbyComboBox_SetCurrentIndex(c.handle, C.int(index))
}

// ---- Layout ----

// Layout represents a box layout
type Layout struct {
	handle C.TabbyLayout
}

// NewVBoxLayout creates a vertical box layout
func NewVBoxLayout() *Layout {
	return &Layout{handle: C.TabbyLayout_CreateVBox()}
}

// NewHBoxLayout creates a horizontal box layout
func NewHBoxLayout() *Layout {
	return &Layout{handle: C.TabbyLayout_CreateHBox()}
}

// AddWidget adds a widget to the layout
func (l *Layout) AddWidget(widget Widget) {
	C.TabbyLayout_AddWidget(l.handle, widget.handle())
}

// AddLayout adds a child layout
func (l *Layout) AddLayout(child *Layout) {
	C.TabbyLayout_AddLayout(l.handle, child.handle)
}

// AddStretch adds a stretch space
func (l *Layout) AddStretch() {
	C.TabbyLayout_AddStretch(l.handle)
}

// ---- Dialog ----

// Dialog represents a dialog window
type Dialog struct {
	handle C.TabbyDialog
}

// NewDialog creates a new dialog
func NewDialog(parent *Window, title string) *Dialog {
	var parentHandle C.TabbyWindow
	if parent != nil {
		parentHandle = parent.handle
	}
	cs := C.CString(title)
	defer C.free(unsafe.Pointer(cs))
	return &Dialog{handle: C.TabbyDialog_Create(parentHandle, cs)}
}

// Destroy destroys the dialog
func (d *Dialog) Destroy() {
	C.TabbyDialog_Destroy(d.handle)
}

// Exec runs the dialog modally. Returns true if accepted.
func (d *Dialog) Exec() bool {
	return C.TabbyDialog_Exec(d.handle) != 0
}

// Show shows the dialog non-modally
func (d *Dialog) Show() {
	C.TabbyDialog_Show(d.handle)
}

// Close closes the dialog
func (d *Dialog) Close() {
	C.TabbyDialog_Close(d.handle)
}

// SetSize sets the dialog size
func (d *Dialog) SetSize(width, height int) {
	C.TabbyDialog_SetSize(d.handle, C.int(width), C.int(height))
}

// SetLayout sets the dialog's layout
func (d *Dialog) SetLayout(layout *Layout) {
	C.TabbyDialog_SetLayout(d.handle, layout.handle)
}

// ---- File Dialogs ----

// GetOpenFileName shows an open file dialog
func GetOpenFileName(parent *Window, caption, dir, filter string) string {
	var parentHandle C.TabbyWindow
	if parent != nil {
		parentHandle = parent.handle
	}
	cc := C.CString(caption)
	defer C.free(unsafe.Pointer(cc))
	cd := C.CString(dir)
	defer C.free(unsafe.Pointer(cd))
	cf := C.CString(filter)
	defer C.free(unsafe.Pointer(cf))

	cs := C.TabbyFileDialog_GetOpenFileName(parentHandle, cc, cd, cf)
	defer C.Tabby_FreeString(cs)
	return C.GoString(cs)
}

// GetSaveFileName shows a save file dialog
func GetSaveFileName(parent *Window, caption, dir, filter string) string {
	var parentHandle C.TabbyWindow
	if parent != nil {
		parentHandle = parent.handle
	}
	cc := C.CString(caption)
	defer C.free(unsafe.Pointer(cc))
	cd := C.CString(dir)
	defer C.free(unsafe.Pointer(cd))
	cf := C.CString(filter)
	defer C.free(unsafe.Pointer(cf))

	cs := C.TabbyFileDialog_GetSaveFileName(parentHandle, cc, cd, cf)
	defer C.Tabby_FreeString(cs)
	return C.GoString(cs)
}

// GetExistingDirectory shows a directory picker dialog
func GetExistingDirectory(parent *Window, caption, dir string) string {
	var parentHandle C.TabbyWindow
	if parent != nil {
		parentHandle = parent.handle
	}
	cc := C.CString(caption)
	defer C.free(unsafe.Pointer(cc))
	cd := C.CString(dir)
	defer C.free(unsafe.Pointer(cd))

	cs := C.TabbyFileDialog_GetExistingDirectory(parentHandle, cc, cd)
	defer C.Tabby_FreeString(cs)
	return C.GoString(cs)
}

// ---- Clipboard ----

// ClipboardText returns the clipboard text
func ClipboardText() string {
	cs := C.TabbyClipboard_Text()
	defer C.Tabby_FreeString(cs)
	return C.GoString(cs)
}

// SetClipboardText sets the clipboard text
func SetClipboardText(text string) {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	C.TabbyClipboard_SetText(cs)
}

// ---- Screen Info ----

// ScreenSize returns the primary screen dimensions
func ScreenSize() (int, int) {
	return int(C.TabbyScreen_Width()), int(C.TabbyScreen_Height())
}

// DevicePixelRatio returns the screen's device pixel ratio
func DevicePixelRatio() float64 {
	return float64(C.TabbyScreen_DevicePixelRatio())
}

// ---- Widget interface ----

// Widget is the interface for all UI widgets
type Widget interface {
	handle() C.TabbyWidget
}

// WidgetBase is a base type that satisfies the Widget interface
type WidgetBase struct {
	h C.TabbyWidget
}

func (w *WidgetBase) handle() C.TabbyWidget { return w.h }

// Widget implementations satisfy the interface
func (t *TabWidget) handle() C.TabbyWidget { return C.TabbyWidget(t.handle) }
func (s *Splitter) handle() C.TabbyWidget { return C.TabbyWidget(s.handle) }
func (t *Terminal) handle() C.TabbyWidget { return C.TabbyWidget(t.handle) }
func (l *Label) handle() C.TabbyWidget     { return C.TabbyWidget(l.handle) }
func (l *LineEdit) handle() C.TabbyWidget  { return C.TabbyWidget(l.handle) }
func (b *Button) handle() C.TabbyWidget    { return C.TabbyWidget(b.handle) }
func (c *ComboBox) handle() C.TabbyWidget  { return C.TabbyWidget(c.handle) }

// ---- Callback storage (prevent GC) ----

var (
	windowCallbacks = make(map[unsafe.Pointer]*Window)
	actionCallbacks = make(map[unsafe.Pointer]*Action)
	buttonCallbacks = make(map[unsafe.Pointer]*Button)
)

func storeWindowCallback(w *Window)   { windowCallbacks[unsafe.Pointer(w)] = w }
func storeActionCallback(a *Action)   { actionCallbacks[unsafe.Pointer(a)] = a }
func storeButtonCallback(b *Button)   { buttonCallbacks[unsafe.Pointer(b)] = b }

// Platform returns the current OS
func Platform() string {
	return runtime.GOOS
}
