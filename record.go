package guac

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FileRecorder store session records to local file with optional gzip compression
type FileRecorder struct {
	writers  map[string]io.WriteCloser
	mu       sync.Mutex
	compress bool
	base     string
}

// connId remove prefixed "$"
func (f *FileRecorder) connId(connId string) string {
	return strings.TrimPrefix(connId, "$")
}

func (f *FileRecorder) openWriter(connId string) (io.WriteCloser, error) {
	var w io.WriteCloser
	var err error
	connId = f.connId(connId)
	filename := filepath.Join(f.base, connId)
	if f.compress {
		filename += ".gz"
	}
	w, err = os.Create(filename)
	if err != nil {
		return nil, err
	}
	if f.compress {
		w = gzip.NewWriter(w)
	}
	return w, nil
}

func (f *FileRecorder) OpenWriter(connId string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	connId = f.connId(connId)
	if _, exists := f.writers[connId]; exists {
		return
	}
	if w, err := f.openWriter(connId); err == nil {
		f.writers[connId] = w
	}
}

func (f *FileRecorder) Close(connId string) {
	f.mu.Lock()
	defer f.mu.Lock()
	connId = f.connId(connId)
	if w, exists := f.writers[connId]; exists {
		_ = w.Close()
		delete(f.writers, connId)
	}
}

func (f *FileRecorder) Write(connId string, data []byte) {
	f.mu.Lock()
	f.mu.Unlock()
	connId = f.connId(connId)
	var err error
	w, exists := f.writers[connId]
	if !exists {
		if w, err = f.openWriter(connId); err != nil {
			return
		}
	}
	_, _ = w.Write(data)
}

func NewFileRecorder(base string, compress bool) *FileRecorder {
	return &FileRecorder{
		writers:  make(map[string]io.WriteCloser),
		compress: compress,
		base:     base,
	}
}
