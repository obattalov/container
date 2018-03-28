package btsbuf

type (
	Iterator interface {
		End() bool
		Get() []byte
		Next()
	}

	Concatenator []byte 
	
	str_arr_it []string
)

// Wraps slice of strings and provide an iterator over it. Can be used for tests
func Wrap(ss []string) Iterator {
	sai := str_arr_it(ss)
	return &sai
}

func (sai *str_arr_it) End() bool {
	return len(*sai) == 0
}

func (sai *str_arr_it) Get() []byte {
	if len(*sai) > 0 {
		return []byte((*sai)[0])
	}
	return nil
}

func (sai *str_arr_it) Next() {
	if len(*sai) > 0 {
		*sai = (*sai)[1:]
	}
}

func (cn *Concatenator) Write(s []byte) {
	cnt := append(*cn, s...)
	cn = &cnt
}