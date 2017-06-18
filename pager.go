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
	ErrUserClose = errors.New("User closed the pager.")
)

type Pager struct {
	Command string

	proc *exec.Cmd
	out  io.WriteCloser
	err  error
}

type flusher interface {
	Flush() error
}

func Open() (*Pager, error) {
	return OpenPager("", nil)
}

func OpenPager(command string, out io.Writer) (*Pager, error) {
	p := &Pager{Command: command}

	cmd := p.Command
	if cmd == "" {
		cmd = os.Getenv("PAGER")
	}
	if cmd == "" {
		return nil, ErrNoCommand
	}

	if out == nil {
		out = os.Stdout
	}

	var err error
	if fl, ok := out.(flusher); ok {
		err = fl.Flush()
		if err != nil {
			return nil, err
		}
	}

	p.proc = exec.Command("sh", "-c", cmd)
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
		return ErrUserClose
	}
	return err
}

func (p *Pager) Write(data []byte) (int, error) {
	var written int
	if p.err == nil {
		written, p.err = p.out.Write(data)
	}
	p.err = translateErr(p.err)
	return written, p.err
}

func (p *Pager) Close() error {
	if p.out == nil {
		return p.err
	}

	p.err = translateErr(p.out.Close())
	procerr := p.proc.Wait()
	if p.err == nil {
		p.err = procerr
	}

	if fl, ok := p.proc.Stdout.(flusher); ok {
		fl.Flush()
	}

	p.out = nil
	return p.err
}

func (p *Pager) Error() error {
	return p.err
}

var _ io.WriteCloser = (*Pager)(nil)
