# Clarity

Clarity reads the error output from a compiler or interpreter and prints the exact line, a plain-English explanation, and some things to try.

I started it because a one-character typo in C++ gave me 40 lines of errors and I couldn't find the problem.

It is written in Go and builds to a single binary. It is a learning project, and it works best on a small set of common errors. See "Honesty scope" below for what that means.

## Example

bad.cpp:5:5 - Undeclared identifier

      3 |  int main() {
      4 |      int x = 5;
 >    5 |      count << x << std::endl;
        |      ^
      6 |      return 0;
      7 |  }

  What happened:
    The compiler found the name 'count' but it was never declared.

  How to fix it:
    - Did you mean 'cout'?
    - Check for a typo in the name.
    - Declare it before this line, or #include the right header.
```

## Install

```
go install github.com/kds1123001/clarity@latest
```

Or build it yourself:

```
git clone https://github.com/kds1123001/clarity.git
cd clarity
make install
```

Tested with Go 1.26 on Ubuntu.

## Usage

Pipe mode, for any command that prints errors:

```
g++ main.cpp -o main 2>&1 | clarity
python3 script.py 2>&1 | clarity
```

Run mode, where Clarity runs the command for you and keeps its exit code:

```
clarity run -- g++ main.cpp -o main
```

Flags:

| Flag | What it does |
| --- | --- |
| `--lang` | Set the language yourself: `cpp`, `python`, `go`, `javascript`, `java`, `rust`. Without it, Clarity guesses. |
| `--no-color` | Turn off colors |
| `--raw` | Also print the original, unedited output |

## How it works

1. **Parse.** Pull the file, line, column, and message out of the raw output. Compiler-style `file:line:col: message` errors use regex. Python tracebacks have their own parser. Linker errors have a fallback.
2. **Match.** Compare the message to a list of regex rules. Each rule has a title, an explanation, and a list of fixes.
3. **Suggest.** For unknown names, compare against a short list of commonly mistyped names for that language (like `cout`, `cin`, `endl` in C++) using edit distance, and offer a "did you mean".
4. **Render.** Print the location, the source lines around it, the explanation, and the fixes.

The rules are in `data.json` as plain JSON, so you can add one without writing Go. `CONTRIBUTING.md` explains the format.

## Honesty scope

I want this README to say only what I can back up.

**What it covers**
- C++
- Python
- Go, JavaScript, java
- Rust
- Any error shaped like `file:line:col: message` still gets a location and a code snippet, even with no rule for it.

It does not cover "any language." If there is no rule for an error, you get the location but no explanation.

**What I have not tested**
- the parser and the rule matching have no automated tests
- Windows
- large projects or very long error output

**Where it can be wrong**
- Rules are regex patterns. A pattern can match the wrong error and give advice that does not fit your problem. Treat the fixes as ideas, not answers, and trust the compiler's message if they disagree.
- The "did you mean" list is small and written by hand, so it misses most names.

**Who wrote it**
i did lol

## Contributing

Adding a rule is the easiest way to help. If an error confused you and Clarity had nothing to say, open an issue with the exact error text, or add a rule to `data.json`. See `CONTRIBUTING.md`.

## Development

```
make build
make test
make fmt
```

## Plans

- tests for the parser, Rust rules, fix CI

## License

MIT. See `LICENSE`.
