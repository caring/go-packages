package writer

import (
	"bufio"
	"context"
	"io"
	"time"

	"go.uber.org/zap/zapcore"
)

type bufferWriterSyncer struct {
	ws           zapcore.WriteSyncer
	bufferWriter *bufio.Writer
	cancel       context.CancelFunc
}

// defaultBufferSize sizes the buffer associated with each WriterSync.
const (
	defaultBufferSize = 256 * 1024

	// defaultFlushInterval means the default flush interval
	defaultFlushInterval = 30 * time.Second
)

// Buffer wraps a WriteSyncer in a buffer to improve performance,
// if bufferSize = 0, we set it to defaultBufferSize
// if flushInterval = 0, we set it to defaultFlushInterval
func Buffer(writer io.Writer, bufferSize int, flushInterval time.Duration) (zapcore.WriteSyncer, io.Closer) {
	ctx, cancel := context.WithCancel(context.Background())

	if bufferSize == 0 {
		bufferSize = defaultBufferSize
	}

	if flushInterval == 0 {
		flushInterval = defaultFlushInterval
	}

	bw := &bufferWriterSyncer{
		bufferWriter: bufio.NewWriterSize(writer, bufferSize),
		cancel:       cancel,
	}

	// bufio is not goroutine safe, so add lock writer here
	ws := zapcore.Lock(bw)

	// flush buffer every interval
	// we do not need exit this goroutine explicitly
	go func() {
		select {
		case <-time.NewTicker(flushInterval).C:
			// the background goroutine just keep syncing
			// until the close func is called.
			_ = ws.Sync()
		case <-ctx.Done():
			return
		}
	}()

	return ws, bw
}

func (s *bufferWriterSyncer) Write(bs []byte) (int, error) {
	// there are some logic internal for bufio.Writer here:
	// 1. when the buffer is enough, data would not be flushed.
	// 2. when the buffer is not enough, data would be flushed as soon as the buffer fills up.
	// this would lead to log spliting, which is not acceptable for log collector
	// so we need to flush bufferWriter before writing the data into bufferWriter
	if len(bs) > s.bufferWriter.Available() && s.bufferWriter.Buffered() > 0 {
		err := s.bufferWriter.Flush()
		if err != nil {
			return 0, err
		}
	}

	return s.bufferWriter.Write(bs)
}

func (s *bufferWriterSyncer) Sync() error {
	return s.bufferWriter.Flush()
}

func (s *bufferWriterSyncer) Close() error {
	s.cancel()
	return s.ws.Sync()
}
