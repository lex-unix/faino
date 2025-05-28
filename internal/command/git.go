package command

import "fmt"

func CommitHash() string {
	return "git rev-parse --short HEAD"
}

func CommitMessage() string {
	return fmt.Sprintf("git log -1 --pretty=%%B")
}
