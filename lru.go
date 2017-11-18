package container

import (
	"time"
)

type (
	element struct {
		prev *element
		next *element
		v    Value
	}

	Value struct {
		size int64
		ts   time.Time
		key  interface{}
		val  interface{}
	}

	// Lru is "least recently used" container, which allows to control the
	// element by size, time of touch, or both. It keeps recently used items
	// near the top of cache.
	Lru struct {
		head    *element
		pool    *element
		kvMap   map[interface{}]*element
		size    int64
		maxSize int64
		maxDur  time.Duration
		cback   LruDeleteCallback
	}

	LruDeleteCallback func(k, v interface{})
	LruCallback       func(k, v interface{}) bool
)

var nilTime = time.Time{}

// NewLru creates new Lru container with maximum size maxSize, and maximum
// time 'to' an element can stay in the cache. cback is a function which is
// invoked when an element is pulled out of the cache. It can be nil
//
// Timeout to could be 0, what means don't use it at all
func NewLru(maxSize int64, to time.Duration, cback LruDeleteCallback) *Lru {
	l := new(Lru)
	l.kvMap = make(map[interface{}]*element)
	l.maxSize = maxSize
	l.maxDur = to
	l.cback = cback
	return l
}

func (l *Lru) Put(k, v interface{}, size int64) {
	e, ok := l.kvMap[k]
	if ok {
		l.delete(e, true)
	}

	tm := l.SweepByTime()
	l.sweepBySize(size)

	if l.pool != nil {
		e = l.pool
		l.pool = nil
	} else {
		e = new(element)
	}
	e.v.key = k
	e.v.val = v
	e.v.ts = tm
	e.v.size = size
	l.head = addToHead(l.head, e)
	l.kvMap[k] = e
	l.size += size
}

func (l *Lru) Get(k interface{}) *Value {
	ts := l.SweepByTime()
	e, ok := l.kvMap[k]
	if ok {
		l.head = removeFromList(l.head, e)
		l.head = addToHead(l.head, e)
		e.v.ts = ts
		return &e.v
	}
	return nil
}

func (l *Lru) Peek(k interface{}) *Value {
	l.SweepByTime()
	e, ok := l.kvMap[k]
	if ok {
		return &e.v
	}
	return nil
}

func (l *Lru) Delete(k interface{}) {
	l.SweepByTime()
	e, ok := l.kvMap[k]
	if ok {
		l.delete(e, true)
	}
}

func (l *Lru) DeleteNoCallback(k interface{}) {
	l.SweepByTime()
	e, ok := l.kvMap[k]
	if ok {
		l.delete(e, false)
	}
}

// Iterate is the container visitor which walks over the elements in LRU order.
// It calls f() for every key-value pair and continues until the f() returns true,
// or all elements are visited.
func (l *Lru) Iterate(f LruCallback) {
	h := l.head
	for h != nil {
		if !f(h.v.key, h.v.val) {
			break
		}
		h = h.next
		if h == l.head {
			break
		}
	}
}

func (l *Lru) Size() int64 {
	return l.size
}

func (l *Lru) Len() int {
	return len(l.kvMap)
}

func (l *Lru) SweepByTime() time.Time {
	if l.maxDur == 0 {
		return nilTime
	}
	tm := time.Now()
	for l.head != nil && tm.Sub(l.head.prev.v.ts) > l.maxDur {
		last := l.head.prev
		l.delete(last, true)
	}
	return tm
}

func (l *Lru) sweepBySize(addSize int64) {
	for l.head != nil && l.size+addSize > l.maxSize {
		last := l.head.prev
		l.delete(last, true)
	}
}

func (l *Lru) delete(e *element, cb bool) {
	l.head = removeFromList(l.head, e)
	l.size -= e.v.size
	l.pool = e
	delete(l.kvMap, e.v.key)
	if cb && l.cback != nil {
		l.cback(e.v.key, e.v.val)
	}
}

func removeFromList(head, e *element) *element {
	if e == head && head.next == head {
		head = nil
	}
	e.prev.next = e.next
	e.next.prev = e.prev
	if e == head {
		head = e.next
	}
	return head
}

// add n to list with head and returns new head
func addToHead(head *element, n *element) *element {
	if n == nil {
		return head
	}
	if head == nil {
		n.prev = n
		n.next = n
		return n
	}
	n.next = head
	n.prev = head.prev
	head.prev = n
	n.prev.next = n
	return n
}

func (v *Value) Key() interface{} {
	return v.key
}

func (v *Value) Val() interface{} {
	return v.val
}

func (v *Value) Size() int64 {
	return v.size
}

func (v *Value) TouchedAt() time.Time {
	return v.ts
}
