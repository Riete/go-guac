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

type Tunnel struct {
	guacd net.Conn
	ws    *websocket.Conn
	err   error
}

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
	return nil
}

func (t *Tunnel) Close() {
	_, _ = t.guacd.Write(NewInstruction("disconnect").Byte())
	_ = t.guacd.Close()
	_ = t.ws.Close()
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
	<-ctx.Done()
	return t.err
}

func NewTunnel(guacd net.Conn, ws *websocket.Conn) *Tunnel {
	return &Tunnel{
		guacd: guacd,
		ws:    ws,
	}
}
