package guac

import (
	"context"
	"testing"
)

func TestFileRecorder_Replay(t *testing.T) {
	r := NewFileRecorder("/tmp", true)
	ch, err := r.Replay(context.Background(), "c65c94f9-61cc-4855-ada1-a2272c53f4b3")
	if err != nil {
		t.Error(err)
		return
	}
	for s := range ch {
		t.Log(s)
	}
}
