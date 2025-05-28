package sshexec

import "fmt"

type pipeError struct {
	fd  fd
	err error
}

func (e pipeError) Error() string {
	var d string
	switch e.fd {
	case fdStdin:
		d = "stdin"
	case fdStdout:
		d = "stdout"
	case fdStderr:
		d = "stderr"
	}
	return fmt.Sprintf("%s pipe: %s", d, e.err)
}

type CommandError struct {
	Host    string
	Msg     string
	Code    int
	Command string
	err     error
}

func (e CommandError) Error() string {
	return fmt.Sprintf("command %s failed with exit code %d", e.Command, e.Code)
}

func (e CommandError) Unwrap() error {
	return e.err
}

func (e CommandError) NotFound() bool {
	return e.Code == 127
}
