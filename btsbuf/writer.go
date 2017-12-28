package btsbuf

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type (
	// Writer allows to split provided slice of bytes on to chunks
	Writer struct {
		buf     []byte
		clsdPos int
		offs    int
		ext     bool
	}
)

// Reset Initializes the writer by provided slice. The extendable flag
// allows the buffer will be re-allocated in case of not enough space in Allocate
// method
func (bbw *Writer) Reset(buf []byte, extendable bool) {
	bbw.buf = buf
	if cap(bbw.buf) > 0 {
		bbw.buf = bbw.buf[:cap(bbw.buf)]
	}
	bbw.offs = 0
	bbw.clsdPos = -1
	bbw.ext = extendable
}

// Allocate reserves ln bytes for data and writes its size before the chunk.
// The method returns slice of allocated bytes. In case of the allocation is not
// possible, it returns error.
//
// extend param defines the caller desire to extend the buffer if the size is
// not big enough or return a error, without extension. If the buffer is not
// extendable, it will return an error in any case when the buffers size is insufficient.
// If the buffer is extendable, it will try to re-allocate the buffer only if the
// extend == true
func (bbw *Writer) Allocate(ln int, extend bool) ([]byte, error) {
	if bbw.clsdPos >= 0 {
		return nil, errors.New("the writer already closed")
	}
	rest := len(bbw.buf) - bbw.offs - ln - 4
	if rest < 0 && (!extend || !bbw.extend(ln+4)) {
		return nil, errors.New(fmt.Sprintf("not enough space - available %d, but needed %d", len(bbw.buf)-bbw.offs, ln+4))
	}
	binary.BigEndian.PutUint32(bbw.buf[bbw.offs:], uint32(ln))
	bbw.offs += ln + 4
	return bbw.buf[bbw.offs-ln : bbw.offs], nil
}

// extend tries to extend the buffer if it is possbile to be able store at least
// ln bytes (including its size)
func (bbw *Writer) extend(ln int) bool {
	if !bbw.ext {
		return false
	}
	nsz := len(bbw.buf) * 3 / 2
	if bbw.offs+ln > nsz {
		nsz = len(bbw.buf) + ln*2
	}
	nb := make([]byte, nsz)
	copy(nb, bbw.buf)
	bbw.buf = nb
	return true
}

// Buf() returns the buffer underlying the writer
func (bbw *Writer) Buf() []byte {
	return bbw.buf
}

// Close puts EOF marker or completes the writing process. Consequentive
// Allocate() calls will return errors. The Close() method returns slice of
// bytes with allocated chunks (without the close marker). The method places the
// close marker into the original slice of bytes for addressing the case
// when it is not completely used so the original bbw.buf can be used for iteration
// with no problems
func (bbw *Writer) Close() ([]byte, error) {
	if bbw.clsdPos >= 0 {
		return bbw.buf[:bbw.clsdPos], nil
	}
	if len(bbw.buf)-bbw.offs < 4 {
		bbw.clsdPos = bbw.offs
		bbw.buf = bbw.buf[:bbw.clsdPos]
		return bbw.buf, nil
	}
	binary.BigEndian.PutUint32(bbw.buf[bbw.offs:], uint32(0xFFFFFFFF))
	bbw.clsdPos = bbw.offs
	bbw.offs = len(bbw.buf)
	return bbw.buf[:bbw.clsdPos], nil
}
