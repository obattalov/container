package btsbuf

import (
	"encoding/binary"
	"errors"
)

type (
	// Reader can iterate over slice of bytes which contains some
	// chunks of bytes and return them. It implements btsbuf.Iterator interface
	Reader struct {
		buf  []byte
		cur  []byte
		offs int
		cnt  int
	}
)

// check will make a check if the buf is properly organized and iteratable
func check(buf []byte) (int, error) {
	cnt := 0
	offs := 0
	for offs < len(buf) {
		if offs > len(buf)-4 {
			return -1, errors.New("wrong offset")
		}

		ln := binary.BigEndian.Uint32(buf[offs:])
		if ln == 0xFFFFFFFF {
			offs = len(buf)
			break
		}
		cnt++
		offs += int(ln) + 4
	}

	if offs == len(buf) {
		return cnt, nil
	}

	return 0, errors.New("broken structure")
}

// Reset initializes Reader and checks whether the provided buf
// is properly organized. Returns an error if the structure is incorrect.
// If the method returns an error End() will return true, Get() will return nil
// and the Next() call will not have any effect
func (bbi *Reader) Reset(buf []byte) error {
	cnt, err := check(buf)
	bbi.buf = nil
	bbi.cur = nil
	bbi.offs = 0
	bbi.cnt = cnt

	if err != nil {
		return err
	}
	bbi.buf = buf
	bbi.fillCur()
	return nil
}

func (bbi *Reader) fillCur() {
	if bbi.offs < len(bbi.buf) {
		ln := binary.BigEndian.Uint32(bbi.buf[bbi.offs:])
		if ln != 0xFFFFFFFF {
			offs := bbi.offs + 4
			bbi.cur = bbi.buf[offs : offs+int(ln)]
			return
		}
	}
	bbi.cur = nil
	bbi.offs = len(bbi.buf)
}

// End returns true if the iterator reaches the end and doesn't have any data,
// Get will return nil when End() returns true
func (bbi *Reader) End() bool {
	return bbi.offs >= len(bbi.buf)
}

// Get returns current element. Get returns first element after initialization
func (bbi *Reader) Get() []byte {
	return bbi.cur
}

// Next switches to the next element. Get() allows to access to the current one.
// Has no effect if the end is reached
func (bbi *Reader) Next() {
	bbi.offs += 4 + len(bbi.cur)
	bbi.fillCur()
}

// Buf returns underlying buffer
func (bbi *Reader) Buf() []byte {
	return bbi.buf
}

// Len returns number of records found in the buf
func (bbi *Reader) Len() int {
	return bbi.cnt
}
