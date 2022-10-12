package log

import (
	"context"
	"io"
	"path/filepath"
	"sync"
	"time"

	"github.com/peterlearn/kratos/pkg/log/internal/filewriter"
)

// level idx
const (
	_debugIdx = iota
	_infoIdx
	_warnIdx
	_errorIdx
	_totalIdx
)

var _fileNames = map[int]string{
	_debugIdx: "debug.log",
	_infoIdx:  "info.log",
	_warnIdx:  "warning.log",
	_errorIdx: "error.log",
}

// FileHandler .
type FileHandler struct {
	render Render
	fws    [_totalIdx]*filewriter.FileWriter
	pool   sync.Pool
}

const defaultLogFilePattern = "[%D %T] [%L] [%S] %q %c %l %u %r %a %M"

// NewFile crete a file logger.
func NewFile(dir string, bufferSize, rotateSize int64, maxLogFile int) *FileHandler {
	// new info writer
	newWriter := func(name string) *filewriter.FileWriter {
		var options []filewriter.Option
		if rotateSize > 0 {
			options = append(options, filewriter.MaxSize(rotateSize))
		}
		if maxLogFile > 0 {
			options = append(options, filewriter.MaxFile(maxLogFile))
		}
		w, err := filewriter.New(filepath.Join(dir, name), options...)
		if err != nil {
			panic(err)
		}
		return w
	}
	handler := &FileHandler{
		render: newPatternRender(defaultLogFilePattern),
		pool:   sync.Pool{New: func() interface{} { return make(map[string]interface{}, 25) }},
	}
	for idx, name := range _fileNames {
		handler.fws[idx] = newWriter(name)
	}
	return handler
}

// Log loggint to file .
func (h *FileHandler) Log(ctx context.Context, lv Level, args ...D) {
	if int32(lv) < c.V {
		return
	}
	d := h.pool.Get().(map[string]interface{})
	defer func() {
		d = make(map[string]interface{}, 25)
		h.pool.Put(d)
	}()
	toMap(d, args...)
	// add extra fields
	addExtraField(ctx, d)
	d[_time] = time.Now().Format(_timeFormat)
	var w io.Writer
	switch lv {
	case _debugLevel:
		w = h.fws[_debugIdx]
	case _warnLevel:
		w = h.fws[_warnIdx]
	case _errorLevel:
		w = h.fws[_errorIdx]
	default:
		w = h.fws[_infoIdx]
	}
	h.render.Render(w, d)
	w.Write([]byte("\n"))
}

// Close log handler
func (h *FileHandler) Close() error {
	for _, fw := range h.fws {
		// ignored error
		fw.Close()
	}
	return nil
}

// SetFormat set log format
func (h *FileHandler) SetFormat(format string) {
	h.render = newPatternRender(format)
}
