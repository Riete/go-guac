package recorder

import (
	"context"
)

type Recorder interface {
	Record(connId string, data []byte)
	Replay(ctx context.Context, connId string) (chan string, error)
	Close(connId string)
}
