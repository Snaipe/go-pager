/* go-pager
 *
 * Copyright (c) 2017 Franklin "Snaipe" Mathieu <me@snai.pe>
 * Use of this source-code is govered by the MIT license, which
 * can be found in the LICENSE file.
 */

package pager

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"syscall"
)

var (
	ErrNoCommand = errors.New("No pager command to execute.")
	ErrClosed    = errors.New("The pager was closed.")
)

// Pager spawns and lets users write to a pager program.
//
// If an error occurs while writing to the pager, all subsequent writes,
// Close, and Error, will return that error.
//
// It implements io.WriteCloser.
type Pager struct {
	proc *exec.Cmd
	out  io.Writer
	err  error
}

type flusher interface {
	Flush() error
}

// Open calls OpenPager("", nil).
func Open() (*Pager, error) {
	return OpenPager("", nil)
}

// OpenPager spawns the pager program from the passed command, redirecting its
// standard output to dst, and returns a new Pager object.
//
// If command is an empty string, then the value of the PAGER enviroment
// variable is used instead. If it is also empty, then ErrNoCommand is returned.
// Any command is valid as long as they are acceptable as an operand to
// "sh -c".
//
// If dst is nil, os.Stdout is used instead.
func OpenPager(command string, dst io.Writer) (*Pager, error) {
	p := &Pager{}

	if command == "" {
		command = os.Getenv("PAGER")
	}
	if command == "" {
		return nil, ErrNoCommand
	}

	if dst == nil {
		dst = os.Stdout
	}

	var err error
	if fl, ok := dst.(flusher); ok {
		err = fl.Flush()
		if err != nil {
			return nil, err
		}
	}

	p.proc = exec.Command("sh", "-c", command)
	p.proc.Stdout = dst
	p.proc.Stderr = os.Stderr
	p.out, err = p.proc.StdinPipe()
	if err != nil {
		return nil, err
	}

	err = p.proc.Start()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func translateErr(err error) error {
	if perr, ok := err.(*os.PathError); ok && perr.Err == syscall.EPIPE {
		return ErrClosed
	}
	return err
}

func (p *Pager) cleanup() {
	if p.proc != nil {
		err := translateErr(p.out.(io.Closer).Close())
		if p.err == nil {
			p.err = err
		}

		err = p.proc.Wait()
		// There is a very good chance that any error that happened
		// during Close or previous writes are caused by an abnormal exit
		// of the pager, so override any error with this.
		if err != nil {
			p.err = err
		}

		if fl, ok := p.proc.Stdout.(flusher); ok {
			fl.Flush()
		}
	}
	p.out = nil
}

// Write writes the contents of data into the pager, and returns the number
// of bytes written. If n < len(data), it also returns an error explaining
// why the write is short.
//
// If Closed was previously called, or if the user closes the read end of the
// pager -- typically by exiting the pager program -- then Write returns
// the ErrClosed error.
func (p *Pager) Write(data []byte) (n int, err error) {
	if p.out == nil {
		return 0, ErrClosed
	}

	if p.err == nil {
		n, err = p.out.Write(data)
		p.err = translateErr(err)
		if p.err == ErrClosed {
			p.cleanup()
		}
		err = p.err
	}
	return
}

// Close closes the write end of the pager and frees all ressources associated
// with it. It returns the last write error that occured, or any error from
// the pager process.
func (p *Pager) Close() error {
	if p.out != nil {
		p.cleanup()
	}
	return p.err
}

// Error returns the last error that occured during a write or Close.
func (p *Pager) Error() error {
	return p.err
}

var _ io.WriteCloser = (*Pager)(nil)
