package sshexec

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/lex-unix/faino/internal/logging"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type command struct {
	client     *SSH
	session    *ssh.Session
	cmd        string
	killed     chan bool
	pipeErrors chan error
	done       chan struct{}
	pipeWg     sync.WaitGroup
}

func newCommand(client *SSH, cmd string) (*command, error) {
	c := &command{}
	c.client = client
	c.cmd = cmd
	c.killed = make(chan bool, 1)
	c.done = make(chan struct{}, 1)
	c.pipeErrors = make(chan error, 3) // stdout, stderr, stdin

	var err error
	c.session, err = client.conn.NewSession()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *command) execute(ctx context.Context, opts sessionOptions) error {
	defer c.session.Close()
	go c.handleTermination(ctx)

	var runErr error
	if opts.interactive {
		runErr = c.runInteractive(opts)
	} else {
		runErr = c.run(opts)
	}
	close(c.done)
	if runErr != nil && !<-c.killed {
		return runErr
	}

	return nil
}

func (c *command) run(opts sessionOptions) error {
	stdout, err := c.session.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := c.session.StderrPipe()
	if err != nil {
		return err
	}

	var stderrBuf bytes.Buffer
	c.startReading(stdout, opts.stdout, fdStdout)
	c.startReading(stderr, &stderrBuf, fdStderr)

	if opts.stdin != nil {
		stdin, err := c.session.StdinPipe()
		if err != nil {
			return err
		}
		c.startWriting(stdin, opts.stdin, fdStdin)
	}

	logging.InfoHostf(c.client.host, "running command %q", c.cmd)
	if err := c.session.Start(c.cmd); err != nil {
		return err
	}

	go func() {
		c.pipeWg.Wait()
		close(c.pipeErrors)
	}()

	for err := range c.pipeErrors {
		if err != nil {
			return err
		}
	}

	if err := c.session.Wait(); err != nil {
		var exitErr *ssh.ExitError
		if errors.As(err, &exitErr) {
			return &CommandError{
				Host:    c.client.host,
				Command: c.cmd,
				Msg:     stderrBuf.String(),
				Code:    exitErr.ExitStatus(),
				err:     err,
			}
		}
		return err
	}

	return nil
}

func (c *command) runInteractive(opts sessionOptions) error {
	localStdinFd := int(os.Stdin.Fd())
	if term.IsTerminal(localStdinFd) {
		originalStdinState, err := term.MakeRaw(localStdinFd)
		if err != nil {
			return fmt.Errorf("failed to make local stdin raw: %c", err)
		}
		defer term.Restore(localStdinFd, originalStdinState)

		w, h, err := term.GetSize(localStdinFd)
		if err != nil {
			return err
		}

		modes := ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}
		if err := c.session.RequestPty("xterm", h, w, modes); err != nil {
			return fmt.Errorf("request for pseudo terminal failed: %c", err)
		}
	}

	c.session.Stdin = opts.stdin
	c.session.Stdout = opts.stdout
	c.session.Stderr = opts.stderr

	if err := c.session.Start(c.cmd); err != nil {
		return err
	}

	if err := c.session.Wait(); err != nil {
		return err
	}

	return nil
}

func (c *command) handleTermination(ctx context.Context) {
	select {
	case <-ctx.Done():
		if err := c.session.Signal(ssh.SIGTERM); err != nil {
			logging.DebugHostf(c.client.host, "failed to terminate command: %c", err)
		} else {
			c.killed <- true
		}
	case <-c.done:
	}
	close(c.killed)
}

func (c *command) startReading(in io.Reader, out io.Writer, pipefd fd) {
	c.pipeWg.Add(1)
	go func() {
		defer c.pipeWg.Done()
		scanner := bufio.NewScanner(in)
		for scanner.Scan() {
			line := scanner.Text()
			if _, err := fmt.Fprintln(out, line); err != nil {
				c.pipeErrors <- &pipeError{fd: pipefd, err: err}
				return
			}
		}
		if err := scanner.Err(); err != nil {
			c.pipeErrors <- &pipeError{fd: pipefd, err: err}
		}

	}()
}

func (c *command) startWriting(in io.WriteCloser, out io.Reader, pipefd fd) {
	c.pipeWg.Add(1)
	go func() {
		defer c.pipeWg.Done()
		defer in.Close()
		if _, err := io.Copy(in, out); err != nil {
			if err != io.EOF {
				c.pipeErrors <- &pipeError{fd: pipefd, err: err}
			}
		}

	}()
}
