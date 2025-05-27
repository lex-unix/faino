package command

import (
	"fmt"

	"al.essio.dev/pkg/shellescape"
)

func Mkdir(dir string) string {
	return fmt.Sprintf("mkdir -p %s", dir)
}

func CreateFileWithContents(file string, contents string) string {
	return fmt.Sprintf("echo %s > %s", shellescape.Quote(contents), file)
}
