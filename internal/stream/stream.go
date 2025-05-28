package stream

import (
	"bufio"
	"io"
	"sync"
)

// LineHandler is a function type that processes a single log line.
type LineHandler func(line []byte)

// StreamErrHandler is a function type that handles errors from the stream processor itself.
type StreamErrHandler func(err error)

// Stream is an io.Writer that processes incoming data line by line
// and invokes a callback for each line.
type Stream struct {
	lineHandler LineHandler
	errHandler  StreamErrHandler
	pipeWriter  *io.PipeWriter
	scanWg      sync.WaitGroup
	closed      bool
	mu          sync.Mutex
}

func New(lineHandler LineHandler, errHandler StreamErrHandler) *Stream {
	pr, pw := io.Pipe()
	s := &Stream{
		lineHandler: lineHandler,
		errHandler:  errHandler,
		pipeWriter:  pw,
	}

	s.scanWg.Add(1)
	go func() {
		defer s.scanWg.Done()
		defer pr.Close()
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			data := make([]byte, len(scanner.Bytes()))
			copy(data, scanner.Bytes())
			s.lineHandler(data)
		}
		if err := scanner.Err(); err != nil {
			s.errHandler(err)
		}
	}()

	return s
}

// Write implements io.Writer. It writes data to the internal pipe.
func (s *Stream) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return 0, io.ErrClosedPipe
	}
	s.mu.Unlock()
	return s.pipeWriter.Write(p)
}

// Close signals that no more data will be written.
// It closes the writer end of the pipe and waits for the scanner goroutine
// to process all buffered data and for the lineHandler to be called for all lines.
func (s *Stream) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	err := s.pipeWriter.Close() // this will cause the scanner to eventually EOF
	s.scanWg.Wait()             // wait for the scanning goroutine to complete
	return err
}
