import "./style.css";
import "./app.css";
import "@xterm/xterm/css/xterm.css";

import { FitAddon } from "@xterm/addon-fit";
import { SearchAddon } from "@xterm/addon-search";
import { WebLinksAddon } from "@xterm/addon-web-links";
import { Terminal } from "@xterm/xterm";
import {
	GetAvailableShells,
	GetDefaultShell,
	GetHomeDir,
	GetHostname,
	GetUsername,
	PTYKill,
	PTYResize,
	PTYSpawn,
	PTYWrite,
	SetWindowTitle,
} from "../wailsjs/go/main/App";
import { EventsOn } from "../wailsjs/runtime/runtime";

// ===== GLOBALS =====
const tabs = [];
let activeTabId = null;
let tabCounter = 0;
let defaultShell = "";
let findVisible = false;

// ===== INIT =====
async function init() {
	defaultShell = await GetDefaultShell();
	buildUI();
	bindGlobalKeys();
	newTab();
}

// ===== BUILD UI =====
function buildUI() {
	document.querySelector("#app").innerHTML = `
        <div id="sidebar">
            <div id="sidebar-header">
                <div class="logo"><span>⌘</span> Tabby</div>
                <button class="btn-icon" id="btn-new-tab" title="New Tab (Ctrl+Shift+T)">+</button>
            </div>
            <div id="tab-list"></div>
            <div id="sidebar-footer">
                <div class="status-dot"></div>
                <div class="status-text">Ready</div>
            </div>
        </div>
        <div id="main-content">
            <div id="welcome">
                <div class="title"><span>Tabby</span> Go</div>
                <div class="shortcuts">
                    <div class="shortcut"><kbd>Ctrl+Shift+T</kbd> New Tab</div>
                    <div class="shortcut"><kbd>Ctrl+W</kbd> Close Tab</div>
                    <div class="shortcut"><kbd>Ctrl+Tab</kbd> Next Tab</div>
                    <div class="shortcut"><kbd>Ctrl+Shift+Tab</kbd> Previous Tab</div>
                    <div class="shortcut"><kbd>Ctrl+Shift+F</kbd> Find</div>
                    <div class="shortcut"><kbd>Alt+1-9</kbd> Switch to Tab N</div>
                </div>
            </div>
        </div>
    `;

	document.getElementById("btn-new-tab").onclick = () => newTab();
}

// ===== TAB CLASS =====
class Tab {
	constructor(shell) {
		this.id = `tab-${Date.now()}-${tabCounter++}`;
		this.ptyId = null;
		this.title = "Local Shell";
		this.shell = shell || defaultShell;
		this.exited = false;

		// Create terminal
		this.term = new Terminal({
			cursorBlink: true,
			cursorStyle: "bar",
			fontFamily:
				'"Cascadia Code", "Fira Code", Consolas, "Courier New", monospace',
			fontSize: 14,
			lineHeight: 1.2,
			theme: {
				background: "#1e1e1e",
				foreground: "#cccccc",
				cursor: "#aeafad",
				selectionBackground: "#264f78",
				black: "#1e1e1e",
				red: "#f44747",
				green: "#6a9955",
				yellow: "#d7ba7d",
				blue: "#569cd6",
				magenta: "#c586c0",
				cyan: "#4ec9b0",
				white: "#cccccc",
				brightBlack: "#666666",
				brightRed: "#f44747",
				brightGreen: "#6a9955",
				brightYellow: "#d7ba7d",
				brightBlue: "#569cd6",
				brightMagenta: "#c586c0",
				brightCyan: "#4ec9b0",
				brightWhite: "#e0e0e0",
			},
			allowProposedApi: true,
			scrollback: 10000,
			allowTransparency: false,
		});

		this.fitAddon = new FitAddon();
		this.searchAddon = new SearchAddon();
		this.webLinksAddon = new WebLinksAddon();

		this.term.loadAddon(this.fitAddon);
		this.term.loadAddon(this.searchAddon);
		this.term.loadAddon(this.webLinksAddon);

		// Create wrapper
		this.wrapper = document.createElement("div");
		this.wrapper.className = "terminal-wrapper";
		this.wrapper.id = this.id;
		document.getElementById("main-content").appendChild(this.wrapper);
		this.term.open(this.wrapper);

		// Create tab list item
		this.tabEl = document.createElement("div");
		this.tabEl.className = "tab-item";
		this.tabEl.innerHTML = `
            <span class="tab-icon">⌘</span>
            <span class="tab-title">${this.title}</span>
            <button class="tab-close">×</button>
        `;
		document.getElementById("tab-list").appendChild(this.tabEl);

		this.tabEl.onclick = (e) => {
			if (!e.target.classList.contains("tab-close")) this.activate();
		};
		this.tabEl.querySelector(".tab-close").onclick = (e) => {
			e.stopPropagation();
			this.close();
		};

		// Right-click context menu on tab
		this.tabEl.oncontextmenu = (e) => {
			e.preventDefault();
			showTabContextMenu(e, this);
		};

		// Terminal input → PTY write
		this.term.onData((data) => {
			if (this.ptyId && !this.exited) {
				PTYWrite(this.ptyId, btoa(data));
			}
		});

		// Terminal title change
		this.term.onTitleChange((title) => {
			if (title && title.length > 0) {
				this.setTitle(title);
			}
		});

		// Listen for PTY data
		EventsOn("pty.data", (params) => {
			const ptyId = params?.ptyId || params?.PTYID;
			if (ptyId === this.ptyId) {
				this.term.write(atob(params.data));
			}
		});

		// Listen for PTY exit
		EventsOn("pty.exit", (params) => {
			const ptyId = params?.ptyId || params?.PTYID;
			if (ptyId === this.ptyId) {
				this.exited = true;
				const code = params.exitCode ?? 0;
				this.term.writeln(
					`\r\n\x1b[1;33m[Process exited with code ${code}. Press Enter to close.]\x1b[0m`,
				);
				this.setTitle(`Exit (${code})`);
				this.tabEl.querySelector(".tab-icon").textContent = "✕";
				this.tabEl.querySelector(".tab-icon").style.color = "#f44747";
			}
		});

		// Handle Enter on exited process to close tab
		this.disposable = this.term.onData((data) => {
			if (this.exited && (data === "\r" || data === "\n")) {
				this.close();
			}
		});
	}

	async spawn() {
		try {
			const result = await PTYSpawn({
				command: this.shell,
				args: [],
				env: {},
				columns: this.term.cols,
				rows: this.term.rows,
			});
			this.ptyId = result.id;

			// Extract shell name for tab title
			const shellName = this.shell.split(/[/\\]/).pop().replace(".exe", "");
			this.setTitle(shellName);
		} catch (err) {
			this.term.writeln(`\x1b[1;31mError: ${err}\x1b[0m`);
		}
	}

	activate() {
		// Hide welcome screen
		const welcome = document.getElementById("welcome");
		if (welcome) welcome.style.display = "none";

		// Deactivate all
		tabs.forEach((t) => {
			t.wrapper.classList.remove("active");
			t.tabEl.classList.remove("active");
		});

		// Activate this
		this.wrapper.classList.add("active");
		this.tabEl.classList.add("active");
		activeTabId = this.id;
		this.term.focus();

		// Fit after activation
		requestAnimationFrame(() => {
			this.fitAddon.fit();
			if (this.ptyId && !this.exited) {
				PTYResize(this.ptyId, this.term.cols, this.term.rows);
			}
		});
	}

	close() {
		if (this.ptyId) {
			PTYKill(this.ptyId, "").catch(() => {});
		}
		this.term.dispose();
		this.wrapper.remove();
		this.tabEl.remove();

		const idx = tabs.indexOf(this);
		if (idx > -1) tabs.splice(idx, 1);

		// Activate another tab or show welcome
		if (activeTabId === this.id) {
			if (tabs.length > 0) {
				const next = tabs[Math.min(idx, tabs.length - 1)];
				next.activate();
			} else {
				activeTabId = null;
				const welcome = document.getElementById("welcome");
				if (welcome) welcome.style.display = "flex";
			}
		}
	}

	setTitle(title) {
		this.title = title;
		const el = this.tabEl.querySelector(".tab-title");
		if (el) el.textContent = title;
		if (activeTabId === this.id) {
			SetWindowTitle(`Tabby — ${title}`);
		}
	}

	findNext(query) {
		if (query) this.searchAddon.findNext(query);
	}

	findPrevious(query) {
		if (query) this.searchAddon.findPrevious(query);
	}
}

// ===== TAB MANAGEMENT =====
function newTab(shell) {
	const tab = new Tab(shell);
	tabs.push(tab);
	tab.activate();
	tab.spawn();
}

function getActiveTab() {
	return tabs.find((t) => t.id === activeTabId);
}

function switchToTab(index) {
	if (index >= 0 && index < tabs.length) {
		tabs[index].activate();
	}
}

function nextTab() {
	const idx = tabs.findIndex((t) => t.id === activeTabId);
	const next = (idx + 1) % tabs.length;
	tabs[next].activate();
}

function prevTab() {
	const idx = tabs.findIndex((t) => t.id === activeTabId);
	const prev = (idx - 1 + tabs.length) % tabs.length;
	tabs[prev].activate();
}

// ===== CONTEXT MENU =====
function showTabContextMenu(e, tab) {
	// Remove any existing context menu
	document.querySelectorAll(".context-menu").forEach((m) => m.remove());

	const menu = document.createElement("div");
	menu.className = "context-menu";
	menu.innerHTML = `
        <div class="context-menu-item" data-action="rename">✏️ Rename Tab</div>
        <div class="context-menu-item" data-action="duplicate">📋 Duplicate</div>
        <div class="context-menu-separator"></div>
        <div class="context-menu-item" data-action="close-others">✕ Close Others</div>
        <div class="context-menu-item" data-action="close">✕ Close <span class="shortcut-label">Ctrl+W</span></div>
    `;
	document.body.appendChild(menu);

	// Position
	const x = Math.min(e.clientX, window.innerWidth - 200);
	const y = Math.min(e.clientY, window.innerHeight - 200);
	menu.style.left = x + "px";
	menu.style.top = y + "px";

	const close = () => menu.remove();
	menu.onclick = async (ev) => {
		const item = ev.target.closest(".context-menu-item");
		if (!item) return;
		const action = item.dataset.action;
		close();
		switch (action) {
			case "rename": {
				const name = prompt("Tab name:", tab.title);
				if (name) tab.setTitle(name);
				break;
			}
			case "duplicate":
				newTab(tab.shell);
				break;
			case "close-others":
				[...tabs].forEach((t) => {
					if (t !== tab) t.close();
				});
				tab.activate();
				break;
			case "close":
				tab.close();
				break;
		}
	};
	setTimeout(() => {
		document.addEventListener("click", close, { once: true });
	}, 10);
}

// ===== KEYBOARD SHORTCUTS =====
function bindGlobalKeys() {
	document.addEventListener("keydown", (e) => {
		const ctrl = e.ctrlKey || e.metaKey;
		const shift = e.shiftKey;

		// Ctrl+Shift+T — New Tab
		if (ctrl && shift && e.key === "T") {
			e.preventDefault();
			newTab();
			return;
		}

		// Ctrl+W — Close Tab
		if (ctrl && !shift && e.key === "w") {
			e.preventDefault();
			const tab = getActiveTab();
			if (tab) tab.close();
			return;
		}

		// Ctrl+Tab — Next Tab
		if (ctrl && e.key === "Tab" && !shift) {
			e.preventDefault();
			nextTab();
			return;
		}

		// Ctrl+Shift+Tab — Prev Tab
		if (ctrl && e.key === "Tab" && shift) {
			e.preventDefault();
			prevTab();
			return;
		}

		// Alt+1-9 — Switch to tab
		if (e.altKey && e.key >= "1" && e.key <= "9") {
			e.preventDefault();
			switchToTab(parseInt(e.key) - 1);
			return;
		}

		// Ctrl+Shift+F — Find
		if (ctrl && shift && e.key === "F") {
			e.preventDefault();
			toggleFind();
			return;
		}

		// Ctrl+Plus/Minus — Font size
		if (ctrl && (e.key === "=" || e.key === "+")) {
			e.preventDefault();
			changeFontSize(1);
			return;
		}
		if (ctrl && e.key === "-") {
			e.preventDefault();
			changeFontSize(-1);
			return;
		}
	});
}

// ===== FIND BAR =====
function toggleFind() {
	let bar = document.getElementById("find-bar");
	if (bar) {
		bar.remove();
		findVisible = false;
		const tab = getActiveTab();
		if (tab) tab.term.focus();
		return;
	}
	findVisible = true;
	bar = document.createElement("div");
	bar.id = "find-bar";
	bar.style.cssText = `
        position: absolute; top: 0; right: 0; z-index: 100;
        background: #2d2d2d; border-bottom: 1px solid #3a3a3a; border-left: 1px solid #3a3a3a;
        padding: 6px 12px; display: flex; gap: 8px; align-items: center; border-radius: 0 0 0 8px;
    `;
	bar.innerHTML = `
        <input type="text" id="find-input" placeholder="Find..." style="
            background: #1e1e1e; border: 1px solid #3a3a3a; color: #ccc; padding: 4px 8px;
            border-radius: 4px; font-size: 13px; width: 200px; outline: none;
        ">
        <button class="btn-icon" id="find-prev" title="Previous (Shift+Enter)">↑</button>
        <button class="btn-icon" id="find-next" title="Next (Enter)">↓</button>
        <button class="btn-icon" id="find-close" title="Close (Esc)">×</button>
    `;
	document.getElementById("main-content").appendChild(bar);

	const input = document.getElementById("find-input");
	input.focus();

	input.onkeydown = (e) => {
		const tab = getActiveTab();
		if (!tab) return;
		if (e.key === "Enter") {
			e.preventDefault();
			if (e.shiftKey) {
				tab.findPrevious(input.value);
			} else {
				tab.findNext(input.value);
			}
		}
		if (e.key === "Escape") {
			toggleFind();
		}
	};

	document.getElementById("find-next").onclick = () => {
		const tab = getActiveTab();
		if (tab) tab.findNext(input.value);
		input.focus();
	};

	document.getElementById("find-prev").onclick = () => {
		const tab = getActiveTab();
		if (tab) tab.findPrevious(input.value);
		input.focus();
	};

	document.getElementById("find-close").onclick = () => toggleFind();
}

// ===== FONT SIZE =====
let fontSize = 14;
function changeFontSize(delta) {
	fontSize = Math.max(8, Math.min(32, fontSize + delta));
	tabs.forEach((tab) => {
		tab.term.options.fontSize = fontSize;
		tab.fitAddon.fit();
		if (tab.ptyId) PTYResize(tab.ptyId, tab.term.cols, tab.term.rows);
	});
}

// ===== WINDOW RESIZE =====
window.addEventListener("resize", () => {
	const tab = getActiveTab();
	if (tab) {
		tab.fitAddon.fit();
		if (tab.ptyId && !tab.exited) {
			PTYResize(tab.ptyId, tab.term.cols, tab.term.rows);
		}
	}
});

// ===== START =====
init();
