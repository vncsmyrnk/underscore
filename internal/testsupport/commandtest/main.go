package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		_, _ = fmt.Fprintln(os.Stderr, "commandtest: missing scenario")
		os.Exit(2)
	}

	switch os.Args[1] {
	case "stdout":
		if len(os.Args) < 3 {
			_, _ = fmt.Fprintln(os.Stderr, "commandtest: stdout requires a payload")
			os.Exit(2)
		}
		_, _ = fmt.Fprint(os.Stdout, os.Args[2])
	case "stderr":
		if len(os.Args) < 3 {
			_, _ = fmt.Fprintln(os.Stderr, "commandtest: stderr requires a payload")
			os.Exit(2)
		}
		_, _ = fmt.Fprint(os.Stderr, os.Args[2])
	case "exit":
		if len(os.Args) < 3 {
			_, _ = fmt.Fprintln(os.Stderr, "commandtest: exit requires a status")
			os.Exit(2)
		}
		var status int
		if _, err := fmt.Sscanf(os.Args[2], "%d", &status); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "commandtest: invalid status %q\n", os.Args[2])
			os.Exit(2)
		}
		os.Exit(status)
	case "cat":
		if _, err := fmt.Fprint(os.Stdout, readAll()); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "commandtest: write stdout: %v\n", err)
			os.Exit(1)
		}
	default:
		_, _ = fmt.Fprintf(os.Stderr, "commandtest: unknown scenario %q\n", os.Args[1])
		os.Exit(2)
	}
}

func readAll() string {
	data, err := os.ReadFile("/dev/stdin")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "commandtest: read stdin: %v\n", err)
		os.Exit(1)
	}

	return string(data)
}
