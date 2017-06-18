/* go-pager
 *
 * Copyright (c) 2017 Franklin "Snaipe" Mathieu <me@snai.pe>
 * Use of this source-code is govered by the MIT license, which
 * can be found in the LICENSE file.
 */

package pager

import (
	"bytes"
	"errors"
	"io"
	"os"
	"syscall"
	"testing"
)

type TestWriter struct {
	bytes.Buffer
	err error
}

func (fw *TestWriter) Write(data []byte) (int, error) {
	if fw.err != nil {
		return 0, fw.err
	}
	return fw.Buffer.Write(data)
}

func (fw *TestWriter) Close() {
	fw.err = &os.PathError{Err: syscall.EPIPE}
}

func NewTestPager(w io.Writer) *Pager {
	return &Pager{out: w}
}

var data = []byte("data")

func TestError(t *testing.T) {
	pager := NewTestPager(nil)

	err := pager.Error()
	if err != nil {
		t.Fatalf("Pager.Error(): expected no errors, but failed (%v).", err)
	}
}

func TestWrite(t *testing.T) {
	var buf TestWriter
	pager := NewTestPager(&buf)

	n, err := pager.Write(data)
	if n < len(data) || err != nil {
		t.Fatalf("Pager.Write(): expected no error, but write returned (%v, %v).", n, err)
	}
	if !bytes.Equal(buf.Bytes(), data) {
		t.Fatalf("Pager.Write(): expected %s to be written, got %s.", string(data), string(buf.Bytes()))
	}

	err = pager.Error()
	if err != nil {
		t.Fatalf("Pager.Error(): expected no errors, but failed (%v).", err)
	}
}

func TestWriteError(t *testing.T) {
	var buf TestWriter
	pager := NewTestPager(&buf)
	buf.err = errors.New("dummy error")

	n, err := pager.Write(data)
	if n > 0 || err == nil {
		t.Fatalf("Pager.Write(): expected error, but call succeeded.")
	}
	if err != buf.err {
		t.Fatalf("Pager.Write(): expected dummy error, but got other error (%v).", err)
	}
}

func TestUserClose(t *testing.T) {
	var buf TestWriter
	pager := NewTestPager(&buf)

	buf.Close()
	n, err := pager.Write(data)
	if n > 0 || err == nil {
		t.Fatalf("Pager.Write(): expected error, but call succeeded.")
	}
	if err != ErrClosed {
		t.Fatalf("Pager.Write(): expected ErrClosed, but got other error (%v).", err)
	}
}

func TestClose(t *testing.T) {
	pager := NewTestPager(nil)

	err := pager.Close()
	if err != nil {
		t.Fatalf("Pager.Close(): expected no errors, but failed (%v).", err)
	}

	err = pager.Error()
	if err != nil {
		t.Fatalf("Pager.Error(): expected no errors, but failed (%v).", err)
	}
}

func TestCloseError(t *testing.T) {
	var buf TestWriter
	pager := NewTestPager(&buf)
	buf.err = errors.New("dummy error")
	pager.Write(data)

	err := pager.Close()
	if err != buf.err {
		t.Fatalf("Pager.Error(): expected dummy error, but got other error (%v).", err)
	}

	err = pager.Error()
	if err != buf.err {
		t.Fatalf("Pager.Error(): expected dummy error, but got other error (%v).", err)
	}
}

func TestWriteAfterClose(t *testing.T) {
	var buf TestWriter
	pager := NewTestPager(&buf)

	err := pager.Close()
	if err != nil {
		t.Fatalf("Pager.Close(): expected no errors, but failed (%v).", err)
	}

	n, err := pager.Write(data)
	if n > 0 || err == nil {
		t.Fatalf("Pager.Write(): expected error, but call succeeded.")
	}
	if err != ErrClosed {
		t.Fatalf("Pager.Write(): expected ErrClosed, but got other error (%v).", err)
	}

	// pager.Write should not change last error after close
	err = pager.Error()
	if err != nil {
		t.Fatalf("Pager.Error(): expected no errors, but failed (%v).", err)
	}
}
