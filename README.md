# Clarity

**Compilers tell you *that* something's wrong. Clarity tells you *where* and *how to fix it.***

Ever run `g++` and get slapped with 40 lines of template gibberish for a one-character typo? Clarity wraps the output of any compiler or interpreter, points at the exact line, explains what actually went wrong in plain English, and gives you concrete steps to fix it — regardless of which language you're working in.

```
✖ bad.cpp:5:5 — Undeclared identifier

      3 │ int main() {
      4 │     int x = 5;
 ▶    5 │     count << x << std::endl;
      │     ^
      6 │     return 0;
      7 │ }

  What happened:
    The compiler ran into the name 'count' but has no record of it ever
    being declared or defined before this point in the file.

  How to fix it:
    • Did you mean 'cout'?
    • Check for a typo in the name.
    • Declare or define it before this line — C++ requires declaration
      before first use.
    • If it comes from the standard library or a third-party header,
      make sure you #include the right header.
```

## Why

Every language's compiler/interpreter has its own dialect of cryptic. `no matching function for call to`, `NullPointerException`, `undefined reference to`, `cannot find symbol` — you learn to translate these over years of experience. Clarity does that translation for you, out of the box, for the errors that trip people up most often.

It's not a linter and it's not a language server. It doesn't need your build system reconfigured. Pipe any command's output into it, or wrap the command itself, and it works.

## Install

```bash
go install github.com/kds1123001/clarity@latest
```

Or build from source:

```bash
git clone https://github.com/kds1123001/clarity.git
cd clarity
make install   # builds and installs to $GOPATH/bin
```

A prebuilt binary requires no runtime dependencies — Clarity is a single static Go binary, so distributing it as a `pip`/`npm`/`brew` wrapper that just downloads the right binary for the platform is straightforward if you want it available through those ecosystems too (see [`CONTRIBUTING.md`](CONTRIBUTING.md)).

## Usage

**Pipe mode** — works with anything that writes errors to stdout/stderr:

```bash
g++ main.cpp -o main 2>&1 | clarity
python3 script.py 2>&1 | clarity
node app.js 2>&1 | clarity
go build ./... 2>&1 | clarity
javac Main.java 2>&1 | clarity
```

**Run mode** — let Clarity run the command for you (useful in scripts/CI, preserves exit codes):

```bash
clarity run -- g++ main.cpp -o main
clarity run -- python3 script.py
```

**Flags:**

| Flag         | Description                                          |
|--------------|-------------------------------------------------------|
| `--lang`     | Force the language (`cpp`, `python`, `go`, `javascript`, `java`, `rust`) instead of auto-detecting |
| `--no-color` | Disable colored output (useful in CI logs)            |
| `--raw`      | Also print the original, unprocessed compiler output  |

## What it understands today

Clarity ships with a growing rule database covering the errors that cost people the most time:

- **C++**: undeclared identifiers, missing members, no matching overload, missing semicolons, linker "undefined reference", missing headers, segfaults
- **Python**: `NameError`, `ModuleNotFoundError`, `IndentationError`, `TypeError`, `IndexError`, `KeyError`, `ZeroDivisionError`, `AttributeError`
- **Go**: undefined identifiers, unused imports, unused variables
- **JavaScript**: `ReferenceError`, property access on `undefined`/`null`, calling a non-function
- **Java**: `cannot find symbol`, `NullPointerException`
- **Generic**: permission denied, file not found, and any `file:line:col: message` style error, even for languages without a dedicated rule set yet

If Clarity doesn't recognize a pattern, it still shows you exactly where the error is (with source context) — it just won't have a plain-English explanation for that specific message yet. That's where you come in.

## Contributing a rule

The whole point of this project is that the community can keep adding coverage for the "there's a problem but nobody tells you where or how to fix it" errors they personally hit. Rules live in [`internal/rules/data.json`](internal/rules/data.json) as plain JSON — no Go code required. See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the format and a walkthrough.

## How it works

1. **Parse** — extract `file`, `line`, `column`, and `message` from raw output using per-language conventions (regex for compiler-style `file:line:col:` errors, a dedicated parser for Python tracebacks, and a fallback for linker-style errors).
2. **Match** — run the message against a small rule database of regex patterns, each mapped to a title, plain-English explanation, and a list of concrete fixes.
3. **Suggest** — for unrecognized identifiers, compare against a curated dictionary of commonly-confused symbols per language (e.g. `cout`/`cin`/`endl` in C++) using edit distance, and offer a "did you mean" when there's a close match.
4. **Render** — print the location, a source snippet with the exact line/column highlighted, the explanation, and fixes.

## Development

```bash
make build   # build ./clarity
make test    # run the test suite
make fmt     # gofmt
```

## License

MIT — see [LICENSE](LICENSE).
            #NO PART OF THIS CODE WAS MADE USING AI NOR LLM#
