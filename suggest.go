// Package parser turns raw, messy compiler/interpreter output into a
// structured list of Diagnostics: file, line, column, message.
package parser

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Diagnostic is one located error or warning.
type Diagnostic struct {
	File     string
	Line     int
	Col      int
	Message  string
	Severity string // "error" or "warning"
	Language string
}

// genericLineRE matches the near-universal "file:line:col: message" or
// "file:line: message" format used by gcc, clang, go, rustc, javac, etc.
var genericLineRE = regexp.MustCompile(
	`^([^\s:][^:]*\.[A-Za-z0-9]+):(\d+):(?:(\d+):)?\s*(?:(fatal error|error|warning|note)\s*:\s*)?(.+)$`,
)

// linkerLineRE matches ld-style lines like:
//
//	main.cpp:(.text+0x9): undefined reference to `doThing()'
//
// These have no line number, only a section offset, so we surface them
// without a source snippet.
var linkerLineRE = regexp.MustCompile(`^([^\s:]+\.[A-Za-z0-9+]+):\([^)]*\):\s*(.+)$`)

var pyFileLineRE = regexp.MustCompile(`^\s*File "([^"]+)", line (\d+), in (.+)$`)
var pyExceptionRE = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_.]*(?:Error|Exception|Warning)):?\s*(.*)$`)

// cmdLanguage maps a compiler/interpreter executable name to a language tag.
var cmdLanguage = map[string]string{
	"g++": "cpp", "gcc": "cpp", "clang": "cpp", "clang++": "cpp", "cc": "cpp", "c++": "cpp",
	"python": "python", "python3": "python", "python2": "python",
	"node": "javascript", "nodejs": "javascript",
	"go":    "go",
	"javac": "java", "java": "java",
	"rustc": "rust", "cargo": "rust",
}

// extLanguage maps a file extension to a language tag, used as a fallback
// when we can't tell from the command name (e.g. piped input).
var extLanguage = map[string]string{
	".cpp": "cpp", ".cc": "cpp", ".cxx": "cpp", ".h": "cpp", ".hpp": "cpp",
	".c":  "cpp",
	".py": "python",
	".go": "go",
	".js": "javascript", ".ts": "javascript", ".jsx": "javascript", ".tsx": "javascript",
	".java": "java",
	".rs":   "rust",
}

// DetectLanguage tries the command name first, then falls back to sniffing
// the output for recognizable markers or file extensions.
func DetectLanguage(cmdName, output string) string {
	base := filepath.Base(cmdName)
	if lang, ok := cmdLanguage[base]; ok {
		return lang
	}
	if strings.Contains(output, "Traceback (most recent call last):") {
		return "python"
	}
	// Fall back to scanning for a recognizable file extension.
	if m := genericLineRE.FindStringSubmatch(output); m != nil {
		ext := filepath.Ext(m[1])
		if lang, ok := extLanguage[ext]; ok {
			return lang
		}
	}
	return "generic"
}

// Parse dispatches to the right strategy based on language.
func Parse(language, output string) []Diagnostic {
	if language == "python" || strings.Contains(output, "Traceback (most recent call last):") {
		if diags := parsePythonTraceback(output); len(diags) > 0 {
			return diags
		}
	}
	return parseGeneric(language, output)
}

func parseGeneric(language, output string) []Diagnostic {
	var diags []Diagnostic
	seen := map[string]bool{}
	for _, rawLine := range strings.Split(output, "\n") {
		line := strings.TrimRight(rawLine, "\r")

		m := genericLineRE.FindStringSubmatch(line)
		if m == nil {
			// Try the linker-style "file:(.section+offset): message" format,
			// common for "undefined reference" errors.
			if lm := linkerLineRE.FindStringSubmatch(line); lm != nil {
				key := lm[1] + ":0:" + lm[2]
				if !seen[key] {
					seen[key] = true
					lang := language
					if lang == "generic" || lang == "" {
						if l, ok := extLanguage[filepath.Ext(lm[1])]; ok {
							lang = l
						} else {
							lang = "generic"
						}
					}
					diags = append(diags, Diagnostic{
						File:     lm[1],
						Line:     0,
						Col:      0,
						Message:  strings.TrimSpace(lm[2]),
						Severity: "error",
						Language: lang,
					})
				}
			}
			continue
		}
		lineNo, _ := strconv.Atoi(m[2])
		col := 0
		if m[3] != "" {
			col, _ = strconv.Atoi(m[3])
		}
		severity := strings.ToLower(m[4])
		if severity == "" {
			severity = "error"
		}
		if severity == "note" {
			continue // skip pure follow-on notes; keep primary diagnostics only
		}
		if severity == "fatal error" {
			severity = "error"
		}
		msg := strings.TrimSpace(m[5])
		lang := language
		if lang == "generic" || lang == "" {
			if l, ok := extLanguage[filepath.Ext(m[1])]; ok {
				lang = l
			} else {
				lang = "generic"
			}
		}
		key := m[1] + ":" + m[2] + ":" + msg
		if seen[key] {
			continue
		}
		seen[key] = true
		diags = append(diags, Diagnostic{
			File:     m[1],
			Line:     lineNo,
			Col:      col,
			Message:  msg,
			Severity: severity,
			Language: lang,
		})
	}
	return diags
}

// parsePythonTraceback extracts the deepest ("most recent") frame and the
// final exception line, since that's almost always what the user needs to
// look at, not the whole call stack.
func parsePythonTraceback(output string) []Diagnostic {
	lines := strings.Split(output, "\n")
	var lastFile string
	var lastLine int
	haveFrame := false

	var diags []Diagnostic

	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		if m := pyFileLineRE.FindStringSubmatch(line); m != nil {
			lastFile = m[1]
			n, _ := strconv.Atoi(m[2])
			lastLine = n
			haveFrame = true
			continue
		}
		// An unindented line starting with an identifier + Error/Exception
		// marks the end of the traceback.
		if strings.TrimSpace(line) == "" {
			continue
		}
		if line[0] == ' ' || line[0] == '\t' {
			continue
		}
		if line == "Traceback (most recent call last):" {
			continue
		}
		if m := pyExceptionRE.FindStringSubmatch(line); m != nil && haveFrame {
			msg := line
			diags = append(diags, Diagnostic{
				File:     lastFile,
				Line:     lastLine,
				Col:      0,
				Message:  msg,
				Severity: "error",
				Language: "python",
			})
			haveFrame = false
		}
	}
	return diags
}
