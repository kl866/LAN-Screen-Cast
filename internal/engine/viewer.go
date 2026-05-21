package engine

import (
	"encoding/json"
	"log"
	"sync"

	"lan-screen-cast/internal/codec"
	"lan-screen-cast/internal/network"
	"lan-screen-cast/internal/protocol"
	"lan-screen-cast/internal/session"
)

type ViewerEngine struct {
	listener *network.Listener
	queue    *session.QueueManager
	fb       *FrameBuffer
	running  bool
	mu       sync.Mutex
}

type FrameBuffer struct {
	mu      sync.RWMutex
	Width   int
	Height  int
	Pixels  []byte
	Updated bool
}

func NewFrameBuffer() *FrameBuffer {
	return &FrameBuffer{
		Width:  1920,
		Height: 1080,
		Pixels: make([]byte, 1920*1080*4),
	}
}

func (fb *FrameBuffer) Resize(w, h int) {
	fb.mu.Lock()
	defer fb.mu.Unlock()
	if w == fb.Width && h == fb.Height {
		return
	}
	fb.Width = w
	fb.Height = h
	fb.Pixels = make([]byte, w*h*4)
}

func (fb *FrameBuffer) Apply(x, y, w, h int, pixels []byte) {
	fb.mu.Lock()
	defer fb.mu.Unlock()
	for row := 0; row < h; row++ {
		srcOff := row * w * 4
		dstOff := (y+row)*fb.Width*4 + x*4
		if dstOff+w*4 <= len(fb.Pixels) && srcOff+w*4 <= len(pixels) {
			copy(fb.Pixels[dstOff:dstOff+w*4], pixels[srcOff:srcOff+w*4])
		}
	}
	fb.Updated = true
}

func (fb *FrameBuffer) Snapshot() (w, h int, pixels []byte) {
	fb.mu.Lock()
	defer fb.mu.Unlock()
	cp := make([]byte, len(fb.Pixels))
	copy(cp, fb.Pixels)
	return fb.Width, fb.Height, cp
}

func (fb *FrameBuffer) HasUpdate() bool {
	fb.mu.RLock()
	defer fb.mu.RUnlock()
	return fb.Updated
}

func (fb *FrameBuffer) ClearUpdate() {
	fb.mu.Lock()
	defer fb.mu.Unlock()
	fb.Updated = false
}

func NewViewerEngine(addr string) (*ViewerEngine, error) {
	ln, err := network.Listen(addr)
	if err != nil {
		return nil, err
	}
	return &ViewerEngine{
		listener: ln,
		queue:    session.NewQueueManager(),
		fb:       NewFrameBuffer(),
		running:  true,
	}, nil
}

func (e *ViewerEngine) Addr() string { return e.listener.Addr().String() }

func (e *ViewerEngine) FrameBuffer() *FrameBuffer { return e.fb }

func (e *ViewerEngine) AcceptLoop() {
	for {
		e.mu.Lock()
		running := e.running
		e.mu.Unlock()
		if !running {
			return
		}

		mc, err := e.listener.Accept()
		if err != nil {
			if running {
				log.Printf("[viewer] accept error: %v", err)
			}
			continue
		}
		go e.handleSender(mc)
	}
}

func (e *ViewerEngine) handleSender(mc *network.MsgConn) {
	typ, payload, err := mc.Read()
	if err != nil || typ != protocol.TypeControl {
		mc.Close()
		return
	}
	var msg protocol.ControlMessage
	json.Unmarshal(payload, &msg)
	if msg.Type != "join" {
		mc.Close()
		return
	}

	sess := &senderSession{
		id:     msg.ID,
		conn:   mc,
		engine: e,
	}
	id := e.queue.Enqueue(sess)
	log.Printf("[viewer] sender %s joined, id=%s", msg.ID, id)

	// Read loop — video blocks and stop messages
	for {
		typ, payload, err := mc.Read()
		if err != nil {
			break
		}
		switch typ {
		case protocol.TypeVideo:
			x, y, w, h, pngData, err := protocol.DecodeVideoBlock(payload)
			if err != nil {
				continue
			}
			decoded, err := codec.DecodeToRGBA(pngData)
			if err != nil {
				continue
			}
			// Auto-resize on first frame (heuristic: large block at origin)
			if x == 0 && y == 0 && w >= 800 && decoded.W == w && decoded.H == h {
				e.fb.Resize(w, h)
			}
			e.fb.Apply(x, y, decoded.W, decoded.H, decoded.Pixels)

		case protocol.TypeControl:
			var ctrl protocol.ControlMessage
			json.Unmarshal(payload, &ctrl)
			if ctrl.Type == "stop" {
				ack, _ := json.Marshal(protocol.ControlMessage{Type: "ack_stop"})
				mc.Send(protocol.TypeControl, ack)
				break
			}
		}
	}

	e.queue.Remove(sess.id)
	mc.Close()
	log.Printf("[viewer] sender %s disconnected", sess.id)
}

type senderSession struct {
	id     string
	conn   *network.MsgConn
	engine *ViewerEngine
	active bool
}

func (s *senderSession) ID() string     { return s.id }
func (s *senderSession) IsActive() bool { return s.active }
func (s *senderSession) Activate() {
	s.active = true
	actMsg, _ := json.Marshal(protocol.ControlMessage{Type: "activate"})
	s.conn.Send(protocol.TypeControl, actMsg)
}
func (s *senderSession) Deactivate() { s.active = false }

func (e *ViewerEngine) DisconnectActive() {
	active := e.queue.Active()
	if active != nil {
		e.queue.Remove(active.ID())
	}
}

func (e *ViewerEngine) Stop() {
	e.mu.Lock()
	e.running = false
	e.mu.Unlock()
	e.listener.Close()
}
