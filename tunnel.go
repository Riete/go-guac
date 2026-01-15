package guac

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/gorilla/websocket"
)

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

type Tunnel struct {
	guacd           net.Conn
	ws              *websocket.Conn
	err             error
	connId          string
	onConnect       func(connId string)
	onReadFromGuacd func(connId string, fromGuacd []byte)
	onReadFromWs    func(connId string, fromWs []byte)
	onDisconnect    func(connId string)
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
func (t *Tunnel) Handshake(config *HandshakeConfig) error {
	br := bufio.NewReader(t.guacd)
	if _, err := t.guacd.Write(config.SelectInstruction().Byte()); err != nil {
		return fmt.Errorf("send select instruction error: %s", err.Error())
	}
	selectResponse, err := br.ReadString(';')
	if err != nil && err != io.EOF {
		return fmt.Errorf("read select instruction response error: %s", err.Error())
	}

	_, err = t.guacd.Write(config.SizeInstruction().Byte())
	if err != nil {
		return fmt.Errorf("send size instruction error: %s", err.Error())
	}
	_, err = t.guacd.Write(config.AudioInstruction().Byte())
	if err != nil {
		return fmt.Errorf("send audio instruction error: %s", err.Error())
	}
	_, err = t.guacd.Write(config.VideoInstruction().Byte())
	if err != nil {
		return fmt.Errorf("send video instruction error: %s", err.Error())
	}
	_, err = t.guacd.Write(config.ImageInstruction().Byte())
	if err != nil {
		return fmt.Errorf("send image instruction error: %s", err.Error())
	}

	argsInstr := Instruction(selectResponse)
	_, err = t.guacd.Write(config.ConnectInstruction(argsInstr.Args()).Byte())
	if err != nil {
		return fmt.Errorf("send connect instruction error: %s", err.Error())
	}
	connectResponse, err := br.ReadString(';')
	if err != nil && err != io.EOF {
		return fmt.Errorf("read connect instruction response error: %s", err.Error())
	}
	readyInstr := Instruction(connectResponse)
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
	_, _ = t.guacd.Write(NewInstruction("disconnect").Byte())
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
	for {
		select {
		case <-ctx.Done():
			return
		default:
			b, err := br.ReadBytes(';')
			if err != nil && err != io.EOF {
				t.setError(fmt.Errorf("read data from guacd error: %s", err.Error()))
				return
			}
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

func (t *Tunnel) Forward(ctx context.Context) error {
	newCtx, cancel := context.WithCancel(ctx)
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
