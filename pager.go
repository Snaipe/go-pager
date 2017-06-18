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

	"golang.org/x/crypto/ssh/terminal"
)

var (
	ErrNoCommand = errors.New("No pager command to execute.")
	ErrClosed    = errors.New("The pager was closed.")
)

type Pager struct {
	proc *exec.Cmd
	out  io.Writer
	err  error
}

type flusher interface {
	Flush() error
}

type filebacked interface {
	Fd() uintptr
}

func isatty(wr io.Writer) bool {
	f, ok := wr.(filebacked)
	return ok && terminal.IsTerminal(int(f.Fd()))
}

func Open() (*Pager, error) {
	return OpenPager("", nil)
}

func OpenPager(command string, dst io.Writer) (*Pager, error) {
	p := &Pager{}

	if command == "" {
		command = os.Getenv("PAGER")
	}
	if command == "" {
		return nil, ErrNoCommand
	}

	if out == nil {
		out = os.Stdout
	}

	if !isatty(out) {
		p.out = out
		return p, nil
	}

	var err error
	if fl, ok := out.(flusher); ok {
		err = fl.Flush()
		if err != nil {
			return nil, err
		}
	}

	p.proc = exec.Command("sh", "-c", command)
	p.proc.Stdout = out
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

func (p *Pager) Write(data []byte) (int, error) {
	if p.out == nil {
		return 0, ErrClosed
	}

	var written int
	if p.err == nil {
		written, p.err = p.out.Write(data)
		p.err = translateErr(p.err)
	}
	return written, p.err
}

func (p *Pager) Close() error {
	if p.out == nil {
		return p.err
	}

	if p.proc != nil {
		p.err = translateErr(p.out.(io.Closer).Close())
		procerr := p.proc.Wait()
		if p.err == nil {
			p.err = procerr
		}

		if fl, ok := p.proc.Stdout.(flusher); ok {
			fl.Flush()
		}
	}

	p.out = nil
	return p.err
}

func (p *Pager) Error() error {
	return p.err
}

var _ io.WriteCloser = (*Pager)(nil)
