package parser

import "testing"

func TestDetectLanguageByCommand(t *testing.T) {
	cases := map[string]string{
		"g++":     "cpp",
		"clang++": "cpp",
		"python3": "python",
		"node":    "javascript",
		"go":      "go",
		"javac":   "java",
		"rustc":   "rust",
	}
	for cmd, want := range cases {
		got := DetectLanguage(cmd, "")
		if got != want {
			t.Errorf("DetectLanguage(%q, \"\") = %q, want %q", cmd, got, want)
		}
	}
}

func TestDetectLanguagePythonTraceback(t *testing.T) {
	out := "Traceback (most recent call last):\n  File \"x.py\", line 1, in <module>\nZeroDivisionError: division by zero\n"
	if got := DetectLanguage("", out); got != "python" {
		t.Errorf("DetectLanguage from traceback = %q, want python", got)
	}
}

func TestParseGenericCppError(t *testing.T) {
	out := "bad.cpp:5:5: error: use of undeclared identifier 'count'\n"
	diags := Parse("cpp", out)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d: %+v", len(diags), diags)
	}
	d := diags[0]
	if d.File != "bad.cpp" || d.Line != 5 || d.Col != 5 {
		t.Errorf("unexpected location: %+v", d)
	}
	if d.Severity != "error" {
		t.Errorf("expected severity error, got %s", d.Severity)
	}
}

func TestParseSkipsNotes(t *testing.T) {
	out := "a.cpp:1:1: error: something bad\na.cpp:1:1: note: see also here\n"
	diags := Parse("cpp", out)
	if len(diags) != 1 {
		t.Fatalf("expected notes to be filtered out, got %d diagnostics", len(diags))
	}
}

func TestParseLinkerError(t *testing.T) {
	out := "main.cpp:(.text+0x9): undefined reference to `doThing()'\ncollect2: error: ld returned 1 exit status\n"
	diags := Parse("cpp", out)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d: %+v", len(diags), diags)
	}
	if diags[0].File != "main.cpp" || diags[0].Line != 0 {
		t.Errorf("unexpected linker diagnostic: %+v", diags[0])
	}
}

func TestParsePythonTracebackDeepestFrame(t *testing.T) {
	out := `Traceback (most recent call last):
  File "bad.py", line 10, in <module>
    main()
  File "bad.py", line 7, in main
    result = divide(numbers[0], numbers[2])
  File "bad.py", line 2, in divide
    return a / b
ZeroDivisionError: division by zero
`
	diags := Parse("python", out)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d: %+v", len(diags), diags)
	}
	d := diags[0]
	if d.File != "bad.py" || d.Line != 2 {
		t.Errorf("expected deepest frame bad.py:2, got %+v", d)
	}
	if d.Message != "ZeroDivisionError: division by zero" {
		t.Errorf("unexpected message: %q", d.Message)
	}
}

func TestParseDedupsRepeatedLines(t *testing.T) {
	out := "a.cpp:1:1: error: same thing\na.cpp:1:1: error: same thing\n"
	diags := Parse("cpp", out)
	if len(diags) != 1 {
		t.Fatalf("expected duplicate lines to be deduped, got %d", len(diags))
	}
}
