package tunnel

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/riete/convert/str"
	"github.com/riete/go-guac/protocol"
)

const minKeepaliveInterval = 30 * time.Second

type TunnelOption func(t *Tunnel)

func WithOnConnect(f func(string)) TunnelOption {
	return func(t *Tunnel) {
		t.onConnect = f
	}
}

func WithOnReadFromGuacd(f func(string, []byte)) TunnelOption {
	return func(t *Tunnel) {
		t.onReadFromGuacd = f
	}
}

func WithOnReadFromWs(f func(string, []byte)) TunnelOption {
	return func(t *Tunnel) {
		t.onReadFromWs = f
	}
}

func WithOnDisconnect(f func(string)) TunnelOption {
	return func(t *Tunnel) {
		t.onDisconnect = f
	}
}

func WithGuacdKeepalive(interval time.Duration) TunnelOption {
	return func(t *Tunnel) {
		if interval < minKeepaliveInterval {
			interval = minKeepaliveInterval
		}
		t.guacdKeepaliveInterval = interval
	}
}

func WithWsKeepalive(interval time.Duration, threshold int64) TunnelOption {
	return func(t *Tunnel) {
		if interval < minKeepaliveInterval {
			interval = minKeepaliveInterval
		}
		if threshold < 1 {
			threshold = 1
		}
		t.wsKeepaliveInterval = interval
		t.wsKeepaliveThreshold = threshold
	}
}

type Tunnel struct {
	guacd                  net.Conn
	ws                     *websocket.Conn
	err                    error
	connId                 string
	guacdKeepaliveInterval time.Duration
	wsKeepaliveInterval    time.Duration
	wsKeepaliveThreshold   int64
	onConnect              func(connId string)
	onReadFromGuacd        func(connId string, fromGuacd []byte)
	onReadFromWs           func(connId string, fromWs []byte)
	onDisconnect           func(connId string)
}

// Handshake performs the complete handshake process.
// The handshake flow is:
//  1. Client sends "select" with the protocol name (vnc, rdp, ssh)
//  2. Server responds with "args" listing required parameters
//  3. Client sends "size" with display dimensions
//  4. Client sends "audio" with supported audio MIME types
//  5. Client sends "video" with supported video MIME types
//  6. Client sends "image" with supported image MIME types
//  7. Client sends "connect" with parameter values (in order from args)
//  8. Server responds with "ready" containing the connection ID
func (t *Tunnel) Handshake(config *protocol.HandshakeConfig) error {
	br := bufio.NewReader(t.guacd)
	if _, err := t.guacd.Write(config.SelectInstruction().Byte()); err != nil {
		return fmt.Errorf("send select instruction error: %s", err.Error())
	}
	selectResponse, err := br.ReadString(';')
	if err != nil && err != io.EOF {
		return fmt.Errorf("read select instruction response error: %s", err.Error())
	}
	argsInstr := protocol.Instruction(selectResponse)
	if err = argsInstr.Error(); err != nil {
		return err
	}

	fullConnectInstr := config.SizeInstruction() + config.AudioInstruction() + config.VideoInstruction() +
		config.ImageInstruction() + config.ConnectInstruction(argsInstr.Args())
	if _, err = t.guacd.Write(fullConnectInstr.Byte()); err != nil {
		return fmt.Errorf("send full connect instruction error: %s", err.Error())
	}
	connectResponse, err := br.ReadString(';')
	if err != nil && err != io.EOF {
		return fmt.Errorf("read connect instruction response error: %s", err.Error())
	}
	readyInstr := protocol.Instruction(connectResponse)
	if err = readyInstr.Error(); err != nil {
		return err
	}
	if len(readyInstr.Args()) == 0 {
		return errors.New("no connection ID received")
	}
	t.connId = readyInstr.Args()[0].Value()
	if t.onConnect != nil {
		t.onConnect(t.connId)
	}
	return nil
}

func (t *Tunnel) ConnId() string {
	return t.connId
}

func (t *Tunnel) Close() {
	_, _ = t.guacd.Write(protocol.Disconnect.Byte())
	_ = t.guacd.Close()
	_ = t.ws.Close()
	if t.onDisconnect != nil {
		t.onDisconnect(t.connId)
	}
	t.connId = ""
}

func (t *Tunnel) setError(err error) {
	if t.err == nil {
		t.err = err
	}
}

func (t *Tunnel) guacdToWs(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	br := bufio.NewReader(t.guacd)
	var once sync.Once
	for {
		select {
		case <-ctx.Done():
			return
		default:
			b, err := br.ReadBytes(';')
			if err == io.EOF {
				return
			}
			if err != nil {
				t.setError(fmt.Errorf("read data from guacd error: %s", err.Error()))
				return
			}
			once.Do(func() {
				// check first instruction after handshake, maybe some error, e.g. CLIENT_UNAUTHORIZED
				instr := protocol.Instruction(str.FromBytes(b))
				if err = instr.Error(); err != nil {
					t.setError(err)
				}
			})
			if t.onReadFromGuacd != nil {
				t.onReadFromGuacd(t.connId, b)
			}
			if err = t.ws.WriteMessage(websocket.TextMessage, b); err != nil {
				t.setError(fmt.Errorf("write data to ws error: %s", err.Error()))
				return
			}
		}
	}
}

func (t *Tunnel) wsToGuacd(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, data, err := t.ws.ReadMessage()
			if err != nil {
				t.setError(fmt.Errorf("read data from ws error: %s", err.Error()))
				return
			}
			if t.onReadFromWs != nil {
				t.onReadFromWs(t.connId, data)
			}
			if _, err = t.guacd.Write(data); err != nil {
				t.setError(fmt.Errorf("write data to guacd error: %s", err.Error()))
				return
			}
		}
	}
}

func (t *Tunnel) guacdKeepalive(ctx context.Context) {
	ticker := time.NewTicker(t.guacdKeepaliveInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = t.guacd.Write(protocol.Nop.Byte())
		}
	}
}

func (t *Tunnel) wsKeepalive(ctx context.Context) {
	ticker := time.NewTicker(t.wsKeepaliveInterval)
	defer ticker.Stop()
	deadline := t.wsKeepaliveInterval * time.Duration(t.wsKeepaliveThreshold)
	_ = t.ws.SetReadDeadline(time.Now().Add(deadline))
	originalPongHandler := t.ws.PongHandler()
	t.ws.SetPongHandler(func(appData string) error {
		_ = t.ws.SetReadDeadline(time.Now().Add(deadline))
		return originalPongHandler(appData)
	})
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = t.ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second))
		}
	}
}

func (t *Tunnel) Forward(ctx context.Context) error {
	newCtx, cancel := context.WithCancel(ctx)
	if t.guacdKeepaliveInterval > 0 {
		go t.guacdKeepalive(newCtx)
	}
	if t.wsKeepaliveInterval > 0 {
		go t.wsKeepalive(newCtx)
	}
	go t.guacdToWs(newCtx, cancel)
	go t.wsToGuacd(newCtx, cancel)
	<-newCtx.Done()
	return t.err
}

func NewTunnel(guacd net.Conn, ws *websocket.Conn, opts ...TunnelOption) *Tunnel {
	t := &Tunnel{
		guacd: guacd,
		ws:    ws,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}
