package engine

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"lan-screen-cast/internal/capture"
	"lan-screen-cast/internal/codec"
	"lan-screen-cast/internal/diff"
	"lan-screen-cast/internal/network"
	"lan-screen-cast/internal/protocol"
)

type SenderEngine struct {
	capturer capture.Capturer
	detector *diff.Detector
	conn     *network.MsgConn
	id       string
	running  bool
	active   bool
	stopCh   chan struct{}
	mu       sync.Mutex
}

type frameAdapter struct {
	width  int
	height int
	pixels []byte
}

func (f *frameAdapter) Width() int     { return f.width }
func (f *frameAdapter) Height() int    { return f.height }
func (f *frameAdapter) Pixels() []byte { return f.pixels }
func (f *frameAdapter) Stride() int    { return f.width * 4 }

func NewSenderEngine(addr string) (*SenderEngine, error) {
	capturer, err := capture.NewGdiCapturer()
	if err != nil {
		return nil, err
	}
	conn, err := network.Dial(addr)
	if err != nil {
		capturer.Release()
		return nil, err
	}
	hostname, _ := os.Hostname()
	return &SenderEngine{
		capturer: capturer,
		detector: diff.NewDetector(diff.DefaultBlockSize),
		conn:     conn,
		id:       hostname,
		stopCh:   make(chan struct{}),
	}, nil
}

func (e *SenderEngine) ID() string { return e.id }

func (e *SenderEngine) Run() error {
	joinMsg, _ := json.Marshal(protocol.ControlMessage{Type: "join", ID: e.id})
	if err := e.conn.Send(protocol.TypeControl, joinMsg); err != nil {
		return err
	}

	for {
		typ, payload, err := e.conn.Read()
		if err != nil {
			return err
		}
		if typ == protocol.TypeControl {
			var msg protocol.ControlMessage
			json.Unmarshal(payload, &msg)
			if msg.Type == "activate" {
				break
			}
			if msg.Type == "queue_pos" {
				log.Printf("[sender] queue position: %d", msg.Pos)
			}
		}
	}

	e.mu.Lock()
	e.active = true
	e.running = true
	e.mu.Unlock()
	log.Printf("[sender] activated, starting stream")

	ticker := time.NewTicker(time.Second / 15)
	defer ticker.Stop()

	for {
		e.mu.Lock()
		running := e.running
		e.mu.Unlock()
		if !running {
			return nil
		}

		select {
		case <-e.stopCh:
			return nil
		case <-ticker.C:
			frame, err := e.capturer.Capture()
			if err != nil {
				log.Printf("[sender] capture error: %v", err)
				continue
			}
			fa := &frameAdapter{
				width:  frame.Width,
				height: frame.Height,
				pixels: frame.Pixels,
			}
			rects := e.detector.Detect(fa)
			for _, r := range rects {
				png, err := codec.EncodeRGBA(r.Pixels, r.W, r.H)
				if err != nil {
					log.Printf("[sender] encode error: %v", err)
					continue
				}
				block := protocol.EncodeVideoBlock(r.X, r.Y, r.W, r.H, png)
				if err := e.conn.Send(protocol.TypeVideo, block); err != nil {
					log.Printf("[sender] send error: %v", err)
					e.mu.Lock()
					e.running = false
					e.mu.Unlock()
					break
				}
			}
		}
	}
}

func (e *SenderEngine) Activate() {
	e.mu.Lock()
	e.active = true
	e.mu.Unlock()
}

func (e *SenderEngine) Deactivate() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.active = false
	if e.running {
		e.running = false
		select {
		case e.stopCh <- struct{}{}:
		default:
		}
	}
}

func (e *SenderEngine) IsActive() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.active
}

func (e *SenderEngine) Stop() {
	e.mu.Lock()
	e.running = false
	e.mu.Unlock()
	select {
	case e.stopCh <- struct{}{}:
	default:
	}
	stopMsg, _ := json.Marshal(protocol.ControlMessage{Type: "stop"})
	e.conn.Send(protocol.TypeControl, stopMsg)
	e.capturer.Release()
	e.conn.Close()
}
