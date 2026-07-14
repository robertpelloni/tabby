import "./style.css";

import "./app.css";

import "@xterm/xterm/css/xterm.css";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { SearchAddon } from "@xterm/addon-search";
import { WebLinksAddon } from "@xterm/addon-web-links";
import { Unicode11Addon } from "@xterm/addon-unicode11";
import { WebglAddon } from "@xterm/addon-webgl";

// Global error handler to prevent WebView2 renderer crashes (sad face emoji)
window.addEventListener("error", (event) => {
	console.error("Global error caught:", event.error);
	event.preventDefault();
});

window.addEventListener("unhandledrejection", (event) => {
	console.error("Unhandled promise rejection:", event.reason);
	event.preventDefault();
});

import {
	PTYSpawn,
	PTYWrite,
	PTYResize,
	PTYKill,
	SSHConnect,
	SSHStartShell,
	SSHWrite,
	SSHResize,
	SSHClose,
	SSHAddForward,
	SSHRemoveForward,
	SSHListForwards,
	HandleHostKeyResponse,
	HandleKeyboardInteractiveResponse,
	SerialOpen,
	SerialWrite,
	SerialClose,
	SerialListPorts,
	TelnetConnect,
	TelnetWrite,
	TelnetResize,
	TelnetClose,
	GetDefaultShell,
	GetAvailableShells,
	GetColorSchemes,
	SetWindowTitle,
	GetSettings,
	SaveSettings,
	ResetSettings,
	SaveSessionState,
	LoadSessionState,
	ClearSessionState,
	GetProfiles,
	SaveProfiles,
	SFTPOpen,
	SFTPList,
	SFTPDownload,
	SFTPUpload,
	SFTPDelete,
	SFTPRename,
	SFTPMkdir,
	SFTPStat,
	SFTPClose,
	SFTPRmdir,
	SFTPReadDir,
	SFTPMkdirAll,
	SFTPChmod,
	SFTPReadlink,
	SFTPSymlink,
	ImportSSHConfig,
	GetUsername,
	GetHostname,
	GetHomeDir,
	GetPlatform,
	GetNotifications,
	GetUnreadNotifications,
	MarkNotificationRead,
	ClearNotifications,
	SelectDirectory,
	StoreCredential,
	GetCredential,
	DeleteCredential,
	IsOSKeyringAvailable,
	CheckForUpdates,
	GetUpdateStatus,
	GetAuditLogPath,
	OpenInBrowser,
} from "../wailsjs/go/main/App";

import {
	EventsOn,
	WindowSetAlwaysOnTop,
	WindowMinimise,
	WindowMaximise,
	WindowUnmaximise,
	WindowIsMaximised,
	WindowToggleMaximise,
	Quit,
	ClipboardSetText,
	ClipboardGetText,
} from "../wailsjs/runtime/runtime";

// ===== GLOBALS =====

const COLOR_SCHEMES = {};

window.__serialDataHandlers = [];
window.__serialExitHandlers = [];

window.__telnetDataHandlers = [];
window.__telnetExitHandlers = [];

let schemeNames = [];

const tabs = [];

let activeTabId = null;

let tabCounter = 0;

let defaultShell = "";

let availableShells = [];

let findVisible = false;

let fontSize = 14;

let activeSplitPane = null;

let settings = {};

let savedProfiles = [];

const broadcastMode = false;

// Broadcast input to all terminals
function broadcastInput(data, sourceTab) {
	if (!broadcastMode) return;
	for (const t of tabs) {
		if (t === sourceTab || t.exited) continue;
		if (t.ptyId) PTYWrite(t.ptyId, btoa(data));
		else if (t.isSSH && t.sshConnectionId && t.sshSessionId)
			SSHWrite({
				connectionId: t.sshConnectionId,
				sessionId: t.sshSessionId,
				data: btoa(data),
			});
		else if (t.isSerial && t.serialId) SerialWrite(t.serialId, btoa(data));
		else if (t.isTelnet && t.telnetConnectionId)
			TelnetWrite(t.telnetConnectionId, btoa(data));
	}
}

// Strip ANSI escape sequences for clean log output
function stripAnsi(str) {
	return str.replace(/\x1b[[0-9;]*[a-zA-Z]/g, "");
}

// Decode base64 string to Uint8Array for proper UTF-8 handling
function b64ToBytes(b64) {
	const binary = atob(b64);
	const bytes = new Uint8Array(binary.length);
	for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
	return bytes;
}

// ===== INIT =====

// Idle connection monitor

setInterval(() => {
	const timeout = settings.IdleTimeout || 0;

	if (timeout <= 0) return;

	tabs.forEach((tab) => {
		if (tab.lastActivity && Date.now() - tab.lastActivity > timeout * 60000) {
			if (
				tab.status === "connected" &&
				(tab.isSSH || tab.isSerial || tab.isTelnet)
			) {
				showToast("Idle timeout: " + tab.title, "info");

				tab.close();
			}
		}
	});
}, 60000);

async function init() {
	defaultShell = await GetDefaultShell();

	try {
		availableShells = await GetAvailableShells();
	} catch (_) {
		availableShells = [];
	}

	try {
		settings = await GetSettings();
	} catch (_) {
		settings = {};
	}

	if (settings.FontSize) fontSize = settings.FontSize;

	// Tab bar drag-and-drop reordering

	const tabBar = document.getElementById("tab-bar");

	if (tabBar) {
		tabBar.addEventListener("dragover", (e) => {
			e.preventDefault();
			const dragging = tabBar.querySelector(".dragging");
			if (dragging) {
				const afterEl = getDragAfterElement(tabBar, e.clientX);
				if (afterEl) tabBar.insertBefore(dragging, afterEl);
				else tabBar.appendChild(dragging);
			}
		});
	}

	function getDragAfterElement(container, x) {
		const elements = [
			...container.querySelectorAll(".tab-item:not(.dragging)"),
		];

		return elements.reduce(
			(closest, child) => {
				const box = child.getBoundingClientRect();

				const offset = x - box.left - box.width / 2;

				if (offset < 0 && offset > closest.offset)
					return { offset, element: child };

				return closest;
			},
			{ offset: Number.NEGATIVE_INFINITY },
		).element;
	}

	const savedCSS = localStorage.getItem("tabby-custom-css");

	// Check for updates silently on startup

	setTimeout(() => CheckForUpdates().catch(() => {}), 5000);

	if (savedCSS) applyCustomCSS(savedCSS);

	// Load color schemes from Go backend

	try {
		const schemes = await GetColorSchemes();

		if (schemes && schemes.length) {
			schemes.forEach((s) => {
				COLOR_SCHEMES[s.Name] = s;
			});

			schemeNames = schemes.map((s) => s.Name);
		}
	} catch (_) {
		schemeNames = [
			"Tabby Default",
			"Tabby Default Light",
			"Dracula",
			"Solarized Dark",
			"Solarized Light",
			"Monokai",
			"Nord",
			"One Half Dark",
			"One Half Light",
			"Gruvbox Dark",
			"Tokyo Night",
			"Catppuccin Mocha",
			"Catppuccin Latte",
			"Ayu Dark",
			"Atom One Light",
			"Batman",
		];
	}

	// Load connection profiles

	try {
		savedProfiles = (await GetProfiles()) || [];
	} catch (_) {
		savedProfiles = [];
	}

	buildUI();

	bindGlobalKeys();

	applySettingsToUI();

	const restored = await restoreSession();

	// Custom titlebar handlers
	const btnClose = document.getElementById("btn-close");
	const btnMaximize = document.getElementById("btn-maximize");
	const btnMinimize = document.getElementById("btn-minimize");

	if (btnClose) {
		btnClose.onclick = () => Quit();
	}
	if (btnMaximize) {
		btnMaximize.onclick = () => {
			WindowIsMaximised().then((isMaximised) => {
				if (isMaximised) {
					WindowUnmaximise();
				} else {
					WindowMaximise();
				}
			});
		};
	}
	if (btnMinimize) {
		btnMinimize.onclick = () => WindowMinimise();
	}

	// Handle maximize/restore state
	const updateMaximizeBtn = () => {
		WindowIsMaximised().then((isMaximised) => {
			const btn = document.getElementById("btn-maximize");
			if (btn) {
				btn.innerHTML = isMaximised
					? '<svg width="12" height="12" viewBox="0 0 12 12"><rect x="2" y="2" width="4" height="4" stroke="currentColor" fill="none" stroke-width="1"/><rect x="6" y="6" width="4" height="4" stroke="currentColor" fill="none" stroke-width="1"/></svg>'
					: '<svg width="12" height="12" viewBox="0 0 12 12"><rect x="2" y="2" width="8" height="8" stroke="currentColor" fill="none" stroke-width="1"/></svg>';
			}
		});
	};
	updateMaximizeBtn();
	window.addEventListener("resize", updateMaximizeBtn);

	// Listen for Wails resize events to keep button state accurate
	setInterval(updateMaximizeBtn, 2000);

	if (!restored) newTab();
}

// ===== COLOR SCHEME HELPERS =====

function getColorSchemeTheme(name) {
	const scheme = COLOR_SCHEMES[name];
	if (!scheme) return null;
	const c = scheme.Colors || [];
	return {
		background: scheme.Background,
		foreground: scheme.Foreground,
		cursor: scheme.Cursor,
		cursorAccent: scheme.CursorAccent || undefined,
		selectionBackground: scheme.Selection || undefined,
		selectionForeground: scheme.SelectionForeground || undefined,
		black: c[0],
		red: c[1],
		green: c[2],
		yellow: c[3],
		blue: c[4],
		magenta: c[5],
		cyan: c[6],
		white: c[7],
		brightBlack: c[8],
		brightRed: c[9],
		brightGreen: c[10],
		brightYellow: c[11],
		brightBlue: c[12],
		brightMagenta: c[13],
		brightCyan: c[14],
		brightWhite: c[15],
	};
}

function isSchemeLight(name) {
	const bg =
		(COLOR_SCHEMES[name] && COLOR_SCHEMES[name].Background) || "#171717";
	return isLightColor(bg);
}

function isLightColor(hex) {
	if (!hex || hex.length < 7 || hex[0] !== "#") return false;
	const r = parseInt(hex.slice(1, 3), 16),
		g = parseInt(hex.slice(3, 5), 16),
		b = parseInt(hex.slice(5, 7), 16);
	return (0.299 * r + 0.587 * g + 0.114 * b) / 255 > 0.5;
}

function applyColorScheme(name) {
	const theme = getColorSchemeTheme(name);
	if (!theme) return;
	tabs.forEach((t) => {
		if (t.term) t.term.options.theme = theme;
	});
	if (isSchemeLight(name)) {
		document.body.classList.add("light-theme");
		document.body.classList.remove("dark-theme");
	} else {
		document.body.classList.add("dark-theme");
		document.body.classList.remove("light-theme");
	}
	renderColorSchemePreview(name);
}

function applyBackgroundColor(color) {
	// Apply to main-content background
	const mc = document.getElementById("main-content");
	if (mc) mc.style.background = color;
	// Apply to all terminal themes
	tabs.forEach((t) => {
		if (t.term && t.term.options.theme) {
			t.term.options.theme = { ...t.term.options.theme, background: color };
		}
	});
}

function renderColorSchemePreview(name) {
	const container = document.getElementById("color-scheme-preview");
	if (!container) return;
	const scheme = COLOR_SCHEMES[name];
	if (!scheme) {
		container.innerHTML = "";
		return;
	}
	const c = scheme.Colors || [];
	const all = [scheme.Background, scheme.Foreground, scheme.Cursor, ...c];
	container.innerHTML = all
		.map(
			(color) =>
				`<div style="width:16px;height:16px;border-radius:3px;background:${color};border:1px solid #3a3a3a;" title="${color}"></div>`,
		)
		.join("");
}

// ===== SSH DIALOG =====

function openSSHDialog() {
	document.getElementById("ssh-dialog").classList.add("active");
	document.getElementById("ssh-connection-mode").value = "direct";
	document.getElementById("ssh-jump-host-group").style.display = "none";
	document.getElementById("ssh-proxy-cmd-group").style.display = "none";
	document.getElementById("ssh-socks-proxy-group").style.display = "none";
	document.getElementById("ssh-http-proxy-group").style.display = "none";
	document.getElementById("ssh-host").focus();
}

function closeSSHDialog() {
	document.getElementById("ssh-dialog").classList.remove("active");
	const t = getActiveTab();
	if (t) t.term.focus();
}
function toggleSSHConnectionMode() {
	const mode = document.getElementById("ssh-connection-mode").value;
	document.getElementById("ssh-jump-host-group").style.display =
		mode === "jumpHost" ? "block" : "none";
	document.getElementById("ssh-proxy-cmd-group").style.display =
		mode === "proxyCommand" ? "block" : "none";
	document.getElementById("ssh-socks-proxy-group").style.display =
		mode === "socksProxy" ? "block" : "none";
	document.getElementById("ssh-http-proxy-group").style.display =
		mode === "httpProxy" ? "block" : "none";
}

async function doSSHConnect() {
	const host = document.getElementById("ssh-host").value.trim();
	const port = parseInt(document.getElementById("ssh-port").value) || 22;
	const user = document.getElementById("ssh-user").value.trim();
	const auth = document.getElementById("ssh-auth").value;
	if (!host) {
		showToast("Host is required", "error");
		return;
	}
	closeSSHDialog();
	showStatus("Connecting to " + host + "...");

	const tab = new Tab(defaultShell);

	tab.connectionType = "ssh";

	tabs.push(tab);

	tab.activate();

	tab.ptyId = null;

	const spinner = document.createElement("div");

	spinner.className = "connecting-spinner";

	spinner.style.cssText =
		"position:absolute;top:50%;left:50%;transform:translate(-50%,-50%);width:40px;height:40px;border:3px solid var(--border);border-top:3px solid var(--accent);border-radius:50%;animation:spin 1s linear infinite;z-index:15;";

	tab.wrapper.appendChild(spinner);

	try {
		const authParams = { type: auth };

		if (auth === "password") {
			authParams.password = document.getElementById("ssh-password").value;
		} else if (auth === "publicKey") {
			const keyPath = document.getElementById("ssh-key-path").value;

			authParams.privateKeyPaths = keyPath ? [keyPath] : [];
		}

		const jumpHostValue =
			document.getElementById("ssh-connection-mode").value === "jumpHost"
				? document.getElementById("ssh-jump-host").value.trim()
				: "";

		let jumpHostParams = null;

		if (jumpHostValue) {
			const jhParts = jumpHostValue.split(":");

			const jhHost = jhParts[0];

			const jhPort = parseInt(jhParts[1]) || 22;

			const jhUser = jhHost.includes("@") ? jhHost.split("@")[0] : user;

			const jhCleanHost = jhHost.includes("@") ? jhHost.split("@")[1] : jhHost;

			jumpHostParams = {
				host: jhCleanHost,
				port: jhPort,
				user: jhUser,
				auth: { type: "agent" },
				keepaliveInterval: 30,
				readyTimeout: 15000,
			};
		}

		const keepalive =
			parseInt(document.getElementById("ssh-keepalive").value) || 30;

		const timeout =
			parseInt(document.getElementById("ssh-timeout").value) || 15;

		const agentForward = document.getElementById("ssh-agent-forward").checked;

		const connectionMode = document.getElementById("ssh-connection-mode").value;
		const x11 = document.getElementById("ssh-x11").checked;
		const skipBanner = document.getElementById("ssh-skip-banner").checked;
		const sshParams = {
			host: host,
			port: port,
			user: user,
			auth: authParams,
			keepaliveInterval: keepalive,
			keepaliveCountMax: 3,
			readyTimeout: timeout * 1000,
			agentForward: agentForward,
			x11: x11,
			skipBanner: skipBanner,
			jumpHost: connectionMode === "jumpHost" ? jumpHostParams : null,
			proxyCommand:
				connectionMode === "proxyCommand"
					? document.getElementById("ssh-proxy-command").value.trim()
					: "",
			socksProxyHost:
				connectionMode === "socksProxy"
					? document.getElementById("ssh-socks-host").value.trim()
					: "",
			socksProxyPort:
				connectionMode === "socksProxy"
					? parseInt(document.getElementById("ssh-socks-port").value) || 1080
					: 0,
			httpProxyHost:
				connectionMode === "httpProxy"
					? document.getElementById("ssh-http-proxy-host").value.trim()
					: "",
			httpProxyPort:
				connectionMode === "httpProxy"
					? parseInt(document.getElementById("ssh-http-proxy-port").value) ||
						3128
					: 0,
		};

		const result = await SSHConnect(sshParams);

		setTabStatus(tab, "connected");
		logConnection(tab, "SSH connected to " + host);

		const sp = tab.wrapper.querySelector(".connecting-spinner");
		if (sp) sp.remove();

		showToast("Connected to " + host, "success");

		tab.sshConnectionId = result.connectionId;

		tab.sessionData = JSON.stringify({
			type: "ssh",
			host,
			port,
			user,
			auth: authParams,
		});
		let jumpLabel = "";

		if (result.jumpChain && result.jumpChain.length > 0) {
			jumpLabel = " (via " + result.jumpChain.join(" -> ") + ")";
		}

		tab.setTitle(user + "@" + host + jumpLabel);
		const shellResult = await SSHStartShell({
			connectionId: result.connectionId,
			columns: tab.term.cols,
			rows: tab.term.rows,
			terminal: "xterm-256color",
		});
		tab.sshSessionId = shellResult.sessionId;
		tab.isSSH = true;
		tab.sshHost = host;
		tab.sshPort = port;
		tab.sshUser = user;
		tab.term.onData((data) => {
			if (data.includes("\n") && data.trim().split("\n").length > 1) {
				if (
					settings.PasteWarning !== false &&
					!confirm(
						"Paste multi-line content? (" +
							data.trim().split("\n").length +
							" lines)",
					)
				)
					return;
			}
			if (tab.sshConnectionId && tab.sshSessionId) {
				SSHWrite({
					connectionId: tab.sshConnectionId,
					sessionId: tab.sshSessionId,
					data: btoa(data),
				});
				broadcastInput(data, tab);
			}
		});

		setupInputProcessing(tab.term, tab);

		const loginScriptEl = document.getElementById("ssh-login-script");

		if (loginScriptEl && loginScriptEl.value.trim())
			runLoginScript(tab, loginScriptEl.value);

		let statusText = "SSH - " + user + "@" + host;

		if (result.jumpChain && result.jumpChain.length > 0) {
			statusText += " via " + result.jumpChain.join(" -> ");
		}

		showStatus(statusText);
		if (
			document.getElementById("ssh-save-profile") &&
			document.getElementById("ssh-save-profile").checked
		) {
			savedProfiles.push({
				id: "ssh-" + Date.now(),
				type: "ssh",
				name: user + "@" + host,
				options: {
					host,
					port,
					user,
					auth,
					privateKeys: authParams.privateKeyPaths || [],
				},
				createdAt: new Date().toISOString(),
				updatedAt: new Date().toISOString(),
			});
			SaveProfiles(savedProfiles).catch(() => {});
			renderProfiles();
		}
	} catch (err) {
		const sp = tab.wrapper.querySelector(".connecting-spinner");
		if (sp) sp.remove();
		showToast("SSH failed: " + err, "error");
		showStatus("SSH failed - " + host);
	}
}

// ===== SERIAL PORT DIALOG =====

function openSerialDialog() {
	document.getElementById("serial-dialog").classList.add("active");

	refreshSerialPorts();
}

function closeSerialDialog() {
	document.getElementById("serial-dialog").classList.remove("active");

	const t = getActiveTab();

	if (t) t.term.focus();
}

async function refreshSerialPorts() {
	const select = document.getElementById("serial-port-select");

	select.innerHTML = '<option value="">-- Scanning... --</option>';

	try {
		const ports = await SerialListPorts();

		select.innerHTML = '<option value="">-- Select Port --</option>';

		if (ports && ports.length) {
			ports.forEach((p) => {
				const opt = document.createElement("option");

				opt.value = p.Name;

				opt.textContent = p.Name;

				select.appendChild(opt);
			});
		} else {
			select.innerHTML = '<option value="">-- No ports found --</option>';
		}
	} catch (err) {
		select.innerHTML = '<option value="">-- Error scanning --</option>';

		showToast("Failed to list serial ports: " + err, "error");
	}
}

async function doSerialConnect() {
	const port = document.getElementById("serial-port-select").value;

	const baud = parseInt(document.getElementById("serial-baud").value) || 115200;

	const dataBits =
		parseInt(document.getElementById("serial-data-bits").value) || 8;

	const stopBits =
		parseInt(document.getElementById("serial-stop-bits").value) || 1;

	const parity = document.getElementById("serial-parity").value;

	if (!port) {
		showToast("Select a serial port", "error");
		return;
	}

	closeSerialDialog();

	showStatus("Connecting to " + port + "...");

	try {
		const tab = new Tab(defaultShell);

		tabs.push(tab);

		tab.activate();

		tab.ptyId = null;

		setTabStatus(tab, "connecting");
		logConnection(tab, "Opening serial port...");

		const result = await SerialOpen({
			port,
			baudRate: baud,
			dataBits,
			stopBits,
			parity,
		});

		setTabStatus(tab, "connected");
		logConnection(tab, "Serial port opened: " + port);
		showToast("Serial connected: " + port, "success");

		tab.serialId = result.ID || result.id;

		tab.isSerial = true;

		tab.serialPort = port;

		tab.serialBaud = baud;

		tab.sessionData = JSON.stringify({ type: "serial", port, baudRate: baud });

		tab.setTitle(port.split("/").pop().split("\\").pop());

		tab.serialDataHandler = (params) => {
			if ((params.serialId || params.SerialID) === tab.serialId)
				tab.term.write(b64ToBytes(params.data || params.Data));
		};

		window.__serialDataHandlers = window.__serialDataHandlers || [];
		window.__serialDataHandlers.push(tab.serialDataHandler);

		tab.serialExitHandler = (params) => {
			if ((params.serialId || params.SerialID) === tab.serialId) {
				tab.exited = true;
				setTabStatus(tab, "disconnected");
				tab.term.writeln(`

[1;33m[Serial port closed][0m`);
				tab.setTitle(tab.title + " [disconnected]");
			}
		};

		window.__serialExitHandlers = window.__serialExitHandlers || [];
		window.__serialExitHandlers.push(tab.serialExitHandler);

		tab.term.onData((data) => {
			if (tab.serialId) {
				SerialWrite(tab.serialId, btoa(data));
				broadcastInput(data, tab);
			}
		});

		setupInputProcessing(tab.term, tab);

		showStatus("Serial - " + port + " @ " + baud);

		if (
			document.getElementById("serial-save-profile") &&
			document.getElementById("serial-save-profile").checked
		) {
			savedProfiles.push({
				id: "serial-" + Date.now(),
				type: "serial",
				name: port + " @ " + baud,
				options: { port, baudRate: baud, dataBits, stopBits, parity },
				createdAt: new Date().toISOString(),
				updatedAt: new Date().toISOString(),
			});

			SaveProfiles(savedProfiles).catch(() => {});

			renderProfiles();
		}
	} catch (err) {
		showToast("Serial failed: " + err, "error");

		showStatus("Serial failed - " + port);
	}
}

// ===== TELNET DIALOG =====

function openTelnetDialog() {
	document.getElementById("telnet-dialog").classList.add("active");

	document.getElementById("telnet-host").focus();
}

function closeTelnetDialog() {
	document.getElementById("telnet-dialog").classList.remove("active");

	const t = getActiveTab();

	if (t) t.term.focus();
}

async function doTelnetConnect() {
	const host = document.getElementById("telnet-host").value.trim();

	const port = parseInt(document.getElementById("telnet-port").value) || 23;

	if (!host) {
		showToast("Host is required", "error");
		return;
	}

	closeTelnetDialog();

	showStatus("Connecting to " + host + ":" + port + "...");

	try {
		const result = await TelnetConnect(host, port);

		const tab = new Tab(defaultShell);

		tabs.push(tab);

		tab.activate();

		tab.ptyId = null;

		setTabStatus(tab, "connected");
		logConnection(tab, "Telnet connected to " + host + ":" + port);
		showToast("Telnet connected: " + host, "success");

		tab.telnetConnectionId = result.ConnectionID || result.connectionId;

		tab.sessionData = JSON.stringify({ type: "telnet", host, port });

		tab.isTelnet = true;

		tab.telnetHost = host;

		tab.telnetPort = port;

		tab.setTitle(host + ":" + port);

		tab.telnetDataHandler = (params) => {
			const cid = params.ConnectionID || params.connectionId;

			if (cid === tab.telnetConnectionId)
				tab.term.write(b64ToBytes(params.Data || params.data));
		};

		window.__telnetDataHandlers = window.__telnetDataHandlers || [];
		window.__telnetDataHandlers.push(tab.telnetDataHandler);

		tab.telnetExitHandler = (params) => {
			const cid = params.ConnectionID || params.connectionId;
			if (cid === tab.telnetConnectionId) {
				tab.exited = true;
				setTabStatus(tab, "disconnected");
				tab.term.writeln(`

[1;33m[Telnet connection closed][0m`);
				tab.setTitle(tab.title + " [disconnected]");
			}
		};

		window.__telnetExitHandlers = window.__telnetExitHandlers || [];
		window.__telnetExitHandlers.push(tab.telnetExitHandler);

		tab.term.onData((data) => {
			if (
				data.includes(String.fromCharCode(10)) &&
				data.trim().split(String.fromCharCode(10)).length > 1
			) {
				if (
					settings.PasteWarning !== false &&
					!confirm(
						"Paste multi-line content? (" +
							data.trim().split(String.fromCharCode(10)).length +
							" lines)",
					)
				)
					return;
			}

			if (tab.telnetConnectionId) {
				TelnetWrite(tab.telnetConnectionId, btoa(data));
				broadcastInput(data, tab);
			}
		});

		setupInputProcessing(tab.term, tab);

		showStatus("Telnet - " + host + ":" + port);

		if (
			document.getElementById("telnet-save-profile") &&
			document.getElementById("telnet-save-profile").checked
		) {
			savedProfiles.push({
				id: "telnet-" + Date.now(),
				type: "telnet",
				name: host + ":" + port,
				options: { host, port },
				createdAt: new Date().toISOString(),
				updatedAt: new Date().toISOString(),
			});

			SaveProfiles(savedProfiles).catch(() => {});

			renderProfiles();
		}
	} catch (err) {
		showToast("Telnet failed: " + err, "error");

		showStatus("Telnet failed - " + host);
	}
}

// ===== PORT FORWARDING DIALOG =====

function openForwardDialog(sshConnectionId) {
	document.getElementById("forward-dialog").classList.add("active");

	document.getElementById("forward-conn-id").value = sshConnectionId || "";

	document.getElementById("forward-type").value = "local";

	document.getElementById("forward-local-addr").value = "";

	document.getElementById("forward-remote-addr").value = "";

	toggleForwardFields();

	loadForwardList(sshConnectionId);
}

function closeForwardDialog() {
	document.getElementById("forward-dialog").classList.remove("active");

	const t = getActiveTab();

	if (t) t.term.focus();
}

function toggleForwardFields() {
	const type = document.getElementById("forward-type").value;

	const lg = document.getElementById("forward-local-group");

	const rg = document.getElementById("forward-remote-group");

	const la = document.getElementById("forward-local-addr");

	if (type === "dynamic") {
		lg.style.display = "block";
		rg.style.display = "none";

		la.placeholder = "localhost:1080 (SOCKS5)";
	} else if (type === "local") {
		lg.style.display = "block";
		rg.style.display = "block";

		la.placeholder = "localhost:8080";

		document.getElementById("forward-remote-addr").placeholder =
			"remotehost:80";
	} else {
		lg.style.display = "block";
		rg.style.display = "block";

		la.placeholder = "remotehost:8080";

		document.getElementById("forward-remote-addr").placeholder = "localhost:80";
	}
}

async function doAddForward() {
	const connId = document.getElementById("forward-conn-id").value;

	const type = document.getElementById("forward-type").value;

	const localAddr = document.getElementById("forward-local-addr").value.trim();

	const remoteAddr = document
		.getElementById("forward-remote-addr")
		.value.trim();

	if (!connId) {
		showToast("No SSH connection", "error");
		return;
	}

	try {
		const params = { connectionId: connId, type };

		if (type === "local") {
			const [lh, lp] = parseAddr(localAddr, "localhost");

			const [rh, rp] = parseAddr(remoteAddr, "localhost");

			params.localHost = lh;
			params.localPort = lp;

			params.remoteHost = rh;
			params.remotePort = rp;
		} else if (type === "remote") {
			const [rh, rp] = parseAddr(localAddr, "0.0.0.0");

			const [lh, lp] = parseAddr(remoteAddr, "localhost");

			params.remoteHost = rh;
			params.remotePort = rp;

			params.localHost = lh;
			params.localPort = lp;
		} else {
			const [lh, lp] = parseAddr(localAddr, "localhost");

			params.localHost = lh;
			params.localPort = lp;
		}

		await SSHAddForward(params);

		showToast("Forward added", "success");

		loadForwardList(connId);
	} catch (err) {
		showToast("Failed to add forward: " + err, "error");
	}
}

async function loadForwardList(connId) {
	const container = document.getElementById("forward-list");

	if (!connId) {
		container.innerHTML = "";
		return;
	}

	try {
		const forwards = await SSHListForwards(connId);

		if (!forwards || forwards.length === 0) {
			container.innerHTML =
				'<div style="color:#666;font-size:12px;">No active forwards</div>';
		} else {
			container.innerHTML = forwards
				.map((f, i) => {
					const ft = f.Type || f.type || "local";

					const icon = ft === "local" ? "-L" : ft === "remote" ? "-R" : "-D";

					const lh = f.LocalHost || f.localHost || "";

					const lp = f.LocalPort || f.localPort || "";

					const rh = f.RemoteHost || f.remoteHost || "";

					const rp = f.RemotePort || f.remotePort || "";

					const label =
						ft === "dynamic"
							? icon + " " + lh + ":" + lp
							: icon + " " + lh + ":" + lp + " \u2192 " + rh + ":" + rp;

					return (
						'<div class="profile-editor-item"><span style="font-size:12px;color:#ccc;">' +
						label +
						'</span><button class="btn-icon forward-remove" data-idx="' +
						i +
						'" title="Remove">\u00d7</button></div>'
					);
				})
				.join("");

			container.querySelectorAll(".forward-remove").forEach((btn) => {
				btn.onclick = async () => {
					const idx = parseInt(btn.dataset.idx);

					try {
						await SSHRemoveForward({ connectionId: connId, forwardIndex: idx });

						showToast("Forward removed", "info");

						loadForwardList(connId);
					} catch (err) {
						showToast("Remove failed: " + err, "error");
					}
				};
			});
		}
	} catch (err) {
		container.innerHTML =
			'<div style="color:#f44747;font-size:12px;">Error: ' + err + "</div>";
	}
}

function parseAddr(addr, defaultHost) {
	const parts = addr.split(":");

	if (parts.length === 2) return [parts[0], parseInt(parts[1]) || 0];

	if (parts.length === 1 && parts[0])
		return [defaultHost, parseInt(parts[0]) || 0];

	return [defaultHost, 0];
}

// ===== SSH AUTH CHALLENGE HANDLERS =====

async function handleKeyboardInteractive(params) {
	try {
		const connID = params.connectionId || params.ConnectionID;

		const name = params.name || "";

		const instruction = params.instruction || "";

		const prompts = params.prompts || [];

		const responses = [];

		for (const p of prompts) {
			const answer = await showPasswordDialog(
				"SSH Authentication",
				p.prompt || "Response:",
			);

			if (answer === null) {
				responses.push("");
			} else {
				responses.push(answer);
			}
		}

		HandleKeyboardInteractiveResponse(connID, responses);
	} catch (err) {
		console.error("Keyboard-interactive error:", err);

		HandleKeyboardInteractiveResponse(
			params.connectionId || params.ConnectionID,
			[],
		);
	}
}

async function handleHostKeyPrompt(params) {
	const connID = params.connectionId || params.ConnectionID;

	const host = params.host || "unknown";

	const keyType = params.keyType || "unknown";

	const fingerprint = params.fingerprint || "unknown";

	const msg =
		"Host key for " +
		host +
		" (" +
		keyType +
		")" +
		String.fromCharCode(10) +
		"Fingerprint: " +
		fingerprint +
		String.fromCharCode(10) +
		String.fromCharCode(10) +
		"Accept this host key?";

	const accepted = confirm(msg);

	HandleHostKeyResponse(connID, accepted);
}

// ===== SSH BANNER AND SERVICE MESSAGE HANDLERS =====

function handleSSHBanner(params) {
	const connID = params.connectionId || params.ConnectionID;

	const message = params.message || params.Message || "";

	if (message) {
		const tab = tabs.find((t) => t.isSSH && t.sshConnectionId === connID);

		if (tab && tab.term) {
			tab.term.writeln(
				String.fromCharCode(13, 10) +
					String.fromCharCode(27) +
					"[1;36m[SSH Banner]" +
					String.fromCharCode(27) +
					"[0m " +
					message,
			);

			logConnection(tab, "SSH Banner: " + message.substring(0, 50));
		}
	}
}

function handleSSHServiceMessage(params) {
	const connID = params.connectionId || params.ConnectionID;

	const message = params.message || params.Message || "";

	if (message) {
		const tab = tabs.find((t) => t.isSSH && t.sshConnectionId === connID);

		if (tab && tab.term) {
			tab.term.writeln(
				String.fromCharCode(13, 10) +
					String.fromCharCode(27) +
					"[1;33m[SSH]" +
					String.fromCharCode(27) +
					"[0m " +
					message,
			);
		}
	}
}

function handleTelnetServiceMessage(params) {
	const connID = params.connectionId || params.ConnectionID;

	const message = params.message || params.Message || "";

	if (message) {
		const tab = tabs.find((t) => t.isTelnet && t.telnetConnectionId === connID);

		if (tab && tab.term) {
			tab.term.writeln(
				String.fromCharCode(13, 10) +
					String.fromCharCode(27) +
					"[1;33m[Telnet]" +
					String.fromCharCode(27) +
					"[0m " +
					message,
			);
		}
	}
}

// ===== HOST KEY VERIFICATION =====

let hostKeyResolve = null;

function showHostKeyDialog(fingerprint, host) {
	return new Promise((resolve) => {
		hostKeyResolve = resolve;

		document.getElementById("hostkey-dialog").classList.add("active");

		document.getElementById("hostkey-message").textContent =
			"The host " + host + " is not in your known hosts. The fingerprint is:";

		document.getElementById("hostkey-fingerprint").textContent = fingerprint;
	});
}

function closeHostKeyDialog(accepted) {
	document.getElementById("hostkey-dialog").classList.remove("active");

	if (hostKeyResolve) {
		hostKeyResolve(accepted);
		hostKeyResolve = null;
	}
}

// ===== SFTP FILE BROWSER =====

let sftpSessionId = null;

let sftpConnectionId = null;

let sftpCurrentPath = "/";

let sftpSelectedFile = null;

let sftpFileData = [];

async function openSFTPBrowser(connectionId) {
	sftpConnectionId = connectionId;

	document.getElementById("sftp-dialog").classList.add("active");

	document.getElementById("sftp-session-label").textContent =
		"SSH: " + connectionId;

	try {
		const result = await SFTPOpen({ connectionId });

		sftpSessionId = result.sessionId || result.SessionID;

		sftpNavigate("/");
	} catch (err) {
		showToast("Failed to open SFTP: " + err, "error");

		closeSFTPBrowser();
	}
}

function closeSFTPBrowser() {
	document.getElementById("sftp-dialog").classList.remove("active");

	if (sftpSessionId) {
		SFTPClose(sftpSessionId).catch(() => {});

		sftpSessionId = null;
	}

	const t = getActiveTab();

	if (t) t.term.focus();
}

async function sftpNavigate(path) {
	if (!sftpSessionId) return;

	if (!path) path = sftpCurrentPath;

	sftpCurrentPath = path;

	document.getElementById("sftp-path").value = path;

	const container = document.getElementById("sftp-file-list");

	container.innerHTML =
		'<div style="padding:20px;text-align:center;color:#666;">Loading...</div>';

	try {
		const files = await SFTPReadDir(sftpSessionId, path);

		sftpFileData = files || [];

		// Sort: directories first, then files, alphabetically

		sftpFileData.sort((a, b) => {
			const aDir = a.IsDir || a.isdir || false;

			const bDir = b.IsDir || b.isdir || false;

			if (aDir && !bDir) return -1;

			if (!aDir && bDir) return 1;

			return (a.Name || a.name || "").localeCompare(b.Name || b.name || "");
		});

		renderSFTPList();
	} catch (err) {
		container.innerHTML =
			'<div style="padding:20px;text-align:center;color:#f44747;">Error: ' +
			err +
			"</div>";
	}
}

function renderSFTPList() {
	const container = document.getElementById("sftp-file-list");

	if (sftpFileData.length === 0) {
		container.innerHTML =
			'<div style="padding:20px;text-align:center;color:#666;">Empty directory</div>';

		return;
	}

	let html = "";

	sftpFileData.forEach((f, i) => {
		const name = f.Name || f.name || "unknown";

		const isDir = f.IsDir || f.isdir || false;

		const size = f.Size || f.size || 0;

		const modTime = f.ModTime || f.modTime || "";

		const perm = f.Mode || f.mode || "";

		const icon = isDir ? "\ud83d\udcc1" : "\ud83d\udcc4";

		const sizeStr = isDir ? "--" : formatBytes(size);

		const selected = sftpSelectedFile === name;

		html +=
			'<div class="sftp-file-item' +
			(selected ? " selected" : "") +
			'" data-idx="' +
			i +
			'">' +
			'<span class="sftp-icon">' +
			icon +
			"</span>" +
			'<span class="sftp-name">' +
			name +
			"</span>" +
			'<span class="sftp-size">' +
			sizeStr +
			"</span>" +
			'<span class="sftp-perm">' +
			perm +
			"</span>" +
			"</div>";
	});

	container.innerHTML = html;

	container.querySelectorAll(".sftp-file-item").forEach((el) => {
		el.onclick = () => {
			const idx = parseInt(el.dataset.idx);

			const f = sftpFileData[idx];

			const name = f.Name || f.name;

			const isDir = f.IsDir || f.isdir || false;

			if (isDir) {
				sftpNavigate(sftpCurrentPath.replace(/\/$/, "") + "/" + name);
			} else {
				sftpSelectedFile = name;

				renderSFTPList();
			}
		};

		el.ondblclick = () => {
			const idx = parseInt(el.dataset.idx);

			const f = sftpFileData[idx];

			const name = f.Name || f.name;

			const isDir = f.IsDir || f.isdir || false;

			if (isDir) {
				sftpNavigate(sftpCurrentPath.replace(/\/$/, "") + "/" + name);
			} else {
				sftpDownloadFile(sftpCurrentPath.replace(/\/$/, "") + "/" + name, name);
			}
		};

		el.oncontextmenu = (e) => {
			e.preventDefault();

			const idx = parseInt(el.dataset.idx);

			sftpSelectedFile = sftpFileData[idx].Name || sftpFileData[idx].name;

			renderSFTPList();

			showSFTPContextMenu(e, sftpFileData[idx]);
		};
	});
}

async function sftpGoUp() {
	if (sftpCurrentPath === "/") return;

	const parts = sftpCurrentPath.replace(/\/$/, "").split("/");

	parts.pop();

	sftpNavigate(parts.length ? parts.join("/") : "/");
}

async function sftpMkdir() {
	const name = prompt("New folder name:");

	if (!name) return;

	try {
		const path = sftpCurrentPath.replace(/\/$/, "") + "/" + name;

		await SFTPMkdir(sftpSessionId, path);

		showToast("Folder created: " + name, "success");

		sftpNavigate(sftpCurrentPath);
	} catch (err) {
		showToast("Failed to create folder: " + err, "error");
	}
}

async function sftpDeleteSelected() {
	if (!sftpSelectedFile) {
		showToast("Select a file first", "info");
		return;
	}

	if (!confirm("Delete " + sftpSelectedFile + "?")) return;

	try {
		const path = sftpCurrentPath.replace(/\/$/, "") + "/" + sftpSelectedFile;

		const file = sftpFileData.find(
			(f) => (f.Name || f.name) === sftpSelectedFile,
		);

		const isDir = file && (file.IsDir || file.isdir || false);

		if (isDir) {
			await SFTPRmdir(sftpSessionId, path);
		} else {
			await SFTPDelete(sftpSessionId, path);
		}

		showToast("Deleted: " + sftpSelectedFile, "success");

		sftpSelectedFile = null;

		sftpNavigate(sftpCurrentPath);
	} catch (err) {
		showToast("Failed to delete: " + err, "error");
	}
}

async function sftpDownloadSelected() {
	if (!sftpSelectedFile) {
		showToast("Select a file first", "info");
		return;
	}

	const path = sftpCurrentPath.replace(/\/$/, "") + "/" + sftpSelectedFile;

	sftpDownloadFile(path, sftpSelectedFile);
}

async function sftpDownloadFile(remotePath, fileName) {
	try {
		const dir = await SelectDirectory("Choose download folder for " + fileName);

		const localPath = dir ? dir.replace(/\/$/, "") + "/" + fileName : fileName;

		const result = await SFTPDownload({
			sessionId: sftpSessionId,
			remotePath: remotePath,
			localPath: localPath,
		});

		showToast("Downloaded: " + fileName + " -> " + localPath, "success");
	} catch (err) {
		showToast("Download failed: " + err, "error");
	}
}

async function sftpUploadFile(event) {
	const file = event.target.files[0];

	if (!file) return;

	try {
		const localPath = file.name;

		const remotePath = sftpCurrentPath.replace(/\/$/, "") + "/" + file.name;

		// Read file as base64

		const reader = new FileReader();

		reader.onload = async () => {
			const b64 = reader.result.split(",")[1];

			try {
				await SFTPUpload({
					sessionId: sftpSessionId,

					remotePath: remotePath,

					data: b64,
				});

				showToast("Uploaded: " + file.name, "success");

				sftpNavigate(sftpCurrentPath);
			} catch (err) {
				showToast("Upload failed: " + err, "error");
			}
		};

		reader.readAsDataURL(file);
	} catch (err) {
		showToast("Upload failed: " + err, "error");
	}

	event.target.value = "";
}

function showSFTPContextMenu(e, file) {
	document.querySelectorAll(".context-menu").forEach((m) => m.remove());

	const name = file.Name || file.name;

	const isDir = file.IsDir || file.isdir || false;

	const menu = document.createElement("div");

	menu.className = "context-menu";

	let items =
		'<div class="context-menu-item" data-action="download">\u2b07 Download</div>';

	if (!isDir) {
		items +=
			'<div class="context-menu-item" data-action="rename">\u2702 Rename</div>';
	}

	items +=
		'<div class="context-menu-item" data-action="delete" style="color:#f44747;">\ud83d\uddd1 Delete</div>';

	menu.innerHTML = items;

	document.body.appendChild(menu);

	if (tab.isSSH && tab.sshConnectionId) {
		const sftpItem = menu.querySelector('[data-action="sftp"]');

		const fwdItem = menu.querySelector('[data-action="forward"]');

		if (sftpItem) sftpItem.style.display = "";

		if (fwdItem) fwdItem.style.display = "";
	}

	if (tab.isSSH && tab.sshConnectionId) {
		const sftpItem = menu.querySelector('[data-action="sftp"]');

		const fwdItem = menu.querySelector('[data-action="forward"]');

		if (sftpItem) sftpItem.style.display = "";

		if (fwdItem) fwdItem.style.display = "";
	}

	menu.style.left = Math.min(e.clientX, window.innerWidth - 180) + "px";

	menu.style.top = Math.min(e.clientY, window.innerHeight - 150) + "px";

	const close = () => menu.remove();

	menu.onclick = async (ev) => {
		const item = ev.target.closest(".context-menu-item");

		if (!item) return;

		close();

		const action = item.dataset.action;

		const path = sftpCurrentPath.replace(/\/$/, "") + "/" + name;

		if (action === "download") {
			await sftpDownloadFile(path, name);
		} else if (action === "rename") {
			const newName = prompt("New name:", name);

			if (newName && newName !== name) {
				const newPath = sftpCurrentPath.replace(/\/$/, "") + "/" + newName;

				try {
					await SFTPRename(sftpSessionId, path, newPath);

					showToast("Renamed to " + newName, "success");

					sftpNavigate(sftpCurrentPath);
				} catch (err) {
					showToast("Rename failed: " + err, "error");
				}
			}
		} else if (action === "delete") {
			if (!confirm("Delete " + name + "?")) return;

			try {
				if (isDir) await SFTPRmdir(sftpSessionId, path);
				else await SFTPDelete(sftpSessionId, path);

				showToast("Deleted: " + name, "success");

				sftpSelectedFile = null;

				sftpNavigate(sftpCurrentPath);
			} catch (err) {
				showToast("Delete failed: " + err, "error");
			}
		}
	};

	setTimeout(
		() => document.addEventListener("click", close, { once: true }),
		10,
	);
}

function formatBytes(bytes) {
	if (bytes === 0) return "0 B";

	const k = 1024;

	const sizes = ["B", "KB", "MB", "GB", "TB"];

	const i = Math.floor(Math.log(bytes) / Math.log(k));

	return parseFloat((bytes / k ** i).toFixed(1)) + " " + sizes[i];
}

// ===== TERMINAL TOOLBAR =====

function buildToolbar(tab) {
	let html = '<div class="terminal-toolbar" id="toolbar-' + tab.id + '">';

	// Toolbar action buttons

	if (tab.isSSH) {
		html +=
			'<button class="toolbar-btn" onclick="openSFTPBrowser(getActiveTab().sshConnectionId)" title="SFTP">U0001f4c2</button>';

		html +=
			'<button class="toolbar-btn" onclick="openForwardDialog(getActiveTab().sshConnectionId)" title="Forward">U0001f504</button>';
	}

	html +=
		'<button class="toolbar-btn" onclick="getActiveTab().copySelection()" title="Copy">C</button>';

	html +=
		'<button class="toolbar-btn" onclick="getActiveTab().pasteFromClipboard()" title="Paste">P</button>';

	html +=
		'<button class="toolbar-btn" onclick="toggleFind()" title="Find">F</button>';

	html +=
		'<button class="toolbar-btn" onclick="toggleBroadcast()" title="Broadcast (Ctrl+Shift+B)" id="btn-broadcast" style="' +
		(broadcastMode ? "color:#007acc;" : "") +
		'">U0001f4e1</button>';

	html +=
		'<button class="toolbar-btn" onclick="clearTerminal()" title="Clear">X</button>';

	html +=
		'<button class="toolbar-btn" onclick="openLogDialog(getActiveTab())" title="View Log">U0001f4dc</button>';

	html += "</div>";

	return html;
}

function escHtml(s) {
	return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

function updateToolbar(tab) {
	if (!tab || !tab.wrapper) return;

	const toolbar = tab.wrapper.querySelector(".terminal-toolbar");

	const div = document.createElement("div");

	div.innerHTML = buildToolbar(tab);

	const newToolbar = div.firstElementChild;

	if (toolbar) toolbar.replaceWith(newToolbar);
	else tab.wrapper.insertBefore(newToolbar, tab.wrapper.firstChild);
}

function clearTerminal() {
	const tab = getActiveTab();

	if (tab) {
		tab.term.clear();
		tab.term.focus();
	}
}

// ===== BROADCAST MODE =====

function toggleBroadcast() {
	broadcastMode = !broadcastMode;
	// Update all toolbar buttons to reflect broadcast state
	for (const t of tabs) {
		const btn = t.wrapper && t.wrapper.querySelector("#btn-broadcast");
		if (btn) {
			btn.style.color = broadcastMode ? "#007acc" : "";
			btn.title = broadcastMode
				? "Broadcasting ON (Ctrl+Shift+B)"
				: "Broadcast (Ctrl+Shift+B)";
		}
	}
	const el = document.getElementById("status-text");
	if (el) {
		el.textContent = broadcastMode
			? "U0001f4e1 Broadcast ON - " +
				tabs.length +
				" tab" +
				(tabs.length !== 1 ? "s" : "")
			: tabs.length + " tab" + (tabs.length !== 1 ? "s" : "");
		clearTimeout(window.__statusTimeout);
		window.__statusTimeout = setTimeout(() => {
			el.textContent = tabs.length + " tab" + (tabs.length !== 1 ? "s" : "");
		}, 3000);
	}
	showToast(broadcastMode ? "Broadcast ON" : "Broadcast OFF", "info");
}

// ===== SESSION LOG =====

function openLogDialog(tab) {
	if (!tab || !tab.logBuffer.length) {
		showToast("No log data available", "info");
		return;
	}
	// Remove existing log dialog if any
	const existing = document.getElementById("log-dialog");
	if (existing) existing.remove();

	const dlg = document.createElement("div");
	dlg.className = "modal-overlay";
	dlg.id = "log-dialog";
	dlg.style.cssText = "z-index:1000;";

	const box = document.createElement("div");
	box.className = "modal-box";
	box.style.cssText =
		"width:700px;max-height:80vh;display:flex;flex-direction:column;";

	const hdr = document.createElement("div");
	hdr.style.cssText =
		"display:flex;justify-content:space-between;align-items:center;margin-bottom:8px;";
	hdr.innerHTML = '<h3 style="margin:0;">U0001f4dc Session Log</h3>';
	box.appendChild(hdr);

	const textarea = document.createElement("textarea");
	textarea.style.cssText =
		"width:100%;flex:1;min-height:300px;background:#1a1a1a;border:1px solid #3a3a3a;color:#ccc;padding:8px;border-radius:4px;font-family:monospace;font-size:12px;resize:none;";
	textarea.readOnly = true;
	textarea.value = stripAnsi(tab.logBuffer.join(""));
	box.appendChild(textarea);

	const actions = document.createElement("div");
	actions.style.cssText =
		"display:flex;gap:8px;justify-content:flex-end;margin-top:8px;";

	const downloadBtn = document.createElement("button");
	downloadBtn.textContent = "U0001f4e5 Download";
	downloadBtn.className = "btn-primary";
	downloadBtn.onclick = () => {
		const blob = new Blob([textarea.value], { type: "text/plain" });
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = "session-" + tab.id + ".log";
		a.click();
		URL.revokeObjectURL(url);
	};
	actions.appendChild(downloadBtn);

	const closeBtn = document.createElement("button");
	closeBtn.textContent = "Close";
	closeBtn.className = "btn-secondary";
	closeBtn.onclick = () => dlg.remove();
	actions.appendChild(closeBtn);

	box.appendChild(actions);
	dlg.appendChild(box);
	dlg.addEventListener("click", (e) => {
		if (e.target === dlg) dlg.remove();
	});
	document.body.appendChild(dlg);
}

// ===== SETTING DESCRIPTIONS =====

const SETTING_DESCRIPTIONS = {
	"s-color-scheme": "Choose a predefined color scheme for the terminal. Each scheme sets the background, text, and ANSI colors.",
	"s-font-family": "Font used in the terminal. Supports fallback fonts separated by commas.",
	"s-font-size": "Base font size in pixels for terminal text.",
	"s-fallback-font": "Fallback font used when the primary font lacks certain glyphs (e.g., Nerd Font icons).",
	"s-font-weight": "Normal text weight. Common values: 400 (normal), 700 (bold).",
	"s-font-weight-bold": "Weight used for bold text in the terminal.",
	"s-line-height": "Spacing between lines of text as a multiplier of font size.",
	"s-line-padding": "Extra vertical padding added to each line.",
	"s-ligatures": "Enable font ligatures for programming symbols like -> or =>.",
	"s-theme": "Color scheme mode: Auto (follows system), Always Dark, or Always Light.",
	"s-opacity": "Terminal background opacity. Lower values make the window more transparent.",
	"s-bg-color": "Custom terminal background color (overrides scheme default).",
	"s-spaciness": "Controls the spacing density of the terminal UI elements.",
	"s-animations": "Enable/disable UI animations throughout the application.",
	"s-cursor-style": "Shape of the terminal cursor: Bar (|), Block (\u2588), or Underline (\u2581).",
	"s-cursor-blink": "Whether the terminal cursor blinks on and off.",
	"s-frontend": "Terminal rendering engine: WebGL (fastest), Canvas, or experimental Block Frontend.",
	"s-draw-bold-bright": "Draw bold text using bright/vivid colors instead of just a heavier weight.",
	"s-min-contrast": "Minimum contrast ratio for text against background (1-21, higher = more contrast).",
	"s-css": "Custom CSS injected into the terminal for advanced styling.",
	"s-shell": "Default shell to launch (leave empty for system default: bash, zsh, pwsh).",
	"s-scrollback": "Number of lines to keep in terminal scrollback buffer.",
	"s-bell": "Terminal bell behavior: off, audible, or visual flash.",
	"s-alt-is-meta": "Treat the Alt key as the Meta modifier (useful for emacs and other TUI apps).",
	"s-scroll-on-input": "Automatically scroll to the bottom when you type.",
	"s-use-conpty": "Use Windows ConPTY for improved pseudoterminal rendering (Windows only).",
	"s-set-comspec": "Set COMSPEC environment variable to the configured shell on launch.",
	"s-copy-on-select": "Automatically copy text to clipboard when it is selected with the mouse.",
	"s-copy-as-html": "Preserve formatting when copying terminal output (e.g., to rich text editors).",
	"s-bracketed-paste": "Enable bracketed paste mode to prevent auto-indent issues when pasting code.",
	"s-session-logging": "Capture all terminal output to a session log that can be viewed via the toolbar button.",
	"s-warn-multiline": "Show a warning dialog when pasting multi-line content.",
	"s-replace-newlines": "Replace newlines with spaces when pasting (useful for pasting into single-line prompts).",
	"s-trim-whitespace": "Trim leading and trailing whitespace from pasted content.",
	"s-right-click": "Behavior when right-clicking in the terminal: context menu, paste, or smart paste.",
	"s-paste-middle-click": "Paste from clipboard when clicking the middle mouse button.",
	"s-word-separator": "Characters treated as word separators for double-click word selection.",
	"s-tab-position": "Position of the tab bar: left side or top of the window.",
	"s-last-tab-closes": "Behavior when the last tab is closed: close window or leave it open.",
	"s-cycle-tabs": "Allow tab cycling from last to first tab and vice versa.",
	"s-hide-close-button": "Hide the close button on individual tabs.",
	"s-pane-resize-step": "Step size (as fraction) for resizing panes with keyboard shortcuts.",
	"s-focus-follows-mouse": "Automatically focus the pane under the mouse cursor.",
	"s-auto-open": "Automatically open a new terminal window on startup.",
	"s-recover-tabs": "Restore tabs from the previous session on startup.",
	"s-frame": "Window frame style: thin, full, or custom frame.",
	"s-dock": "Dock the terminal to a screen edge (off, top, bottom, left, right).",
	"s-dock-hide-blur": "Hide the docked terminal when it loses focus.",
	"s-dock-on-top": "Keep the docked terminal always on top of other windows.",
	"s-hide-tray": "Hide the system tray icon.",
	"s-language": "UI language (leave empty for system default).",
	"s-analytics": "Send anonymous usage analytics to help improve the application.",
	"s-auto-updates": "Automatically check for and install updates.",
	"s-experimental": "Enable experimental features that may be unstable.",
	"s-ssh-warn-close": "Show a confirmation dialog when closing an active SSH connection.",
	"s-ssh-verify-keys": "Verify SSH host keys against known_hosts on first connection.",
	"s-ssh-disable-title": "Prevent the remote host from changing the terminal title.",
	"s-ssh-agent-type": "Type of SSH agent to use: auto-detect, pageant, or custom.",
	"s-ssh-agent-path": "Path to the SSH agent socket or executable.",
	"s-ssh-x11": "X11 display for SSH X11 forwarding (e.g., localhost:0).",
	"s-serial-baud": "Baud rate for serial port communication (bits per second).",
	"s-serial-data-bits": "Number of data bits per serial frame (typically 7 or 8).",
	"s-serial-stop-bits": "Number of stop bits per serial frame (1 or 2).",
	"s-serial-parity": "Parity checking mode for serial communication.",
	"s-serial-flow": "Flow control method for serial connections: none, hardware (RTS/CTS), or software (XON/XOFF).",
};

// ===== COMMAND PALETTE =====

const PALETTE_COMMANDS = [
	{
		id: "new-tab",
		label: "New Tab",
		icon: "+",
		action: () => {
			newTab();
		},
	},

	{
		id: "close-tab",
		label: "Close Tab",
		icon: "x",
		action: () => {
			var t = getActiveTab();
			if (t) t.close();
		},
	},

	{
		id: "ssh-connect",
		label: "SSH Connect...",
		icon: "#",
		action: () => {
			openSSHDialog();
		},
	},

	{
		id: "serial-connect",
		label: "Serial Port...",
		icon: "~",
		action: () => {
			openSerialDialog();
		},
	},

	{
		id: "telnet-connect",
		label: "Telnet Connect...",
		icon: "=",
		action: () => {
			openTelnetDialog();
		},
	},

	{
		id: "import-ssh-config",
		label: "Import SSH Config",
		icon: "@",
		action: () => {
			importSSHConfig();
		},
	},

	{
		id: "toggle-settings",
		label: "Settings",
		icon: "*",
		action: () => {
			toggleSettings();
		},
	},

	{
		id: "split-horizontal",
		label: "Split Horizontal",
		icon: "-",
		action: () => {
			splitPane("horizontal");
		},
	},

	{
		id: "split-vertical",
		label: "Split Vertical",
		icon: "|",
		action: () => {
			splitPane("vertical");
		},
	},

	{
		id: "find",
		label: "Find in Terminal",
		icon: "?",
		action: () => {
			toggleFind();
		},
	},

	{
		id: "clear-terminal",
		label: "Clear Terminal",
		icon: "!",
		action: () => {
			clearTerminal();
		},
	},

	{
		id: "copy",
		label: "Copy Selection",
		icon: "C",
		action: () => {
			var t = getActiveTab();
			if (t) t.copySelection();
		},
	},

	{
		id: "paste",
		label: "Paste from Clipboard",
		icon: "V",
		action: () => {
			var t = getActiveTab();
			if (t) t.pasteFromClipboard();
		},
	},

	{
		id: "zoom-in",
		label: "Increase Font Size",
		icon: "+",
		action: () => {
			fontSize = Math.min(32, fontSize + 1);
			tabs.forEach((t) => {
				t.term.options.fontSize = fontSize;
			});
			showToast("Font: " + fontSize, "info");
		},
	},

	{
		id: "zoom-out",
		label: "Decrease Font Size",
		icon: "-",
		action: () => {
			fontSize = Math.max(8, fontSize - 1);
			tabs.forEach((t) => {
				t.term.options.fontSize = fontSize;
			});
			showToast("Font: " + fontSize, "info");
		},
	},

	{
		id: "zoom-reset",
		label: "Reset Font Size",
		icon: "0",
		action: () => {
			fontSize = 14;
			tabs.forEach((t) => {
				t.term.options.fontSize = fontSize;
			});
			showToast("Font: 14", "info");
		},
	},

	{
		id: "toggle-fullscreen",
		label: "Toggle Fullscreen",
		icon: "F",
		action: () => {
			if (document.fullscreenElement) document.exitFullscreen();
			else document.documentElement.requestFullscreen();
		},
	},

	{
		id: "sftp-browser",
		label: "SFTP File Browser",
		icon: "F",
		action: () => {
			var t = getActiveTab();
			if (t && t.isSSH) openSFTPBrowser(t.sshConnectionId);
			else showToast("Requires active SSH tab", "info");
		},
	},

	{
		id: "port-forward",
		label: "Port Forwarding",
		icon: "P",
		action: () => {
			var t = getActiveTab();
			if (t && t.isSSH) openForwardDialog(t.sshConnectionId);
			else showToast("Requires active SSH tab", "info");
		},
	},

	{
		id: "save-session",
		label: "Save Session State",
		icon: "S",
		action: () => {
			saveSession();
		},
	},

	{
		id: "reload-colors",
		label: "Reload Color Schemes",
		icon: "R",
		action: () => {
			reloadColorSchemes();
		},
	},

	{
		id: "tab-search",
		label: "Search Tabs...",
		icon: "T",
		action: () => {
			toggleTabSearch();
		},
	},

	{
		id: "export-profiles",
		label: "Export Profiles...",
		icon: "↑",
		action: () => {
			exportProfiles();
		},
	},

	{
		id: "import-profiles",
		label: "Import Profiles...",
		icon: "↓",
		action: () => {
			importProfiles();
		},
	},

	{
		id: "open-settings",
		label: "Open Settings",
		icon: "⚙",
		action: () => {
			openSettingsPanel();
		},
	},

	{
		id: "connection-log",
		label: "View Connection Log",
		icon: "📋",
		action: () => {
			showConnectionLog();
		},
	},

	{
		id: "always-on-top",
		label: "Toggle Always on Top",
		icon: "📌",
		action: () => {
			toggleAlwaysOnTop();
		},
	},

	{
		id: "about",
		label: "About Tabby Go",
		icon: "i",
		action: () => {
			showAboutDialog();
		},
	},

	{
		id: "run-snippet",
		label: "Run Snippet...",
		icon: ">",
		action: () => {
			runSnippet();
		},
	},

	{
		id: "save-snippet",
		label: "Save Snippet...",
		icon: "S",
		action: () => {
			saveSnippet();
		},
	},

	{
		id: "check-updates",
		label: "Check for Updates",
		icon: "↑",
		action: () => {
			checkForUpdates();
		},
	},

	{
		id: "audit-log",
		label: "Open Audit Log",
		icon: "A",
		action: () => {
			openAuditLog();
		},
	},

	{
		id: "close-all-tabs",
		label: "Close All Tabs",
		icon: "X",
		action: () => {
			closeAllTabs();
		},
	},

	{
		id: "close-other-tabs",
		label: "Close Other Tabs",
		icon: "X",
		action: () => {
			closeOtherTabs();
		},
	},

	{
		id: "duplicate-profile",
		label: "Duplicate Active Profile",
		icon: "D",
		action: () => {
			duplicateActiveProfile();
		},
	},

	{
		id: "notifications",
		label: "Show Notifications",
		icon: "!",
		action: () => {
			showNotificationCenter();
		},
	},

	{
		id: "custom-css",
		label: "Edit Custom CSS",
		icon: "C",
		action: () => {
			editCustomCSS();
		},
	},

	{
		id: "save-credential",
		label: "Save Credential to Keychain",
		icon: "K",
		action: () => {
			saveCredentialDialog();
		},
	},

	{
		id: "get-credential",
		label: "Get Credential from Keychain",
		icon: "K",
		action: () => {
			getCredentialDialog();
		},
	},
];

var paletteVisible = false;

var paletteSelectedIdx = 0;

function toggleCommandPalette() {
	paletteVisible = !paletteVisible;

	var el = document.getElementById("command-palette");

	if (paletteVisible) {
		el.classList.add("active");

		var input = document.getElementById("cmd-palette-input");

		input.value = "";

		input.focus();

		paletteSelectedIdx = 0;

		filterCommandPalette();
	} else {
		el.classList.remove("active");

		var tab = getActiveTab();

		if (tab) tab.term.focus();
	}
}

function filterCommandPalette() {
	var query = (
		document.getElementById("cmd-palette-input").value || ""
	).toLowerCase();

	var container = document.getElementById("cmd-palette-items");

	var filtered = PALETTE_COMMANDS.filter((cmd) =>
		cmd.label.toLowerCase().includes(query),
	);

	paletteSelectedIdx = 0;

	container.innerHTML = filtered
		.map(
			(cmd, i) =>
				'<div class="palette-item' +
				(i === 0 ? " selected" : "") +
				'" data-id="' +
				cmd.id +
				'">' +
				'<span class="palette-icon">' +
				cmd.icon +
				"</span>" +
				'<span class="palette-label">' +
				cmd.label +
				"</span></div>",
		)
		.join("");

	container.querySelectorAll(".palette-item").forEach((el) => {
		el.onclick = () => {
			executePaletteCommand(el.dataset.id);
		};
	});
}

function handlePaletteKey(e) {
	var items = document.querySelectorAll(".palette-item");

	if (e.key === "Escape") {
		e.preventDefault();

		toggleCommandPalette();
	} else if (e.key === "ArrowDown") {
		e.preventDefault();
		paletteSelectedIdx = Math.min(paletteSelectedIdx + 1, items.length - 1);
		items.forEach((el, i) => {
			el.classList.toggle("selected", i === paletteSelectedIdx);
		});
	} else if (e.key === "ArrowUp") {
		e.preventDefault();
		paletteSelectedIdx = Math.max(paletteSelectedIdx - 1, 0);
		items.forEach((el, i) => {
			el.classList.toggle("selected", i === paletteSelectedIdx);
		});
	} else if (e.key === "Enter") {
		e.preventDefault();
		var sel = items[paletteSelectedIdx];
		if (sel) sel.click();
	}
}

function executePaletteCommand(id) {
	var cmd = PALETTE_COMMANDS.find((c) => c.id === id);

	if (cmd) {
		toggleCommandPalette();
		cmd.action();
	}
}

async function reloadColorSchemes() {
	try {
		const schemes = await GetColorSchemes();

		if (schemes && schemes.length) {
			schemes.forEach((s) => {
				COLOR_SCHEMES[s.Name] = s;
			});

			schemeNames = schemes.map((s) => s.Name);

			showToast("Loaded " + schemes.length + " color schemes", "success");
		}
	} catch (err) {
		showToast("Failed to reload schemes: " + err, "error");
	}
}

// ===== SSH CONFIG IMPORT =====

async function importSSHConfig() {
	try {
		const result = await ImportSSHConfig();

		if (result && result.length > 0) {
			let imported = 0;

			let skipped = 0;

			for (var i = 0; i < result.length; i++) {
				var host = result[i];

				var opts = host.options || host.Options || {};

				var h = opts.host || opts.Host || "";

				var p = opts.port || opts.Port || 22;

				var u = opts.user || opts.User || "root";

				var name = u + "@" + h;

				var exists = savedProfiles.find(
					(pr) => pr.name === name && pr.type === "ssh",
				);

				if (exists) {
					skipped++;
					continue;
				}

				savedProfiles.push({
					id: "ssh-import-" + Date.now() + "-" + imported,

					type: "ssh",

					name: name,

					options: {
						host: h,
						port: p,
						user: u,
						auth:
							opts.privateKeys && opts.privateKeys.length
								? "publicKey"
								: "agent",
						privateKeys: opts.privateKeys || opts.PrivateKeys || [],
						jumpHost: opts.jumpHost || opts.JumpHost || "",
					},

					createdAt: new Date().toISOString(),

					updatedAt: new Date().toISOString(),
				});

				imported++;
			}

			await SaveProfiles(savedProfiles);

			renderProfiles();

			showToast(
				"Imported " +
					imported +
					" hosts" +
					(skipped ? ", skipped " + skipped + " duplicates" : ""),
				"success",
			);
		} else {
			showToast("No SSH hosts found in config", "info");
		}
	} catch (err) {
		showToast("SSH config import failed: " + err, "error");
	}
}

// ===== PROFILE EDITOR =====

function editProfile(profile) {
	const modal = document.getElementById("profile-edit-dialog");

	if (!modal) return;

	modal.classList.add("active");

	document.getElementById("edit-profile-id").value = profile.id;

	document.getElementById("edit-profile-type").value = profile.type || "ssh";

	document.getElementById("edit-profile-name").value = profile.name || "";

	document.getElementById("edit-profile-group").value = profile.group || "";
	// Populate group suggestions datalist for the group input
	(() => {
		const groupsSet = new Set();
		savedProfiles.forEach((p) => {
			if (p.group) groupsSet.add(p.group);
		});
		const datalist = document.getElementById("existing-groups");
		if (datalist) {
			datalist.innerHTML = Array.from(groupsSet)
				.map((g) => `<option value="${g}"></option>`)
				.join("");
		}
	})();

	const opts = profile.options || {};

	document.getElementById("edit-profile-host").value = opts.host || "";

	document.getElementById("edit-profile-port").value = opts.port || 22;

	document.getElementById("edit-profile-user").value = opts.user || "";

	document.getElementById("edit-profile-auth").value = opts.auth || "agent";

	document.getElementById("edit-profile-key").value =
		(opts.privateKeys && opts.privateKeys[0]) || "";

	document.getElementById("edit-profile-jump").value = opts.jumpHost || "";

	document.getElementById("edit-profile-login-script").value =
		opts.loginScript || "";
}

function closeProfileEditor() {
	document.getElementById("profile-edit-dialog").classList.remove("active");

	const t = getActiveTab();
	if (t) t.term.focus();
}

function saveProfileEdit() {
	const id = document.getElementById("edit-profile-id").value;

	const profile = savedProfiles.find((p) => p.id === id);

	if (!profile) return;

	profile.name = document.getElementById("edit-profile-name").value.trim();

	profile.group = document.getElementById("edit-profile-group").value.trim();

	if (document.getElementById("edit-profile-type"))
		profile.type = document.getElementById("edit-profile-type").value;

	if (!profile.options) profile.options = {};

	profile.options.host = document
		.getElementById("edit-profile-host")
		.value.trim();

	profile.options.port =
		parseInt(document.getElementById("edit-profile-port").value) || 22;

	profile.options.user = document
		.getElementById("edit-profile-user")
		.value.trim();

	profile.options.auth = document.getElementById("edit-profile-auth").value;

	const keyPath = document.getElementById("edit-profile-key").value.trim();

	profile.options.privateKeys = keyPath ? [keyPath] : [];

	profile.options.jumpHost = document
		.getElementById("edit-profile-jump")
		.value.trim();

	profile.options.loginScript = document
		.getElementById("edit-profile-login-script")
		.value.trim();

	profile.updatedAt = new Date().toISOString();

	SaveProfiles(savedProfiles).catch(() => {});

	renderProfiles();

	closeProfileEditor();

	showToast("Profile saved: " + profile.name, "success");
}

// ===== TAB SEARCH =====

let tabSearchVisible = false;

function toggleTabSearch() {
	tabSearchVisible = !tabSearchVisible;

	const el = document.getElementById("tab-search-bar");

	if (tabSearchVisible) {
		el.classList.add("active");

		const input = document.getElementById("tab-search-input");

		input.value = "";

		input.focus();

		filterTabs();
	} else {
		el.classList.remove("active");

		const t = getActiveTab();
		if (t) t.term.focus();
	}
}

function filterTabs() {
	const query = (
		document.getElementById("tab-search-input").value || ""
	).toLowerCase();

	const results = document.getElementById("tab-search-results");

	if (!results) return;

	const filtered = tabs.filter((t) =>
		(t.title || "").toLowerCase().includes(query),
	);

	results.innerHTML = filtered
		.map((t, i) => {
			const type = t.isSSH
				? "SSH"
				: t.isSerial
					? "SER"
					: t.isTelnet
						? "TEL"
						: "LOCAL";

			return (
				'<div class="tab-search-item" data-tab-id="' +
				t.id +
				'">' +
				'<span class="tab-search-type">' +
				type +
				"</span>" +
				'<span class="tab-search-title">' +
				(t.title || "shell") +
				"</span></div>"
			);
		})
		.join("");

	results.querySelectorAll(".tab-search-item").forEach((el) => {
		el.onclick = () => {
			const tab = tabs.find((t) => t.id === el.dataset.tabId);

			if (tab) {
				tab.activate();
				toggleTabSearch();
			}
		};
	});
}

// ===== LOGIN SCRIPTS =====

function runLoginScript(tab, script) {
	if (!script || !script.trim()) return;

	const lines = script.split("\n");

	const delay = 500; // Initial delay after connection

	lines.forEach((line, i) => {
		setTimeout(
			() => {
				const trimmed = line.trim();

				if (!trimmed || trimmed.startsWith("#")) return; // Skip comments and empty lines

				if (tab.isSSH && tab.sshConnectionId && tab.sshSessionId) {
					SSHWrite({
						connectionId: tab.sshConnectionId,
						sessionId: tab.sshSessionId,
						data: btoa(trimmed + "\r"),
					});
				} else if (tab.ptyId && !tab.exited) {
					PTYWrite(tab.ptyId, btoa(trimmed + "\r"));
				} else if (tab.telnetConnectionId) {
					TelnetWrite(tab.telnetConnectionId, btoa(trimmed + "\r"));
				}
			},
			delay + i * 300,
		);
	});
}

// ===== INPUT PROCESSING =====

// Right-click paste support

function setupInputProcessing(term, tab) {
	// Handle right-click paste (only if RightClick setting allows it)
	// If clipboard has content, paste it; otherwise copy if selection
	term.element.addEventListener("contextmenu", (e) => {
		e.preventDefault();

		if (settings.RightClick === "off") return;

		const sel = term.getSelection();

		// If there's a selection and RightClick is set to copy or clipboard-aware, copy it
		if (
			sel &&
			(settings.RightClick === "clipboard" || settings.RightClick === "menu")
		) {
			ClipboardSetText(sel).then(() => {
				showToast("Copied", "success");
			});
			term.clearSelection();
			return;
		}

		// Otherwise paste if clipboard has content
		if (
			settings.RightClick === "paste" ||
			settings.RightClick === "clipboard"
		) {
			tab.pasteFromClipboard();
		}
	});

	// Handle copy on select

	if (settings.CopyOnSelect) {
		term.onSelectionChange(() => {
			const sel = term.getSelection();

			if (sel) ClipboardSetText(sel).catch(() => {});
		});
	}
}

// ===== ABOUT DIALOG =====

async function showAboutDialog() {
	try {
		const platform = await GetPlatform();

		const home = await GetHomeDir();

		const hostname = await GetHostname();

		const username = await GetUsername();

		const shell = await GetDefaultShell();

		const info =
			"Tabby Go v" +
			(platform.version || "1.0.0") +
			"\r\n" +
			"Platform: " +
			(platform.os || "unknown") +
			"/" +
			(platform.arch || "unknown") +
			"\r\n" +
			"User: " +
			username +
			"@" +
			hostname +
			"\r\n" +
			"Home: " +
			home +
			"\r\n" +
			"Shell: " +
			shell;

		alert(info);
	} catch (err) {
		alert(
			"Tabby Go v1.0.0\r\nA modern terminal emulator built with Go + Wails",
		);
	}
}

// ===== NOTIFICATION CENTER =====

async function showNotificationCenter() {
	const panel = document.getElementById("notification-center");

	if (panel) {
		panel.classList.toggle("active");

		if (panel.classList.contains("active")) {
			try {
				const notifs = await GetNotifications();

				const list = document.getElementById("notification-list");

				if (list) {
					if (!notifs || notifs.length === 0) {
						list.innerHTML =
							'<div style="color:#666;font-size:12px;text-align:center;padding:20px;">No notifications</div>';
					} else {
						list.innerHTML = notifs
							.map((n) => {
								const levelColor =
									n.Level === 2
										? "#f44747"
										: n.Level === 1
											? "#e8a84c"
											: "#4ca8e8";

								const time = new Date(n.Timestamp).toLocaleTimeString();

								return (
									'<div class="notif-item" style="padding:8px 12px;border-bottom:1px solid #2a2a2a;">' +
									'<div style="display:flex;justify-content:space-between;align-items:center;">' +
									'<span style="font-size:12px;color:' +
									levelColor +
									';font-weight:600;">' +
									(n.Title || "Notification") +
									"</span>" +
									'<span style="font-size:10px;color:#666;">' +
									time +
									"</span></div>" +
									'<div style="font-size:11px;color:#aaa;margin-top:4px;">' +
									(n.Message || "") +
									"</div></div>"
								);
							})
							.join("");
					}
				}
			} catch (err) {
				console.error("Failed to load notifications:", err);
			}
		}
	}
}

// ===== SETTINGS PANEL =====

function openSettingsPanel() {
	let panel = document.getElementById("settings-panel");

	if (panel) {
		panel.remove();
		return;
	}

	panel = document.createElement("div");

	panel.id = "settings-panel";

	panel.className = "modal-overlay";

	const dlg = document.createElement("div");

	dlg.className = "modal-dialog";

	dlg.style.maxWidth = "600px";

	dlg.style.maxHeight = "80vh";

	dlg.style.overflowY = "auto";

	const hdr = document.createElement("div");

	hdr.className = "modal-header";

	hdr.innerHTML = "<h3>Settings</h3>";

	const body = document.createElement("div");

	body.className = "modal-body";

	body.innerHTML = `

    <div class="setting-group"><label>Terminal Font</label><input type="text" id="settings-font" class="text-input" value="${settings.Font || "Cascadia Mono, Consolas, monospace"}"></div>

    <div class="setting-group"><label>Font Size (px)</label><input type="number" id="settings-font-size" class="text-input" value="${settings.FontSize || 14}" min="8" max="32"></div>

    <div class="setting-group"><label>Scrollback Lines</label><input type="number" id="settings-scrollback" class="text-input" value="${settings.Scrollback || 25000}" min="100" max="100000"></div>

    <div class="setting-group"><label>Cursor Style</label><select id="settings-cursor" class="text-input"><option value="block" ${settings.CursorStyle === "block" ? "selected" : ""}>Block</option><option value="underline" ${settings.CursorStyle === "underline" ? "selected" : ""}>Underline</option><option value="bar" ${settings.CursorStyle === "bar" ? "selected" : ""}>Bar</option></select></div>

    <div class="setting-group"><label>Cursor Blink</label><input type="checkbox" id="settings-cursor-blink" ${settings.CursorBlink !== false ? "checked" : ""}></div>

    <div class="setting-group"><label>Idle Timeout (minutes, 0=off)</label><input type="number" id="settings-idle-timeout" class="text-input" value="${settings.IdleTimeout || 0}" min="0" max="1440"></div>

    <div class="setting-group"><label>Color Scheme</label><select id="settings-color-scheme" class="text-input">${Object.keys(
			colorSchemes,
		)
			.map(
				(s) =>
					'<option value="' +
					s +
					'" ' +
					(settings.ColorScheme === s ? "selected" : "") +
					">" +
					s +
					"</option>",
			)
			.join("")}</select></div>

    <div class="setting-group"><label>Shell (Windows)</label><input type="text" id="settings-shell" class="text-input" value="${settings.DefaultShell || ""}" placeholder="Auto-detect"></div>

  `;

	const ftr = document.createElement("div");

	ftr.className = "modal-footer";

	const saveBtn = document.createElement("button");

	saveBtn.className = "btn btn-primary";
	saveBtn.textContent = "Save";

	const cancelBtn = document.createElement("button");

	cancelBtn.className = "btn btn-secondary";
	cancelBtn.textContent = "Cancel";

	ftr.appendChild(saveBtn);

	ftr.appendChild(cancelBtn);

	dlg.appendChild(hdr);

	dlg.appendChild(body);

	dlg.appendChild(ftr);

	panel.appendChild(dlg);

	document.body.appendChild(panel);

	cancelBtn.onclick = () => panel.remove();

	panel.onclick = (e) => {
		if (e.target === panel) panel.remove();
	};

	saveBtn.onclick = () => {
		settings.Font = document.getElementById("settings-font").value;

		settings.FontSize =
			parseInt(document.getElementById("settings-font-size").value) || 14;

		settings.Scrollback =
			parseInt(document.getElementById("settings-scrollback").value) || 25000;

		settings.CursorStyle = document.getElementById("settings-cursor").value;

		settings.CursorBlink = document.getElementById(
			"settings-cursor-blink",
		).checked;

		settings.IdleTimeout =
			parseInt(document.getElementById("settings-idle-timeout").value) || 0;

		settings.ColorScheme = document.getElementById(
			"settings-color-scheme",
		).value;

		settings.DefaultShell = document.getElementById("settings-shell").value;

		applySettings(settings);

		SaveSettings(settings).catch(() => {});

		showToast("Settings saved", "success");

		panel.remove();
	};
}

// ===== CUSTOM CSS =====

function editCustomCSS() {
	const css = prompt(
		"Enter custom CSS (applied on top of color scheme):",
		localStorage.getItem("tabby-custom-css") || "",
	);

	if (css === null) return;

	localStorage.setItem("tabby-custom-css", css);

	applyCustomCSS(css);

	showToast("Custom CSS applied", "success");
}

function applyCustomCSS(css) {
	let el = document.getElementById("tabby-custom-css");

	if (!el) {
		el = document.createElement("style");
		el.id = "tabby-custom-css";
		document.head.appendChild(el);
	}

	el.textContent = css;
}

// ===== SFTP DRAG & DROP =====

async function sftpHandleDrop(event) {
	const files = event.dataTransfer.files;

	if (!files || files.length === 0) return;

	for (const file of files) {
		const reader = new FileReader();

		reader.onload = async () => {
			const b64 = reader.result.split(",")[1];

			const remotePath = sftpCurrentPath.replace(/\/$/, "") + "/" + file.name;

			try {
				await SFTPUpload({
					sessionId: sftpSessionId,
					remotePath: remotePath,
					data: b64,
				});

				showToast("Uploaded: " + file.name, "success");

				sftpNavigate(sftpCurrentPath);
			} catch (err) {
				showToast("Upload failed: " + err, "error");
			}
		};

		reader.readAsDataURL(file);
	}
}

// ===== PROFILE GROUP RENDERING =====

function renderProfileGroups() {
	const groups = getProfileGroups();

	const list = document.getElementById("profiles-list");

	if (!list) return;

	let html = "";

	Object.keys(groups)
		.sort()
		.forEach((groupName) => {
			if (groupName !== "Ungrouped" || Object.keys(groups).length > 1) {
				const collapsed =
					localStorage.getItem("tabby-group-collapsed-" + groupName) === "true";

				const arrow = collapsed ? "▸" : "▾";

				html +=
					'<div class="profile-group-header" data-group="' +
					groupName +
					'">' +
					arrow +
					" " +
					groupName +
					' <span class="group-count">(' +
					groups[groupName].length +
					")</span></div>";

				if (collapsed)
					html += '<div class="profile-group-items" style="display:none">';
				else html += '<div class="profile-group-items">';
			} else {
				html += '<div class="profile-group-items">';
			}

			groups[groupName].forEach((p) => {
				const icon =
					p.type === "ssh"
						? "🔐"
						: p.type === "serial"
							? "📡"
							: p.type === "telnet"
								? "🌐"
								: "⌘";

				html +=
					'<div class="profile-item" data-profile-id="' +
					p.id +
					'" title="' +
					p.name +
					'"><span class="profile-icon">' +
					icon +
					'</span><span class="profile-name">' +
					p.name +
					"</span></div>";
			});

			html += "</div>";
		});

	list.innerHTML = html;

	list.querySelectorAll(".profile-group-header").forEach((h) => {
		h.onclick = (e) => {
			const items = h.nextElementSibling;

			if (items) {
				const collapsed = items.style.display === "none";

				items.style.display = collapsed ? "" : "none";

				localStorage.setItem(
					"tabby-group-collapsed-" + h.dataset.group,
					collapsed ? "false" : "true",
				);

				renderProfileGroups();
			}
		};
	});

	list.querySelectorAll(".profile-item").forEach((el) => {
		el.onclick = () => {
			const profile = savedProfiles.find((p) => p.id === el.dataset.profileId);

			if (profile) connectProfile(profile);
		};
	});
}

function closeOtherTabs(keepTab) {
	const tabs = [...openTabs];

	tabs.forEach((t) => {
		if (t !== keepTab) closeTab(t);
	});
}

function duplicateTab(tab) {
	if (!tab) return;

	const newT = newTab(tab.title);

	if (newT && tab.connectionType) {
		newT.connectionType = tab.connectionType;
	}
}

// ===== DUPLICATE PROFILE =====

function duplicateActiveProfile() {
	const activeTab = getActiveTab();

	if (!activeTab) {
		showToast("No active tab", "error");
		return;
	}

	const matchingProfile = savedProfiles.find((p) => p.name === activeTab.title);

	if (matchingProfile) {
		const id = "profile-" + Date.now();

		const dup = JSON.parse(JSON.stringify(matchingProfile));

		dup.id = id;

		dup.name = matchingProfile.name + " (copy)";

		dup.createdAt = new Date().toISOString();

		dup.updatedAt = new Date().toISOString();

		savedProfiles.push(dup);

		SaveProfiles(savedProfiles).catch(() => {});

		renderProfiles();

		showToast("Duplicated: " + dup.name, "success");
	} else {
		showToast("No matching profile found", "error");
	}
}

// ===== CONNECTION LOG VIEWER =====

// ===== CONNECTION LOG =====

function setTabStatus(tab, status) {
	if (!tab) return;

	tab.status = status;
}

function logConnection(tab, message) {
	if (!tab) return;

	tab.connectionLog.push({ time: new Date().toISOString(), message });

	if (tab.connectionLog.length > 500) tab.connectionLog.shift();
}

function closeTab(tab) {
	tab.close();
}

function closeAllTabs(keepTab) {
	if (!keepTab) {
		[...tabs].forEach((t) => t.close());
	} else {
		tabs.forEach((t) => {
			if (t !== keepTab) t.close();
		});
	}
}

function getProfileGroups() {
	const groups = {};

	savedProfiles.forEach((p) => {
		const group = p.group || "Ungrouped";

		if (!groups[group]) groups[group] = [];

		groups[group].push(p);
	});

	return groups;
}

function applySettings() {
	if (!settings) return;

	tabs.forEach((t) => {
		if (t.term) {
			if (settings.FontFamily) t.term.options.fontFamily = settings.FontFamily;

			if (settings.FontSize) t.term.options.fontSize = settings.FontSize;

			if (settings.LineHeight)
				t.term.options.lineHeight = parseFloat(settings.LineHeight);

			if (settings.CursorStyle)
				t.term.options.cursorStyle = settings.CursorStyle;

			if (settings.CursorBlink !== undefined)
				t.term.options.cursorBlink = settings.CursorBlink;

			if (settings.Scrollback) t.term.options.scrollback = settings.Scrollback;

			t.fitAddon.fit();

			if (t.ptyId && !t.exited) PTYResize(t.ptyId, t.term.cols, t.term.rows);

			if (t.isSSH && t.sshConnectionId && t.sshSessionId)
				SSHResize({
					connectionId: t.sshConnectionId,
					sessionId: t.sshSessionId,
					columns: t.term.cols,
					rows: t.term.rows,
				});
		}
	});
}

function showConnectionLog() {
	let panel = document.getElementById("conn-log-panel");

	if (panel) {
		panel.remove();
		return;
	}

	panel = document.createElement("div");

	panel.id = "conn-log-panel";
	panel.className = "modal-overlay";

	const dlg = document.createElement("div");

	dlg.className = "modal-dialog";

	dlg.style.cssText = "max-width:700px;max-height:80vh;";

	const hdr = document.createElement("div");

	hdr.className = "modal-header";

	hdr.innerHTML = "<h3>Connection Log</h3>";

	const body = document.createElement("div");

	body.className = "modal-body";

	body.style.cssText = "max-height:60vh;overflow-y:auto;";

	const allLogs = [];

	tabs.forEach((t) => {
		if (t.connectionLog) allLogs.push(...t.connectionLog);
	});

	if (allLogs.length === 0) {
		body.innerHTML =
			'<p style="color:var(--text-secondary);text-align:center;">No connection events</p>';
	} else {
		let logHtml = '<div class="connection-log-entries">';

		allLogs
			.sort((a, b) => b.time - a.time)
			.forEach((entry) => {
				const time = new Date(entry.time).toLocaleTimeString();

				const typeClass =
					entry.type === "error"
						? "log-error"
						: entry.type === "connect"
							? "log-connect"
							: "log-info";

				logHtml +=
					'<div class="log-entry ' +
					typeClass +
					'"><span class="log-time">' +
					time +
					'</span><span class="log-msg">' +
					entry.message +
					"</span></div>";
			});

		logHtml += "</div>";

		body.innerHTML = logHtml;
	}

	const ftr = document.createElement("div");

	ftr.className = "modal-footer";

	const closeBtn = document.createElement("button");

	closeBtn.className = "btn btn-secondary";
	closeBtn.textContent = "Close";

	const clearBtn = document.createElement("button");

	clearBtn.className = "btn btn-danger";
	clearBtn.textContent = "Clear Log";
	clearBtn.style.marginRight = "auto";

	ftr.appendChild(clearBtn);
	ftr.appendChild(closeBtn);

	dlg.appendChild(hdr);
	dlg.appendChild(body);
	dlg.appendChild(ftr);

	panel.appendChild(dlg);
	document.body.appendChild(panel);

	closeBtn.onclick = () => panel.remove();

	panel.onclick = (e) => {
		if (e.target === panel) panel.remove();
	};

	clearBtn.onclick = () => {
		tabs.forEach((t) => {
			if (t.connectionLog) t.connectionLog = [];
		});
		body.innerHTML =
			'<p style="color:var(--text-secondary);text-align:center;">Log cleared</p>';
	};
}

// ===== UPDATER =====

async function checkForUpdates() {
	try {
		const status = await CheckForUpdates();

		if (status && status.updateAvailable) {
			showToast("Update available: v" + status.latestVersion, "success");
		} else if (status && !status.updateAvailable) {
			showToast(
				"You are on the latest version (v" + status.currentVersion + ")",
				"info",
			);
		} else if (status && status.error) {
			showToast("Update check failed: " + status.error, "error");
		}
	} catch (err) {
		showToast("Update check failed: " + err, "error");
	}
}

async function openAuditLog() {
	try {
		const path = await GetAuditLogPath();

		if (path) {
			showToast("Audit log: " + path, "info");
		} else {
			showToast("Audit logging not available", "error");
		}
	} catch (err) {
		showToast("Failed: " + err, "error");
	}
}

// ===== TAB COLOR LABELS =====

function setTabColor(tab, color) {
	if (!tab || !tab.tabEl) return;

	if (color) {
		tab.tabEl.style.borderLeft = "3px solid " + color;

		tab.colorLabel = color;
	} else {
		tab.tabEl.style.borderLeft = "";

		tab.colorLabel = "";
	}
}

// ===== SNIPPETS =====

const savedSnippets = JSON.parse(
	localStorage.getItem("tabby-snippets") || "[]",
);

function saveSnippet() {
	const name = prompt("Snippet name:");

	if (!name) return;

	const cmd = prompt("Snippet command:");

	if (!cmd) return;

	savedSnippets.push({ id: "snippet-" + Date.now(), name, command: cmd });

	localStorage.setItem("tabby-snippets", JSON.stringify(savedSnippets));

	showToast("Snippet saved: " + name, "success");
}

function runSnippet() {
	if (savedSnippets.length === 0) {
		showToast("No saved snippets", "info");
		return;
	}

	const names = savedSnippets
		.map((s, i) => i + 1 + ". " + s.name + ": " + s.command)
		.join("\n");

	const choice = prompt("Choose snippet:\n" + names);

	if (!choice) return;

	const idx = parseInt(choice) - 1;

	if (idx >= 0 && idx < savedSnippets.length) {
		const snippet = savedSnippets[idx];

		const tab = getActiveTab();

		if (tab) {
			tab.term.paste(snippet.command + "\r");

			showToast("Running: " + snippet.name, "info");
		} else {
			showToast("No active tab", "error");
		}
	}
}

// ===== PASSWORD DIALOG =====

function showPasswordDialog(title, message) {
	return new Promise((resolve) => {
		const overlay = document.createElement("div");

		overlay.className = "modal-overlay";

		const dlg = document.createElement("div");

		dlg.className = "modal-dialog";

		const hdr = document.createElement("div");

		hdr.className = "modal-header";

		hdr.innerHTML = "<h3>" + title + "</h3>";

		const body = document.createElement("div");

		body.className = "modal-body";

		body.innerHTML =
			"<p>" +
			message +
			'</p><input type="password" id="password-dialog-input" class="text-input" autofocus>';

		const ftr = document.createElement("div");

		ftr.className = "modal-footer";

		const cancelBtn = document.createElement("button");

		cancelBtn.className = "btn btn-secondary";

		cancelBtn.textContent = "Cancel";

		const okBtn = document.createElement("button");

		okBtn.className = "btn btn-primary";

		okBtn.textContent = "Connect";

		ftr.appendChild(cancelBtn);

		ftr.appendChild(okBtn);

		dlg.appendChild(hdr);

		dlg.appendChild(body);

		dlg.appendChild(ftr);

		overlay.appendChild(dlg);

		document.body.appendChild(overlay);

		const input = overlay.querySelector("#password-dialog-input");

		input.focus();

		input.addEventListener("keydown", (e) => {
			if (e.key === "Enter") {
				resolve(input.value);
				overlay.remove();
			}
			if (e.key === "Escape") {
				resolve(null);
				overlay.remove();
			}
		});

		okBtn.onclick = () => {
			resolve(input.value);
			overlay.remove();
		};

		cancelBtn.onclick = () => {
			resolve(null);
			overlay.remove();
		};
	});
}

// ===== KEYCHAIN =====

async function saveCredentialDialog() {
	const key = prompt("Enter credential key (e.g. ssh:host:user):");

	if (!key) return;

	const value = prompt("Enter credential value:");

	if (!value) return;

	try {
		await StoreCredential(key, value);

		showToast("Credential saved to keychain", "success");
	} catch (err) {
		showToast("Failed to save: " + err, "error");
	}
}

async function getCredentialDialog() {
	const key = prompt("Enter credential key to retrieve:");

	if (!key) return;

	try {
		const value = await GetCredential(key);

		if (value)
			showToast(
				"Credential: " + key + " = " + value.substring(0, 20) + "...",
				"success",
			);
		else showToast("Credential not found", "error");
	} catch (err) {
		showToast("Failed to retrieve: " + err, "error");
	}
}

// ===== PROFILE EXPORT/IMPORT =====

function exportProfiles() {
	const d = JSON.stringify(savedProfiles, null, 2);

	const blob = new Blob([d], { type: "application/json" });

	const url = URL.createObjectURL(blob);

	const a = document.createElement("a");

	a.href = url;

	a.download =
		"tabby-profiles-" + new Date().toISOString().slice(0, 10) + ".json";

	a.click();

	URL.revokeObjectURL(url);

	showToast("Profiles exported", "success");
}

function importProfiles() {
	const input = document.createElement("input");

	input.type = "file";

	input.accept = ".json";

	input.onchange = (e) => {
		const file = e.target.files[0];

		if (!file) return;

		const reader = new FileReader();

		reader.onload = (ev) => {
			try {
				const imported = JSON.parse(ev.target.result);

				if (!Array.isArray(imported)) throw new Error("Invalid format");

				let count = 0;

				imported.forEach((p) => {
					if (!savedProfiles.find((sp) => sp.id === p.id)) {
						savedProfiles.push(p);

						count++;
					}
				});

				SaveProfiles(savedProfiles).catch(() => {});

				renderProfiles();

				showToast("Imported " + count + " profiles", "success");
			} catch (err) {
				showToast("Import failed: " + err.message, "error");
			}
		};

		reader.readAsText(file);
	};

	input.click();
}

// ===== PROFILES =====

function renderProfiles() {
	const section = document.getElementById("profiles-section");
	const list = document.getElementById("profiles-list");
	const editor = document.getElementById("profiles-editor-list");
	if (!savedProfiles || savedProfiles.length === 0) {
		if (section) section.style.display = "none";
		if (editor)
			editor.innerHTML =
				'<div style="color:#666;font-size:12px;">No saved profiles yet.</div>';
		return;
	}
	if (section) section.style.display = "block";
	if (list) {
		renderProfileGroups();
		list.querySelectorAll(".profile-item").forEach((el) => {
			el.ondblclick = () => {
				const p = savedProfiles.find((x) => x.id === el.dataset.profileId);
				if (p) connectProfile(p);
			};
		});
	}
	if (editor) {
		editor.innerHTML = savedProfiles
			.map((p) => {
				const icon =
					p.type === "ssh"
						? "🔐"
						: p.type === "serial"
							? "📡"
							: p.type === "telnet"
								? "🌐"
								: "⌘";
				return `<div class="profile-editor-item"><span>${icon} ${p.name}</span><button class="btn-icon profile-edit" data-id="${p.id}" title="Edit">✎</button><button class="btn-icon profile-duplicate" data-id="${p.id}" title="Duplicate">➕</button>\u270e</button><button class="btn-icon profile-delete" data-id="${p.id}" title="Delete">\u00d7</button></div>`;
			})
			.join("");
		editor.querySelectorAll(".profile-edit").forEach((btn) => {
			btn.onclick = (e) => {
				e.stopPropagation();
				const id = btn.dataset.id;
				const profile = savedProfiles.find((p) => p.id === id);
				if (profile) editProfile(profile);
			};
		});
		editor.querySelectorAll(".profile-duplicate").forEach((btn) => {
			btn.onclick = (e) => {
				e.stopPropagation();

				const id = btn.dataset.id;

				const profile = savedProfiles.find((p) => p.id === id);

				if (profile) {
					const dup = JSON.parse(JSON.stringify(profile));

					dup.id = "profile-" + Date.now();

					dup.name = profile.name + " (copy)";

					dup.createdAt = new Date().toISOString();

					dup.updatedAt = new Date().toISOString();

					savedProfiles.push(dup);

					SaveProfiles(savedProfiles).catch(() => {});

					renderProfiles();

					showToast("Duplicated: " + dup.name, "success");
				}
			};
		});

		editor.querySelectorAll(".profile-delete").forEach((btn) => {
			btn.onclick = (e) => {
				e.stopPropagation();
				const id = btn.dataset.id;
				savedProfiles = savedProfiles.filter((p) => p.id !== id);
				SaveProfiles(savedProfiles).catch(() => {});
				renderProfiles();
			};
		});
	}
}

async function connectProfile(profile) {
	if (profile.type === "ssh") {
		const opts = profile.options;
		openSSHDialog();
		document.getElementById("ssh-host").value = opts.host || "";
		document.getElementById("ssh-port").value = opts.port || 22;
		document.getElementById("ssh-user").value = opts.user || "";
		document.getElementById("ssh-auth").value = opts.auth || "agent";
		document.getElementById("ssh-auth").dispatchEvent(new Event("change"));
		if (opts.auth === "password")
			document.getElementById("ssh-password").value = opts.password || "";
		if (
			opts.auth === "publicKey" &&
			opts.privateKeys &&
			opts.privateKeys.length
		)
			document.getElementById("ssh-key-path").value = opts.privateKeys[0];
	} else if (profile.type === "serial") {
		openSerialDialog();

		setTimeout(() => {
			const opts = profile.options;

			refreshSerialPorts();

			document.getElementById("serial-baud").value = opts.baudRate || 115200;

			document.getElementById("serial-data-bits").value = opts.dataBits || 8;

			document.getElementById("serial-stop-bits").value = opts.stopBits || 1;

			document.getElementById("serial-parity").value = opts.parity || "none";
		}, 300);
	} else if (profile.type === "telnet") {
		openTelnetDialog();

		const opts = profile.options;

		document.getElementById("telnet-host").value = opts.host || "";

		document.getElementById("telnet-port").value = opts.port || 23;
	} else if (profile.type === "local") {
		newTab(
			(profile.options && profile.options.shell) ||
				(profile.options && profile.options.command),
		);
	}
}

function addProfile() {
	const id = "profile-" + Date.now();
	savedProfiles.push({
		id,
		type: "ssh",
		name: "New SSH Profile",
		group: "",
		options: { host: "", port: 22, user: "", auth: "agent" },
		createdAt: new Date().toISOString(),
		updatedAt: new Date().toISOString(),
	});
	SaveProfiles(savedProfiles).catch(() => {});
	renderProfiles();
	showToast("Profile added", "info");
}

// ===== BUILD UI =====

function buildUI() {
	document.querySelector("#app").innerHTML = `

    <div id="sidebar">

        <div id="sidebar-header">

            <div class="logo"><span class="logo-accent">⌘</span> Tabby <small style="font-size:9px;color:#666;font-weight:400;">go</small></div>

            <div style="display:flex;gap:4px;align-items:center;">
                <button class="btn-icon" id="btn-new-tab" title="New Tab / Connect" style="font-size:18px;font-weight:bold;">+</button> <button class="btn-icon" id="btn-settings" title="Settings (Ctrl+,)">&#9881;</button>
            </div>

        </div>

        <div id="tab-list"></div>

        <div id="profiles-section" style="border-top:1px solid #2b2b2b;padding:8px 0;max-height:200px;overflow-y:auto;display:none;">

            <div style="padding:0 12px 4px;font-size:10px;color:#666;text-transform:uppercase;letter-spacing:0.5px;">Profiles</div>

            <div id="profiles-list"></div>

        </div>

    </div>

    <div id="main-content">

        <div id="welcome">
 <div class="welcome-screen">
 <div class="tabby-logo"><span class="logo-accent">⌘</span> Tabby <small>go</small></div>
 <p style="color:#777;font-size:14px;margin:8px 0 24px;">A modern terminal for a modern age</p>
 <div class="quick-connect-grid">
 <div class="quick-connect-btn" onclick="newTab()"><span class="qc-icon">⌘</span><div><div class="qc-name">Local Shell</div><div class="qc-desc">Open a terminal</div></div></div>
 <div class="quick-connect-btn" onclick="openSSHDialog()"><span class="qc-icon">🔐</span><div><div class="qc-name">SSH Connect</div><div class="qc-desc">Connect to a remote host</div></div></div>
 <div class="quick-connect-btn" onclick="openSerialDialog()"><span class="qc-icon">📡</span><div><div class="qc-name">Serial Port</div><div class="qc-desc">Connect to hardware</div></div></div>
 <div class="quick-connect-btn" onclick="openTelnetDialog()"><span class="qc-icon">🌐</span><div><div class="qc-name">Telnet</div><div class="qc-desc">Connect to a telnet server</div></div></div>
 </div>
 <div class="shortcuts" style="margin-top:40px;">
 <div class="shortcut"><kbd>Ctrl+Shift+T</kbd> New Tab</div>
 <div class="shortcut"><kbd>Ctrl+W</kbd> Close Tab</div>
 <div class="shortcut"><kbd>Ctrl+Tab</kbd> Next Tab</div>
 <div class="shortcut"><kbd>Ctrl+Shift+P</kbd> Command Palette</div>
 <div class="shortcut"><kbd>Ctrl+Shift+F</kbd> Find</div>
 <div class="shortcut"><kbd>Ctrl+</kbd> Split Vertical</div>
 <div class="shortcut"><kbd>Ctrl+,</kbd> Settings</div>
 </div>
 </div>
 </div>
<div id="settings-panel">

        <div id="settings-header">

            <h2>⚙ Settings</h2>

            <button id="settings-close" class="btn-icon">×</button>

        </div>

        <div id="settings-tabs">

            <button class="settings-tab active" data-tab="appearance">🎨 Appearance</button>

            <button class="settings-tab" data-tab="terminal">⌨ Terminal</button>

            <button class="settings-tab" data-tab="clipboard">📋 Clipboard</button>

            <button class="settings-tab" data-tab="mouse">🖱 Mouse</button>

            <button class="settings-tab" data-tab="tabs">📑 Tabs</button>

            <button class="settings-tab" data-tab="startup">🚀 Startup</button>

            <button class="settings-tab" data-tab="ssh">🔐 SSH</button>

            <button class="settings-tab" data-tab="serial">📡 Serial</button>

            <button class="settings-tab" data-tab="profiles">📁 Profiles</button>

        </div>

        <div id="settings-content">

            <!-- Appearance -->

            <div class="settings-page active" id="settings-appearance">

                <h3>Color Scheme</h3>

                <div class="setting-group"><label>Terminal Color Scheme</label><select id="s-color-scheme"></select></div>

                <div id="color-scheme-preview" style="display:flex;flex-wrap:wrap;gap:3px;padding:8px 0;"></div>

                <h3>Font</h3>

                <div class="setting-group"><label>Font Family</label><input type="text" id="s-font-family" placeholder="Cascadia Code, Fira Code, Consolas..."></div>

                <div class="setting-group"><label>Font Size</label>

                    <div class="slider-container"><input type="range" id="s-font-size" min="8" max="48" value="14"><span class="slider-label" id="s-font-size-val">14</span>px</div>

                </div>

                <div class="setting-group"><label>Fallback Font</label><input type="text" id="s-fallback-font" placeholder="e.g. Nerd Font"></div>

                <div class="setting-group"><label>Normal Font Weight</label><input type="number" id="s-font-weight" min="100" max="900" step="100" value="400"></div>

                <div class="setting-group"><label>Bold Font Weight</label><input type="number" id="s-font-weight-bold" min="100" max="900" step="100" value="700"></div>

                <div class="setting-group"><label>Line Height</label>

                    <div class="slider-container"><input type="range" id="s-line-height" min="1.0" max="2.0" step="0.05" value="1.2"><span class="slider-label" id="s-line-height-val">1.2</span>×</div>

                </div>

                <div class="setting-group"><label>Line Padding</label><input type="number" id="s-line-padding" min="0" max="10" value="0"></div>

                <div class="setting-group"><label>Enable Font Ligatures</label><div class="toggle-container"><input type="checkbox" id="s-ligatures"><label for="s-ligatures" class="toggle-label"></label></div></div>

                <h3>Theme & Colors</h3>

                <div class="setting-group"><label>Color Scheme Mode</label>

                    <select id="s-theme"><option value="auto">Auto (Follow System)</option><option value="dark">Always Dark</option><option value="light">Always Light</option></select>

                </div>

                <div class="setting-group"><label>Opacity</label>

                    <div class="slider-container"><input type="range" id="s-opacity" min="0.3" max="1.0" step="0.05" value="1.0"><span class="slider-label" id="s-opacity-val">1.0</span></div>

                </div>

                <div class="setting-group"><label>Background Color</label><input type="color" id="s-bg-color" value="#000000" style="width:60px;height:30px;border:1px solid #3a3a3a;border-radius:4px;cursor:pointer;"></div>

                <div class="setting-group"><label>UI Spacing</label>

                    <select id="s-spaciness"><option value="1">Compact</option><option value="2">Normal</option><option value="3">Spacious</option></select>

                </div>

                <div class="setting-group"><label>Enable Animations</label><div class="toggle-container"><input type="checkbox" id="s-animations" checked><label for="s-animations" class="toggle-label"></label></div></div>

                <h3>Cursor</h3>

                <div class="setting-group"><label>Cursor Shape</label>

                    <select id="s-cursor-style"><option value="bar">Bar (|)</option><option value="block">Block (█)</option><option value="underline">Underline (▁)</option></select>

                </div>

                <div class="setting-group"><label>Blink Cursor</label><div class="toggle-container"><input type="checkbox" id="s-cursor-blink" checked><label for="s-cursor-blink" class="toggle-label"></label></div></div>

                <h3>Rendering</h3>

                <div class="setting-group"><label>Terminal Frontend</label>

                    <select id="s-frontend"><option value="xterm-webgl">xterm (WebGL)</option><option value="xterm">xterm (Canvas)</option><option value="block">Block Frontend (Experimental)</option></select>

                </div>

                <div class="setting-group"><label>Draw Bold Text in Bright Colors</label><div class="toggle-container"><input type="checkbox" id="s-draw-bold-bright" checked><label for="s-draw-bold-bright" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Minimum Contrast Ratio</label><input type="number" id="s-min-contrast" min="1" max="21" step="0.5" value="4"></div>

                <h3>Custom CSS</h3>

                <div class="setting-group"><textarea id="s-css" rows="4" placeholder="/* Custom CSS */"></textarea></div>

            </div>

            <!-- Terminal -->

            <div class="settings-page" id="settings-terminal">

                <h3>Shell</h3>

                <div class="setting-group"><label>Default Shell</label>

                    <select id="s-shell"><option value="">Auto-detect</option></select>

                </div>

                <h3>Scrollback</h3>

                <div class="setting-group"><label>Scrollback Lines</label><input type="number" id="s-scrollback" min="100" max="1000000" step="1000" value="25000"></div>

                <h3>Bell</h3>

                <div class="setting-group"><label>Terminal Bell</label>

                    <select id="s-bell"><option value="off">Off</option><option value="visual">Visual</option><option value="audible">Audible</option></select>

                </div>

                <h3>Keyboard</h3>

                <div class="setting-group"><label>Use Alt as Meta key</label><div class="toggle-container"><input type="checkbox" id="s-alt-is-meta"><label for="s-alt-is-meta" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Scroll on Input</label><div class="toggle-container"><input type="checkbox" id="s-scroll-on-input" checked><label for="s-scroll-on-input" class="toggle-label"></label></div></div>

                <h3>Windows</h3>

                <div class="setting-group"><label>Use ConPTY</label><div class="toggle-container"><input type="checkbox" id="s-use-conpty" checked><label for="s-use-conpty" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Set Tabby as %COMSPEC%</label><div class="toggle-container"><input type="checkbox" id="s-set-comspec"><label for="s-set-comspec" class="toggle-label"></label></div></div>

                <h3>Toolbar</h3>

            </div>

            <!-- Clipboard -->

            <div class="settings-page" id="settings-clipboard">

                <h3>Clipboard</h3>

                <div class="setting-group"><label>Copy on Select</label><div class="toggle-container"><input type="checkbox" id="s-copy-on-select"><label for="s-copy-on-select" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Copy with Formatting (HTML)</label><div class="toggle-container"><input type="checkbox" id="s-copy-as-html" checked><label for="s-copy-as-html" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Bracketed Paste</label><div class="toggle-container"><input type="checkbox" id="s-bracketed-paste" checked><label for="s-bracketed-paste" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Session Logging</label><div class="toggle-container"><input type="checkbox" id="s-session-logging"><label for="s-session-logging" class="toggle-label"></label></div><span style="color:#666;font-size:11px;margin-top:2px;">Capture all terminal output to a session log</span></div>

                <div class="setting-group"><label>Warn on Multi-line Paste</label><div class="toggle-container"><input type="checkbox" id="s-warn-multiline" checked><label for="s-warn-multiline" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Replace Line Breaks with Spaces on Paste</label><div class="toggle-container"><input type="checkbox" id="s-replace-newlines"><label for="s-replace-newlines" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Trim Whitespace and Newlines</label><div class="toggle-container"><input type="checkbox" id="s-trim-whitespace" checked><label for="s-trim-whitespace" class="toggle-label"></label></div></div>

            </div>

            <!-- Mouse -->

            <div class="settings-page" id="settings-mouse">

                <h3>Mouse</h3>

                <div class="setting-group"><label>Right Click</label>

                    <select id="s-right-click"><option value="off">Off</option><option value="menu">Context Menu</option><option value="paste">Paste</option><option value="clipboard">Paste if No Selection, Else Copy</option></select>

                </div>

                <div class="setting-group"><label>Paste on Middle-Click</label><div class="toggle-container"><input type="checkbox" id="s-paste-middle-click" checked><label for="s-paste-middle-click" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Word Separators</label><input type="text" id="s-word-separator" value=" ()[]{}'&quot;"></div>

            </div>

            <!-- Tabs -->

            <div class="settings-page" id="settings-tabs">

                <h3>Tabs</h3>

                <div class="setting-group"><label>Tab Position</label>

                    <select id="s-tab-position"><option value="left">Left</option><option value="right">Right</option><option value="top">Top</option><option value="bottom">Bottom</option></select>

                </div>

                <div class="setting-group"><label>Last Tab Closes Window</label><div class="toggle-container"><input type="checkbox" id="s-last-tab-closes"><label for="s-last-tab-closes" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Cycle Tabs (wrap around)</label><div class="toggle-container"><input type="checkbox" id="s-cycle-tabs" checked><label for="s-cycle-tabs" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Hide Close Button</label><div class="toggle-container"><input type="checkbox" id="s-hide-close-button"><label for="s-hide-close-button" class="toggle-label"></label></div></div>

                <h3>Split Panes</h3>

                <div class="setting-group"><label>Pane Resize Step</label><input type="number" id="s-pane-resize-step" min="0.01" max="1.0" step="0.01" value="0.1"></div>

                <div class="setting-group"><label>Focus Follows Mouse</label><div class="toggle-container"><input type="checkbox" id="s-focus-follows-mouse"><label for="s-focus-follows-mouse" class="toggle-label"></label></div></div>

            </div>

            <!-- Startup -->

            <div class="settings-page" id="settings-startup">

                <h3>Startup</h3>

                <div class="setting-group"><label>Auto-open Terminal on Start</label><div class="toggle-container"><input type="checkbox" id="s-auto-open" checked><label for="s-auto-open" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Restore Tabs on Restart</label><div class="toggle-container"><input type="checkbox" id="s-recover-tabs" checked><label for="s-recover-tabs" class="toggle-label"></label></div></div>

                <h3>Window</h3>

                <div class="setting-group"><label>Window Frame</label>

                    <select id="s-frame"><option value="thin">Thin</option><option value="none">None (Borderless)</option><option value="native">Native</option></select>

                </div>

                <div class="setting-group"><label>Quake-style Dock</label>

                    <select id="s-dock"><option value="off">Off</option><option value="bottom">Bottom</option><option value="top">Top</option><option value="left">Left</option><option value="right">Right</option></select>

                </div>

                <div class="setting-group"><label>Hide Dock on Blur</label><div class="toggle-container"><input type="checkbox" id="s-dock-hide-blur"><label for="s-dock-hide-blur" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Dock Always on Top</label><div class="toggle-container"><input type="checkbox" id="s-dock-on-top" checked><label for="s-dock-on-top" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Hide Tray Icon</label><div class="toggle-container"><input type="checkbox" id="s-hide-tray"><label for="s-hide-tray" class="toggle-label"></label></div></div>

                <h3>Misc</h3>

                <div class="setting-group"><label>Language</label><input type="text" id="s-language" placeholder="Auto"></div>

                <div class="setting-group"><label>Enable Analytics</label><div class="toggle-container"><input type="checkbox" id="s-analytics" checked><label for="s-analytics" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Automatic Updates</label><div class="toggle-container"><input type="checkbox" id="s-auto-updates" checked><label for="s-auto-updates" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Experimental Features</label><div class="toggle-container"><input type="checkbox" id="s-experimental"><label for="s-experimental" class="toggle-label"></label></div></div>

            </div>

            <!-- SSH -->

            <div class="settings-page" id="settings-ssh">

                <h3>SSH</h3>

                <div class="setting-group"><label>Warn When Closing Active Connections</label><div class="toggle-container"><input type="checkbox" id="s-ssh-warn-close"><label for="s-ssh-warn-close" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Verify Host Keys When Connecting</label><div class="toggle-container"><input type="checkbox" id="s-ssh-verify-keys" checked><label for="s-ssh-verify-keys" class="toggle-label"></label></div></div>

                <div class="setting-group"><label>Disable Dynamic Title (SSH)</label><div class="toggle-container"><input type="checkbox" id="s-ssh-disable-title" checked><label for="s-ssh-disable-title" class="toggle-label"></label></div></div>

                <h3>Agent</h3>

                <div class="setting-group"><label>SSH Agent Type</label>

                    <select id="s-ssh-agent-type"><option value="auto">Automatic</option><option value="pageant">Pageant</option><option value="pipe">Named Pipe</option></select>

                </div>

                <div class="setting-group"><label>Agent Pipe Path</label><input type="text" id="s-ssh-agent-path" placeholder="Default: \\\\.\\pipe\\openssh-ssh-agent"></div>

                <h3>X11 Forwarding</h3>

                <div class="setting-group"><label>Override X11 Display</label><input type="text" id="s-ssh-x11" placeholder="Auto-detect"></div>

            </div>

            <!-- Serial -->

            <div class="settings-page" id="settings-serial">

                <h3>Serial Port Defaults</h3>

                <div class="setting-group"><label>Baud Rate</label>

                    <select id="s-serial-baud">

                        <option value="9600">9600</option><option value="19200">19200</option><option value="38400">38400</option>

                        <option value="57600">57600</option><option value="115200" selected>115200</option><option value="230400">230400</option>

                        <option value="460800">460800</option><option value="921600">921600</option><option value="1000000">1000000</option>

                    </select>

                </div>

                <div class="setting-group"><label>Data Bits</label>

                    <select id="s-serial-data-bits"><option value="5">5</option><option value="6">6</option><option value="7">7</option><option value="8" selected>8</option></select>

                </div>

                <div class="setting-group"><label>Stop Bits</label>

                    <select id="s-serial-stop-bits"><option value="1" selected>1</option><option value="2">2</option></select>

                </div>

                <div class="setting-group"><label>Parity</label>

                    <select id="s-serial-parity"><option value="none" selected>None</option><option value="even">Even</option><option value="odd">Odd</option></select>

                </div>

                <div class="setting-group"><label>Flow Control</label>

                    <select id="s-serial-flow"><option value="none" selected>None</option><option value="hardware">Hardware (RTS/CTS)</option><option value="software">Software (XON/XOFF)</option></select>

                </div>

            </div>

            <div class="settings-page" id="settings-profiles">

                <h3>Saved Profiles</h3>

                <div id="profiles-editor-list" style="margin-bottom:12px;"></div>

                <button class="btn-primary" id="btn-add-profile" style="width:100%;margin-bottom:8px;">+ Add SSH Profile</button>

            </div>

        </div>

        <div id="settings-actions">

            <button id="btn-reset" class="btn-secondary">Reset to Defaults</button>

            <button id="btn-save" class="btn-primary">Save Settings</button>

        </div>

    </div>

    </div>`;

	// Populate color scheme dropdown

	const schemeSelect = document.getElementById("s-color-scheme");

	schemeNames.forEach((name) => {
		const opt = document.createElement("option");
		opt.value = name;
		opt.textContent = name;
		schemeSelect.appendChild(opt);
	});

	schemeSelect.onchange = () => applyColorScheme(schemeSelect.value);

	// Populate shell dropdown

	const shellSelect = document.getElementById("s-shell");

	availableShells.forEach((s) => {
		const opt = document.createElement("option");

		opt.value = s;

		opt.textContent = `${s.split(/[/\\]/).pop().replace(".exe", "")}  (${s})`;

		shellSelect.appendChild(opt);
	});

	// Render profiles

	renderProfiles();

	// Settings tab navigation

	document.querySelectorAll(".settings-tab").forEach((btn) => {
		btn.onclick = () => {
			document
				.querySelectorAll(".settings-tab")
				.forEach((b) => b.classList.remove("active"));

			document
				.querySelectorAll(".settings-page")
				.forEach((p) => p.classList.remove("active"));

			btn.classList.add("active");

			document
				.getElementById(`settings-${btn.dataset.tab}`)
				.classList.add("active");
		};
	});

	// Live-update sliders

	document.getElementById("s-font-size").oninput = (e) => {
		document.getElementById("s-font-size-val").textContent = e.target.value;

		applyFontSize(parseInt(e.target.value));
	};

	document.getElementById("s-line-height").oninput = (e) => {
		document.getElementById("s-line-height-val").textContent = parseFloat(
			e.target.value,
		).toFixed(2);

		tabs.forEach((t) => {
			if (t.term) {
				t.term.options.lineHeight = parseFloat(e.target.value);
				t.fitAddon.fit();
			}
		});
	};

	document.getElementById("s-opacity").oninput = (e) => {
		document.getElementById("s-opacity-val").textContent = parseFloat(
			e.target.value,
		).toFixed(2);

		document.getElementById("main-content").style.opacity = parseFloat(
			e.target.value,
		);
	};

	// Background color picker
	document.getElementById("s-bg-color").oninput = (e) => {
		applyBackgroundColor(e.target.value);
	};

	// SSH auth toggle

	document.getElementById("ssh-auth").onchange = (e) => {
		document.getElementById("ssh-password-group").style.display =
			e.target.value === "password" ? "block" : "none";
		document.getElementById("ssh-key-group").style.display =
			e.target.value === "publicKey" ? "block" : "none";
		document.getElementById("ssh-passphrase-group").style.display =
			e.target.value === "publicKey" ? "block" : "none";
	};

	// Button bindings

	document.getElementById("btn-new-tab").onclick = (e) => showNewTabDropdown(e);

	document.getElementById("btn-settings").onclick = () => toggleSettings();

	document.getElementById("serial-refresh").onclick = () =>
		refreshSerialPorts();

	document.getElementById("serial-cancel").onclick = () => closeSerialDialog();

	document.getElementById("serial-connect").onclick = () => doSerialConnect();

	document.getElementById("telnet-cancel").onclick = () => closeTelnetDialog();

	document.getElementById("telnet-connect").onclick = () => doTelnetConnect();

	document.getElementById("forward-cancel").onclick = () =>
		closeForwardDialog();

	document.getElementById("forward-add").onclick = () => doAddForward();

	document.getElementById("forward-type").onchange = () =>
		toggleForwardFields();

	document.getElementById("hostkey-accept").onclick = () =>
		closeHostKeyDialog(true);

	document.getElementById("hostkey-reject").onclick = () =>
		closeHostKeyDialog(false);

	document.getElementById("sftp-go-up").onclick = () => sftpGoUp();

	document.getElementById("sftp-refresh").onclick = () =>
		sftpNavigate(sftpCurrentPath);

	document.getElementById("sftp-go").onclick = () =>
		sftpNavigate(document.getElementById("sftp-path").value.trim());

	document.getElementById("sftp-path").onkeydown = (e) => {
		if (e.key === "Enter")
			sftpNavigate(document.getElementById("sftp-path").value.trim());
	};

	document.getElementById("sftp-mkdir-btn").onclick = () => sftpMkdir();

	document.getElementById("sftp-upload-btn").onclick = () =>
		document.getElementById("sftp-upload-input").click();

	document.getElementById("sftp-upload-input").onchange = (e) =>
		sftpUploadFile(e);

	document.getElementById("sftp-download-btn").onclick = () =>
		sftpDownloadSelected();

	document.getElementById("sftp-delete-btn").onclick = () =>
		sftpDeleteSelected();

	document.getElementById("sftp-dialog").ondragover = (e) => {
		e.preventDefault();
		e.dataTransfer.dropEffect = "copy";
	};

	document.getElementById("sftp-dialog").ondrop = (e) => {
		e.preventDefault();
		sftpHandleDrop(e);
	};

	document.getElementById("sftp-close-btn").onclick = () => closeSFTPBrowser();

	// Import SSH config is available via Command Palette

	document.getElementById("cmd-palette-input").oninput = () =>
		filterCommandPalette();

	document.getElementById("cmd-palette-input").onkeydown = (e) =>
		handlePaletteKey(e);

	document.getElementById("settings-close").onclick = () => hideSettings();

	document.getElementById("btn-save").onclick = () => saveSettingsFromUI();

	document.getElementById("ssh-cancel").onclick = () => closeSSHDialog();

	document.getElementById("ssh-connect").onclick = () => doSSHConnect();

	document.getElementById("btn-add-profile").onclick = () => addProfile();

	document.getElementById("edit-profile-save").onclick = () =>
		saveProfileEdit();

	document.getElementById("edit-profile-cancel").onclick = () =>
		closeProfileEditor();

	document.getElementById("tab-search-input").oninput = () => filterTabs();

	document.getElementById("tab-search-input").onkeydown = (e) => {
		if (e.key === "Escape") toggleTabSearch();
	};

	document.getElementById("btn-reset").onclick = () => doResetSettings();
}

// ===== SETTINGS APPLY/SAVE =====

// Inject descriptions below each setting in the settings panel
function applySettingDescriptions() {
	// Remove any existing descriptions first
	document.querySelectorAll(".setting-description").forEach((el) => el.remove());

	Object.entries(SETTING_DESCRIPTIONS).forEach(([id, desc]) => {
		const el = document.getElementById(id);
		if (!el) return;

		let group = el.closest(".setting-group");
		if (!group) {
			// Handle settings inside a slider-container wrapper
			const slider = el.closest(".slider-container");
			if (slider) group = slider.closest(".setting-group");
		}
		if (!group) return;

		const descEl = document.createElement("div");
		descEl.className = "setting-description";
		descEl.textContent = desc;
		group.appendChild(descEl);
	});
}

function applySettingsToUI() {
	const s = settings;

	const set = (id, val) => {
		const el = document.getElementById(id);
		if (el) el.value = val ?? "";
	};

	const check = (id, val) => {
		const el = document.getElementById(id);
		if (el) el.checked = !!val;
	};

	// Appearance

	set("s-color-scheme", s.ColorScheme || "Tabby Default");

	if (s.ColorScheme) applyColorScheme(s.ColorScheme);

	set("s-font-family", s.FontFamily);

	set("s-font-size", s.FontSize || 14);

	document.getElementById("s-font-size-val").textContent = s.FontSize || 14;

	if (s.FontSize) applyFontSize(s.FontSize);

	set("s-fallback-font", s.FallbackFont);

	set("s-font-weight", s.FontWeight || 400);

	set("s-font-weight-bold", s.FontWeightBold || 700);

	set("s-line-height", s.LineHeight || 1.2);

	document.getElementById("s-line-height-val").textContent =
		s.LineHeight || 1.2;

	set("s-line-padding", s.LinePadding || 0);

	check("s-ligatures", s.Ligatures);

	set("s-theme", s.Theme || "dark");

	applyTheme(s.Theme || "dark");

	set("s-opacity", s.Opacity ?? 1.0);

	document.getElementById("s-opacity-val").textContent = (
		s.Opacity ?? 1.0
	).toFixed(2);

	// Background color
	const bgColor = s.BackgroundColor || "#000000";
	const bgInput = document.getElementById("s-bg-color");
	if (bgInput) bgInput.value = bgColor;
	applyBackgroundColor(bgColor);

	set("s-spaciness", s.Spaciness || 1);

	check("s-animations", s.Animations ?? true);

	set("s-cursor-style", s.CursorStyle || "bar");

	check("s-cursor-blink", s.CursorBlink ?? true);

	set("s-frontend", s.Frontend || "xterm-webgl");

	check("s-draw-bold-bright", s.DrawBoldTextInBrightColors ?? true);

	set("s-min-contrast", s.MinimumContrastRatio ?? 4);

	set("s-css", s.CSS || "");

	// Terminal

	set("s-shell", s.Shell || "");

	set("s-scrollback", s.Scrollback || 25000);

	set("s-bell", s.Bell || "off");

	check("s-alt-is-meta", s.AltIsMeta);

	check("s-scroll-on-input", s.ScrollOnInput ?? true);

	check("s-use-conpty", s.UseConPTY ?? true);

	check("s-set-comspec", s.SetComSpec);

	// Clipboard

	check("s-copy-on-select", s.CopyOnSelect);

	check("s-copy-as-html", s.CopyAsHTML ?? true);

	check("s-bracketed-paste", s.BracketedPaste ?? true);
	check("s-session-logging", s.SessionLogging ?? false);

	check("s-warn-multiline", s.WarnOnMultilinePaste ?? true);

	check("s-replace-newlines", s.ReplaceNewlinesOnPaste);

	check("s-trim-whitespace", s.TrimWhitespaceOnPaste ?? true);

	// Mouse

	set("s-right-click", s.RightClick || "menu");

	check("s-paste-middle-click", s.PasteOnMiddleClick ?? true);

	set("s-word-separator", s.WordSeparator || " ()[]{}'\"");

	// Tabs

	set("s-tab-position", s.TabPosition || "left");

	check("s-last-tab-closes", s.LastTabClosesWindow);

	check("s-cycle-tabs", s.CycleTabs ?? true);

	check("s-hide-close-button", s.HideCloseButton);

	set("s-pane-resize-step", s.PaneResizeStep ?? 0.1);

	check("s-focus-follows-mouse", s.FocusFollowsMouse);

	// Startup

	check("s-auto-open", s.AutoOpen ?? true);

	check("s-recover-tabs", s.RecoverTabs ?? true);

	set("s-frame", s.Frame || "thin");

	set("s-dock", s.Dock || "off");

	check("s-dock-hide-blur", s.DockHideOnBlur);

	check("s-dock-on-top", s.DockAlwaysOnTop ?? true);

	check("s-hide-tray", s.HideTray);

	set("s-language", s.Language || "");

	check("s-analytics", s.EnableAnalytics ?? true);

	check("s-auto-updates", s.EnableAutomaticUpdates ?? true);

	check("s-experimental", s.EnableExperimentalFeatures);

	// SSH

	check("s-ssh-warn-close", s.SSHWarnOnClose);

	check("s-ssh-verify-keys", s.SSHVerifyHostKeys ?? true);

	check("s-ssh-disable-title", s.SSHDisableDynamicTitle ?? true);

	set("s-ssh-agent-type", s.SSHAgentType || "auto");

	set("s-ssh-agent-path", s.SSHAgentPath || "");

	set("s-ssh-x11", s.SSHX11Display || "");

	// Serial

	set("s-serial-baud", s.SerialBaudRate || 115200);

	set("s-serial-data-bits", s.SerialDataBits || 8);

	set("s-serial-stop-bits", s.SerialStopBits || 1);

	set("s-serial-parity", s.SerialParity || "none");

	set("s-serial-flow", s.SerialFlowControl || "none");

	applySettingDescriptions();
}

function applyFontSize(size) {
	fontSize = size;

	tabs.forEach((t) => {
		if (t.term) {
			t.term.options.fontSize = size;
			t.fitAddon.fit();
			if (t.ptyId) PTYResize(t.ptyId, t.term.cols, t.term.rows);
			if (t.isSSH && t.sshConnectionId && t.sshSessionId)
				SSHResize({
					connectionId: t.sshConnectionId,
					sessionId: t.sshSessionId,
					columns: t.term.cols,
					rows: t.term.rows,
				});
		}
	});
}

function applyTheme(theme) {
	if (theme === "light") {
		document.body.classList.add("light-theme");
		document.body.classList.remove("dark-theme");
	} else if (theme === "dark") {
		document.body.classList.add("dark-theme");
		document.body.classList.remove("light-theme");
	} else {
		document.body.classList.remove("light-theme");
		document.body.classList.remove("dark-theme");
	}

	// Update xterm themes

	const isDark = !isSchemeLight(settings.ColorScheme || "Tabby Default");

	const schemeTheme = getColorSchemeTheme(
		settings.ColorScheme || "Tabby Default",
	);
	tabs.forEach((t) => {
		if (t.term && schemeTheme) t.term.options.theme = schemeTheme;
	});
}

const FALLBACK_DARK = {
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
};

async function saveSettingsFromUI() {
	const s = {
		ColorScheme:
			document.getElementById("s-color-scheme").value || "Tabby Default",

		FontFamily: document.getElementById("s-font-family").value.trim(),

		FontSize: parseInt(document.getElementById("s-font-size").value) || 14,

		FallbackFont: document.getElementById("s-fallback-font").value.trim(),

		FontWeight: parseInt(document.getElementById("s-font-weight").value) || 400,

		FontWeightBold:
			parseInt(document.getElementById("s-font-weight-bold").value) || 700,

		LineHeight:
			parseFloat(document.getElementById("s-line-height").value) || 1.2,

		LinePadding: parseInt(document.getElementById("s-line-padding").value) || 0,

		Ligatures: document.getElementById("s-ligatures").checked,

		Theme: document.getElementById("s-theme").value || "dark",

		Opacity: parseFloat(document.getElementById("s-opacity").value) || 1.0,
		BackgroundColor: document.getElementById("s-bg-color").value || "#000000",

		Spaciness: parseInt(document.getElementById("s-spaciness").value) || 1,

		Animations: document.getElementById("s-animations").checked,

		CursorStyle: document.getElementById("s-cursor-style").value || "bar",

		CursorBlink: document.getElementById("s-cursor-blink").checked,

		Frontend: document.getElementById("s-frontend").value || "xterm-webgl",

		DrawBoldTextInBrightColors:
			document.getElementById("s-draw-bold-bright").checked,

		MinimumContrastRatio:
			parseFloat(document.getElementById("s-min-contrast").value) || 4,

		CSS: document.getElementById("s-css").value,

		Shell: document.getElementById("s-shell").value || "",

		Scrollback:
			parseInt(document.getElementById("s-scrollback").value) || 25000,

		Bell: document.getElementById("s-bell").value || "off",

		AltIsMeta: document.getElementById("s-alt-is-meta").checked,

		ScrollOnInput: document.getElementById("s-scroll-on-input").checked,

		UseConPTY: document.getElementById("s-use-conpty").checked,

		SetComSpec: document.getElementById("s-set-comspec").checked,

		CopyOnSelect: document.getElementById("s-copy-on-select").checked,

		CopyAsHTML: document.getElementById("s-copy-as-html").checked,

		BracketedPaste: document.getElementById("s-bracketed-paste").checked,
		SessionLogging: document.getElementById("s-session-logging").checked,

		WarnOnMultilinePaste: document.getElementById("s-warn-multiline").checked,

		ReplaceNewlinesOnPaste:
			document.getElementById("s-replace-newlines").checked,

		TrimWhitespaceOnPaste: document.getElementById("s-trim-whitespace").checked,

		RightClick: document.getElementById("s-right-click").value || "menu",

		PasteOnMiddleClick: document.getElementById("s-paste-middle-click").checked,

		WordSeparator: document.getElementById("s-word-separator").value,

		TabPosition: document.getElementById("s-tab-position").value || "left",

		LastTabClosesWindow: document.getElementById("s-last-tab-closes").checked,

		CycleTabs: document.getElementById("s-cycle-tabs").checked,

		HideCloseButton: document.getElementById("s-hide-close-button").checked,

		PaneResizeStep:
			parseFloat(document.getElementById("s-pane-resize-step").value) || 0.1,

		FocusFollowsMouse: document.getElementById("s-focus-follows-mouse").checked,

		AutoOpen: document.getElementById("s-auto-open").checked,

		RecoverTabs: document.getElementById("s-recover-tabs").checked,

		Frame: document.getElementById("s-frame").value || "thin",

		Dock: document.getElementById("s-dock").value || "off",

		DockHideOnBlur: document.getElementById("s-dock-hide-blur").checked,

		DockAlwaysOnTop: document.getElementById("s-dock-on-top").checked,

		HideTray: document.getElementById("s-hide-tray").checked,

		Language: document.getElementById("s-language").value.trim(),

		EnableAnalytics: document.getElementById("s-analytics").checked,

		EnableAutomaticUpdates: document.getElementById("s-auto-updates").checked,

		EnableExperimentalFeatures:
			document.getElementById("s-experimental").checked,

		SSHWarnOnClose: document.getElementById("s-ssh-warn-close").checked,

		SSHVerifyHostKeys: document.getElementById("s-ssh-verify-keys").checked,

		SSHDisableDynamicTitle: document.getElementById("s-ssh-disable-title")
			.checked,

		SSHAgentType: document.getElementById("s-ssh-agent-type").value || "auto",

		SSHAgentPath: document.getElementById("s-ssh-agent-path").value.trim(),

		SSHX11Display: document.getElementById("s-ssh-x11").value.trim(),

		SerialBaudRate:
			parseInt(document.getElementById("s-serial-baud").value) || 115200,

		SerialDataBits:
			parseInt(document.getElementById("s-serial-data-bits").value) || 8,

		SerialStopBits:
			parseInt(document.getElementById("s-serial-stop-bits").value) || 1,

		SerialParity: document.getElementById("s-serial-parity").value || "none",

		SerialFlowControl: document.getElementById("s-serial-flow").value || "none",
	};

	try {
		await SaveSettings(s);
		settings = s;
		applySettingsToUI();
		showToast("Settings saved", "success");
	} catch (_) {
		showToast("Failed to save settings", "error");
	}

	hideSettings();
}

GetUsername()
	.then((u) => {
		const el = document.getElementById("ssh-user");
		if (el && !el.value) el.value = u;
	})
	.catch(() => {});

function doResetSettings() {
	ResetSettings()
		.then(() => {
			settings = {};
			fontSize = 14;
			applyFontSize(14);
			applyTheme("dark");
			applySettingsToUI();
			showToast("Reset to defaults", "info");
		})

		.catch(() => {
			showToast("Failed to reset", "error");
		});
}

function toggleSettings() {
	const p = document.getElementById("settings-panel");
	if (p.classList.contains("active")) {
		hideSettings();
		return;
	}
	applySettingsToUI();
	p.classList.add("active");
}

function hideSettings() {
	document.getElementById("settings-panel").classList.remove("active");
	const tab = getActiveTab();
	if (tab) tab.term.focus();
}

// ===== WINDOW CONTROLS =====

let alwaysOnTop = false;

async function toggleAlwaysOnTop() {
	alwaysOnTop = !alwaysOnTop;

	try {
		await WindowSetAlwaysOnTop(alwaysOnTop);

		showToast(alwaysOnTop ? "Always on top: ON" : "Always on top: OFF", "info");
	} catch (e) {
		showToast("Window control not available", "error");
	}
}

// ===== TOAST =====

function showToast(message, type = "info") {
	const existing = document.querySelector(".toast");
	if (existing) existing.remove();
	const toast = document.createElement("div");
	toast.className = `toast ${type}`;
	toast.textContent = message;
	document.body.appendChild(toast);
	requestAnimationFrame(() => toast.classList.add("visible"));
	setTimeout(() => {
		toast.classList.remove("visible");
		setTimeout(() => toast.remove(), 300);
	}, 2500);
}

// ===== NEW TAB DROPDOWN =====

function showNewTabDropdown(e) {
	e.stopPropagation();
	document.querySelectorAll("#new-tab-dropdown").forEach((d) => d.remove());
	const dropdown = document.createElement("div");
	dropdown.id = "new-tab-dropdown";
	let html = "";
	// Connection types
	html += '<div class="dropdown-section-label">New</div>';
	html +=
		'<div class="shell-item" data-action="new-tab"><span class="shell-name">⎘ Local Shell</span><span class="shell-path">' +
		defaultShell +
		"</span></div>";
	availableShells.forEach((s) => {
		const name = s
			.replace(String.fromCharCode(92), "/")
			.split("/")
			.pop()
			.replace(".exe", "");
		html +=
			'<div class="shell-item" data-action="new-tab" data-shell="' +
			s +
			'"><span class="shell-name">' +
			name +
			'</span><span class="shell-path">' +
			s +
			"</span></div>";
	});
	html += '<div class="dropdown-separator"></div>';
	html +=
		'<div class="shell-item" data-action="ssh"><span class="shell-name">🔐 SSH Connect</span><span class="shell-path">Remote server</span></div>';
	html +=
		'<div class="shell-item" data-action="serial"><span class="shell-name">📡 Serial Port</span><span class="shell-path">Hardware device</span></div>';
	html +=
		'<div class="shell-item" data-action="telnet"><span class="shell-name">🌐 Telnet</span><span class="shell-path">Telnet server</span></div>';
	html += '<div class="dropdown-separator"></div>';
	html +=
		'<div class="shell-item" data-action="command-palette"><span class="shell-name">⚙ Command Palette</span><span class="shell-path">Ctrl+Shift+P</span></div>';
	html +=
		'<div class="shell-item" data-action="notifications"><span class="shell-name">🔔 Notifications</span></div>';
	html +=
		'<div class="shell-item" data-action="settings"><span class="shell-name">⚙ Settings</span><span class="shell-path">Ctrl+,</span></div>';
	dropdown.innerHTML = html;
	const rect = e.currentTarget.getBoundingClientRect();
	dropdown.style.position = "fixed";
	dropdown.style.left = rect.left + "px";
	dropdown.style.top = rect.bottom + 4 + "px";
	document.body.appendChild(dropdown);
	dropdown.onclick = (ev) => {
		const item = ev.target.closest(".shell-item");
		if (!item) return;
		dropdown.remove();
		const action = item.dataset.action;
		const shell = item.dataset.shell;
		if (action === "new-tab") newTab(shell || undefined);
		else if (action === "ssh") openSSHDialog();
		else if (action === "serial") openSerialDialog();
		else if (action === "telnet") openTelnetDialog();
		else if (action === "command-palette") toggleCommandPalette();
		else if (action === "notifications") showNotificationCenter();
		else if (action === "settings") toggleSettings();
	};
	setTimeout(() => {
		document.addEventListener("click", function handler(ev) {
			if (!dropdown.contains(ev.target)) {
				dropdown.remove();
				document.removeEventListener("click", handler);
			}
		});
	}, 10);
}

// ===== SCROLL TO BOTTOM =====

function addScrollToBottom(term, container) {
	const btn = document.createElement("button");

	btn.className = "scroll-to-bottom-btn";

	btn.textContent = "\u2193";

	btn.title = "Scroll to bottom";

	btn.style.cssText =
		"position:absolute;bottom:30px;right:20px;width:32px;height:32px;border-radius:50%;border:1px solid var(--border);background:var(--bg-secondary);color:var(--text);cursor:pointer;display:none;z-index:10;font-size:16px;line-height:32px;text-align:center;opacity:0.8;";

	container.style.position = "relative";

	container.appendChild(btn);

	term.onScroll(() => {
		const atBottom =
			term.buffer.active.viewportY >= term.buffer.active.length - term.rows;

		btn.style.display = atBottom ? "none" : "block";
	});

	btn.onclick = () => term.scrollToBottom();

	return btn;
}

// ===== ZMODEM =====

const zmodemActive = false;

function setupZmodem(term, tab) {
	if (!term.zmodemAttach) return;

	try {
		term.zmodemAttach({
			sendTerminal: (data) => {
				if (tab.ptyId && !tab.exited) PTYWrite(tab.ptyId, btoa(data));
				else if (tab.isSSH && tab.sshConnectionId)
					SSHWrite({
						connectionId: tab.sshConnectionId,
						sessionId: tab.sshSessionId,
						data: btoa(data),
					});
			},

			senderAction: (xfer) => {
				const fileInput = document.createElement("input");

				fileInput.type = "file";

				fileInput.onchange = () => {
					const file = fileInput.files[0];

					if (file) {
						const reader = new FileReader();

						reader.onload = () => {
							const bytes = new Uint8Array(reader.result);

							xfer.send(bytes);

							showToast("Sent: " + file.name, "success");
						};

						reader.readAsArrayBuffer(file);
					}
				};

				fileInput.click();
			},

			receiverAction: (xfer) => {
				const offered = xfer.files.map((f) => f.name).join(", ");

				if (confirm("Accept Zmodem file(s): " + offered + "?")) {
					xfer.accept();

					xfer.on("complete", () => {
						showToast("Zmodem transfer complete", "success");
					});
				} else {
					xfer.skip();
				}
			},
		});
	} catch (e) {
		/* zmodem not available */
	}
}

// ===== TAB CLASS =====

class Tab {
	constructor(shell) {
		this.id = `tab-${Date.now()}-${tabCounter++}`;

		this.ptyId = null;
		this.title = "Shell";
		this.userRenamed = false;
		this.shell = shell || defaultShell;
		this.workingDir = "";
		this.exited = false;
		this.status = "disconnected";
		this.connectionType = "local";
		this.connectionLog = [];
		this.logBuffer = [];
		this.loggingEnabled = settings.SessionLogging || false;
		this.lastActivity = Date.now();
		this.isSSH = false;
		this.sshConnectionId = null;
		this.sshSessionId = null;
		this.sshHost = "";
		this.sshPort = 22;
		this.sshUser = "";
		this.isSerial = false;
		this.serialId = null;
		this.serialPort = "";
		this.serialDataHandler = null;
		this.serialExitHandler = null;
		this.isTelnet = false;
		this.telnetConnectionId = null;
		this.telnetHost = "";
		this.telnetPort = 23;
		this.telnetDataHandler = null;
		this.telnetExitHandler = null;

		const fontFamily =
			settings.FontFamily ||
			'"Cascadia Code","Fira Code",Consolas,"Courier New","Segoe UI Emoji",monospace';

		const lineHeight = settings.LineHeight || 1.2;

		const scrollback = settings.Scrollback || 25000;

		const cursorStyle = settings.CursorStyle || "bar";

		const cursorBlink = settings.CursorBlink ?? true;

		const colorScheme = settings.ColorScheme || "Tabby Default";
		const theme = getColorSchemeTheme(colorScheme);

		const fontWeight = settings.FontWeight || 400;

		const fontWeightBold = settings.FontWeightBold || 700;

		this.term = new Terminal({
			cursorBlink,
			cursorStyle,
			fontFamily,
			fontSize,
			fontWeight,
			fontWeightBold,

			lineHeight,
			allowProposedApi: true,
			scrollback,

			bellStyle: settings.Bell || "off",

			theme: theme || FALLBACK_DARK,
		});

		this.fitAddon = new FitAddon();
		this.searchAddon = new SearchAddon();
		this.webLinksAddon = new WebLinksAddon();

		this.term.loadAddon(this.fitAddon);
		this.term.loadAddon(this.searchAddon);
		this.term.loadAddon(this.webLinksAddon);
		this.term.loadAddon(new Unicode11Addon());
		this.term.unicode.activeVersion = "11";

		// WebGL renderer for better Unicode/emoji support
		try {
			this.term.loadAddon(new WebglAddon());
		} catch (e) {
			console.warn("WebGL addon failed, falling back to canvas:", e);
		}

		// Handle terminal resize events from xterm.js
		this.term.onResize(({ cols, rows }) => {
			if (this.ptyId && !this.exited) PTYResize(this.ptyId, cols, rows);
			if (this.isSSH && this.sshConnectionId && this.sshSessionId)
				SSHResize({
					connectionId: this.sshConnectionId,
					sessionId: this.sshSessionId,
					columns: cols,
					rows: rows,
				});
			if (this.isTelnet && this.telnetConnectionId)
				TelnetResize(this.telnetConnectionId, cols, rows);
		});

		this.term.registerLinkProvider({
			provideLinks: (y, callback) => {
				const line = this.term.buffer.active.getLine(y - 1);

				if (!line) {
					callback(undefined);
					return;
				}

				const text = line.translateToString(true);

				const urlRegex = /https?:\/\/[^\s)\]}>]+/g;

				let match;
				const links = [];

				while ((match = urlRegex.exec(text)) !== null) {
					links.push({
						text: match[0],

						range: {
							start: { x: match.index + 1, y },
							end: { x: match.index + match[0].length, y },
						},

						activate: () => OpenInBrowser(match[0]),

						decorations: { underline: true, cursorPointer: true },
					});
				}

				callback(links.length ? links : undefined);
			},
		});

		this.wrapper = document.createElement("div");
		this.wrapper.style.position = "relative";
		this.wrapper.className = "terminal-wrapper";

		this.wrapper.id = this.id;

		document.getElementById("main-content").appendChild(this.wrapper);
		this.term.open(this.wrapper);

		addScrollToBottom(this.term, this.wrapper);

		this.tabEl = document.createElement("div");
		this.tabEl.className = "tab-item";
		this.tabEl.draggable = true;

		this.tabEl.addEventListener("dragstart", (e) => {
			e.dataTransfer.setData("text/plain", this.id);

			this.tabEl.classList.add("dragging");
		});

		this.tabEl.addEventListener("dragend", () => {
			this.tabEl.classList.remove("dragging");
		});
		this.tabEl.dataset.tabId = this.id;

		this.tabEl.innerHTML = `<span class="tab-title">${this.title}</span><button class="tab-close">×</button>`;

		document.getElementById("tab-list").appendChild(this.tabEl);

		this.tabEl.onclick = (e) => {
			if (!e.target.classList.contains("tab-close")) this.activate();
		};

		// Double-click tab title to rename
		this.tabEl.querySelector(".tab-title").ondblclick = (e) => {
			e.stopPropagation();
			const n = prompt("Tab name:", this.title);
			if (n) {
				this.userRenamed = true;
				this.setTitle(n);
			}
		};

		this.tabEl.querySelector(".tab-close").onclick = (e) => {
			e.stopPropagation();
			this.close();
		};

		this.tabEl.oncontextmenu = (e) => {
			e.preventDefault();
			showTabContextMenu(e, this);
		};

		this.term.element.addEventListener("contextmenu", (e) => {
			e.preventDefault();

			const sel = this.term.getSelection();

			const m = document.createElement("div");

			m.className = "context-menu";

			m.style.cssText =
				"position:fixed;left:" +
				e.clientX +
				"px;top:" +
				e.clientY +
				"px;z-index:999;";

			if (sel)
				m.innerHTML +=
					'<div class="context-menu-item" data-action="copy">Copy Selection</div>';

			m.innerHTML +=
				'<div class="context-menu-item" data-action="paste">Paste</div>';

			if (sel)
				m.innerHTML +=
					'<div class="context-menu-item" data-action="search">Search Web</div>';

			m.innerHTML +=
				'<div class="context-menu-item" data-action="clear">Clear Terminal</div>';

			m.innerHTML +=
				'<div class="context-menu-item" data-action="select-all">Select All</div>';

			m.innerHTML +=
				'<div class="context-menu-item" data-action="reset">Reset Terminal</div>';

			document.body.appendChild(m);

			const rm = () => m.remove();

			m.querySelectorAll(".context-menu-item").forEach((it) => {
				it.onclick = () => {
					const a = it.dataset.action;

					if (a === "copy" && sel) ClipboardSetText(sel);

					if (a === "paste")
						navigator.clipboard
							.readText()
							.then((t) => this.term.paste(t))
							.catch(() => {});

					if (a === "search" && sel)
						OpenInBrowser(
							"https://www.google.com/search?q=" + encodeURIComponent(sel),
						);

					if (a === "clear") this.term.clear();

					if (a === "select-all") this.term.selectAll();

					if (a === "reset") this.term.reset();

					rm();
				};
			});

			setTimeout(
				() => document.addEventListener("click", rm, { once: true }),
				10,
			);
		});

		this.term.onTitleChange((title) => {
			if (title && title.trim() && !this.userRenamed) {
				this.setTitle(title.trim());
				SetWindowTitle(title.trim());
			}
		});

		this.lastActivity = Date.now();
		this.term.onData((data) => {
			if (this.ptyId && !this.exited) {
				PTYWrite(this.ptyId, btoa(data));
				broadcastInput(data, this);
			}
		});

		setupInputProcessing(this.term, this);

		this.dataHandler = (params) => {
			const pid = params.ptyId ?? params.PTYID;
			const sid = params.sessionId ?? params.SessionID;
			const serid = params.serialId ?? params.SerialID;
			const cid = params.connectionId ?? params.ConnectionID;
			if (pid && pid === this.ptyId) {
				const bytes = b64ToBytes(params.data);
				this.term.write(bytes);
				if (this.loggingEnabled)
					this.logBuffer.push(new TextDecoder().decode(bytes));
				this.lastActivity = Date.now();
			} else if (this.isSSH && sid && sid === this.sshSessionId) {
				const bytes = b64ToBytes(params.data);
				this.term.write(bytes);
				if (this.loggingEnabled)
					this.logBuffer.push(new TextDecoder().decode(bytes));
				this.lastActivity = Date.now();
			} else if (this.isSerial && serid && serid === this.serialId) {
				const bytes = b64ToBytes(params.data);
				this.term.write(bytes);
				if (this.loggingEnabled)
					this.logBuffer.push(new TextDecoder().decode(bytes));
				this.lastActivity = Date.now();
			} else if (this.isTelnet && cid && cid === this.telnetConnectionId) {
				const bytes = b64ToBytes(params.data);
				this.term.write(bytes);
				if (this.loggingEnabled)
					this.logBuffer.push(new TextDecoder().decode(bytes));
				this.lastActivity = Date.now();
			}
		};

		this.exitHandler = (params) => {
			const pid = params.ptyId ?? params.PTYID;
			const sid = params.sessionId ?? params.SessionID;
			const cid = params.connectionId ?? params.ConnectionID;
			let matched = false;
			if (pid && pid === this.ptyId) matched = true;
			else if (this.isSSH && sid && sid === this.sshSessionId) matched = true;
			else if (this.isSSH && cid && cid === this.sshConnectionId)
				matched = true;
			if (matched) {
				this.exited = true;
				setTabStatus(this, "disconnected");
				const code = params.exitCode ?? 0;
				this.term.writeln(
					`\r\n\x1b[1;33m[Process exited — code ${code}]\x1b[0m`,
				);
				this.setTitle(`Exit (${code})`);
			}
		};

		window.__ptyDataHandlers = window.__ptyDataHandlers || [];
		window.__ptyExitHandlers = window.__ptyExitHandlers || [];

		window.__ptyDataHandlers = window.__ptyDataHandlers || [];
		window.__ptyDataHandlers.push(this.dataHandler);
		window.__ptyExitHandlers = window.__ptyExitHandlers || [];
		window.__ptyExitHandlers.push(this.exitHandler);
	}

	async spawn() {
		try {
			setTabStatus(this, "connecting");
			logConnection(this, "Spawning shell: " + this.shell);
			// Fit first to get correct cols/rows
			this.fitAddon.fit();
			console.log("[spawn] PTYSpawn:", {
				command: this.shell,
				cols: this.term.cols,
				rows: this.term.rows,
			});
			const result = await PTYSpawn({
				command: this.shell,
				args: [],
				env: {},
				cwd: this.workingDir || undefined,
				columns: this.term.cols,
				rows: this.term.rows,
			});
			console.log("[spawn] PTYSpawn result:", result);
			this.ptyId = result.id;
			setTabStatus(this, "connected");
			logConnection(this, "Shell started: " + this.shell);
			const name = this.shell.split(/[/\\]/).pop().replace(".exe", "");
			this.setTitle(name);
			showStatus(`Connected — ${name}`);
			// Delayed fit to ensure correct dimensions after DOM layout
			setTimeout(() => {
				this.fitAddon.fit();
				if (this.ptyId && !this.exited)
					PTYResize(this.ptyId, this.term.cols, this.term.rows);
			}, 100);
		} catch (err) {
			this.term.writeln(`\x1b[1;31mFailed to spawn shell: ${err}\x1b[0m`);
			showToast(`Shell spawn failed: ${err}`, "error");
		}
	}

	activate() {
		const w = document.getElementById("welcome");
		if (w) w.style.display = "none";

		tabs.forEach((t) => {
			t.wrapper.classList.remove("active");
			t.tabEl.classList.remove("active");
		});

		this.wrapper.classList.add("active");
		this.tabEl.classList.add("active");
		activeTabId = this.id;
		updateToolbar(this);

		this.term.focus();

		// Fit synchronously so cols/rows are correct for spawn()
		this.fitAddon.fit();
		if (this.ptyId && !this.exited)
			PTYResize(this.ptyId, this.term.cols, this.term.rows);

		if (this.isTelnet && this.telnetConnectionId)
			TelnetResize(this.telnetConnectionId, this.term.cols, this.term.rows);

		saveSession();
	}

	close() {
		window.__ptyDataHandlers = (window.__ptyDataHandlers || []).filter(
			(h) => h !== this.dataHandler,
		);

		window.__ptyExitHandlers = (window.__ptyExitHandlers || []).filter(
			(h) => h !== this.exitHandler,
		);

		if (this.ptyId) PTYKill(this.ptyId, "").catch(() => {});
		if (this.isSSH && this.sshConnectionId)
			SSHClose({
				connectionId: this.sshConnectionId,
				sessionId: this.sshSessionId,
			}).catch(() => {});

		if (this.isSerial && this.serialId) {
			SerialClose(this.serialId).catch(() => {});
			window.__serialDataHandlers = (window.__serialDataHandlers || []).filter(
				(h) => h !== this.serialDataHandler,
			);

			window.__serialExitHandlers = (window.__serialExitHandlers || []).filter(
				(h) => h !== this.serialExitHandler,
			);
		}

		if (this.isTelnet && this.telnetConnectionId) {
			TelnetClose(this.telnetConnectionId).catch(() => {});
			window.__telnetDataHandlers = (window.__telnetDataHandlers || []).filter(
				(h) => h !== this.telnetDataHandler,
			);

			window.__telnetExitHandlers = (window.__telnetExitHandlers || []).filter(
				(h) => h !== this.telnetExitHandler,
			);
		}

		this.term.dispose();
		this.wrapper.remove();
		this.tabEl.remove();

		const idx = tabs.indexOf(this);
		if (idx > -1) tabs.splice(idx, 1);

		if (activeTabId === this.id) {
			if (tabs.length > 0) tabs[Math.min(idx, tabs.length - 1)].activate();
			else {
				activeTabId = null;
				const w = document.getElementById("welcome");
				if (w) w.style.display = "flex";
			}
		}

		saveSession();
	}

	setTitle(title) {
		this.title = title;
		const el = this.tabEl.querySelector(".tab-title");
		if (el) el.textContent = title;
		if (activeTabId === this.id) SetWindowTitle(`Tabby — ${title}`);
	}

	findNext(q) {
		if (q) this.searchAddon.findNext(q);
	}

	findPrevious(q) {
		if (q) this.searchAddon.findPrevious(q);
	}

	copySelection() {
		const sel = this.term.getSelection();
		if (sel)
			navigator.clipboard
				.writeText(sel)
				.then(() => showToast("Copied", "success"));
	}

	async pasteFromClipboard() {
		try {
			const text = await ClipboardGetText();
			if (text) {
				if (this.isSSH && this.sshConnectionId && this.sshSessionId)
					SSHWrite({
						connectionId: this.sshConnectionId,
						sessionId: this.sshSessionId,
						data: btoa(text),
					});
				else if (this.isSerial && this.serialId)
					SerialWrite(this.serialId, btoa(text));
				else if (this.isTelnet && this.telnetConnectionId)
					TelnetWrite(this.telnetConnectionId, btoa(text));
				else if (this.ptyId && !this.exited) PTYWrite(this.ptyId, btoa(text));
			}
		} catch (_) {
			showToast("Clipboard access denied", "error");
		}
	}
}

// ===== PTY EVENTS =====

EventsOn("pty.data", (params) => {
	(window.__ptyDataHandlers || []).forEach((h) => h(params));
});

EventsOn("pty.exit", (params) => {
	(window.__ptyExitHandlers || []).forEach((h) => h(params));
});

EventsOn("ssh.data", (params) => {
	(window.__ptyDataHandlers || []).forEach((h) => h(params));
});

EventsOn("ssh.exit", (params) => {
	(window.__ptyExitHandlers || []).forEach((h) => h(params));
});

EventsOn("serial.data", (params) => {
	(window.__serialDataHandlers || []).forEach((h) => h(params));
});

EventsOn("serial.exit", (params) => {
	(window.__serialExitHandlers || []).forEach((h) => h(params));
});

EventsOn("telnet.data", (params) => {
	(window.__telnetDataHandlers || []).forEach((h) => h(params));
});

EventsOn("telnet.exit", (params) => {
	(window.__telnetExitHandlers || []).forEach((h) => h(params));
});

EventsOn("ssh.keyboardInteractive", (params) => {
	handleKeyboardInteractive(params);
});

EventsOn("ssh.hostKeyPrompt", (params) => {
	handleHostKeyPrompt(params);
});

EventsOn("ssh.banner", (params) => {
	handleSSHBanner(params);
});

EventsOn("ssh.serviceMessage", (params) => {
	handleSSHServiceMessage(params);
});

EventsOn("telnet.serviceMessage", (params) => {
	handleTelnetServiceMessage(params);
});

EventsOn("menu:new-tab", () => newTab());

EventsOn("menu:settings", () => openSettingsPanel());

EventsOn("menu:copy", () => {
	const t = getActiveTab();
	if (t && t.term) document.execCommand("copy");
});

EventsOn("menu:paste", () => {
	const t = getActiveTab();
	if (t && t.term) ClipboardGetText().then((text) => t.term.paste(text));
});

EventsOn("menu:select-all", () => {
	const t = getActiveTab();
	if (t && t.term) t.term.selectAll();
});

EventsOn("menu:command-palette", () => toggleCommandPalette());

EventsOn("menu:about", () => showAboutDialog());

// ===== TAB MANAGEMENT =====

function newTab(shell) {
	const effectiveShell = shell || defaultShell;
	console.log(
		"[newTab] Creating tab with shell:",
		effectiveShell,
		"defaultShell:",
		defaultShell,
	);
	const tab = new Tab(effectiveShell);
	tabs.push(tab);
	tab.activate();
	tab.spawn();
	console.log("[newTab] Tab created:", tab.id, "tabs count:", tabs.length);
	return tab;
}

function getActiveTab() {
	return tabs.find((t) => t.id === activeTabId);
}

function switchToTab(i) {
	if (i >= 0 && i < tabs.length) tabs[i].activate();
}

function nextTab() {
	const i = tabs.findIndex((t) => t.id === activeTabId);
	tabs[(i + 1) % tabs.length].activate();
}

function prevTab() {
	const i = tabs.findIndex((t) => t.id === activeTabId);
	tabs[(i - 1 + tabs.length) % tabs.length].activate();
}

function reorderTab(draggedId, targetId) {
	const fromIdx = tabs.findIndex((t) => t.id === draggedId);
	const toIdx = tabs.findIndex((t) => t.id === targetId);

	if (fromIdx === -1 || toIdx === -1 || fromIdx === toIdx) return;

	const [moved] = tabs.splice(fromIdx, 1);
	tabs.splice(toIdx, 0, moved);

	const list = document.getElementById("tab-list");
	list.innerHTML = "";
	tabs.forEach((t) => list.appendChild(t.tabEl));

	saveSession();
}

// ===== TERMINAL RESIZE =====

function resizeAllTerminals() {
	tabs.forEach((tab) => {
		if (tab.term && !tab.exited) {
			setTimeout(() => tab.term.fitAddon.fit(), 50);
		}
	});
}

// ===== SPLIT PANE =====

class SplitPane {
	constructor(orientation) {
		this.orientation = orientation;
		this.panes = [];
		this.element = document.createElement("div");
		this.element.className = `split-container ${orientation}`;
		this.dividers = [];
	}

	addPane(tab, flex) {
		const pe = document.createElement("div");
		pe.className = "pane";
		pe.style.flex = `${flex} 1 0%`;
		pe.appendChild(tab.wrapper);
		tab.wrapper.classList.add("active");
		this.panes.push({ tab, element: pe, flex });
		this.element.appendChild(pe);
		if (this.panes.length > 1) this._addDivider();
		return pe;
	}

	_addDivider() {
		const d = document.createElement("div");
		d.className = "splitter";
		if (this.orientation === "vertical") d.classList.add("horizontal");
		this.element.appendChild(d);
		this.dividers.push(d);

		let dragging = false,
			startPos = 0,
			sf0 = 50,
			sf1 = 50;

		const onStart = (e) => {
			if (this.panes.length !== 2) return;
			e.preventDefault();
			dragging = true;
			startPos = this.orientation === "vertical" ? e.clientY : e.clientX;
			sf0 = this.panes[0].flex;
			sf1 = this.panes[1].flex;
			d.classList.add("dragging");
		};

		const onMove = (e) => {
			if (!dragging) return;
			e.preventDefault();
			const pos = this.orientation === "vertical" ? e.clientY : e.clientX;
			const rect = this.element.getBoundingClientRect();
			const total = this.orientation === "vertical" ? rect.height : rect.width;
			const pct = ((pos - startPos) / total) * 100;
			const f0 = Math.max(20, Math.min(80, sf0 + pct));
			this.panes[0].flex = f0;
			this.panes[1].flex = 100 - f0;
			this.panes[0].element.style.flex = `${f0} 1 0%`;
			this.panes[1].element.style.flex = `${100 - f0} 1 0%`;
		};

		const onEnd = () => {
			if (!dragging) return;
			dragging = false;
			d.classList.remove("dragging");
			this.panes.forEach((p) => {
				p.tab.fitAddon.fit();
				if (p.tab.ptyId)
					PTYResize(p.tab.ptyId, p.tab.term.cols, p.tab.term.rows);
				if (p.tab.isSSH && p.tab.sshConnectionId && p.tab.sshSessionId)
					SSHResize({
						connectionId: p.tab.sshConnectionId,
						sessionId: p.tab.sshSessionId,
						columns: p.tab.term.cols,
						rows: p.tab.term.rows,
					});
			});
		};

		d.addEventListener("mousedown", onStart);
		document.addEventListener("mousemove", onMove);
		document.addEventListener("mouseup", onEnd);
	}

	removePane(tab) {
		const idx = this.panes.findIndex((p) => p.tab === tab);
		if (idx === -1) return null;
		this.panes[idx].element.remove();
		this.panes.splice(idx, 1);
		if (this.dividers.length > 0) this.dividers.pop()?.remove();
		if (this.panes.length === 1) {
			this.element.remove();
			return this.panes[0].tab;
		}
		return null;
	}
}

function splitPane(orientation) {
	if (orientation === "horizontal") splitHorizontally();
	else splitVertically();
}

function splitVertically() {
	const tab = getActiveTab();
	if (!tab) return;
	const mc = document.getElementById("main-content");
	const w = document.getElementById("welcome");
	if (w) w.style.display = "none";
	const idx = tabs.indexOf(tab);
	if (idx > -1) tabs.splice(idx, 1);
	tab.wrapper.remove();
	tab.tabEl.remove();
	const sp = new SplitPane("vertical");
	sp.addPane(tab, 50);
	const t2 = new Tab(tab.shell);
	tabs.push(t2);
	sp.addPane(t2, 50);
	mc.appendChild(sp.element);
	activeSplitPane = sp;
	t2.spawn();
	tab.activate();
	showToast("Split vertically", "info");
}

function splitHorizontally() {
	const tab = getActiveTab();
	if (!tab) return;
	const mc = document.getElementById("main-content");
	const w = document.getElementById("welcome");
	if (w) w.style.display = "none";
	const idx = tabs.indexOf(tab);
	if (idx > -1) tabs.splice(idx, 1);
	tab.wrapper.remove();
	tab.tabEl.remove();
	const sp = new SplitPane("horizontal");
	sp.addPane(tab, 50);
	const t2 = new Tab(tab.shell);
	tabs.push(t2);
	sp.addPane(t2, 50);
	mc.appendChild(sp.element);
	activeSplitPane = sp;
	t2.spawn();
	tab.activate();
	showToast("Split horizontally", "info");
}

function removeSplit() {
	if (!activeSplitPane) return;
	const tab = getActiveTab();
	if (!tab) return;
	const remaining = activeSplitPane.removePane(tab);
	if (tab.ptyId) PTYKill(tab.ptyId, "").catch(() => {});
	tab.term.dispose();
	const idx = tabs.indexOf(tab);
	if (idx > -1) tabs.splice(idx, 1);
	const mc = document.getElementById("main-content");
	if (remaining) {
		mc.appendChild(remaining.wrapper);
		document.getElementById("tab-list").appendChild(remaining.tabEl);
		tabs.push(remaining);
		activeSplitPane = null;
		remaining.activate();
	} else if (tabs.length === 0) {
		activeSplitPane = null;
		const w = document.getElementById("welcome");
		if (w) w.style.display = "flex";
	}
}

function closeAllSplits() {
	if (!activeSplitPane) return;
	[...activeSplitPane.panes].forEach((p) => {
		if (p.tab.ptyId) PTYKill(p.tab.ptyId, "").catch(() => {});
		p.tab.term.dispose();
		const idx = tabs.indexOf(p.tab);
		if (idx > -1) tabs.splice(idx, 1);
	});
	activeSplitPane.element.remove();
	activeSplitPane = null;
	if (tabs.length === 0) {
		const w = document.getElementById("welcome");
		if (w) w.style.display = "flex";
	} else tabs[0].activate();
	showToast("All splits closed", "info");
}

// ===== TAB CONTEXT MENU =====

function showTabContextMenu(e, tab) {
	document.querySelectorAll(".context-menu").forEach((m) => m.remove());

	const menu = document.createElement("div");
	menu.className = "context-menu";

	menu.innerHTML =
		'<div class="context-menu-item" data-action="rename">Rename</div>' +
		'<div class="context-menu-item" data-action="duplicate">Duplicate</div>' +
		'<div class="context-menu-separator"></div>' +
		'<div class="context-menu-item" data-action="reconnect">Reconnect</div>' +
		'<div class="context-menu-separator"></div>' +
		'<div class="context-menu-item" data-action="sftp" style="display:none;">SFTP</div>' +
		'<div class="context-menu-item" data-action="forward" style="display:none;">Port Forward</div>' +
		'<div class="context-menu-separator"></div>' +
		'<div class="context-menu-item" data-action="color-red">Red</div>' +
		'<div class="context-menu-item" data-action="color-green">Green</div>' +
		'<div class="context-menu-item" data-action="color-blue">Blue</div>' +
		'<div class="context-menu-item" data-action="color-yellow">Yellow</div>' +
		'<div class="context-menu-item" data-action="color-reset">Reset Color</div>' +
		'<div class="context-menu-separator"></div>' +
		'<div class="context-menu-item" data-action="close-others">Close Others</div>' +
		'<div class="context-menu-item" data-action="close-right">Close to Right</div>' +
		'<div class="context-menu-item" data-action="close-all">Close All</div>' +
		'<div class="context-menu-item" data-action="close">Close</div>';

	document.body.appendChild(menu);

	menu.style.left = Math.min(e.clientX, window.innerWidth - 180) + "px";
	menu.style.top = Math.min(e.clientY, window.innerHeight - 200) + "px";

	const close = () => menu.remove();

	menu.onclick = (ev) => {
		const item = ev.target.closest(".context-menu-item");
		if (!item) return;
		close();

		switch (item.dataset.action) {
			case "rename": {
				const n = prompt("Tab name:", tab.title);
				if (n) {
					tab.userRenamed = true;
					tab.setTitle(n);
				}
				break;
			}

			case "duplicate":
				newTab(tab.shell);
				break;

			case "close-others": {
				[...tabs].forEach((t) => {
					if (t !== tab) t.close();
				});
				tab.activate();
				break;
			}

			case "close-right": {
				const idx = tabs.indexOf(tab);
				for (let i = tabs.length - 1; i > idx; i--) tabs[i].close();
				break;
			}

			case "forward":
				if (tab.isSSH && tab.sshConnectionId)
					openForwardDialog(tab.sshConnectionId);
				break;

			case "sftp":
				if (tab.isSSH && tab.sshConnectionId)
					openSFTPBrowser(tab.sshConnectionId);
				break;

			case "reconnect":
				reconnectTab(tab);
				break;

			case "close-all":
				closeAllTabs();
				break;

			case "color-red":
				setTabColor(tab, "#f44747");
				break;

			case "color-green":
				setTabColor(tab, "#4caf50");
				break;

			case "color-blue":
				setTabColor(tab, "#4ca8e8");
				break;

			case "color-yellow":
				setTabColor(tab, "#e8a84c");
				break;

			case "color-reset":
				setTabColor(tab, "");
				break;

			case "close":
				tab.close();
				break;
		}
	};

	setTimeout(
		() => document.addEventListener("click", close, { once: true }),
		10,
	);
}

// ===== TAB BADGES =====

function setTabBadge(tab, count) {
	if (!tab || !tab.tabEl) return;

	let badge = tab.tabEl.querySelector(".tab-badge");

	if (count > 0) {
		if (!badge) {
			badge = document.createElement("span");

			badge.className = "tab-badge";

			badge.style.cssText =
				"position:absolute;top:-4px;right:16px;min-width:16px;height:16px;border-radius:8px;background:#f44747;color:#fff;font-size:10px;line-height:16px;text-align:center;padding:0 4px;z-index:5;";

			tab.tabEl.style.position = "relative";

			tab.tabEl.appendChild(badge);
		}

		badge.textContent = count > 99 ? "99+" : count;
	} else if (badge) {
		badge.remove();
	}
}

// ===== FIND BAR =====

function toggleFind() {
	let bar = document.getElementById("find-bar");
	if (bar) {
		bar.remove();
		findVisible = false;
		const t = getActiveTab();
		if (t) t.term.focus();
		return;
	}

	findVisible = true;
	bar = document.createElement("div");
	bar.id = "find-bar";

	bar.style.cssText =
		"position:absolute;top:0;right:0;z-index:100;background:#2d2d2d;border-bottom:1px solid #3a3a3a;border-left:1px solid #3a3a3a;padding:6px 12px;display:flex;gap:8px;align-items:center;border-radius:0 0 0 8px;";

	bar.innerHTML = `<input type="text" id="find-input" placeholder="Find..." style="background:#1e1e1e;border:1px solid #3a3a3a;color:#ccc;padding:4px 8px;border-radius:4px;font-size:13px;width:200px;outline:none;"><button class="btn-icon" id="find-prev">↑</button><button class="btn-icon" id="find-next">↓</button><button class="btn-icon" id="find-close">×</button>`;

	document.getElementById("main-content").appendChild(bar);

	const input = document.getElementById("find-input");
	input.focus();

	input.onkeydown = (e) => {
		const t = getActiveTab();
		if (!t) return;
		if (e.key === "Enter") {
			e.preventDefault();
			e.shiftKey ? t.findPrevious(input.value) : t.findNext(input.value);
		}
		if (e.key === "Escape") toggleFind();
	};

	document.getElementById("find-next").onclick = () => {
		const t = getActiveTab();
		if (t) t.findNext(input.value);
		input.focus();
	};

	document.getElementById("find-prev").onclick = () => {
		const t = getActiveTab();
		if (t) t.findPrevious(input.value);
		input.focus();
	};

	document.getElementById("find-close").onclick = () => toggleFind();
}

// ===== RECONNECT =====

async function reconnectTab(tab) {
	tab.exited = false;

	if (tab.sessionData) {
		try {
			const session = JSON.parse(tab.sessionData);

			if (session.type === "ssh") {
				showToast("Reconnecting to " + session.host + "...", "info");

				setTabStatus(tab, "connecting");

				const result = await SSHConnect({
					host: session.host,

					port: session.port || 22,

					user: session.user,

					auth: session.auth,

					keepaliveInterval: 30,

					keepaliveCountMax: 3,

					readyTimeout: 15000,
				});

				tab.sshConnectionId = result.connectionId;

				tab.isSSH = true;

				tab.sshHost = session.host;

				tab.sshPort = session.port || 22;

				tab.sshUser = session.user;

				const shellResult = await SSHStartShell({
					connectionId: result.connectionId,

					columns: tab.term.cols,

					rows: tab.term.rows,

					terminal: "xterm-256color",
				});

				tab.sshSessionId = shellResult.sessionId;

				setTabStatus(tab, "connected");

				let jumpLabel = "";

				if (result.jumpChain && result.jumpChain.length > 0)
					jumpLabel = " (via " + result.jumpChain.join(" -> ") + ")";

				tab.setTitle(session.user + "@" + session.host + jumpLabel);

				tab.term.onData((data) => {
					if (tab.sshConnectionId && tab.sshSessionId) {
						SSHWrite({
							connectionId: tab.sshConnectionId,
							sessionId: tab.sshSessionId,
							data: btoa(data),
						});
						broadcastInput(data, tab);
					}
				});

				setupInputProcessing(tab.term, tab);

				showStatus("SSH - " + session.user + "@" + session.host);

				showToast("Reconnected to " + session.host, "success");
			} else if (session.type === "telnet") {
				showToast("Reconnecting to " + session.host + "...", "info");

				setTabStatus(tab, "connecting");

				const result = await TelnetConnect(session.host, session.port || 23);

				tab.telnetConnectionId = result.ConnectionID || result.connectionId;

				tab.isTelnet = true;

				tab.telnetHost = session.host;

				tab.telnetPort = session.port || 23;

				setTabStatus(tab, "connected");

				tab.setTitle(session.host + ":" + (session.port || 23));

				tab.telnetDataHandler = (params) => {
					const cid = params.ConnectionID || params.connectionId;

					if (cid === tab.telnetConnectionId)
						tab.term.write(b64ToBytes(params.Data || params.data));
				};

				window.__telnetDataHandlers = window.__telnetDataHandlers || [];

				window.__telnetDataHandlers.push(tab.telnetDataHandler);

				tab.telnetExitHandler = (params) => {
					const cid = params.ConnectionID || params.connectionId;

					if (cid === tab.telnetConnectionId) {
						tab.exited = true;

						setTabStatus(tab, "disconnected");

						tab.term.writeln("\x1b[1;33m[Telnet connection closed]\x1b[0m");

						tab.setTitle(tab.title + " [disconnected]");
					}
				};

				window.__telnetExitHandlers = window.__telnetExitHandlers || [];

				window.__telnetExitHandlers.push(tab.telnetExitHandler);

				tab.term.onData((data) => {
					if (tab.telnetConnectionId) {
						TelnetWrite(tab.telnetConnectionId, btoa(data));
						broadcastInput(data, tab);
					}
				});

				setupInputProcessing(tab.term, tab);

				showStatus("Telnet - " + session.host + ":" + (session.port || 23));

				showToast("Reconnected to " + session.host, "success");
			} else if (session.type === "serial") {
				showToast("Reconnecting to " + session.port + "...", "info");

				setTabStatus(tab, "connecting");

				const result = await SerialOpen({
					port: session.port,
					baudRate: session.baudRate || 115200,
					dataBits: 8,
					stopBits: 1,
					parity: "none",
				});

				tab.serialId = result.ID || result.id;

				tab.isSerial = true;

				tab.serialPort = session.port;

				tab.serialBaud = session.baudRate || 115200;

				setTabStatus(tab, "connected");

				tab.setTitle(
					session.port.split("/").pop().split("\\").pop() +
						" @ " +
						(session.baudRate || 115200) +
						" baud",
				);

				tab.serialDataHandler = (params) => {
					if ((params.serialId || params.SerialID) === tab.serialId)
						tab.term.write(b64ToBytes(params.data || params.Data));
				};

				window.__serialDataHandlers = window.__serialDataHandlers || [];

				window.__serialDataHandlers.push(tab.serialDataHandler);

				tab.serialExitHandler = (params) => {
					if ((params.serialId || params.SerialID) === tab.serialId) {
						tab.exited = true;

						setTabStatus(tab, "disconnected");

						tab.term.writeln("\x1b[1;33m[Serial port closed]\x1b[0m");

						tab.setTitle(tab.title + " [disconnected]");
					}
				};

				window.__serialExitHandlers = window.__serialExitHandlers || [];

				window.__serialExitHandlers.push(tab.serialExitHandler);

				tab.term.onData((data) => {
					if (tab.serialId) {
						SerialWrite(tab.serialId, btoa(data));
						broadcastInput(data, tab);
					}
				});

				setupInputProcessing(tab.term, tab);

				showStatus(
					"Serial - " + session.port + " @ " + (session.baudRate || 115200),
				);

				showToast("Reconnected to " + session.port, "success");
			} else {
				tab.spawn();

				showToast("Reconnecting local shell...", "info");
			}
		} catch (e) {
			showToast("Reconnect failed: " + e.message, "error");

			setTabStatus(tab, "disconnected");

			tab.exited = true;
		}
	} else {
		tab.spawn();
	}

	saveSession();
}

// ===== SESSION PERSISTENCE =====

function saveSession() {
	try {
		const tabStates = tabs.map((t) => ({
			Shell: t.shell,

			Title: t.title,

			Active: t.id === activeTabId,

			Type: t.isSSH
				? "ssh"
				: t.isSerial
					? "serial"
					: t.isTelnet
						? "telnet"
						: "local",

			Host: t.isSSH
				? t.sshHost
				: t.isTelnet
					? t.telnetHost
					: t.isSerial
						? t.serialPort
						: "",

			Port: t.isSSH ? t.sshPort : t.isTelnet ? t.telnetPort : 0,

			User: t.isSSH ? t.sshUser : "",

			BaudRate: t.isSerial ? t.serialBaud : 0,

			SerialPort: t.isSerial ? t.serialPort : "",

			WorkingDir: t.isSSH || t.isSerial || t.isTelnet ? "" : t.workingDir,

			Exited: t.exited,
		}));

		SaveSessionState(tabStates).catch(() => {});
	} catch (_) {}
}

async function restoreSession() {
	try {
		const state = await LoadSessionState();

		if (!state || !state.Tabs || state.Tabs.length === 0) return false;

		let activated = false;

		for (const saved of state.Tabs) {
			const tabType = saved.Type || "local";

			if (tabType === "ssh" && saved.Host && !saved.Exited) {
				const tab = new Tab(saved.Shell || defaultShell);

				tabs.push(tab);
				tab.activate();

				tab.setTitle(
					(saved.User || "root") + "@" + saved.Host + " [reconnecting...]",
				);

				tab.term.writeln(
					"\x1b[1;33m[Reconnecting to " +
						(saved.User || "root") +
						"@" +
						saved.Host +
						"...]\x1b[0m",
				);

				try {
					setTabStatus(tab, "connecting");
					logConnection(
						tab,
						"SSH connecting to " + saved.Host + ":" + (saved.Port || 22),
					);
					const result = await SSHConnect({
						host: saved.Host,
						port: saved.Port || 22,
						user: saved.User || "root",
						auth: { type: "agent" },
						keepaliveInterval: 30,
						keepaliveCountMax: 3,
						readyTimeout: 15000,
					});
					setTabStatus(tab, "connected");
					logConnection(tab, "SSH connected");

					tab.sshConnectionId = result.connectionId;

					tab.sshHost = saved.Host;
					tab.sshPort = saved.Port || 22;
					tab.sshUser = saved.User || "root";
					tab.isSSH = true;

					const shellResult = await SSHStartShell({
						connectionId: result.connectionId,
						columns: tab.term.cols,
						rows: tab.term.rows,
						terminal: "xterm-256color",
					});

					tab.sshSessionId = shellResult.sessionId;

					let jumpLabel = "";

					if (result.jumpChain && result.jumpChain.length > 0)
						jumpLabel = " (via " + result.jumpChain.join(" -> ") + ")";

					tab.setTitle((saved.User || "root") + "@" + saved.Host + jumpLabel);

					tab.term.writeln("\x1b[1;32m[Reconnected]\x1b[0m");

					tab.term.onData((data) => {
						if (tab.sshConnectionId && tab.sshSessionId) {
							SSHWrite({
								connectionId: tab.sshConnectionId,
								sessionId: tab.sshSessionId,
								data: btoa(data),
							});
							broadcastInput(data, tab);
						}
					});

					showStatus("SSH - " + (saved.User || "root") + "@" + saved.Host);

					showToast("Reconnected to " + saved.Host, "success");
				} catch (err) {
					tab.term.writeln("\x1b[1;31m[Reconnect failed: " + err + "]\x1b[0m");

					tab.setTitle(
						(saved.User || "root") + "@" + saved.Host + " [disconnected]",
					);

					tab.exited = true;
				}

				if (saved.Active) activated = true;
			} else if (tabType === "telnet" && saved.Host && !saved.Exited) {
				const tab = new Tab(saved.Shell || defaultShell);

				tabs.push(tab);
				tab.activate();

				tab.setTitle(
					saved.Host + ":" + (saved.Port || 23) + " [reconnecting...]",
				);

				try {
					const result = await TelnetConnect(saved.Host, saved.Port || 23);

					tab.telnetConnectionId = result.ConnectionID || result.connectionId;

					tab.telnetHost = saved.Host;
					tab.telnetPort = saved.Port || 23;
					tab.isTelnet = true;

					tab.setTitle(saved.Host + ":" + (saved.Port || 23));

					tab.telnetDataHandler = (params) => {
						const cid = params.ConnectionID || params.connectionId;
						if (cid === tab.telnetConnectionId)
							tab.term.write(b64ToBytes(params.Data || params.data));
					};

					window.__telnetDataHandlers.push(tab.telnetDataHandler);

					tab.term.onData((data) => {
						if (tab.telnetConnectionId) {
							TelnetWrite(tab.telnetConnectionId, btoa(data));
							broadcastInput(data, tab);
						}
					});

					showStatus("Telnet - " + saved.Host + ":" + (saved.Port || 23));

					showToast("Reconnected to " + saved.Host, "success");
				} catch (err) {
					tab.term.writeln("\x1b[1;31m[Reconnect failed: " + err + "]\x1b[0m");

					tab.exited = true;
				}

				if (saved.Active) activated = true;
			} else if (tabType === "serial" && saved.Host && !saved.Exited) {
				const tab = new Tab(saved.Shell || defaultShell);

				tabs.push(tab);

				tab.activate();

				tab.setTitle(saved.Host + " [reconnecting...]");

				try {
					const result = await SerialOpen({
						port: saved.Host,
						baudRate: saved.BaudRate || 115200,
						dataBits: 8,
						stopBits: 1,
						parity: "none",
					});

					tab.serialId = result.id || result.ID;

					tab.serialPort = saved.Host;

					tab.serialBaud = saved.BaudRate || 115200;

					tab.isSerial = true;

					tab.setTitle(
						saved.Host + " @ " + (saved.BaudRate || 115200) + " baud",
					);

					tab.term.writeln("[1;32m[Reconnected][0m");

					tab.term.onData((data) => {
						if (tab.serialId) {
							SerialWrite(tab.serialId, btoa(data));
							broadcastInput(data, tab);
						}
					});

					showStatus("Serial - " + saved.Host);

					showToast("Reconnected to " + saved.Host, "success");
				} catch (err) {
					tab.term.writeln("[1;31m[Reconnect failed: " + err + "][0m");

					tab.exited = true;
				}

				if (saved.Active) activated = true;
			} else {
				const tab = new Tab(saved.Shell);

				tab.workingDir = saved.WorkingDir || "";

				tabs.push(tab);
				tab.spawn();

				if (saved.Active) {
					tab.activate();
					activated = true;
				} else if (saved.Title) tab.setTitle(saved.Title);
			}
		}

		if (!activated && tabs.length > 0) tabs[0].activate();

		return true;
	} catch (_) {
		return false;
	}
}

function showStatus(msg) {
	const el = document.getElementById("status-text");
	if (el) {
		el.textContent = msg;
		clearTimeout(window.__statusTimeout);
		window.__statusTimeout = setTimeout(() => {
			el.textContent = `${tabs.length} tab${tabs.length !== 1 ? "s" : ""}`;
		}, 3000);
	}
}

// ===== KEYBOARD SHORTCUTS =====

function bindGlobalKeys() {
	document.addEventListener("keydown", (e) => {
		const ctrl = e.ctrlKey || e.metaKey;
		const shift = e.shiftKey;

		const inInput =
			document.activeElement &&
			(document.activeElement.tagName === "INPUT" ||
				document.activeElement.tagName === "TEXTAREA" ||
				document.activeElement.tagName === "SELECT");

		if (ctrl && shift && e.key === "P") {
			e.preventDefault();
			toggleCommandPalette();
			return;
		}

		if (ctrl && shift && e.key === "F") {
			e.preventDefault();
			toggleTabSearch();
			return;
		}

		if (ctrl && shift && e.key === "S") {
			e.preventDefault();
			openSerialDialog();
			return;
		}

		if (ctrl && shift && e.key === "N") {
			e.preventDefault();
			openTelnetDialog();
			return;
		}

		if (ctrl && !shift && e.key === ",") {
			e.preventDefault();
			toggleSettings();
			return;
		}

		if (ctrl && shift && (e.key === "T" || e.key === "t")) {
			e.preventDefault();
			newTab();
			return;
		}

		if (ctrl && !shift && e.key === "w" && !inInput) {
			e.preventDefault();
			const t = getActiveTab();
			if (t) t.close();
			return;
		}

		if (ctrl && e.key === "Tab" && !shift) {
			e.preventDefault();
			nextTab();
			return;
		}

		if (ctrl && e.key === "Tab" && shift) {
			e.preventDefault();
			prevTab();
			return;
		}

		if (e.altKey && e.key >= "1" && e.key <= "9") {
			e.preventDefault();
			switchToTab(parseInt(e.key) - 1);
			return;
		}

		if (ctrl && shift && e.key === "L") {
			e.preventDefault();
			showConnectionLog();
		}

		if (ctrl && shift && e.key === "O") {
			e.preventDefault();
			openSettingsPanel();
		}

		if (ctrl && shift && e.key === "E") {
			e.preventDefault();
			exportProfiles();
		}

		if (ctrl && shift && e.key === "B") {
			e.preventDefault();
			toggleBroadcast();
		}

		if (ctrl && !shift && e.key === "\\") {
			e.preventDefault();
			splitVertically();
			return;
		}

		if (ctrl && shift && e.key === "\\") {
			e.preventDefault();
			splitHorizontally();
			return;
		}

		if (ctrl && shift && e.key === "W") {
			e.preventDefault();
			removeSplit();
			return;
		}

		if (ctrl && shift && (e.key === "F" || e.key === "f")) {
			e.preventDefault();
			toggleFind();
			return;
		}

		if (ctrl && shift && (e.key === "C" || e.key === "c") && !inInput) {
			e.preventDefault();
			const t = getActiveTab();
			if (t) t.copySelection();
			return;
		}

		// Ctrl+C: copy if selection, otherwise send to terminal
		if (ctrl && (e.key === "c" || e.key === "C") && !shift && !inInput) {
			const t = getActiveTab();
			if (t) {
				const sel = t.term.getSelection();
				if (sel) {
					// Selection exists: copy and deselect, DON'T send to terminal
					e.preventDefault();
					ClipboardSetText(sel).then(() => {
						showToast("Copied", "success");
					});
					t.term.clearSelection();
					return;
				}
				// No selection: send Ctrl+C to terminal
			}
		}

		if (ctrl && shift && (e.key === "V" || e.key === "v") && !inInput) {
			e.preventDefault();
			const t = getActiveTab();
			if (t) t.pasteFromClipboard();
			return;
		}

		if (ctrl && (e.key === "=" || e.key === "+")) {
			e.preventDefault();
			applyFontSize(Math.min(48, fontSize + 1));
			return;
		}

		if (ctrl && e.key === "-") {
			e.preventDefault();
			applyFontSize(Math.max(8, fontSize - 1));
			return;
		}

		if (ctrl && e.key === "0") {
			e.preventDefault();
			applyFontSize(14);
			return;
		}
	});

	window.addEventListener("focus", () => {
		const tab = getActiveTab();
		if (tab && !tab.exited) tab.term.focus();
	});
}

window.addEventListener("resize", () => {
	const tab = getActiveTab();
	if (tab) {
		tab.fitAddon.fit();
		if (tab.ptyId && !tab.exited)
			PTYResize(tab.ptyId, tab.term.cols, tab.term.rows);
		if (tab.isSSH && tab.sshConnectionId && tab.sshSessionId)
			SSHResize({
				connectionId: tab.sshConnectionId,
				sessionId: tab.sshSessionId,
				columns: tab.term.cols,
				rows: tab.term.rows,
			});

		if (tab.isTelnet && tab.telnetConnectionId)
			TelnetResize(tab.telnetConnectionId, tab.term.cols, tab.term.rows);
	}

	if (activeSplitPane)
		activeSplitPane.panes.forEach((p) => {
			p.tab.fitAddon.fit();
			if (p.tab.ptyId && !p.tab.exited)
				PTYResize(p.tab.ptyId, p.tab.term.cols, p.tab.term.rows);

			if (p.tab.isSSH && p.tab.sshConnectionId && p.tab.sshSessionId)
				SSHResize({
					connectionId: p.tab.sshConnectionId,
					sessionId: p.tab.sshSessionId,
					columns: p.tab.term.cols,
					rows: p.tab.term.rows,
				});

			if (p.tab.isTelnet && p.tab.telnetConnectionId)
				TelnetResize(
					p.tab.telnetConnectionId,
					p.tab.term.cols,
					p.tab.term.rows,
				);
		});
});

// ResizeObserver for reliable terminal fitting when main-content changes size
const resizeObserver = new ResizeObserver(() => {
	const tab = getActiveTab();
	if (tab && tab.term) {
		tab.fitAddon.fit();
	}
});
// Will observe main-content once buildUI runs
setTimeout(() => {
	const mc = document.getElementById("main-content");
	if (mc) resizeObserver.observe(mc);
}, 100);

// Idle connection monitor

setInterval(() => {
	const timeout = settings.IdleTimeout || 0;

	if (timeout <= 0) return;

	tabs.forEach((tab) => {
		if (tab.lastActivity && Date.now() - tab.lastActivity > timeout * 60000) {
			if (tab.status === "connected" && (tab.isSSH || tab.telnetConnectionId)) {
				showToast("Idle timeout: " + tab.title, "info");

				tab.close();
			}
		}
	});
}, 60000);

window.addEventListener("beforeunload", () => {
	saveSession();
});

init();
