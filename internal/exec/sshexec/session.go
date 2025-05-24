package sshexec

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/lex-unix/faino/internal/logging"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func (s *SSH) run(session *ssh.Session, cmd string, opts sessionOptions) error {
	time.Sleep(2 * time.Second)
	var wg sync.WaitGroup
	errCh := make(chan error, 3) // buffer size of 3 for stdout, stderr, stdin
	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	wg.Add(2)
	go read(&wg, fdStdout, stdout, opts.stdout, errCh)
	go read(&wg, fdStderr, stderr, opts.stderr, errCh)

	if opts.stdin != nil {
		stdin, err := session.StdinPipe()
		if err != nil {
			return err
		}
		wg.Add(1)
		go write(&wg, fdStdin, stdin, opts.stdin, errCh)
	}

	logging.InfoHostf(s.host, "running command %q", cmd)
	if err := session.Start(cmd); err != nil {
		return err
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	if err := session.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *SSH) runInteractiveSession(session *ssh.Session, cmd string, opts sessionOptions) error {
	localStdinFd := int(os.Stdin.Fd())
	if term.IsTerminal(localStdinFd) {
		originalStdinState, err := term.MakeRaw(localStdinFd)
		if err != nil {
			return fmt.Errorf("failed to make local stdin raw: %s", err)
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
		if err := session.RequestPty("xterm", h, w, modes); err != nil {
			return fmt.Errorf("request for pseudo terminal failed: %s", err)
		}
	}

	session.Stdin = opts.stdin
	session.Stdout = opts.stdout
	session.Stderr = opts.stderr

	if err := session.Start(cmd); err != nil {
		return err
	}

	if err := session.Wait(); err != nil {
		return err
	}

	return nil
}

func write(wg *sync.WaitGroup, pipefd fd, in io.WriteCloser, out io.Reader, errCh chan<- error) {
	defer wg.Done()
	defer in.Close()
	if _, err := io.Copy(in, out); err != nil {
		if err != io.EOF {
			errCh <- &pipeError{fd: fdStdin, err: err}
		}
	}
}

func read(wg *sync.WaitGroup, pipefd fd, in io.Reader, out io.Writer, errCh chan<- error) {
	defer wg.Done()
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		if _, err := fmt.Fprintln(out, line); err != nil {
			errCh <- &pipeError{fd: pipefd, err: err}
			return
		}
	}
	if err := scanner.Err(); err != nil {
		errCh <- &pipeError{fd: pipefd, err: err}
	}
}
