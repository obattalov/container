package btsbuf

import (
	"testing"
)

func TestStrArrIt(t *testing.T) {
	sa := []string{"a", "b"}
	sai := str_arr_it(sa)
	var it Iterator
	it = &sai

	if string(sai.Get()) != "a" || sai.End() {
		t.Fatal("Expecting a and not End()")
	}
	sai.Next()
	if string(sai.Get()) != "b" || sai.End() {
		t.Fatal("Expecting b and not End()")
	}
	for i := 0; i < 10; i++ {
		it.Next()
		if !sai.End() || sai.Get() != nil {
			t.Fatal("should be end now")
		}
	}
}
