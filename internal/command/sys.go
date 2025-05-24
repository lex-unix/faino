package command

import "fmt"

func Mkdir(dir string) string {
	return fmt.Sprintf("mkdir -p %s", dir)
}

func CreateFileWithContents(file string, contents string) string {
	return fmt.Sprintf("echo %q > %s", contents, file)
}
