package particle

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"cleanroomorcontrol/internal/model"
)

type Opener interface {
	Open(path string) (io.WriteCloser, error)
}

type FileOpener struct{}

func (FileOpener) Open(path string) (io.WriteCloser, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}

type RecordWriter struct {
	mu      sync.Mutex
	dir     string
	opener  Opener
	open    int
	closed  int
	bytes   map[model.RoomID]int
	handles map[model.RoomID]io.WriteCloser
	buffers map[model.RoomID]*bufio.Writer
}

func NewRecordWriter(dir string, opener Opener) *RecordWriter {
	return &RecordWriter{
		dir:     dir,
		opener:  opener,
		bytes:   make(map[model.RoomID]int),
		handles: make(map[model.RoomID]io.WriteCloser),
		buffers: make(map[model.RoomID]*bufio.Writer),
	}
}

func (w *RecordWriter) Append(room model.RoomID, sample model.ParticleSample) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.opener == nil {
		w.opener = FileOpener{}
	}
	writer, ok := w.buffers[room]
	if !ok {
		handle, err := w.opener.Open(filepath.Join(w.dir, string(room)+".samples"))
		if err != nil {
			return err
		}
		w.open++
		writer = bufio.NewWriter(handle)
		w.handles[room] = handle
		w.buffers[room] = writer
	}
	line := fmt.Sprintf("%s|%s|%d|%.3f|%s\n", sample.Point, room, sample.Count, sample.Volume, sample.At.Format(time.RFC3339))
	if _, err := writer.WriteString(line); err != nil {
		return err
	}
	w.bytes[room] += len(line)
	if w.bytes[room] > 65536 {
		return writer.Flush()
	}
	return nil
}

func (w *RecordWriter) Flush(room model.RoomID) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	writer, ok := w.buffers[room]
	if !ok {
		return nil
	}
	return writer.Flush()
}

func (w *RecordWriter) Rotate(room model.RoomID) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if writer, ok := w.buffers[room]; ok {
		if err := writer.Flush(); err != nil {
			return err
		}
		delete(w.buffers, room)
	}
	if handle, ok := w.handles[room]; ok {
		if err := handle.Close(); err != nil {
			w.closed++
			delete(w.handles, room)
			return err
		}
		w.closed++
		delete(w.handles, room)
	}
	delete(w.bytes, room)
	return nil
}

func (w *RecordWriter) Counters() (int, int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.open, w.closed
}
