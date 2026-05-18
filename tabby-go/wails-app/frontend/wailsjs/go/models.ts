export namespace api {
	
	export class PTYSpawnParams {
	    id: string;
	    command: string;
	    args?: string[];
	    env?: Record<string, string>;
	    cwd?: string;
	    columns: number;
	    rows: number;
	
	    static createFrom(source: any = {}) {
	        return new PTYSpawnParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.command = source["command"];
	        this.args = source["args"];
	        this.env = source["env"];
	        this.cwd = source["cwd"];
	        this.columns = source["columns"];
	        this.rows = source["rows"];
	    }
	}
	export class PTYSpawnResult {
	    id: string;
	    pid: number;
	
	    static createFrom(source: any = {}) {
	        return new PTYSpawnResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.pid = source["pid"];
	    }
	}
	export class PortForwardInfo {
	    id: string;
	    type: string;
	    host: string;
	    port: number;
	    targetAddress?: string;
	    targetPort?: number;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PortForwardInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.targetAddress = source["targetAddress"];
	        this.targetPort = source["targetPort"];
	        this.active = source["active"];
	    }
	}
	export class PortForwardParams {
	    connectionId: string;
	    type: string;
	    host: string;
	    port: number;
	    targetAddress: string;
	    targetPort: number;
	
	    static createFrom(source: any = {}) {
	        return new PortForwardParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.type = source["type"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.targetAddress = source["targetAddress"];
	        this.targetPort = source["targetPort"];
	    }
	}
	export class PortForwardRemoveParams {
	    connectionId: string;
	    forwardId: string;
	
	    static createFrom(source: any = {}) {
	        return new PortForwardRemoveParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.forwardId = source["forwardId"];
	    }
	}
	export class PortForwardResult {
	    forwardId: string;
	
	    static createFrom(source: any = {}) {
	        return new PortForwardResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.forwardId = source["forwardId"];
	    }
	}
	export class SFTPDownloadParams {
	    sessionId: string;
	    remotePath: string;
	    localPath: string;
	
	    static createFrom(source: any = {}) {
	        return new SFTPDownloadParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.remotePath = source["remotePath"];
	        this.localPath = source["localPath"];
	    }
	}
	export class SFTPFile {
	    name: string;
	    fullPath: string;
	    size: number;
	    mode: number;
	    modTime: string;
	    isDir: boolean;
	    isSymlink: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SFTPFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.fullPath = source["fullPath"];
	        this.size = source["size"];
	        this.mode = source["mode"];
	        this.modTime = source["modTime"];
	        this.isDir = source["isDir"];
	        this.isSymlink = source["isSymlink"];
	    }
	}
	export class SFTPListParams {
	    sessionId: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new SFTPListParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.path = source["path"];
	    }
	}
	export class SFTPOpenParams {
	    connectionId: string;
	
	    static createFrom(source: any = {}) {
	        return new SFTPOpenParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	    }
	}
	export class SFTPOpenResult {
	    sessionId: string;
	
	    static createFrom(source: any = {}) {
	        return new SFTPOpenResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	    }
	}
	export class SFTPTransferResult {
	    bytesTransferred: number;
	
	    static createFrom(source: any = {}) {
	        return new SFTPTransferResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bytesTransferred = source["bytesTransferred"];
	    }
	}
	export class SFTPUploadParams {
	    sessionId: string;
	    localPath: string;
	    remotePath: string;
	
	    static createFrom(source: any = {}) {
	        return new SFTPUploadParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.localPath = source["localPath"];
	        this.remotePath = source["remotePath"];
	    }
	}
	export class SSHAlgorithms {
	    hmac?: string[];
	    kex?: string[];
	    cipher?: string[];
	    serverHostKey?: string[];
	    compression?: string[];
	
	    static createFrom(source: any = {}) {
	        return new SSHAlgorithms(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hmac = source["hmac"];
	        this.kex = source["kex"];
	        this.cipher = source["cipher"];
	        this.serverHostKey = source["serverHostKey"];
	        this.compression = source["compression"];
	    }
	}
	export class SSHAuthParams {
	    type: string;
	    password?: string;
	    privateKey?: string;
	    privateKeyPaths?: string[];
	    agentSocketPath?: string;
	    agentType?: string;
	    keyboardInteractivePassthrough?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SSHAuthParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.password = source["password"];
	        this.privateKey = source["privateKey"];
	        this.privateKeyPaths = source["privateKeyPaths"];
	        this.agentSocketPath = source["agentSocketPath"];
	        this.agentType = source["agentType"];
	        this.keyboardInteractivePassthrough = source["keyboardInteractivePassthrough"];
	    }
	}
	export class SSHCloseParams {
	    connectionId: string;
	    sessionId?: string;
	
	    static createFrom(source: any = {}) {
	        return new SSHCloseParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.sessionId = source["sessionId"];
	    }
	}
	export class SSHConnectParams {
	    host: string;
	    port?: number;
	    user: string;
	    auth: SSHAuthParams;
	    keepaliveInterval?: number;
	    keepaliveCountMax?: number;
	    readyTimeout?: number;
	    agentForward?: boolean;
	    x11?: boolean;
	    x11Display?: string;
	    jumpHost?: SSHConnectParams;
	    algorithms?: SSHAlgorithms;
	    proxyCommand?: string;
	    socksProxyHost?: string;
	    socksProxyPort?: number;
	    httpProxyHost?: string;
	    httpProxyPort?: number;
	    environment?: Record<string, string>;
	    verifyHostKey?: boolean;
	    knownHostsPath?: string;
	    skipBanner?: boolean;
	    password?: string;
	
	    static createFrom(source: any = {}) {
	        return new SSHConnectParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.port = source["port"];
	        this.user = source["user"];
	        this.auth = this.convertValues(source["auth"], SSHAuthParams);
	        this.keepaliveInterval = source["keepaliveInterval"];
	        this.keepaliveCountMax = source["keepaliveCountMax"];
	        this.readyTimeout = source["readyTimeout"];
	        this.agentForward = source["agentForward"];
	        this.x11 = source["x11"];
	        this.x11Display = source["x11Display"];
	        this.jumpHost = this.convertValues(source["jumpHost"], SSHConnectParams);
	        this.algorithms = this.convertValues(source["algorithms"], SSHAlgorithms);
	        this.proxyCommand = source["proxyCommand"];
	        this.socksProxyHost = source["socksProxyHost"];
	        this.socksProxyPort = source["socksProxyPort"];
	        this.httpProxyHost = source["httpProxyHost"];
	        this.httpProxyPort = source["httpProxyPort"];
	        this.environment = source["environment"];
	        this.verifyHostKey = source["verifyHostKey"];
	        this.knownHostsPath = source["knownHostsPath"];
	        this.skipBanner = source["skipBanner"];
	        this.password = source["password"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SSHConnectionResult {
	    connectionId: string;
	    serverVersion: string;
	    remoteAddress: string;
	    banner?: string;
	    authMethods: string[];
	    jumpChain?: string[];
	
	    static createFrom(source: any = {}) {
	        return new SSHConnectionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.serverVersion = source["serverVersion"];
	        this.remoteAddress = source["remoteAddress"];
	        this.banner = source["banner"];
	        this.authMethods = source["authMethods"];
	        this.jumpChain = source["jumpChain"];
	    }
	}
	export class SSHResizeParams {
	    connectionId: string;
	    sessionId: string;
	    columns: number;
	    rows: number;
	
	    static createFrom(source: any = {}) {
	        return new SSHResizeParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.sessionId = source["sessionId"];
	        this.columns = source["columns"];
	        this.rows = source["rows"];
	    }
	}
	export class SSHSessionParams {
	    connectionId: string;
	    columns: number;
	    rows: number;
	    terminal?: string;
	    command?: string;
	    agentForward?: boolean;
	    x11?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SSHSessionParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.columns = source["columns"];
	        this.rows = source["rows"];
	        this.terminal = source["terminal"];
	        this.command = source["command"];
	        this.agentForward = source["agentForward"];
	        this.x11 = source["x11"];
	    }
	}
	export class SSHSessionResult {
	    sessionId: string;
	
	    static createFrom(source: any = {}) {
	        return new SSHSessionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	    }
	}
	export class SSHWriteParams {
	    connectionId: string;
	    sessionId: string;
	    data: string;
	
	    static createFrom(source: any = {}) {
	        return new SSHWriteParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.sessionId = source["sessionId"];
	        this.data = source["data"];
	    }
	}
	export class SerialOpenParams {
	    id: string;
	    port: string;
	    baudRate: number;
	    dataBits?: number;
	    stopBits?: number;
	    parity?: string;
	    flowControl?: string;
	
	    static createFrom(source: any = {}) {
	        return new SerialOpenParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.port = source["port"];
	        this.baudRate = source["baudRate"];
	        this.dataBits = source["dataBits"];
	        this.stopBits = source["stopBits"];
	        this.parity = source["parity"];
	        this.flowControl = source["flowControl"];
	    }
	}
	export class SerialOpenResult {
	    id: string;
	
	    static createFrom(source: any = {}) {
	        return new SerialOpenResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	    }
	}
	export class SerialPortInfo {
	    name: string;
	    manufacturer?: string;
	    product?: string;
	    serialNumber?: string;
	    vid?: string;
	    pid?: string;
	
	    static createFrom(source: any = {}) {
	        return new SerialPortInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.manufacturer = source["manufacturer"];
	        this.product = source["product"];
	        this.serialNumber = source["serialNumber"];
	        this.vid = source["vid"];
	        this.pid = source["pid"];
	    }
	}

}

export namespace colorscheme {
	
	export class ColorScheme {
	    name: string;
	    foreground: string;
	    background: string;
	    cursor: string;
	    cursorAccent?: string;
	    selection?: string;
	    selectionForeground?: string;
	    colors: string[];
	
	    static createFrom(source: any = {}) {
	        return new ColorScheme(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.foreground = source["foreground"];
	        this.background = source["background"];
	        this.cursor = source["cursor"];
	        this.cursorAccent = source["cursorAccent"];
	        this.selection = source["selection"];
	        this.selectionForeground = source["selectionForeground"];
	        this.colors = source["colors"];
	    }
	}

}

export namespace notification {
	
	export class Notification {
	    id: string;
	    level: number;
	    title: string;
	    message: string;
	    // Go type: time
	    timestamp: any;
	    read: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Notification(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.level = source["level"];
	        this.title = source["title"];
	        this.message = source["message"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.read = source["read"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace profile {
	
	export class ConnectionProfile {
	    id: string;
	    type: string;
	    name: string;
	    group?: string;
	    icon?: string;
	    color?: string;
	    disableDynamicTitle?: boolean;
	    options: any;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.name = source["name"];
	        this.group = source["group"];
	        this.icon = source["icon"];
	        this.color = source["color"];
	        this.disableDynamicTitle = source["disableDynamicTitle"];
	        this.options = source["options"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace session {
	
	export class TabState {
	    shell: string;
	    title: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TabState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.shell = source["shell"];
	        this.title = source["title"];
	        this.active = source["active"];
	    }
	}
	export class SessionState {
	    tabs: TabState[];
	    version: number;
	
	    static createFrom(source: any = {}) {
	        return new SessionState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tabs = this.convertValues(source["tabs"], TabState);
	        this.version = source["version"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace settings {
	
	export class Settings {
	    color_scheme: string;
	    font_size: number;
	    font_family: string;
	    fallback_font: string;
	    font_weight: number;
	    font_weight_bold: number;
	    line_height: number;
	    line_padding: number;
	    ligatures: boolean;
	    theme: string;
	    css: string;
	    opacity: number;
	    spaciness: number;
	    animations: boolean;
	    shell: string;
	    scrollback: number;
	    cursor_style: string;
	    cursor_blink: boolean;
	    bell: string;
	    frontend: string;
	    draw_bold_text_in_bright_colors: boolean;
	    minimum_contrast_ratio: number;
	    alt_is_meta: boolean;
	    scroll_on_input: boolean;
	    copy_on_select: boolean;
	    copy_as_html: boolean;
	    bracketed_paste: boolean;
	    warn_on_multiline_paste: boolean;
	    replace_newlines_on_paste: boolean;
	    trim_whitespace_on_paste: boolean;
	    right_click: string;
	    paste_on_middle_click: boolean;
	    word_separator: string;
	    tab_position: string;
	    last_tab_closes_window: boolean;
	    cycle_tabs: boolean;
	    hide_close_button: boolean;
	    show_tab_profile_icon: boolean;
	    pane_resize_step: number;
	    focus_follows_mouse: boolean;
	    auto_open: boolean;
	    recover_tabs: boolean;
	    frame: string;
	    dock: string;
	    dock_hide_on_blur: boolean;
	    dock_always_on_top: boolean;
	    vibrancy: boolean;
	    hide_tray: boolean;
	    ssh_warn_on_close: boolean;
	    ssh_verify_host_keys: boolean;
	    ssh_agent_type: string;
	    ssh_agent_path: string;
	    ssh_x11_display: string;
	    ssh_disable_dynamic_title: boolean;
	    serial_baud_rate: number;
	    serial_data_bits: number;
	    serial_stop_bits: number;
	    serial_parity: string;
	    serial_flow_control: string;
	    use_conpty: boolean;
	    set_comspec: boolean;
	    language: string;
	    enable_analytics: boolean;
	    enable_automatic_updates: boolean;
	    enable_experimental_features: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.color_scheme = source["color_scheme"];
	        this.font_size = source["font_size"];
	        this.font_family = source["font_family"];
	        this.fallback_font = source["fallback_font"];
	        this.font_weight = source["font_weight"];
	        this.font_weight_bold = source["font_weight_bold"];
	        this.line_height = source["line_height"];
	        this.line_padding = source["line_padding"];
	        this.ligatures = source["ligatures"];
	        this.theme = source["theme"];
	        this.css = source["css"];
	        this.opacity = source["opacity"];
	        this.spaciness = source["spaciness"];
	        this.animations = source["animations"];
	        this.shell = source["shell"];
	        this.scrollback = source["scrollback"];
	        this.cursor_style = source["cursor_style"];
	        this.cursor_blink = source["cursor_blink"];
	        this.bell = source["bell"];
	        this.frontend = source["frontend"];
	        this.draw_bold_text_in_bright_colors = source["draw_bold_text_in_bright_colors"];
	        this.minimum_contrast_ratio = source["minimum_contrast_ratio"];
	        this.alt_is_meta = source["alt_is_meta"];
	        this.scroll_on_input = source["scroll_on_input"];
	        this.copy_on_select = source["copy_on_select"];
	        this.copy_as_html = source["copy_as_html"];
	        this.bracketed_paste = source["bracketed_paste"];
	        this.warn_on_multiline_paste = source["warn_on_multiline_paste"];
	        this.replace_newlines_on_paste = source["replace_newlines_on_paste"];
	        this.trim_whitespace_on_paste = source["trim_whitespace_on_paste"];
	        this.right_click = source["right_click"];
	        this.paste_on_middle_click = source["paste_on_middle_click"];
	        this.word_separator = source["word_separator"];
	        this.tab_position = source["tab_position"];
	        this.last_tab_closes_window = source["last_tab_closes_window"];
	        this.cycle_tabs = source["cycle_tabs"];
	        this.hide_close_button = source["hide_close_button"];
	        this.show_tab_profile_icon = source["show_tab_profile_icon"];
	        this.pane_resize_step = source["pane_resize_step"];
	        this.focus_follows_mouse = source["focus_follows_mouse"];
	        this.auto_open = source["auto_open"];
	        this.recover_tabs = source["recover_tabs"];
	        this.frame = source["frame"];
	        this.dock = source["dock"];
	        this.dock_hide_on_blur = source["dock_hide_on_blur"];
	        this.dock_always_on_top = source["dock_always_on_top"];
	        this.vibrancy = source["vibrancy"];
	        this.hide_tray = source["hide_tray"];
	        this.ssh_warn_on_close = source["ssh_warn_on_close"];
	        this.ssh_verify_host_keys = source["ssh_verify_host_keys"];
	        this.ssh_agent_type = source["ssh_agent_type"];
	        this.ssh_agent_path = source["ssh_agent_path"];
	        this.ssh_x11_display = source["ssh_x11_display"];
	        this.ssh_disable_dynamic_title = source["ssh_disable_dynamic_title"];
	        this.serial_baud_rate = source["serial_baud_rate"];
	        this.serial_data_bits = source["serial_data_bits"];
	        this.serial_stop_bits = source["serial_stop_bits"];
	        this.serial_parity = source["serial_parity"];
	        this.serial_flow_control = source["serial_flow_control"];
	        this.use_conpty = source["use_conpty"];
	        this.set_comspec = source["set_comspec"];
	        this.language = source["language"];
	        this.enable_analytics = source["enable_analytics"];
	        this.enable_automatic_updates = source["enable_automatic_updates"];
	        this.enable_experimental_features = source["enable_experimental_features"];
	    }
	}

}

export namespace telnet {
	
	export class TelnetConnectResult {
	    connectionId: string;
	    host: string;
	    port: number;
	
	    static createFrom(source: any = {}) {
	        return new TelnetConnectResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.host = source["host"];
	        this.port = source["port"];
	    }
	}

}

export namespace updater {
	
	export class Asset {
	    name: string;
	    browser_download_url: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new Asset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.browser_download_url = source["browser_download_url"];
	        this.size = source["size"];
	    }
	}
	export class ReleaseInfo {
	    tag_name: string;
	    name: string;
	    body: string;
	    html_url: string;
	    // Go type: time
	    published_at: any;
	    prerelease: boolean;
	    draft: boolean;
	    assets: Asset[];
	
	    static createFrom(source: any = {}) {
	        return new ReleaseInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tag_name = source["tag_name"];
	        this.name = source["name"];
	        this.body = source["body"];
	        this.html_url = source["html_url"];
	        this.published_at = this.convertValues(source["published_at"], null);
	        this.prerelease = source["prerelease"];
	        this.draft = source["draft"];
	        this.assets = this.convertValues(source["assets"], Asset);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UpdateStatus {
	    currentVersion: string;
	    latestVersion: string;
	    updateAvailable: boolean;
	    releaseInfo?: ReleaseInfo;
	    error?: string;
	    // Go type: time
	    checkedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new UpdateStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.updateAvailable = source["updateAvailable"];
	        this.releaseInfo = this.convertValues(source["releaseInfo"], ReleaseInfo);
	        this.error = source["error"];
	        this.checkedAt = this.convertValues(source["checkedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

