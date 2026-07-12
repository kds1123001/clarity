# Contributing

The most valuable contribution to this project is a new rule for an error you personally got stuck on. You don't need to touch any Go code.

## Adding a rule

Rules live in [`internal/rules/data.json`](internal/rules/data.json), as a plain JSON array. Each entry looks like this:

```json
{
  "id": "cpp-undeclared-identifier",
  "language": "cpp",
  "match": "use of undeclared identifier '([^']+)'|'([^']+)' was not declared in this scope",
  "title": "Undeclared identifier",
  "explanation": "The compiler ran into the name '{{sym}}' but has no record of it ever being declared or defined before this point in the file.",
  "fixes": [
    "Check for a typo in the name.",
    "Declare or define it before this line."
  ]
}
```

Field reference:

| Field         | Meaning                                                                 |
|---------------|--------------------------------------------------------------------------|
| `id`          | Unique, kebab-case identifier, prefixed with the language (e.g. `python-key-error`) |
| `language`    | One of `cpp`, `python`, `go`, `javascript`, `java`, `rust`, or `generic` for language-agnostic patterns |
| `match`       | A Go regular expression ([RE2 syntax](https://github.com/google/re2/wiki/Syntax)) matched against the diagnostic message. Use a capture group around the most relevant symbol (identifier, filename, etc.) — the first non-empty group becomes `{{sym}}`. |
| `title`       | Short label shown next to the file:line (e.g. "Undeclared identifier")   |
| `explanation` | One or two plain-English sentences. Use `{{sym}}` to interpolate the captured symbol. |
| `fixes`       | An ordered list of concrete, actionable steps. Most specific/likely fix first. |

### Guidelines for good rules

- **Write for someone who's never seen this error before.** Assume no prior knowledge of the term used in the raw error message.
- **Be concrete.** "Check your code" is not a fix. "Check that argument types match the function's declaration" is.
- **Order fixes by likelihood.** Put the most common cause first.
- **Test your regex** against the *exact* error text your compiler/interpreter produces — error wording varies between compiler versions (e.g. gcc vs clang phrase the same C++ error differently), so it's fine — encouraged, even — to add multiple alternation branches (`|`) covering different compilers' wording for the same underlying problem.
- **One rule, one root cause.** If two different messages need substantially different fix lists, that's two rules.

### Testing your rule

```bash
go test ./internal/rules/...
```

Add a small test in `internal/rules/rules_test.go` asserting your new rule matches its target message and extracts the right symbol. Then try it end-to-end:

```bash
go build -o /tmp/clarity .
echo "your_language_here:12:3: error: <exact message your rule targets>" | /tmp/clarity
```

## Adding a new language

1. Add file-extension mappings in `internal/parser/parser.go` (`extLanguage`) and a command-name mapping in `cmdLanguage` if the language's compiler/interpreter has a recognizable executable name.
2. If the language's error format doesn't fit the generic `file:line:col: message` pattern (like Python's multi-line tracebacks), add a dedicated parser function, following the pattern of `parsePythonTraceback`.
3. Add rules for that language's most common/confusing errors in `data.json`.
4. Optionally add a symbol dictionary in `internal/suggest/suggest.go` for "did you mean" suggestions.

## Packaging for other ecosystems

Clarity is a dependency-free static Go binary, so wrapping it for `pip`, `npm`, `brew`, etc. mostly means shipping a small installer script/package that downloads the right release binary for the platform and puts it on `PATH`. PRs adding these wrappers (as separate subdirectories, e.g. `packaging/pip/`, `packaging/npm/`) are welcome.

## Code changes

For changes to the parsing/matching/rendering logic itself:

```bash
make fmt
make test
```

Please keep the CLI dependency-free (standard library only) — that's part of what makes a single-binary, zero-install tool possible.
