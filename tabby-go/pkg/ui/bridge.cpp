// Tabby UI Bridge — C++ implementation wrapping BTK widgets.
// This file implements the C API declared in bridge.h using BTK's C++ classes.
//
// Build: compiled as a static library linked into the Go binary via CGo.

#include "bridge.h"

#include <qapplication.h>
#include <qmainwindow.h>
#include <qwidget.h>
#include <qtabwidget.h>
#include <qsplitter.h>
#include <qmenubar.h>
#include <qmenu.h>
#include <qaction.h>
#include <qtoolbar.h>
#include <qlabel.h>
#include <qlineedit.h>
#include <qpushbutton.h>
#include <qcombobox.h>
#include <qstatusbar.h>
#include <qlayout.h>
#include <qdialog.h>
#include <qfiledialog.h>
#include <qclipboard.h>
#include <qscreen.h>
#include <qguiapplication.h>
#include <qdesktopservices.h>
#include <qstandardpaths.h>
#include <qstyle.h>
#include <qstylefactory.h>
#include <qfont.h>
#include <qevent.h>
#include <qscrollbar.h>
#include <qtextcodec.h>
#include <qdatetime.h>

#include <string>
#include <vector>
#include <cstring>
#include <cstdio>

// ---- Helper for allocating strings that Go can free ----
static thread_local std::vector<char*> allocatedStrings;

static char* allocateString(const std::string& s) {
    char* result = new char[s.size() + 1];
    std::memcpy(result, s.c_str(), s.size() + 1);
    allocatedStrings.push_back(result);
    return result;
}

// ---- Callback wrapper types ----
struct TabbyCallbackData {
    TabbyCloseCallback closeCb;
    TabbySizeCallback sizeCb;
    TabbyDataCallback dataCb;
    TabbyMenuCallback menuCb;
    TabbyInputCallback inputCb;
    void* userData;
};

static TabbyCallbackData* createCallbackData() {
    return new TabbyCallbackData{nullptr, nullptr, nullptr, nullptr, nullptr, nullptr};
}

// ---- Application ----

TabbyApp TabbyApp_Create(int argc, char** argv) {
    // BTK requires argc/argv to persist
    int* pargc = new int(argc);
    char** pargv = new char*[argc];
    for (int i = 0; i < argc; i++) {
        pargv[i] = new char[std::strlen(argv[i]) + 1];
        std::strcpy(pargv[i], argv[i]);
    }

    QApplication* app = new QApplication(*pargc, pargv);
    return static_cast<TabbyApp>(app);
}

int TabbyApp_Run(TabbyApp app) {
    QApplication* qapp = static_cast<QApplication*>(app);
    return qapp->exec();
}

void TabbyApp_Quit(TabbyApp app) {
    QApplication* qapp = static_cast<QApplication*>(app);
    qapp->quit();
}

void TabbyApp_Destroy(TabbyApp app) {
    QApplication* qapp = static_cast<QApplication*>(app);
    delete qapp;
}

void TabbyApp_SetStyle(TabbyApp app, const char* styleName) {
    QApplication* qapp = static_cast<QApplication*>(app);
    QApplication::setStyle(QString::fromUtf8(styleName));
}

void TabbyApp_ProcessEvents(TabbyApp app) {
    QApplication* qapp = static_cast<QApplication*>(app);
    qapp->processEvents();
}

void TabbyApp_RunOnUIThread(TabbyApp app, TabbyCallback cb, void* userData) {
    // Post a lambda to the event loop
    QApplication* qapp = static_cast<QApplication*>(app);
    QMetaObject::invokeMethod(qapp, [cb, userData]() {
        if (cb) cb(userData);
    }, Qt::QueuedConnection);
}

void TabbyApp_SetDarkMode(TabbyApp app, int darkMode) {
    // Set dark palette if available
    QApplication* qapp = static_cast<QApplication*>(app);
    if (darkMode) {
        QPalette darkPalette;
        darkPalette.setColor(QPalette::Window, QColor(53, 53, 53));
        darkPalette.setColor(QPalette::WindowText, Qt::white);
        darkPalette.setColor(QPalette::Base, QColor(35, 35, 35));
        darkPalette.setColor(QPalette::AlternateBase, QColor(53, 53, 53));
        darkPalette.setColor(QPalette::Text, Qt::white);
        darkPalette.setColor(QPalette::Button, QColor(53, 53, 53));
        darkPalette.setColor(QPalette::ButtonText, Qt::white);
        darkPalette.setColor(QPalette::BrightText, Qt::red);
        darkPalette.setColor(QPalette::Link, QColor(42, 130, 218));
        darkPalette.setColor(QPalette::Highlight, QColor(42, 130, 218));
        darkPalette.setColor(QPalette::HighlightedText, QColor(35, 35, 35));
        qapp->setPalette(darkPalette);
    }
}

void TabbyApp_SetFont(TabbyApp app, const char* family, int size) {
    QApplication* qapp = static_cast<QApplication*>(app);
    QFont font(QString::fromUtf8(family), size);
    qapp->setFont(font);
}

const char* TabbyApp_SettingsPath() {
    QString path = QStandardPaths::writableLocation(QStandardPaths::AppDataLocation);
    return allocateString(path.toStdString());
}

// ---- Main Window ----

class TabbyMainWindow : public QMainWindow {
public:
    TabbyCallbackData* closeData;
    TabbyCallbackData* resizeData;

    TabbyMainWindow(QWidget* parent = nullptr)
        : QMainWindow(parent), closeData(nullptr), resizeData(nullptr) {}

    ~TabbyMainWindow() {
        delete closeData;
        delete resizeData;
    }

    void closeEvent(QCloseEvent* event) override {
        if (closeData && closeData->closeCb) {
            closeData->closeCb(closeData->userData);
        }
        QMainWindow::closeEvent(event);
    }

    void resizeEvent(QResizeEvent* event) override {
        if (resizeData && resizeData->sizeCb) {
            resizeData->sizeCb(width(), height(), resizeData->userData);
        }
        QMainWindow::resizeEvent(event);
    }
};

TabbyWindow TabbyWindow_Create(TabbyApp app) {
    TabbyMainWindow* win = new TabbyMainWindow();
    win->setAttribute(Qt::WA_DeleteOnClose, false);
    return static_cast<TabbyWindow>(win);
}

void TabbyWindow_Destroy(TabbyWindow win) {
    TabbyMainWindow* w = static_cast<TabbyMainWindow*>(win);
    delete w;
}

void TabbyWindow_Show(TabbyWindow win) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    w->show();
}

void TabbyWindow_Hide(TabbyWindow win) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    w->hide();
}

void TabbyWindow_SetTitle(TabbyWindow win, const char* title) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    w->setWindowTitle(QString::fromUtf8(title));
}

void TabbyWindow_SetSize(TabbyWindow win, int width, int height) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    w->resize(width, height);
}

void TabbyWindow_SetMinimumSize(TabbyWindow win, int width, int height) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    w->setMinimumSize(width, height);
}

void TabbyWindow_GetSize(TabbyWindow win, int* width, int* height) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    *width = w->width();
    *height = w->height();
}

void TabbyWindow_SetPosition(TabbyWindow win, int x, int y) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    w->move(x, y);
}

void TabbyWindow_Maximize(TabbyWindow win) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    w->showMaximized();
}

void TabbyWindow_Minimize(TabbyWindow win) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    w->showMinimized();
}

void TabbyWindow_Restore(TabbyWindow win) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    w->showNormal();
}

void TabbyWindow_SetFullScreen(TabbyWindow win, int fullScreen) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    if (fullScreen) {
        w->showFullScreen();
    } else {
        w->showNormal();
    }
}

void TabbyWindow_Close(TabbyWindow win) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    w->close();
}

void TabbyWindow_OnClose(TabbyWindow win, TabbyCloseCallback cb, void* userData) {
    TabbyMainWindow* w = static_cast<TabbyMainWindow*>(win);
    if (!w->closeData) {
        w->closeData = createCallbackData();
    }
    w->closeData->closeCb = cb;
    w->closeData->userData = userData;
}

void TabbyWindow_OnResize(TabbyWindow win, TabbySizeCallback cb, void* userData) {
    TabbyMainWindow* w = static_cast<TabbyMainWindow*>(win);
    if (!w->resizeData) {
        w->resizeData = createCallbackData();
    }
    w->resizeData->sizeCb = cb;
    w->resizeData->userData = userData;
}

void TabbyWindow_SetCentralWidget(TabbyWindow win, TabbyWidget widget) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    QWidget* qwidget = static_cast<QWidget*>(widget);
    w->setCentralWidget(qwidget);
}

void TabbyWindow_SetMenuBar(TabbyWindow win, TabbyMenuBar menuBar) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    QMenuBar* mb = static_cast<QMenuBar*>(menuBar);
    w->setMenuBar(mb);
}

TabbyMenuBar TabbyWindow_MenuBar(TabbyWindow win) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    return static_cast<TabbyMenuBar>(w->menuBar());
}

TabbyStatusBar TabbyWindow_StatusBar(TabbyWindow win) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    return static_cast<TabbyStatusBar>(w->statusBar());
}

void TabbyStatusBar_ShowMessage(TabbyStatusBar bar, const char* message, int timeoutMs) {
    QStatusBar* sb = static_cast<QStatusBar*>(bar);
    sb->showMessage(QString::fromUtf8(message), timeoutMs);
}

void TabbyStatusBar_SetPermanentWidget(TabbyStatusBar bar, TabbyWidget widget) {
    QStatusBar* sb = static_cast<QStatusBar*>(bar);
    QWidget* w = static_cast<QWidget*>(widget);
    sb->addPermanentWidget(w);
}

TabbyToolBar TabbyWindow_AddToolBar(TabbyWindow win, const char* title) {
    QMainWindow* w = static_cast<QMainWindow*>(win);
    QToolBar* tb = w->addToolBar(QString::fromUtf8(title));
    return static_cast<TabbyToolBar>(tb);
}

void TabbyToolBar_AddAction(TabbyToolBar toolbar, TabbyAction action) {
    QToolBar* tb = static_cast<QToolBar*>(toolbar);
    QAction* a = static_cast<QAction*>(action);
    tb->addAction(a);
}

void TabbyToolBar_AddSeparator(TabbyToolBar toolbar) {
    QToolBar* tb = static_cast<QToolBar*>(toolbar);
    tb->addSeparator();
}

// ---- Menu System ----

TabbyMenuBar TabbyMenuBar_Create(TabbyWindow win) {
    // QMainWindow creates its menu bar automatically
    return TabbyWindow_MenuBar(win);
}

TabbyMenu TabbyMenuBar_AddMenu(TabbyMenuBar menuBar, const char* title) {
    QMenuBar* mb = static_cast<QMenuBar*>(menuBar);
    QMenu* menu = mb->addMenu(QString::fromUtf8(title));
    return static_cast<TabbyMenu>(menu);
}

TabbyAction TabbyMenu_AddAction(TabbyMenu menu, const char* text) {
    QMenu* m = static_cast<QMenu*>(menu);
    QAction* a = m->addAction(QString::fromUtf8(text));
    return static_cast<TabbyAction>(a);
}

TabbyAction TabbyMenu_AddSeparator(TabbyMenu menu) {
    QMenu* m = static_cast<QMenu*>(menu);
    QAction* a = m->addSeparator();
    return static_cast<TabbyAction>(a);
}

void TabbyAction_SetShortcut(TabbyAction action, const char* shortcut) {
    QAction* a = static_cast<QAction*>(action);
    a->setShortcut(QKeySequence(QString::fromUtf8(shortcut)));
}

void TabbyAction_SetCheckable(TabbyAction action, int checkable) {
    QAction* a = static_cast<QAction*>(action);
    a->setCheckable(checkable != 0);
}

void TabbyAction_SetChecked(TabbyAction action, int checked) {
    QAction* a = static_cast<QAction*>(action);
    a->setChecked(checked != 0);
}

int TabbyAction_IsChecked(TabbyAction action) {
    QAction* a = static_cast<QAction*>(action);
    return a->isChecked() ? 1 : 0;
}

void TabbyAction_SetEnabled(TabbyAction action, int enabled) {
    QAction* a = static_cast<QAction*>(action);
    a->setEnabled(enabled != 0);
}

void TabbyAction_SetToolTip(TabbyAction action, const char* tip) {
    QAction* a = static_cast<QAction*>(action);
    a->setToolTip(QString::fromUtf8(tip));
}

void TabbyAction_OnTriggered(TabbyAction action, TabbyMenuCallback cb, void* userData) {
    QAction* a = static_cast<QAction*>(action);
    TabbyCallbackData* data = createCallbackData();
    data->menuCb = cb;
    data->userData = userData;
    QObject::connect(a, &QAction::triggered, [data](bool) {
        if (data->menuCb) data->menuCb(data->userData);
    });
}

// ---- Tab Widget ----

TabbyTabWidget TabbyTabWidget_Create() {
    QTabWidget* tabs = new QTabWidget();
    tabs->setDocumentMode(true);
    return static_cast<TabbyTabWidget>(tabs);
}

void TabbyTabWidget_Destroy(TabbyTabWidget tabs) {
    QTabWidget* t = static_cast<QTabWidget*>(tabs);
    delete t;
}

int TabbyTabWidget_AddTab(TabbyTabWidget tabs, TabbyWidget widget, const char* label) {
    QTabWidget* t = static_cast<QTabWidget*>(tabs);
    QWidget* w = static_cast<QWidget*>(widget);
    return t->addTab(w, QString::fromUtf8(label));
}

void TabbyTabWidget_RemoveTab(TabbyTabWidget tabs, int index) {
    QTabWidget* t = static_cast<QTabWidget*>(tabs);
    t->removeTab(index);
}

void TabbyTabWidget_SetTabText(TabbyTabWidget tabs, int index, const char* text) {
    QTabWidget* t = static_cast<QTabWidget*>(tabs);
    t->setTabText(index, QString::fromUtf8(text));
}

void TabbyTabWidget_SetCurrentIndex(TabbyTabWidget tabs, int index) {
    QTabWidget* t = static_cast<QTabWidget*>(tabs);
    t->setCurrentIndex(index);
}

int TabbyTabWidget_CurrentIndex(TabbyTabWidget tabs) {
    QTabWidget* t = static_cast<QTabWidget*>(tabs);
    return t->currentIndex();
}

int TabbyTabWidget_Count(TabbyTabWidget tabs) {
    QTabWidget* t = static_cast<QTabWidget*>(tabs);
    return t->count();
}

void TabbyTabWidget_SetTabsClosable(TabbyTabWidget tabs, int closable) {
    QTabWidget* t = static_cast<QTabWidget*>(tabs);
    t->setTabsClosable(closable != 0);
}

void TabbyTabWidget_SetMovable(TabbyTabWidget tabs, int movable) {
    QTabWidget* t = static_cast<QTabWidget*>(tabs);
    t->setMovable(movable != 0);
}

void TabbyTabWidget_OnTabCloseRequested(TabbyTabWidget tabs, TabbyMenuCallback cb, void* userData) {
    QTabWidget* t = static_cast<QTabWidget*>(tabs);
    TabbyCallbackData* data = createCallbackData();
    data->menuCb = cb;
    data->userData = userData;
    QObject::connect(t, &QTabWidget::tabCloseRequested, [data](int index) {
        if (data->menuCb) data->menuCb(data->userData);
    });
}

void TabbyTabWidget_OnCurrentChanged(TabbyTabWidget tabs, TabbyMenuCallback cb, void* userData) {
    QTabWidget* t = static_cast<QTabWidget*>(tabs);
    TabbyCallbackData* data = createCallbackData();
    data->menuCb = cb;
    data->userData = userData;
    QObject::connect(t, &QTabWidget::currentChanged, [data](int index) {
        if (data->menuCb) data->menuCb(data->userData);
    });
}

// ---- Splitter ----

TabbySplitter TabbySplitter_Create(int orientation) {
    Qt::Orientation orient = (orientation == 1) ? Qt::Vertical : Qt::Horizontal;
    QSplitter* splitter = new QSplitter(orient);
    return static_cast<TabbySplitter>(splitter);
}

void TabbySplitter_Destroy(TabbySplitter splitter) {
    QSplitter* s = static_cast<QSplitter*>(splitter);
    delete s;
}

void TabbySplitter_AddWidget(TabbySplitter splitter, TabbyWidget widget) {
    QSplitter* s = static_cast<QSplitter*>(splitter);
    QWidget* w = static_cast<QWidget*>(widget);
    s->addWidget(w);
}

void TabbySplitter_SetSizes(TabbySplitter splitter, int* sizes, int count) {
    QSplitter* s = static_cast<QSplitter*>(splitter);
    QList<int> list;
    for (int i = 0; i < count; i++) {
        list.append(sizes[i]);
    }
    s->setSizes(list);
}

void TabbySplitter_GetSizes(TabbySplitter splitter, int* sizes, int* count) {
    QSplitter* s = static_cast<QSplitter*>(splitter);
    QList<int> list = s->sizes();
    *count = list.size();
    for (int i = 0; i < list.size() && i < *count; i++) {
        sizes[i] = list[i];
    }
}

// ---- Terminal Widget (placeholder - basic QWidget with input handling) ----

class TabbyTerminalRender : public QWidget {
public:
    TabbyTerminalRender(QWidget* parent = nullptr)
        : QWidget(parent), cols(80), rows(24), scrollback(10000),
          dataCb(nullptr), sizeCb(nullptr), titleCb(nullptr), callbackData(nullptr)
    {
        setFocusPolicy(Qt::StrongFocus);
        setMinimumSize(200, 100);
        setAttribute(Qt::WA_InputMethodEnabled, true);
        QFont font("Consolas", 10);
        setFont(font);
    }

    int cols, rows, scrollback;
    TabbyDataCallback dataCb;
    TabbySizeCallback sizeCb;
    TabbyDataCallback titleCb;
    void* callbackData;

    void paintEvent(QPaintEvent*) override {
        QPainter painter(this);
        painter.fillRect(rect(), QColor(30, 30, 30));
        painter.setPen(Qt::white);
        painter.setFont(font());

        // Render placeholder text
        painter.drawText(rect(), Qt::AlignCenter,
            QString("Terminal %1x%2").arg(cols).arg(rows));
    }

    void keyPressEvent(QKeyEvent* event) override {
        if (dataCb && callbackData) {
            QString text = event->text();
            int key = event->key();
            QByteArray data;

            if (text.isEmpty()) {
                // Handle special keys
                switch (key) {
                    case Qt::Key_Return:
                    case Qt::Key_Enter:
                        data = "\r";
                        break;
                    case Qt::Key_Backspace:
                        data = "\x7f";
                        break;
                    case Qt::Key_Tab:
                        data = "\t";
                        break;
                    case Qt::Key_Escape:
                        data = "\x1b";
                        break;
                    default:
                        QWidget::keyPressEvent(event);
                        return;
                }
            } else {
                data = text.toUtf8();
            }

            dataCb(data.constData(), data.size(), callbackData);
        }
    }

    void resizeEvent(QResizeEvent* event) override {
        QWidget::resizeEvent(event);
        // Calculate terminal dimensions based on font metrics
        QFontMetrics fm(font());
        int newCols = width() / fm.horizontalAdvance('M');
        int newRows = height() / fm.lineSpacing();
        if (newCols > 0 && newRows > 0) {
            cols = newCols;
            rows = newRows;
            if (sizeCb && callbackData) {
                sizeCb(cols, rows, callbackData);
            }
        }
        update();
    }
};

TabbyTerminalWidget TabbyTerminalWidget_Create() {
    TabbyTerminalRender* term = new TabbyTerminalRender();
    return static_cast<TabbyTerminalWidget>(term);
}

void TabbyTerminalWidget_Destroy(TabbyTerminalWidget term) {
    TabbyTerminalRender* t = static_cast<TabbyTerminalRender*>(term);
    delete t;
}

void TabbyTerminalWidget_SetSize(TabbyTerminalWidget term, int cols, int rows) {
    TabbyTerminalRender* t = static_cast<TabbyTerminalRender*>(term);
    t->cols = cols;
    t->rows = rows;
}

void TabbyTerminalWidget_GetSize(TabbyTerminalWidget term, int* cols, int* rows) {
    TabbyTerminalRender* t = static_cast<TabbyTerminalRender*>(term);
    *cols = t->cols;
    *rows = t->rows;
}

void TabbyTerminalWidget_Write(TabbyTerminalWidget term, const char* data, int len) {
    // TODO: Full terminal emulation rendering
    // For now, trigger a repaint
    TabbyTerminalRender* t = static_cast<TabbyTerminalRender*>(term);
    t->update();
}

void TabbyTerminalWidget_SetColorScheme(TabbyTerminalWidget term, const char* json) {
    // TODO: Parse color scheme JSON and apply
}

void TabbyTerminalWidget_SetFont(TabbyTerminalWidget term, const char* family, int size) {
    TabbyTerminalRender* t = static_cast<TabbyTerminalRender*>(term);
    QFont font(QString::fromUtf8(family), size);
    t->setFont(font);
    t->update();
}

void TabbyTerminalWidget_SetScrollback(TabbyTerminalWidget term, int lines) {
    TabbyTerminalRender* t = static_cast<TabbyTerminalRender*>(term);
    t->scrollback = lines;
}

void TabbyTerminalWidget_Clear(TabbyTerminalWidget term) {
    TabbyTerminalRender* t = static_cast<TabbyTerminalRender*>(term);
    t->update();
}

void TabbyTerminalWidget_OnInput(TabbyTerminalWidget term, TabbyDataCallback cb, void* userData) {
    TabbyTerminalRender* t = static_cast<TabbyTerminalRender*>(term);
    t->dataCb = cb;
    t->callbackData = userData;
}

void TabbyTerminalWidget_OnResize(TabbyTerminalWidget term, TabbySizeCallback cb, void* userData) {
    TabbyTerminalRender* t = static_cast<TabbyTerminalRender*>(term);
    t->sizeCb = cb;
    t->callbackData = userData;
}

void TabbyTerminalWidget_OnTitleChanged(TabbyTerminalWidget term, TabbyDataCallback cb, void* userData) {
    TabbyTerminalRender* t = static_cast<TabbyTerminalRender*>(term);
    t->titleCb = cb;
}

void TabbyTerminalWidget_SetFocus(TabbyTerminalWidget term) {
    TabbyTerminalRender* t = static_cast<TabbyTerminalRender*>(term);
    t->setFocus();
}

int TabbyTerminalWidget_HasFocus(TabbyTerminalWidget term) {
    TabbyTerminalRender* t = static_cast<TabbyTerminalRender*>(term);
    return t->hasFocus() ? 1 : 0;
}

// ---- Basic Widgets ----

TabbyLabel TabbyLabel_Create(const char* text) {
    QLabel* label = new QLabel(QString::fromUtf8(text));
    return static_cast<TabbyLabel>(label);
}

void TabbyLabel_SetText(TabbyLabel label, const char* text) {
    QLabel* l = static_cast<QLabel*>(label);
    l->setText(QString::fromUtf8(text));
}

TabbyLineEdit TabbyLineEdit_Create() {
    QLineEdit* edit = new QLineEdit();
    return static_cast<TabbyLineEdit>(edit);
}

void TabbyLineEdit_SetText(TabbyLineEdit edit, const char* text) {
    QLineEdit* e = static_cast<QLineEdit*>(edit);
    e->setText(QString::fromUtf8(text));
}

const char* TabbyLineEdit_Text(TabbyLineEdit edit) {
    QLineEdit* e = static_cast<QLineEdit*>(edit);
    return allocateString(e->text().toStdString());
}

void TabbyLineEdit_SetPlaceholder(TabbyLineEdit edit, const char* text) {
    QLineEdit* e = static_cast<QLineEdit*>(edit);
    e->setPlaceholderText(QString::fromUtf8(text));
}

void TabbyLineEdit_SetEchoMode(TabbyLineEdit edit, int mode) {
    QLineEdit* e = static_cast<QLineEdit*>(edit);
    e->setEchoMode(static_cast<QLineEdit::EchoMode>(mode));
}

void TabbyLineEdit_OnTextChanged(TabbyLineEdit edit, TabbyDataCallback cb, void* userData) {
    QLineEdit* e = static_cast<QLineEdit*>(edit);
    TabbyCallbackData* data = createCallbackData();
    data->dataCb = cb;
    data->userData = userData;
    QObject::connect(e, &QLineEdit::textChanged, [data](const QString& text) {
        if (data->dataCb) {
            QByteArray utf8 = text.toUtf8();
            data->dataCb(utf8.constData(), utf8.size(), data->userData);
        }
    });
}

void TabbyLineEdit_OnReturnPressed(TabbyLineEdit edit, TabbyMenuCallback cb, void* userData) {
    QLineEdit* e = static_cast<QLineEdit*>(edit);
    TabbyCallbackData* data = createCallbackData();
    data->menuCb = cb;
    data->userData = userData;
    QObject::connect(e, &QLineEdit::returnPressed, [data]() {
        if (data->menuCb) data->menuCb(data->userData);
    });
}

TabbyButton TabbyButton_Create(const char* text) {
    QPushButton* btn = new QPushButton(QString::fromUtf8(text));
    return static_cast<TabbyButton>(btn);
}

void TabbyButton_SetText(TabbyButton btn, const char* text) {
    QPushButton* b = static_cast<QPushButton*>(btn);
    b->setText(QString::fromUtf8(text));
}

void TabbyButton_SetCheckable(TabbyButton btn, int checkable) {
    QPushButton* b = static_cast<QPushButton*>(btn);
    b->setCheckable(checkable != 0);
}

void TabbyButton_SetChecked(TabbyButton btn, int checked) {
    QPushButton* b = static_cast<QPushButton*>(btn);
    b->setChecked(checked != 0);
}

void TabbyButton_OnClicked(TabbyButton btn, TabbyMenuCallback cb, void* userData) {
    QPushButton* b = static_cast<QPushButton*>(btn);
    TabbyCallbackData* data = createCallbackData();
    data->menuCb = cb;
    data->userData = userData;
    QObject::connect(b, &QPushButton::clicked, [data](bool) {
        if (data->menuCb) data->menuCb(data->userData);
    });
}

TabbyComboBox TabbyComboBox_Create() {
    QComboBox* combo = new QComboBox();
    return static_cast<TabbyComboBox>(combo);
}

void TabbyComboBox_AddItem(TabbyComboBox combo, const char* text) {
    QComboBox* c = static_cast<QComboBox*>(combo);
    c->addItem(QString::fromUtf8(text));
}

void TabbyComboBox_Clear(TabbyComboBox combo) {
    QComboBox* c = static_cast<QComboBox*>(combo);
    c->clear();
}

int TabbyComboBox_CurrentIndex(TabbyComboBox combo) {
    QComboBox* c = static_cast<QComboBox*>(combo);
    return c->currentIndex();
}

void TabbyComboBox_SetCurrentIndex(TabbyComboBox combo, int index) {
    QComboBox* c = static_cast<QComboBox*>(combo);
    c->setCurrentIndex(index);
}

void TabbyComboBox_OnCurrentIndexChanged(TabbyComboBox combo, TabbyMenuCallback cb, void* userData) {
    QComboBox* c = static_cast<QComboBox*>(combo);
    TabbyCallbackData* data = createCallbackData();
    data->menuCb = cb;
    data->userData = userData;
    QObject::connect(c, static_cast<void(QComboBox::*)(int)>(&QComboBox::currentIndexChanged),
        [data](int index) {
            if (data->menuCb) data->menuCb(data->userData);
        });
}

// ---- Layout ----

TabbyLayout TabbyLayout_CreateVBox() {
    QVBoxLayout* layout = new QVBoxLayout();
    return static_cast<TabbyLayout>(layout);
}

TabbyLayout TabbyLayout_CreateHBox() {
    QHBoxLayout* layout = new QHBoxLayout();
    return static_cast<TabbyLayout>(layout);
}

void TabbyLayout_AddWidget(TabbyLayout layout, TabbyWidget widget) {
    QBoxLayout* l = static_cast<QBoxLayout*>(layout);
    QWidget* w = static_cast<QWidget*>(widget);
    l->addWidget(w);
}

void TabbyLayout_AddLayout(TabbyLayout layout, TabbyLayout child) {
    QBoxLayout* l = static_cast<QBoxLayout*>(layout);
    QLayout* c = static_cast<QLayout*>(child);
    l->addLayout(c);
}

void TabbyLayout_AddStretch(TabbyLayout layout) {
    QBoxLayout* l = static_cast<QBoxLayout*>(layout);
    l->addStretch();
}

void TabbyWidget_SetLayout(TabbyWidget widget, TabbyLayout layout) {
    QWidget* w = static_cast<QWidget*>(widget);
    QLayout* l = static_cast<QLayout*>(layout);
    w->setLayout(l);
}

// ---- Dialog ----

TabbyDialog TabbyDialog_Create(TabbyWindow parent, const char* title) {
    QWidget* p = parent ? static_cast<QWidget*>(parent) : nullptr;
    QDialog* dlg = new QDialog(p);
    dlg->setWindowTitle(QString::fromUtf8(title));
    return static_cast<TabbyDialog>(dlg);
}

void TabbyDialog_Destroy(TabbyDialog dlg) {
    QDialog* d = static_cast<QDialog*>(dlg);
    delete d;
}

int TabbyDialog_Exec(TabbyDialog dlg) {
    QDialog* d = static_cast<QDialog*>(dlg);
    return d->exec() == QDialog::Accepted ? 1 : 0;
}

void TabbyDialog_Show(TabbyDialog dlg) {
    QDialog* d = static_cast<QDialog*>(dlg);
    d->show();
}

void TabbyDialog_Close(TabbyDialog dlg) {
    QDialog* d = static_cast<QDialog*>(dlg);
    d->close();
}

void TabbyDialog_SetSize(TabbyDialog dlg, int width, int height) {
    QDialog* d = static_cast<QDialog*>(dlg);
    d->resize(width, height);
}

void TabbyDialog_SetLayout(TabbyDialog dlg, TabbyLayout layout) {
    QDialog* d = static_cast<QDialog*>(dlg);
    QLayout* l = static_cast<QLayout*>(layout);
    d->setLayout(l);
}

// ---- File Dialog ----

const char* TabbyFileDialog_GetOpenFileName(TabbyWindow parent, const char* caption, const char* dir, const char* filter) {
    QWidget* p = parent ? static_cast<QWidget*>(parent) : nullptr;
    QString result = QFileDialog::getOpenFileName(p,
        QString::fromUtf8(caption),
        QString::fromUtf8(dir),
        QString::fromUtf8(filter));
    return allocateString(result.toStdString());
}

const char* TabbyFileDialog_GetSaveFileName(TabbyWindow parent, const char* caption, const char* dir, const char* filter) {
    QWidget* p = parent ? static_cast<QWidget*>(parent) : nullptr;
    QString result = QFileDialog::getSaveFileName(p,
        QString::fromUtf8(caption),
        QString::fromUtf8(dir),
        QString::fromUtf8(filter));
    return allocateString(result.toStdString());
}

const char* TabbyFileDialog_GetExistingDirectory(TabbyWindow parent, const char* caption, const char* dir) {
    QWidget* p = parent ? static_cast<QWidget*>(parent) : nullptr;
    QString result = QFileDialog::getExistingDirectory(p,
        QString::fromUtf8(caption),
        QString::fromUtf8(dir));
    return allocateString(result.toStdString());
}

// ---- Utility ----

void Tabby_FreeString(const char* s) {
    if (s) {
        // Find and remove from allocated strings
        for (auto it = allocatedStrings.begin(); it != allocatedStrings.end(); ++it) {
            if (*it == s) {
                allocatedStrings.erase(it);
                delete[] s;
                return;
            }
        }
    }
}

const char* TabbyClipboard_Text() {
    QClipboard* clipboard = QApplication::clipboard();
    return allocateString(clipboard->text().toStdString());
}

void TabbyClipboard_SetText(const char* text) {
    QClipboard* clipboard = QApplication::clipboard();
    clipboard->setText(QString::fromUtf8(text));
}

int TabbyScreen_Width() {
    QScreen* screen = QApplication::primaryScreen();
    return screen ? screen->geometry().width() : 1920;
}

int TabbyScreen_Height() {
    QScreen* screen = QApplication::primaryScreen();
    return screen ? screen->geometry().height() : 1080;
}

double TabbyScreen_DevicePixelRatio() {
    QScreen* screen = QApplication::primaryScreen();
    return screen ? screen->devicePixelRatio() : 1.0;
}
