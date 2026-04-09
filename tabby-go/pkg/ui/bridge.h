#ifndef TABBY_UI_BRIDGE_H
#define TABBY_UI_BRIDGE_H

// Tabby UI Bridge — C API wrapping BTK C++ widgets for CGo access.
// This header provides a flat C API that exposes BTK's widget system
// to Go via CGo. The implementation (bridge.cpp) wraps BTK C++ classes.

#ifdef __cplusplus
extern "C" {
#endif

// Opaque handle types
typedef void* TabbyApp;
typedef void* TabbyWindow;
typedef void* TabbyWidget;
typedef void* TabbyTerminalWidget;
typedef void* TabbyTabWidget;
typedef void* TabbySplitter;
typedef void* TabbyMenuBar;
typedef void* TabbyMenu;
typedef void* TabbyAction;
typedef void* TabbyToolBar;
typedef void* TabbyLabel;
typedef void* TabbyLineEdit;
typedef void* TabbyButton;
typedef void* TabbyComboBox;
typedef void* TabbyStatusBar;
typedef void* TabbyTreeView;
typedef void* TabbyListView;
typedef void* TabbyDialog;
typedef void* TabbyLayout;
typedef void* TabbySettingsWidget;

// Callback types
typedef void (*TabbyCallback)(void* userData);
typedef void (*TabbyDataCallback)(const char* data, int len, void* userData);
typedef void (*TabbySizeCallback)(int cols, int rows, void* userData);
typedef void (*TabbyCloseCallback)(void* userData);
typedef void (*TabbyMenuCallback)(void* userData);
typedef void (*TabbyInputCallback)(const char* text, int len, void* userData);

// ---- Application Lifecycle ----

// TabbyApp_Create creates the BTK application instance.
// Must be called before any other UI function.
TabbyApp TabbyApp_Create(int argc, char** argv);

// TabbyApp_Run starts the application event loop. Blocks until exit.
int TabbyApp_Run(TabbyApp app);

// TabbyApp_Quit requests the application to exit.
void TabbyApp_Quit(TabbyApp app);

// TabbyApp_Destroy cleans up the application.
void TabbyApp_Destroy(TabbyApp app);

// TabbyApp_SetStyle sets the application style (e.g., "Fusion", "Windows").
void TabbyApp_SetStyle(TabbyApp app, const char* styleName);

// ---- Main Window ----

TabbyWindow TabbyWindow_Create(TabbyApp app);
void TabbyWindow_Destroy(TabbyWindow win);
void TabbyWindow_Show(TabbyWindow win);
void TabbyWindow_Hide(TabbyWindow win);
void TabbyWindow_SetTitle(TabbyWindow win, const char* title);
void TabbyWindow_SetSize(TabbyWindow win, int width, int height);
void TabbyWindow_SetMinimumSize(TabbyWindow win, int width, int height);
void TabbyWindow_GetSize(TabbyWindow win, int* width, int* height);
void TabbyWindow_SetPosition(TabbyWindow win, int x, int y);
void TabbyWindow_Maximize(TabbyWindow win);
void TabbyWindow_Minimize(TabbyWindow win);
void TabbyWindow_Restore(TabbyWindow win);
void TabbyWindow_SetFullScreen(TabbyWindow win, int fullScreen);
void TabbyWindow_Close(TabbyWindow win);

// Set callbacks
void TabbyWindow_OnClose(TabbyWindow win, TabbyCloseCallback cb, void* userData);
void TabbyWindow_OnResize(TabbyWindow win, TabbySizeCallback cb, void* userData);

// Central widget
void TabbyWindow_SetCentralWidget(TabbyWindow win, TabbyWidget widget);

// Menu bar
void TabbyWindow_SetMenuBar(TabbyWindow win, TabbyMenuBar menuBar);
TabbyMenuBar TabbyWindow_MenuBar(TabbyWindow win);

// Status bar
TabbyStatusBar TabbyWindow_StatusBar(TabbyWindow win);
void TabbyStatusBar_ShowMessage(TabbyStatusBar bar, const char* message, int timeoutMs);
void TabbyStatusBar_SetPermanentWidget(TabbyStatusBar bar, TabbyWidget widget);

// Tool bar
TabbyToolBar TabbyWindow_AddToolBar(TabbyWindow win, const char* title);
void TabbyToolBar_AddAction(TabbyToolBar toolbar, TabbyAction action);
void TabbyToolBar_AddSeparator(TabbyToolBar toolbar);

// ---- Menu System ----

TabbyMenuBar TabbyMenuBar_Create(TabbyWindow win);
TabbyMenu TabbyMenuBar_AddMenu(TabbyMenuBar menuBar, const char* title);
TabbyAction TabbyMenu_AddAction(TabbyMenu menu, const char* text);
TabbyAction TabbyMenu_AddSeparator(TabbyMenu menu);
void TabbyAction_SetShortcut(TabbyAction action, const char* shortcut);
void TabbyAction_SetCheckable(TabbyAction action, int checkable);
void TabbyAction_SetChecked(TabbyAction action, int checked);
int TabbyAction_IsChecked(TabbyAction action);
void TabbyAction_SetEnabled(TabbyAction action, int enabled);
void TabbyAction_SetToolTip(TabbyAction action, const char* tip);
void TabbyAction_OnTriggered(TabbyAction action, TabbyMenuCallback cb, void* userData);

// ---- Tab Widget ----

TabbyTabWidget TabbyTabWidget_Create();
void TabbyTabWidget_Destroy(TabbyTabWidget tabs);
int TabbyTabWidget_AddTab(TabbyTabWidget tabs, TabbyWidget widget, const char* label);
void TabbyTabWidget_RemoveTab(TabbyTabWidget tabs, int index);
void TabbyTabWidget_SetTabText(TabbyTabWidget tabs, int index, const char* text);
void TabbyTabWidget_SetCurrentIndex(TabbyTabWidget tabs, int index);
int TabbyTabWidget_CurrentIndex(TabbyTabWidget tabs);
int TabbyTabWidget_Count(TabbyTabWidget tabs);
void TabbyTabWidget_SetTabsClosable(TabbyTabWidget tabs, int closable);
void TabbyTabWidget_SetMovable(TabbyTabWidget tabs, int movable);
void TabbyTabWidget_OnTabCloseRequested(TabbyTabWidget tabs, TabbyMenuCallback cb, void* userData);
void TabbyTabWidget_OnCurrentChanged(TabbyTabWidget tabs, TabbyMenuCallback cb, void* userData);

// ---- Splitter ----

TabbySplitter TabbySplitter_Create(int orientation); // 0=horizontal, 1=vertical
void TabbySplitter_Destroy(TabbySplitter splitter);
void TabbySplitter_AddWidget(TabbySplitter splitter, TabbyWidget widget);
void TabbySplitter_SetSizes(TabbySplitter splitter, int* sizes, int count);
void TabbySplitter_GetSizes(TabbySplitter splitter, int* sizes, int* count);

// ---- Terminal Widget ----
// The terminal widget renders terminal output using BTK's painting system
// and forwards keyboard/mouse input.

TabbyTerminalWidget TabbyTerminalWidget_Create();
void TabbyTerminalWidget_Destroy(TabbyTerminalWidget term);
void TabbyTerminalWidget_SetSize(TabbyTerminalWidget term, int cols, int rows);
void TabbyTerminalWidget_GetSize(TabbyTerminalWidget term, int* cols, int* rows);
void TabbyTerminalWidget_Write(TabbyTerminalWidget term, const char* data, int len);
void TabbyTerminalWidget_SetColorScheme(TabbyTerminalWidget term, const char* json);
void TabbyTerminalWidget_SetFont(TabbyTerminalWidget term, const char* family, int size);
void TabbyTerminalWidget_SetScrollback(TabbyTerminalWidget term, int lines);
void TabbyTerminalWidget_Clear(TabbyTerminalWidget term);

// Input callbacks
void TabbyTerminalWidget_OnInput(TabbyTerminalWidget term, TabbyDataCallback cb, void* userData);
void TabbyTerminalWidget_OnResize(TabbyTerminalWidget term, TabbySizeCallback cb, void* userData);
void TabbyTerminalWidget_OnTitleChanged(TabbyTerminalWidget term, TabbyDataCallback cb, void* userData);

// Focus
void TabbyTerminalWidget_SetFocus(TabbyTerminalWidget term);
int TabbyTerminalWidget_HasFocus(TabbyTerminalWidget term);

// ---- Basic Widgets ----

TabbyLabel TabbyLabel_Create(const char* text);
void TabbyLabel_SetText(TabbyLabel label, const char* text);

TabbyLineEdit TabbyLineEdit_Create();
void TabbyLineEdit_SetText(TabbyLineEdit edit, const char* text);
const char* TabbyLineEdit_Text(TabbyLineEdit edit);
void TabbyLineEdit_SetPlaceholder(TabbyLineEdit edit, const char* text);
void TabbyLineEdit_SetEchoMode(TabbyLineEdit edit, int mode); // 0=normal, 1=password, 2=noEcho, 3=passwordOnEdit
void TabbyLineEdit_OnTextChanged(TabbyLineEdit edit, TabbyDataCallback cb, void* userData);
void TabbyLineEdit_OnReturnPressed(TabbyLineEdit edit, TabbyMenuCallback cb, void* userData);

TabbyButton TabbyButton_Create(const char* text);
void TabbyButton_SetText(TabbyButton btn, const char* text);
void TabbyButton_SetCheckable(TabbyButton btn, int checkable);
void TabbyButton_SetChecked(TabbyButton btn, int checked);
void TabbyButton_OnClicked(TabbyButton btn, TabbyMenuCallback cb, void* userData);

TabbyComboBox TabbyComboBox_Create();
void TabbyComboBox_AddItem(TabbyComboBox combo, const char* text);
void TabbyComboBox_Clear(TabbyComboBox combo);
int TabbyComboBox_CurrentIndex(TabbyComboBox combo);
void TabbyComboBox_SetCurrentIndex(TabbyComboBox combo, int index);
void TabbyComboBox_OnCurrentIndexChanged(TabbyComboBox combo, TabbyMenuCallback cb, void* userData);

// ---- Layout ----

TabbyLayout TabbyLayout_CreateVBox();
TabbyLayout TabbyLayout_CreateHBox();
void TabbyLayout_AddWidget(TabbyLayout layout, TabbyWidget widget);
void TabbyLayout_AddLayout(TabbyLayout layout, TabbyLayout child);
void TabbyLayout_AddStretch(TabbyLayout layout);
void TabbyWidget_SetLayout(TabbyWidget widget, TabbyLayout layout);

// ---- Dialog ----

TabbyDialog TabbyDialog_Create(TabbyWindow parent, const char* title);
void TabbyDialog_Destroy(TabbyDialog dlg);
int TabbyDialog_Exec(TabbyDialog dlg); // modal, returns accepted=1, rejected=0
void TabbyDialog_Show(TabbyDialog dlg);
void TabbyDialog_Close(TabbyDialog dlg);
void TabbyDialog_SetSize(TabbyDialog dlg, int width, int height);
void TabbyDialog_SetLayout(TabbyDialog dlg, TabbyLayout layout);

// ---- File Dialog ----

// Returns allocated string, caller must free with Tabby_FreeString
const char* TabbyFileDialog_GetOpenFileName(TabbyWindow parent, const char* caption, const char* dir, const char* filter);
const char* TabbyFileDialog_GetSaveFileName(TabbyWindow parent, const char* caption, const char* dir, const char* filter);
const char* TabbyFileDialog_GetExistingDirectory(TabbyWindow parent, const char* caption, const char* dir);

// ---- Utility ----

void Tabby_FreeString(const char* s);

// Process pending events (for integrating with Go's event loop)
void TabbyApp_ProcessEvents(TabbyApp app);

// Run a function on the UI thread
void TabbyApp_RunOnUIThread(TabbyApp app, TabbyCallback cb, void* userData);

// Clipboard
const char* TabbyClipboard_Text();
void TabbyClipboard_SetText(const char* text);

// Screen info
int TabbyScreen_Width();
int TabbyScreen_Height();
double TabbyScreen_DevicePixelRatio();

// Theming
void TabbyApp_SetDarkMode(TabbyApp app, int darkMode);
void TabbyApp_SetFont(TabbyApp app, const char* family, int size);

// Settings path
const char* TabbyApp_SettingsPath();

#ifdef __cplusplus
}
#endif

#endif // TABBY_UI_BRIDGE_H
