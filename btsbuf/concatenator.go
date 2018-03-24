package btsbuf

type (
	Concatenator struct {
		buf       []byte
		bootstrap [128]byte
	}
)

func (c *Concatenator) Reset(buf []byte) {
	c.buf = buf
	if c.buf != nil {
		c.buf = c.buf[:0]
	}
}

func (c *Concatenator) Write(buf []byte) {
	sz := len(buf)
	if sz == 0 {
		return
	}

	c.grow(sz)
	pos := len(c.buf)
	c.buf = c.buf[:pos+sz]
	copy(c.buf[pos:pos+sz], buf)
}

func (c *Concatenator) Buf() []byte {
	return c.buf
}

func (c *Concatenator) CopyBuf() []byte {
	if len(c.buf) == 0 {
		return c.buf
	}
	res := make([]byte, len(c.buf))
	copy(res, c.buf)
	return res
}

func (c *Concatenator) grow(sz int) {
	if c.buf == nil {
		c.buf = c.bootstrap[:0]
	}

	if cap(c.buf)-len(c.buf) >= sz {
		return
	}

	nsz := cap(c.buf) * 3 / 2
	if len(c.buf)+sz > nsz {
		nsz = len(c.buf) + sz*2
	}

	nb := make([]byte, nsz)
	copy(nb, c.buf)
	c.buf = nb[:len(c.buf)]
}
