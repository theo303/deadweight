# deadweight

**deadweight** is a tool that detects unused code using the [Language Server Protocol](https://microsoft.github.io/language-server-protocol/) (LSP).

The idea is simple: any LSP-compatible language server already knows how to enumerate symbols and find their references. deadweight leverages that to identify symbols with no real callers.

> **Current status:** only Go is supported, via `gopls`. The approach is language-agnostic by design and could be extended to any language with an LSP server.

---

## How it works

1. **Symbol discovery** — deadweight sends `textDocument/documentSymbol` requests to the language server for each source file, collecting all symbols (functions, types, struct fields, etc.) recursively.
2. **Reference lookup** — for each symbol, it sends a `textDocument/references` request to find every location where that symbol is used.
3. **Dead code detection** — symbols that are not referenced are considered unused and reported.

---

## Requirements

- **Go 1.22+**
- **gopls** installed and available in your `PATH`:

```bash
go install golang.org/x/tools/gopls@latest
```

---

## Installation

```bash
go install github.com/theo303/deadweight/cmd/deadweight@latest
```

Or clone and build locally:

```bash
git clone https://github.com/theo303/deadweight.git
cd deadweight
go build ./cmd/deadweight
```

---

## Usage

Run deadweight from the root of your project:

```bash
cd /path/to/your/project
deadweight
```

It will analyze all Go files in the current directory.

### Example output

```
INFO MyHandler (Function) internal/api/handler.go:42:1
INFO OldMiddleware (Function) internal/middleware/legacy.go:17:1
INFO UserStatus (String) internal/model/user.go:88:5
```

Each line shows the symbol name, its kind, and its location (`file:line:column`).

---

## Configuration

deadweight can be configured with a YAML file to filter out symbols you don't want to track. You can set the config file using the `-c` flag. If this flag is not provided deadweight will also look for a `.deadweight.yaml` file at the current directory root.

```bash
deadweight -c deadweight.yml
```

### Configuration file format

```yaml
# Ignore embedded struct fields (e.g. sync.Mutex embedded in a struct)
ignore-embedded-fields: true

# Symbols in this section are ignored, references are not checked for them
ignore-symbols:
  # Ignore common interface methods
  - kinds:
      - Method
    names:
      # name of method format: '(Type).Method', so we can use glob pattern.
      - "*.Error"
      - "*.String"
      - "*.MarshalJSON"
      - "*.UnmarshalJSON"

  # Ignore all exported types (useful for library packages)
  - kinds:
      - Struct
      - Interface
    names:
      - "[A-Z]*"

  - kinds:
      - Function
    names:
      - "main"
      - "init"
```

### Available symbol kinds

The `kinds` field accepts any of the LSP symbol kind names:

`File`, `Module`, `Namespace`, `Package`, `Class`, `Method`, `Property`, `Field`, `Constructor`, `Enum`, `Interface`, `Function`, `Variable`, `Constant`, `String`, `Number`, `Boolean`, `Array`, `Object`, `Key`, `Null`, `EnumMember`, `Struct`, `Event`, `Operator`, `TypeParameter`

### Name matching

Names support glob patterns using Go's [`filepath.Match`](https://pkg.go.dev/path/filepath#Match) syntax:

- `*` matches any sequence of non-separator characters
- `?` matches any single character
- `[abc]` matches a character class

---

## License

MIT — see [LICENSE](LICENSE).
