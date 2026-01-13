package guac

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/gorilla/websocket"
)

type Tunnel struct {
	guacd net.Conn
	ws    *websocket.Conn
	err   error
}

func (t *Tunnel) Handshake(config *Config) error {
	config.SetScreen()
	br := bufio.NewReader(t.guacd)

	selectInstr := NewInstruction("select", config.Protocol)
	if _, err := t.guacd.Write(selectInstr.Byte()); err != nil {
		return fmt.Errorf("send select instruction error: %s", err.Error())
	}
	s, err := br.ReadString(';')
	if err != nil && err != io.EOF {
		return fmt.Errorf("read select instruction response error: %s", err.Error())
	}

	argsInstr := Instruction(s)
	var argsInstrArgs []string
	for _, i := range argsInstr.Args() {
		argsInstrArgs = append(argsInstrArgs, config.Params[i.Value()])
	}

	_, err = t.guacd.Write(NewInstruction(
		"size",
		strconv.Itoa(config.ScreenWidth),
		strconv.Itoa(config.ScreenHeight),
		strconv.Itoa(config.ScreenDpi)).Byte(),
	)
	if err != nil {
		return fmt.Errorf("send size instruction error: %s", err.Error())
	}
	_, err = t.guacd.Write(NewInstruction("audio").Byte())
	if err != nil {
		return fmt.Errorf("send audio instruction error: %s", err.Error())
	}
	_, err = t.guacd.Write(NewInstruction("video").Byte())
	if err != nil {
		return fmt.Errorf("send video instruction error: %s", err.Error())
	}
	_, err = t.guacd.Write(NewInstruction("image").Byte())
	if err != nil {
		return fmt.Errorf("send image instruction error: %s", err.Error())
	}
	_, err = t.guacd.Write(NewInstruction("connect", argsInstrArgs...).Byte())
	if err != nil {
		return fmt.Errorf("send connect instruction error: %s", err.Error())
	}
	s, err = br.ReadString(';')
	if err != nil && err != io.EOF {
		return fmt.Errorf("read connect instruction response error: %s", err.Error())
	}
	readyInstr := Instruction(s)
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

func NewTunnel(netConn net.Conn, wsConn *websocket.Conn) *Tunnel {
	return &Tunnel{
		guacd: netConn,
		ws:    wsConn,
	}
}
