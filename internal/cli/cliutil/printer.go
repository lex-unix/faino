package cliutil

import (
	"fmt"
	"io"

	"github.com/lex-unix/faino/internal/app"
)

func PrintOutput(output app.HostOutput, out io.Writer) {
	for host, result := range output {
		fmt.Fprintf(out, "App Host %s:\n", host)
		fmt.Fprint(out, result)
		fmt.Fprintln(out)
	}
}
