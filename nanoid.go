// Copyright (c) 2021 Handle
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package nanoid

import (
	"crypto/rand"
	"errors"
	"io"
	"math"
	"sync"
)

// DefaultAlphabet is a default alphabet.
const DefaultAlphabet = "ModuleSymbhasOwnPr-0123456789ABCDEFGHNRVfgctiUvz_KqYTJkLxpZXIjQW"

// MaxAlphabetSize represents the maximum number of characters in the alphabet.
const MaxAlphabetSize = 256

// bufSliceSize represents the size of the buffer byte slice. See the
// comments section of the bufSlicePool constants for details.
const bufSliceSize = 128

// bufSlicePool is a pool of buffer byte slices for reading Nano IDs.
//
// This pool reduces the number of heap memory allocations by recycling and
// reusing allocated byte slices.
//
// The length of the byte slices in the pool is fixed at 128. If the
// generated Nano ID size is less than 72, the required buffer size will
// not exceed 128 bytes. Generally, we do not generate Nano IDs larger than
// 72 in size, so we can allocate buffers from the pool to reduce the
// number of heap memory allocations.
//
// This pool will allocate a byte slice pointer.
var bufSlicePool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, bufSliceSize)
		return &buf
	},
}

// zeroString represents an empty string that does not contain any characters.
var zeroString string

// Reader is a reader for generating Nano IDs and has implemented the
// io.Reader interface.
type Reader struct {
	rander   io.Reader
	alphabet string
	mask     int
}

// initializeMask initializes the mask using the alphabet. The caller must
// ensure that the alphabet is valid, otherwise it will cause panic.
func (r *Reader) initializeMask() {
	if len(r.alphabet) == 0 {
		// This should not be the case.
		panic("nanoid: alphabet invalid")
	}

	size := len(r.alphabet)
	clz32 := 31 - (int(math.Log(float64(size-1|1))/math.Ln2) | 0) | 0
	r.mask = (2 << (31 - clz32)) - 1
}

// getRandomSize calculates and returns a random number size.
//
// The random number size is determined by the given Nano ID size and the
// reader's mask, alphabet size, and magic number 1.6.
//
// The caller must ensure that the given Nano ID size and the alphabet
// is valid, otherwise, it will cause panic.
func (r *Reader) getRandomSize(size int) int {
	if len(r.alphabet) == 0 {
		// This should not be the case.
		panic("nanoid: alphabet invalid")
	}
	if r.mask <= 0 {
		// This should not be the case.
		panic("nanoid: mask invalid")
	}
	x := (1.6 * float64(r.mask*size)) / float64(len(r.alphabet))
	return int(math.Ceil(x))
}

// Read generates a Nano ID using the alphabet and stores it to the given
// byte slice, then returns the actual number of bytes generated and any
// errors encountered.
func (r *Reader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, io.ErrShortBuffer
	}

	if len(r.alphabet) == 0 {
		_, err = r.rander.Read(p)
		if err != nil {
			return 0, err
		}
		for index := 0; index < len(p); index++ {
			p[index] = DefaultAlphabet[p[index]&63]
		}
		return len(p), nil
	}

	var random []byte
	size := r.getRandomSize(len(p))
	if size <= bufSliceSize {
		pointer := bufSlicePool.Get().(*[]byte)
		defer bufSlicePool.Put(pointer)
		random = (*pointer)[:size]
	} else {
		random = make([]byte, size)
	}

	for {
		_, err = r.rander.Read(random)
		if err != nil {
			return n, err
		}
		for index := range random {
			position := random[index] & byte(r.mask)
			if position < byte(len(r.alphabet)) {
				p[n] = r.alphabet[position]
				n++
				if n == len(p) {
					return n, err
				}
			}
		}
	}
}

// ReaderOption is an option for the Nano ID reader. See the comments
// section of the Reader structure for details.
type ReaderOption func(r *Reader) error

// WithAlphabet uses the given string as the alphabet used to generate
// the Nano ID.
//
// The number of characters contained in the given alphabet must be less
// than or equal to the value of the MaxAlphabetSize constant.
func WithAlphabet(alphabet string) ReaderOption {
	return func(r *Reader) error {
		if len(alphabet) == 0 {
			return errors.New("alphabet invalid")
		}
		if len(alphabet) > MaxAlphabetSize {
			return errors.New("alphabet is too long")
		}
		r.alphabet = alphabet
		r.initializeMask()
		return nil
	}
}

// WithRandReader uses the given reader as the random number reader used
// to generate the Nano ID.
func WithRandReader(rander io.Reader) ReaderOption {
	return func(r *Reader) error {
		if rander == nil {
			return errors.New("nil rand reader")
		}
		r.rander = rander
		return nil
	}
}

// NewReader creates and returns a reader for generating Nano IDs.
func NewReader(options ...ReaderOption) (r io.Reader, err error) {
	reader := &Reader{
		rander: rand.Reader,
	}
	for _, option := range options {
		err = option(reader)
		if err != nil {
			return nil, err
		}
	}
	return reader, nil
}

// defaultReader is a reader that is used by default to generate Nano IDs.
var defaultReader = &Reader{
	rander: rand.Reader,
}

// Read generates a Nano ID using the alphabet and stores it to the given
// byte slice, then returns the actual number of bytes generated and any
// errors encountered.
func Read(p []byte) (n int, err error) { return defaultReader.Read(p) }

// NewWithSize returns the new Nano ID generated using the default alphabet
// and any errors encountered. The size of the generated Nano ID depends on
// the given size.
func NewWithSize(size int) (id string, err error) {
	if size < 1 {
		return zeroString, errors.New("size is too small")
	}

	var buf []byte
	if size <= bufSliceSize {
		pointer := bufSlicePool.Get().(*[]byte)
		defer bufSlicePool.Put(pointer)
		buf = (*pointer)[:size]
	} else {
		buf = make([]byte, size)
	}

	_, err = Read(buf)
	if err != nil {
		return zeroString, err
	}
	return string(buf), nil
}

// New returns the new Nano ID generated using the default alphabet
// and any errors encountered. The size of the generated Nano ID is 21.
func New() (id string, err error) { return NewWithSize(21) }
