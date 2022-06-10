package stream

import (
	"sync"
)

type Ring struct {
	buf []int64
	pos int
	mut *sync.Mutex
}

func New(size int) *Ring {
	return &Ring{buf: make([]int64, size, size), pos: 0, mut: &sync.Mutex{}}
}

func (r *Ring) Put(cell int64) {
	r.mut.Lock()
	defer r.mut.Unlock()
	r.buf[r.pos] = cell
	r.pos = (r.pos + 1) % cap(r.buf)
}

func (r *Ring) GetAll() []int64 {
	r.mut.Lock()
	defer r.mut.Unlock()
	result := make([]int64, cap(r.buf))
	index := 0
	for i := r.pos; i < cap(r.buf); i++ {
		result[index] = r.buf[i]
		index++
	}
	for i := 0; i < r.pos; i++ {
		result[index] = r.buf[i]
		index++
	}
	return result
}
