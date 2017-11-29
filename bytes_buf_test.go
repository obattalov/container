package container

import (
	"encoding/binary"
	"math/rand"
	"reflect"
	"testing"
)

func TestBtsBufWriterEmpty(t *testing.T) {
	var bbw BtsBufWriter
	var buf [100]byte
	bbw.Reset(buf[:], false)
	bf, err := bbw.Close()
	if err != nil || len(bf) != 0 {
		t.Fatal("should be empty")
	}

	bf, err = bbw.Close()
	if err != nil || len(bf) != 0 {
		t.Fatal("should be empty from second attempt")
	}
}

func TestBtsBufWriterAllocate(t *testing.T) {
	var bbw BtsBufWriter
	var buf [100]byte
	bbw.Reset(buf[:], false)
	bf, err := bbw.Allocate(20)
	if bf == nil || len(bf) != 20 || err != nil {
		t.Fatal("Should be able to allocate 20 bytes err=", err)
	}

	bf, err = bbw.Allocate(20)
	if bf == nil || len(bf) != 20 || err != nil {
		t.Fatal("Should be able to allocate 20 bytes again err=", err)
	}

	bf, err = bbw.Allocate(60)
	if err == nil {
		t.Fatal("Should not be able to allocate 60 bytes err=", err)
	}

	if bbw.offs != 48 {
		t.Fatal("expecting offs=48, but it is ", bbw.offs)
	}

	bf, err = bbw.Close()
	if err != nil || len(bf) != 48 {
		t.Fatal("should be 48 bytes lenght")
	}

	v := binary.BigEndian.Uint32(buf[48:])
	if v != 0xFFFFFFFF {
		t.Fatal("No marker")
	}
}

func TestBtsBufWriterAllocateExt(t *testing.T) {
	var bbw BtsBufWriter
	bbw.Reset(nil, true)
	for i := 0; i < 100; i++ {
		_, err := bbw.Allocate(0)
		if err != nil {
			t.Fatal("Should be extendable err=", err)
		}
	}
	if len(bbw.Buf()) < 400 {
		t.Fatal("Expecting at least 400 bytes in length, but it is ", len(bbw.Buf()))
	}

	bbw.Reset(nil, true)
	for i := 0; i < 100; i++ {
		_, err := bbw.Allocate(100)
		if err != nil {
			t.Fatal("Should be extendable err=", err)
		}
	}
	if len(bbw.Buf()) < 10400 {
		t.Fatal("Expecting at least 10400 bytes in length, but it is ", len(bbw.Buf()))
	}
}

func TestInsufficientAllocate(t *testing.T) {
	var bbw BtsBufWriter
	var buf [12]byte
	bbw.Reset(buf[:], false)
	bf, err := bbw.Allocate(4)
	if bf == nil || len(bf) != 4 || err != nil {
		t.Fatal("Should be able to allocate 4 bytes err=", err)
	}

	bbw.Reset(buf[:], false)
	bf, err = bbw.Allocate(2)
	if bf == nil || len(bf) != 2 || err != nil {
		t.Fatal("Should be able to allocate 2 bytes err=", err)
	}

	bbw.Reset(buf[:], false)
	bf, err = bbw.Allocate(8)
	if bf == nil || len(bf) != 8 || err != nil {
		t.Fatal("Should be able to allocate 8 bytes err=", err)
	}

	bbw.Reset(buf[:], false)
	bf, err = bbw.Allocate(5)
	if bf != nil || err == nil {
		t.Fatal("Should not be able to allocate 5 bytes err=", err)
	}
	bbw.Reset(buf[:], false)
	bf, err = bbw.Allocate(25)
	if bf != nil || err == nil {
		t.Fatal("Should not be able to allocate 25 bytes err=", err)
	}
}

func TestEmptyBufIterator(t *testing.T) {
	var bbi BtsBufIterator
	if !bbi.End() || bbi.Get() != nil {
		t.Fatal("Empty iterator check fail")
	}
	bbi.Next()
	if !bbi.End() || bbi.Get() != nil {
		t.Fatal("Empty iterator check fail (2)")
	}

	var buf []byte
	err := bbi.Reset(buf)
	if err != nil {
		t.Fatal("should not be problem with empty reset")
	}

	bbi.Next()
	if !bbi.End() || bbi.Get() != nil {
		t.Fatal("Empty iterator check fail (3)")
	}
}

func TestResetBtsBufIterator(t *testing.T) {
	var bbi BtsBufIterator
	var buf [20]byte
	err := bbi.Reset(buf[:4])
	if err != nil {
		t.Fatal("should not be problem with empty reset, err=", err)
	}
	err = bbi.Reset(buf[:])
	if err != nil {
		t.Fatal("should not be problem with empty reset(2), err=", err)
	}

	err = bbi.Reset(buf[:8])
	if err != nil {
		t.Fatal("should not be problem with empty reset(3), err=", err)
	}

	err = bbi.Reset(buf[:7])
	if err == nil {
		t.Fatal("should be problem with empty reset, but it is not")
	}
	err = bbi.Reset(buf[:2])
	if err == nil {
		t.Fatal("should be problem with empty reset, but it is not")
	}

	binary.BigEndian.PutUint32(buf[:], 3)
	binary.BigEndian.PutUint32(buf[7:], 0xFFFFFFFF)
	err = bbi.Reset(buf[:])
	if err != nil {
		t.Fatal("should not be problem with empty reset(4), err=", err)
	}

	binary.BigEndian.PutUint32(buf[7:], 0x12FFFFFF)
	err = bbi.Reset(buf[:])
	if err == nil {
		t.Fatal("should be problem with empty reset, but it is not (2)")
	}
}

func TestBtsBufIterator(t *testing.T) {
	var bbw BtsBufWriter
	var src [92]byte

	rand.Read(src[:])
	bbw.Reset(nil, true)
	var sz int
	for offs := 0; offs < len(src); offs += sz {
		sz = 20 + offs/10
		bw, err := bbw.Allocate(sz)
		if len(bw) != sz || err != nil {
			t.Fatal("Something goes wrong with allocation err=", err)
		}
		copy(bw, src[offs:])
	}

	bw, err := bbw.Close()
	if len(bw) != 108 || err != nil {
		t.Fatal("Something goes wrong with writing")
	}

	var bbi BtsBufIterator
	err = bbi.Reset(bbw.Buf())
	if err != nil {
		t.Fatal("Expecting no problems with the dst buf, but err=", err)
	}

	if bbi.Len() != 4 {
		t.Fatal("Expected len is 4, but got ", bbi.Len())
	}

	offs := 0
	for !bbi.End() {
		bw = bbi.Get()
		sz = 20 + offs/10
		if len(bw) != sz {
			t.Fatal("Wrong buf size ", len(bw), ", but expected ", sz)
		}
		if !reflect.DeepEqual(bw, src[offs:offs+sz]) {
			t.Fatal("wrong data read. Expected size sz=", sz, ", offs=", offs)
		}
		offs += sz
		bbi.Next()
	}
}

func TestBtsBufIteratorEven(t *testing.T) {
	var bbw BtsBufWriter
	var dst [16]byte
	var src [8]byte

	rand.Read(src[:])
	bbw.Reset(dst[:], false)
	for offs := 0; offs < len(src); offs += 4 {
		bw, err := bbw.Allocate(4)
		if len(bw) != 4 || err != nil {
			t.Fatal("Something goes wrong with allocation err=", err)
		}
		copy(bw, src[offs:])
	}

	bw, err := bbw.Close()
	if len(bw) != 16 || err != nil {
		t.Fatal("Something goes wrong with writing")
	}

	var bbi BtsBufIterator
	err = bbi.Reset(dst[:])
	if err != nil {
		t.Fatal("Expecting no problems with the dst buf, but err=", err)
	}
	if bbi.Len() != 2 {
		t.Fatal("Expected len is 2, but got ", bbi.Len())
	}

	offs := 0
	its := 0
	for !bbi.End() {
		its++
		bw = bbi.Get()
		if len(bw) != 4 {
			t.Fatal("Wrong buf size ", len(bw), ", but expected 4")
		}
		if !reflect.DeepEqual(bw, src[offs:offs+4]) {
			t.Fatal("wrong data read. Expected size 4, offs=", offs)
		}
		offs += 4
		bbi.Next()
	}
	if its != 2 {
		t.Fatal("should be iterate over 2 elems")
	}
}
