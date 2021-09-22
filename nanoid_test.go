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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkRead(b *testing.B) {
	b.SetBytes(21)
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		buf := make([]byte, 21)
		for p.Next() {
			_, _ = Read(buf)
		}
	})
}

func BenchmarkNew(b *testing.B) {
	b.SetBytes(21)
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_, _ = New()
		}
	})
}

func newTestReader(t *testing.T, opts ...ReaderOption) *Reader {
	r, err := NewReader(opts...)
	assert.NoError(t, err, "Unexpected error")
	assert.NotNil(t, r, "Unexpected nil reader")
	assert.IsType(t, &Reader{}, r, "Unexpected reader type")
	return r.(*Reader)
}

func TestNewReader(t *testing.T) {
	reader := newTestReader(t)
	assert.Equal(t, zeroString, reader.alphabet, "Unexpected alphabet")
	assert.Equal(t, 0, reader.mask, "Unexpected mask")
	assert.Equal(t, rand.Reader, reader.rander, "Unexpected rand reader")

	_, err := NewReader(WithAlphabet(zeroString))
	assert.Error(t, err, "Unexpected nil error")
}

const customAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz-"

func TestWithAlphabet(t *testing.T) {
	reader := newTestReader(t, WithAlphabet(customAlphabet))
	assert.Equal(t, customAlphabet, reader.alphabet, "Unexpected alphabet")
	assert.NotZero(t, reader.mask, "Unexpected mask")

	err := WithAlphabet(zeroString)(nil)
	assert.Error(t, err, "Unexpected nil error")

	alphabet := make([]byte, MaxAlphabetSize+1)
	_, _ = rand.Read(alphabet)

	err = WithAlphabet(string(alphabet))(nil)
	assert.Error(t, err, "Unexpected nil error")
}

type customRandReader struct {
	data []byte
}

func (r *customRandReader) Read(p []byte) (n int, err error) {
	if r.data == nil {
		return 0, io.EOF
	}
	nc := copy(p, r.data)
	if nc < len(p) {
		return nc, io.EOF
	}
	return nc, nil
}

func newCustomRandReader(size int) (r *customRandReader) {
	r = &customRandReader{
		data: make([]byte, size),
	}
	for index := 0; ; {
		for count := 0; count < 255; count++ {
			r.data[index] = byte(count)
			index++
			if index == size {
				return r
			}
		}
	}
}

func TestWithRandReader(t *testing.T) {
	rander := &customRandReader{}
	reader := newTestReader(t, WithRandReader(rander))
	assert.Equal(t, rander, reader.rander, "Unexpected rander")

	err := WithRandReader(nil)(nil)
	assert.Error(t, err, "Unexpected nil error")
}

func TestReaderInitializeMask(t *testing.T) {
	reader := newTestReader(t, WithAlphabet(customAlphabet))
	assert.Equal(t, 63, reader.mask, "Unexpected mask")

	reader.alphabet = zeroString
	assert.Panics(t, func() {
		reader.initializeMask()
	}, "Unexpected not panic")
}

func TestReaderGetRandomSize(t *testing.T) {
	reader := newTestReader(t, WithAlphabet(customAlphabet))
	size := reader.getRandomSize(21)
	assert.Equal(t, 34, size, "Unexpected random size")

	reader.mask = 0
	assert.Panics(t, func() {
		reader.getRandomSize(21)
	}, "Unexpected not panic")

	reader.alphabet = zeroString
	assert.Panics(t, func() {
		reader.getRandomSize(21)
	}, "Unexpected not panic")
}

func TestReaderRead(t *testing.T) {
	rander := newCustomRandReader(256)
	reader := newTestReader(t, WithRandReader(rander))

	buf := make([]byte, 21)
	nr, err := reader.Read(buf)
	assert.NoError(t, err, "Unexpected read error")
	assert.Equal(t, len(buf), nr, "Unexpected read size")
	assert.Equal(t, DefaultAlphabet[:len(buf)], string(buf), "Unexpected read data")

	_, err = reader.Read(nil)
	assert.Error(t, err, "Unexpected nil error")

	rander.data = nil
	_, err = reader.Read(buf)
	assert.Error(t, err, "Unexpected nil error")
}

func TestReaderCustomAlphabetRead(t *testing.T) {
	rander := newCustomRandReader(1024)
	reader := newTestReader(t, WithRandReader(rander), WithAlphabet(customAlphabet))

	buf := make([]byte, 21)
	nr, err := reader.Read(buf)
	assert.NoError(t, err, "Unexpected read error")
	assert.Equal(t, len(buf), nr, "Unexpected read size")
	assert.Equal(t, customAlphabet[:len(buf)], string(buf), "Unexpected read data")

	buf = make([]byte, 128)
	nr, err = reader.Read(buf)
	assert.NoError(t, err, "Unexpected read error")
	assert.Equal(t, len(buf), nr, "Unexpected read size")
	assert.Equal(t, customAlphabet, string(buf[:64]), "Unexpected read data")
	assert.Equal(t, customAlphabet, string(buf[64:]), "Unexpected read data")

	rander.data = nil
	_, err = reader.Read(buf)
	assert.Error(t, err, "Unexpected nil error")
}

func TestNewWithSize(t *testing.T) {
	defaultReader.rander = rand.Reader

	id, err := NewWithSize(21)
	assert.NoError(t, err, "Unexpected error")
	assert.Equal(t, 21, len(id), "Unexpected id size")

	id, err = NewWithSize(256)
	assert.NoError(t, err, "Unexpected error")
	assert.Equal(t, 256, len(id), "Unexpected id size")

	defaultReader.rander = &customRandReader{}
	_, err = NewWithSize(21)
	assert.Error(t, err, "Unexpected nil error")

	_, err = NewWithSize(0)
	assert.Error(t, err, "Unexpected nil error")
}

func TestNew(t *testing.T) {
	defaultReader.rander = rand.Reader

	id, err := New()
	assert.NoError(t, err, "Unexpected error")
	assert.Equal(t, 21, len(id), "Unexpected id size")
}
