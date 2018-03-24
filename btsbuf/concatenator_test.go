package btsbuf

import (
	"math/rand"
	"testing"

	"reflect"
)

func TestNil(t *testing.T) {
	var c Concatenator

	if c.Buf() != nil {
		t.Fatal("expecting nil c.Buf()")
	}

	c.Reset(nil)
	if c.Buf() != nil {
		t.Fatal("expecting nil c.Buf()")
	}

	c.Write(nil)
	if c.Buf() != nil {
		t.Fatal("expecting nil c.Buf()")
	}
}

func TestGrow(t *testing.T) {
	var c Concatenator
	b := make([]byte, cap(c.bootstrap))
	rand.Read(b)

	c.Write(b)
	if !reflect.DeepEqual(b, c.Buf()) || !reflect.DeepEqual(b, c.bootstrap[:]) {
		t.Fatal("Should be same what is written \nb=", b, "\nc.bootstrap=", c.bootstrap)
	}

	b1 := make([]byte, cap(c.bootstrap))
	rand.Read(b1)

	c.Write(b1)
	if !reflect.DeepEqual(b1, c.Buf()[len(b):]) || !reflect.DeepEqual(b, c.Buf()[:len(b)]) || len(b1)+len(b) != len(c.Buf()) {
		t.Fatal("Should be same what is written")
	}

	c.Reset(b)
	if len(c.Buf()) != 0 || cap(c.Buf()) != cap(b) {
		t.Fatal("wrong state after reset")
	}
}
