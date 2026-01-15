package record

import (
	"bufio"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FileRecorder store session records to local file with optional gzip compression
type FileRecorder struct {
	writers  map[string]io.Writer
	closers  map[string][]io.Closer
	mu       sync.Mutex
	compress bool
	base     string
}

// connId remove prefixed "$"
func (f *FileRecorder) connId(connId string) string {
	return strings.TrimPrefix(connId, "$")
}

func (f *FileRecorder) filename(connId string) string {
	filename := filepath.Join(f.base, f.connId(connId))
	if f.compress {
		filename += ".gz"
	}
	return filename
}

func (f *FileRecorder) open(connId string) (io.WriteCloser, error) {
	var w io.WriteCloser
	var err error
	connId = f.connId(connId)
	w, err = os.Create(f.filename(connId))
	if err != nil {
		return nil, err
	}
	fileCloser := w
	if f.compress {
		w = gzip.NewWriter(w)
		// close gzip first
		f.closers[connId] = []io.Closer{w, fileCloser}
	} else {
		f.closers[connId] = []io.Closer{fileCloser}
	}
	f.writers[connId] = w
	return w, nil
}

func (f *FileRecorder) Open(connId string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, _ = f.open(connId)
}

func (f *FileRecorder) Close(connId string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	connId = f.connId(connId)
	if closers, exists := f.closers[connId]; exists {
		for _, c := range closers {
			_ = c.Close()
		}
		delete(f.writers, connId)
		delete(f.closers, connId)
	}
}

func (f *FileRecorder) Write(connId string, data []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	connId = f.connId(connId)
	var err error
	w, exists := f.writers[connId]
	if !exists {
		if w, err = f.open(connId); err != nil {
			return
		}
	}
	_, _ = w.Write(data)
	if gw, ok := w.(*gzip.Writer); ok {
		_ = gw.Flush()
	}
}

func (f *FileRecorder) Replay(ctx context.Context, connId string) (chan string, error) {
	filename := f.filename(connId)
	var r io.ReadCloser
	var closers []io.Closer
	var err error
	r, err = os.Open(filename)
	if err != nil {
		return nil, err
	}
	closers = []io.Closer{r}
	if f.compress {
		var gr io.ReadCloser
		gr, err = gzip.NewReader(r)
		if err != nil {
			_ = r.Close()
			return nil, err
		} else {
			// close gzip first
			closers = []io.Closer{gr, r}
			r = gr
		}
	}
	ch := make(chan string, 64)

	go func() {
		defer func() {
			close(ch)
			for _, c := range closers {
				_ = c.Close()
			}
		}()
		br := bufio.NewReader(r)
		var line string
		for {
			select {
			case <-ctx.Done():
				return
			default:
				line, err = br.ReadString(';')
				if err != nil {
					return
				}
				ch <- line
			}
		}
	}()
	return ch, nil
}

func NewFileRecorder(base string, compress bool) *FileRecorder {
	_ = os.MkdirAll(base, 0755)
	return &FileRecorder{
		writers:  make(map[string]io.Writer),
		closers:  make(map[string][]io.Closer),
		compress: compress,
		base:     base,
	}
}
