// Clarity turns cryptic compiler and interpreter errors — from any
// language — into plain-English explanations that say exactly where the
// problem is and how to fix it.
//
// Usage:
//
//	g++ main.cpp -o main 2>&1 | clarity
//	python bad.py 2>&1 | clarity
//	clarity run -- g++ main.cpp -o main
//	clarity run -- python bad.py
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/kds1123001/clarity/internal/parser"
	"github.com/kds1123001/clarity/internal/report"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("clarity", flag.ContinueOnError)
	lang := fs.String("lang", "", "force the language (cpp, python, go, javascript, java, rust)")
	noColor := fs.Bool("no-color", false, "disable colored output")
	showRaw := fs.Bool("raw", false, "also print the original, unprocessed output")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `clarity — turns cryptic compiler/runtime errors into clear, actionable fixes

Usage:
  <build or run command> 2>&1 | clarity [flags]
  clarity run [flags] -- <command> [args...]

Flags:`)
		fs.PrintDefaults()
	}

	// Support "clarity run -- cmd args..." by splitting args manually,
	// since flag stops parsing at the first non-flag token otherwise.
	if len(args) > 0 && args[0] == "run" {
		return runSubcommand(args[1:], *lang, *noColor, *showRaw)
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if isStdinPiped() {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "clarity: failed reading stdin:", err)
			return 1
		}
		output := string(data)
		if *showRaw {
			fmt.Print(output)
			fmt.Println()
		}
		language := *lang
		if language == "" {
			language = parser.DetectLanguage("", output)
		}
		return renderAndExit(output, language, !*noColor)
	}

	fs.Usage()
	return 2
}

func runSubcommand(rest []string, forcedLang string, noColor, showRaw bool) int {
	fs := flag.NewFlagSet("clarity run", flag.ContinueOnError)
	lang := fs.String("lang", forcedLang, "force the language")
	nc := fs.Bool("no-color", noColor, "disable colored output")
	raw := fs.Bool("raw", showRaw, "also print the original, unprocessed output")

	// Find "--" separator to split clarity's own flags from the target command.
	sepIdx := -1
	for i, a := range rest {
		if a == "--" {
			sepIdx = i
			break
		}
	}
	if sepIdx == -1 {
		fmt.Fprintln(os.Stderr, "clarity run: expected -- before the command, e.g. clarity run -- g++ main.cpp")
		return 2
	}
	if err := fs.Parse(rest[:sepIdx]); err != nil {
		return 2
	}
	cmdArgs := rest[sepIdx+1:]
	if len(cmdArgs) == 0 {
		fmt.Fprintln(os.Stderr, "clarity run: no command given after --")
		return 2
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	runErr := cmd.Run()

	output := buf.String()
	if *raw {
		fmt.Print(output)
		fmt.Println()
	}

	language := *lang
	if language == "" {
		language = parser.DetectLanguage(cmdArgs[0], output)
	}

	code := renderAndExit(output, language, !*nc)
	if runErr == nil {
		return 0
	}
	if code == 0 {
		// The command failed but we found nothing to explain — still
		// surface a non-zero exit so scripts/CI notice.
		return 1
	}
	return code
}

// renderAndExit parses+prints diagnostics and returns a suggested process
// exit code: 0 if nothing was found, 1 if diagnostics were printed.
func renderAndExit(output, language string, color bool) int {
	diags := parser.Parse(language, output)
	n := report.Print(os.Stdout, diags, report.Options{Color: color})
	if n == 0 {
		if len(output) == 0 {
			fmt.Println("clarity: no output to analyze — the command may have succeeded.")
			return 0
		}
		fmt.Println("clarity: didn't recognize an error location in this output. Showing it as-is:")
		fmt.Println()
		fmt.Print(output)
		return 1
	}
	return 1
}

func isStdinPiped() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}
