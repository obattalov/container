package container

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type (
	// BtsBufWriter allows to split provided slice of bytes on to chunks
	BtsBufWriter struct {
		buf     []byte
		clsdPos int
		offs    int
		ext     bool
	}

	// BtsBufIterator can iterate over slice of bytes which contains some
	// chunks of bytes and return them
	BtsBufIterator struct {
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

	return -1, errors.New("broken structure")
}

// Reset initializes BtsBufIterator and checks whether the provided buf
// is properly organized. Returns an error if the structure is incorrect.
// If the method returns an error End() will return true, Get() will return nil
// and the Next() call will not have any effect
func (bbi *BtsBufIterator) Reset(buf []byte) error {
	cnt, err := check(buf)
	if err != nil {
		return err
	}
	bbi.buf = buf
	bbi.cur = nil
	bbi.offs = 0
	bbi.cnt = cnt
	bbi.fillCur()
	return nil
}

func (bbi *BtsBufIterator) fillCur() {
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
func (bbi *BtsBufIterator) End() bool {
	return bbi.offs >= len(bbi.buf)
}

// Get returns current element. Get returns first element after initialization
func (bbi *BtsBufIterator) Get() []byte {
	return bbi.cur
}

// Next switches to the next element. Get() allows to access to the current one.
// Has no effect if the end is reached
func (bbi *BtsBufIterator) Next() {
	bbi.offs += 4 + len(bbi.cur)
	bbi.fillCur()
}

// Buf returns underlying buffer
func (bbi *BtsBufIterator) Buf() []byte {
	return bbi.buf
}

// Len returns number of records found in the buf
func (bbi *BtsBufIterator) Len() int {
	return bbi.cnt
}

// Reset Initializes the writer by provided slice. The extendable flag
// allows the buffer will be re-allocated in case of not enough space in Allocate
// method
func (bbw *BtsBufWriter) Reset(buf []byte, extendable bool) {
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
// If the buffer is extendable, it will try to re-allocate the buffer to make
// the request possible
func (bbw *BtsBufWriter) Allocate(ln int) ([]byte, error) {
	if bbw.clsdPos >= 0 {
		return nil, errors.New("the writer already closed")
	}
	rest := len(bbw.buf) - bbw.offs - ln - 4
	if rest < 0 && !bbw.extend(ln+4) {
		return nil, errors.New(fmt.Sprintf("not enough space - available %d, but needed %d", len(bbw.buf)-bbw.offs, ln+4))
	}
	binary.BigEndian.PutUint32(bbw.buf[bbw.offs:], uint32(ln))
	bbw.offs += ln + 4
	return bbw.buf[bbw.offs-ln : bbw.offs], nil
}

// extend tries to extend the buffer if it is possbile to be able store at least
// ln bytes (including its size)
func (bbw *BtsBufWriter) extend(ln int) bool {
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
func (bbw *BtsBufWriter) Buf() []byte {
	return bbw.buf
}

// Close puts EOF marker or completes the writing process. Consequentive
// Allocate() calls will return errors. The Close() method returns slice of
// bytes with allocated chunks (without the close marker). The method places the
// close marker into the original slice of bytes for addressing the case
// when it is not completely used so the original bbw.buf can be used for iteration
// with no problems
func (bbw *BtsBufWriter) Close() ([]byte, error) {
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
