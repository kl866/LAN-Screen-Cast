# LAN Screen Cast

A lightweight LAN screen casting tool for Windows, written in pure Go with no CGo dependencies. Share your screen to another PC on the same local network with low latency.

## Features

- **Low latency** — diff-based dirty block detection with xxhash, only changed regions are encoded and transmitted
- **Pure Go + Win32 API** — no CGo, no external dependencies beyond Go standard library and xxhash
- **Minimal bandwidth** — 16×16 block-level change detection, PNG-encoded dirty rects only
- **Sender queue** — multiple senders can connect; one active at a time, others wait in queue
- **IP memory** — sender remembers the last connected IP address
- **Custom UI** — native Win32 windows with owner-draw styling, Chinese/English text
- **Log to file** — both sender and viewer write logs to local files, console window hidden

## Architecture

```
Sender                              Viewer
┌──────────────┐                    ┌──────────────┐
│  GDI Capture │                    │  FrameBuffer │
│       ↓      │                    │       ↓      │
│  Diff Detect │── TCP ──────────→ │  PNG Decode  │
│       ↓      │   (custom proto)   │       ↓      │
│  PNG Encode  │                    │  GDI Render  │
└──────────────┘                    └──────────────┘
```

- **Capture**: GDI `BitBlt` into a 32-bit DIBSection for direct memory access
- **Diff**: 16×16 block hashing with xxhash, comparing against previous frame
- **Codec**: PNG encode dirty rects (best speed), decode on viewer side
- **Network**: Custom binary protocol with 9-byte header (magic + type + length), TCP transport
- **Session**: Queue manager with activate/deactivate, position tracking

## Project Structure

```
lan-screen-cast/
├── cmd/
│   ├── sender/main.go      # Sender entry point
│   └── viewer/main.go      # Viewer entry point
├── internal/
│   ├── capture/            # GDI screen capture (Windows)
│   ├── codec/              # PNG encode/decode
│   ├── diff/               # Dirty block detection (xxhash)
│   ├── engine/             # Sender/viewer core logic
│   ├── network/            # TCP client/server with custom protocol
│   ├── protocol/           # Wire protocol & control messages
│   ├── session/            # Sender queue manager
│   └── ui/                 # Win32 UI (sender & viewer windows)
├── sender.ico              # Sender application icon
├── viewer.ico              # Viewer application icon
├── go.mod
└── go.sum
```

## Quick Start

### Prerequisites

- Windows 10 or later
- Go 1.23+
- [windres](https://sourceware.org/binutils/docs/binutils/windres.html) (for embedding icons, optional)

### Build

```bash
# Build sender
cd cmd/sender
windres -o rsrc.syso sender.rc   # embed icon (optional)
go build -ldflags "-H windowsgui" -o sender.exe .

# Build viewer
cd cmd/viewer
windres -o rsrc.syso viewer.rc   # embed icon (optional)
go build -ldflags "-H windowsgui" -o viewer.exe .
```

The `-H windowsgui` flag hides the console window. Logs are written to `sender.log` / `viewer.log` next to the executables.

Pre-built `rsrc.syso` files are included in the repository, so `windres` is optional — icons will be embedded as-is.

## Usage

### Viewer (接收端)

在要观看投屏的电脑上启动 viewer.exe，等待发送端连接。

```bash
viewer.exe -port :9527
```

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-port` | 监听端口 | `:9527` |

启动后会出现一个窗口，显示 **"LAN Screen Cast"** 标题和本机监听地址（如 `192.168.1.100:9527`）。

窗口下方会显示蓝色的地址信息，发送端需要填写这个地址来连接。

**提示**：如果窗口显示灰色背景，说明还没有发送端连接，等待连接中。

### Sender (发送端)

在要共享屏幕的电脑上启动 sender.exe，连接到 viewer。

```bash
sender.exe
```

启动后会出现一个 **500×380** 的窗口，界面包含：

| 区域 | 说明 |
|------|------|
| **标题栏** | 蓝色背景，显示 "LAN Screen Cast - 发送端" |
| **IP 地址** | 输入 viewer 的 IP 地址，如 `192.168.1.100` |
| **端口号** | viewer 的监听端口，默认 `9527` |
| **状态提示** | 显示当前连接状态（等待连接、投屏中等） |
| **开始投屏 按钮** | 蓝色按钮，点击后连接到 viewer 并开始投屏 |
| **停止投屏 按钮** | 红色按钮，点击后断开连接，停止投屏 |

### 使用步骤

1. **在接收端电脑上**，双击运行 `viewer.exe`，记下窗口显示的 IP 地址和端口
2. **在发送端电脑上**，双击运行 `sender.exe`
3. 在发送端窗口中输入接收端的 IP 地址和端口
4. 点击 **"开始投屏"** 按钮
5. 接收端窗口随即显示发送端的屏幕画面
6. 需要停止时，点击发送端的 **"停止投屏"** 按钮

> **注意**：两台电脑需要在同一局域网内（能互相 ping 通）。

### 多发送端队列

当有多个发送端同时连接到一个 viewer 时：

- **第一个**连接的发送端立即激活，开始投屏
- **后续**发送端进入等待队列，窗口会显示当前排队位置
- 当前投屏端断开后，队列中的**下一个**自动激活
- 每个发送端只能看到自己的排队状态

### IP 记忆

发送端会自动保存上次成功连接的 IP 地址。配置文件 `sender.conf` 保存在 sender.exe 同目录下，下次启动时自动填充。

### 日志文件

两个程序都会在 exe 同目录下生成日志文件：

| 文件 | 说明 |
|------|------|
| `sender.log` | 发送端日志，记录连接状态和错误信息 |
| `viewer.log` | 接收端日志，仅记录错误信息 |

遇到问题时可以查看日志文件排查。

### 如何查看本机 IP

如果不确定 viewer 的 IP 地址，可以在命令行中运行：

```bash
ipconfig
```

找到 IPv4 地址，通常是 `192.168.x.x` 或 `10.x.x.x` 格式。

viewer 窗口标题栏也会自动显示检测到的本机局域网 IP 地址。

## How It Works

1. Sender connects to viewer via TCP
2. Sender sends a `join` control message with screen dimensions
3. Viewer acknowledges with `activate` (or sends queue position if another sender is active)
4. Sender captures screen at ~30 FPS, detects dirty blocks, PNG-encodes them, sends via custom protocol
5. Viewer decodes and applies each block to a frame buffer, renders via `StretchBlt`

### Protocol

Custom binary protocol with 9-byte header:

```
| Magic (4B) | Type (1B) | Payload Length (4B) | Payload (variable) |
```

- Magic: `LSC\x01`
- Types: `0x01` (control), `0x02` (video)
- Max payload: 100 MB

## Performance

- Screen capture via GDI `BitBlt` at physical resolution (DESKTOPHORZRES/DESKTOPVERTRES)
- Block-level diffing reduces bandwidth to changed regions only
- ~30 FPS capture rate, ~60 FPS viewer render with `COLORONCOLOR` stretch mode
- Direct DIBSection memory access avoids extra copies on capture side

## Limitations

- Windows only (Win32 GDI API)
- Single display capture (primary monitor)
- No audio transmission
- No encryption — intended for trusted LAN environments
- PNG encoding trades CPU for bandwidth; H.264 would be more efficient for full-screen video

## License

MIT
