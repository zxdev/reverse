package reverse

import (
	"bytes"
	"errors"
	"io"
)

const (
	// DefaultChunkSize is the default value for the option ChunkSize
	DefaultChunkSize = 1024 // 1K

	// DefaultBufferSize is the default value for the option BufferSize which
	// also limits the max line size that can be processed before ErrOverflow
	DefaultBufferSize = 1 << 20 // 1MB

	// DefaultIgnoreEmptyLine is the default behavior for option IgnoreEmptyLine
	DefaultIgnoreEmptyLine = true
)

var (
	// ErrOverflow reports that the line is longer than the internal read buffer
	// and allows the error state condition Err() to report it was not io.EOF
	ErrOverflow = errors.New("line overflow")
)

// Scanner is a reverse (LIFO) tail-to-head line scanner
type Scanner struct {
	r      io.ReaderAt // r is the input to read from
	last   int         // last is the index of the last read line
	offset int         // offset is the index within buf1
	bn     int         // block buffer size
	err    error       // err is the encountered error (if any)
	opt    Options     // opt are the configuration Options
	line   []byte      // line stores the extraction \n segment bytes
	buf1   []byte      // buf1 stores the read but not yet returned chunk bytes
	buf2   []byte      // buf2 stores the retrograde read block chunk bytes
}

// Options configures scanner parameters
type Options struct {
	// ChunkSize configures the size of the byte chunk that is read from the input
	ChunkSize int

	// BufferSize configures the maximum byte size limit of the internal buffer
	// Note: This also indirectly limits the max line size
	BufferSize int

	// IgnoreEmptyLine will skip returning empty line content until io.EOF
	IgnoreEmptyLine bool
}

// NewScanner returns a new reverse (LIFO) tail-to-head line Scanner, because the scanner
// reads retrograde lines (tail-to-head toward toward zero) the size parameter must be
// non-zero or a nil Scanner will be returned, nil for Options applies default values
func NewScanner(r io.ReaderAt, size int, opt *Options) *Scanner {

	if size < 1 { // since io.ReaderAt interface has no Len() method it makes size a requirement
		return nil
	}

	if opt == nil { // use default option values for configuration
		opt = &Options{DefaultChunkSize, DefaultBufferSize, DefaultIgnoreEmptyLine}
	}

	if opt.BufferSize < opt.ChunkSize { // sanity check
		opt.BufferSize = opt.ChunkSize
	}

	return &Scanner{r: r, last: size, opt: *opt}
}

// Scan reads the next line until the head is reached which signals termination
// by returning false with an io.EOF (or reverse.ErrOverflow) for any error state,
// and once an error state is encountered it continues to be returned until reset
func (s *Scanner) Scan() bool {

	if s.err != nil { // failover; prior state
		return false
	}

	for {

		// copy byte segment to line and truncate
		s.offset = bytes.LastIndexByte(s.buf1, '\n')
		if s.offset >= 0 {
			s.line, s.buf1 = noCR(s.buf1[s.offset+1:]), s.buf1[:s.offset]
			if s.opt.IgnoreEmptyLine && len(s.line) == 0 {
				continue
			}
			return true
		}

		// no more bytes to process; head reached
		if s.last == 0 {
			s.line = noCR(s.buf1)
			s.err = io.EOF // set io.EOF status
			return true
		}

		// refill the block buffer with chunks until \n is found and
		// then consume the chunk buffer extracting lines from chunk

		// check chunk size and n index; reset on chunk underflow
		if s.opt.ChunkSize > s.last {
			s.opt.ChunkSize = s.last
		}
		s.last -= s.opt.ChunkSize // walk n index

		// check block buffer size/growth for overflow
		s.bn = s.opt.ChunkSize + len(s.buf1)
		if s.bn > s.opt.BufferSize {
			s.line = []byte{}
			s.err = ErrOverflow // yikes, that a bit much
			return false
		}

		// read block buffer sizer; once grown does not shrink
		if cap(s.buf2) >= s.bn {
			s.buf2 = s.buf2[:s.opt.ChunkSize]
		} else {
			s.buf2 = make([]byte, s.opt.ChunkSize, s.bn)
		}

		// io.ReadAt attempts to read a full block buffer
		_, s.err = s.r.ReadAt(s.buf2, int64(s.last))
		switch s.err {
		case nil:
			s.buf1, s.buf2 = append(s.buf2, s.buf1...), s.buf1
		case io.EOF:
			s.line = noCR(s.buf1)
			return true
		default: // yikes, unexpected error state
			return false
		}

	}
}

// noCR will remove a terminal \r from the byte slice
func noCR(b []byte) []byte {

	if len(b) > 0 && b[len(b)-1] == '\r' {
		return b[0 : len(b)-1]
	}

	return b
}

// Text returns the string form of the current line read from data source
func (s *Scanner) Text() string { return string(s.line) }

// Bytes returns the byte slice form of the current line read from data source
func (s *Scanner) Bytes() []byte { return s.line }

// Len returns the len of the current line waiting in the extraction buffer
func (s *Scanner) Len() int { return len(s.line) }

// IsEmpty returns a boolean reporting if the current line is empty
func (s *Scanner) IsEmpty() bool { return len(s.line) > 0 }

// Err returns the most recent Scan() error state encountered (if any)
func (s *Scanner) Err() error { return s.err }
