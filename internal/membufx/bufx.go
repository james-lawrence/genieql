package membufx

import (
	"errors"
	"io"
)

// Implemented interfaces
var _ io.ReadSeeker = &BufferReaderAt{}
var _ io.ReaderAt = &BufferReaderAt{}

var ErrInvalidSeek = errors.New("invalid seek")
var ErrInvalidParam = errors.New("invalid parameter")
var ErrPartialRead = errors.New("partial read")

func NewMemBuffer(b []byte) *BufferReaderAt {
	return &BufferReaderAt{
		Buf: b,
	}
}

type BufferReaderAt struct {
	Buf []byte
	off int64
}

func (b *BufferReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 || off > b.size() {
		return -1, ErrInvalidSeek
	}

	rd := copy(p, b.Buf[off:])
	if rd < len(p) {
		// docs say this must return an error if we read less than p
		if rd == int(b.size()-off) {
			return rd, io.EOF
		}
		return rd, ErrPartialRead
	}

	return rd, nil
}

func (b *BufferReaderAt) Read(p []byte) (n int, err error) {
	if b.off >= b.size() {
		return 0, io.EOF
	}

	rd := copy(p, b.Buf[b.off:])
	if rd > 0 {
		b.off += int64(rd)
	}

	return rd, nil
}

func (b *BufferReaderAt) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		if offset < 0 || offset > b.size() {
			return -1, ErrInvalidSeek
		}
		b.off = offset
	case io.SeekCurrent:
		newOff := b.off + offset
		if newOff < 0 || newOff > b.size() {
			return -1, ErrInvalidSeek
		}
		b.off = newOff
	case io.SeekEnd:
		newOff := b.size() + offset
		if newOff < 0 || newOff > b.size() {
			return -1, ErrInvalidSeek
		}
		b.off = newOff
	default:
		return -1, ErrInvalidParam
	}

	return b.off, nil
}

func (b *BufferReaderAt) size() int64 {
	return int64(len(b.Buf))
}
