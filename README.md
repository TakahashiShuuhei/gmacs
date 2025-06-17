# gmacs - Go Emacs-like Editor

A minimal Emacs-like text editor implemented in Go with Lua configuration support.

## 🚀 Features

- **Basic text editing**: Character insertion, cursor movement, text deletion
- **File operations**: Open, save, and write files
- **Emacs-like key bindings**: C-x C-f (find-file), C-x C-s (save-buffer), etc.
- **Lua configuration**: Complete Lua scripting support with package management
- **Package system**: D案 architecture for natural package loading
- **Extensible**: Clean API for adding commands and key bindings

## 📦 Installation

```bash
# Clone the repository
git clone https://github.com/TakahashiShuuhei/gmacs.git
cd gmacs

# Build the editor
go build ./cmd/gmacs

# Run gmacs
./gmacs [filename]
```

## ⚙️ Configuration

gmacs uses Lua for configuration. The configuration file is automatically loaded from:

```
~/.config/gmacs/init.lua
```

### Basic Configuration Example

```lua
-- ~/.config/gmacs/init.lua

-- Set variables
gmacs.set_variable("theme", "dark")

-- Register custom commands
function my_hello()
    gmacs.message("Hello from Lua!")
end
gmacs.register_command("hello", my_hello, "Say hello")

-- Set key bindings
gmacs.global_set_key("C-h", "hello")

-- Package management (D案 architecture)
gmacs.use_package("github.com/user/ruby-mode", {
    ruby_path = "/usr/local/bin/ruby",
    auto_indent = true
})

-- Package APIs are available immediately after declaration
gmacs.global_set_key("C-c C-d", "ruby-show-doc")
```

### Available Lua APIs

#### Core Functions
- `gmacs.message(text)` - Display message in minibuffer
- `gmacs.set_variable(name, value)` - Set configuration variable
- `gmacs.register_command(name, function, description)` - Register custom command
- `gmacs.global_set_key(keyseq, command)` - Bind key to command

#### Package Management
- `gmacs.use_package(url, config)` - Declare package with optional configuration
- `gmacs.use_package(url, version)` - Declare package with specific version

#### Key Sequence Format
- `C-x` - Ctrl+x
- `M-x` - Meta+x (Alt+x)
- `C-x C-f` - Ctrl+x followed by Ctrl+f
- Standard key names: `return`, `tab`, `backspace`, `delete`, `up`, `down`, `left`, `right`

## 🎯 Default Key Bindings

| Key Binding | Command | Description |
|-------------|---------|-------------|
| `C-f` / `→` | forward-char | Move cursor forward |
| `C-b` / `←` | backward-char | Move cursor backward |
| `C-n` / `↓` | next-line | Move cursor down |
| `C-p` / `↑` | previous-line | Move cursor up |
| `C-d` | delete-char | Delete character at cursor |
| `Backspace` | backward-delete-char | Delete character before cursor |
| `C-x C-f` | find-file | Open file |
| `C-x C-s` | save-buffer | Save current buffer |
| `C-x C-w` | write-file | Save buffer to new file |
| `C-x C-c` | quit | Quit gmacs |
| `M-x` | execute-extended-command | Execute command by name |

## 📁 Project Structure

```
gmacs/
├── cmd/gmacs/          # Main executable
├── internal/
│   ├── buffer/         # Text buffer implementation
│   ├── command/        # Command system
│   ├── config/         # Lua configuration system
│   ├── cursor/         # Cursor management
│   ├── display/        # UI and terminal handling
│   ├── input/          # Keyboard input processing
│   ├── keymap/         # Key binding system
│   ├── package/        # Package management
│   └── window/         # Window management
├── examples/           # Configuration examples
├── design/            # Design documents
└── README.md
```

## 🔧 Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/config/
go test ./internal/package/

# Run with verbose output
go test -v ./...
```

### Package Development

gmacs supports external packages written in Go. Packages must implement the `Package` interface:

```go
type Package interface {
    GetInfo() PackageInfo
    Initialize() error
    Cleanup() error
    IsEnabled() bool
    Enable() error
    Disable() error
}
```

For Lua API extensions, also implement `LuaAPIExtender`:

```go
type LuaAPIExtender interface {
    Package
    ExtendLuaAPI(luaTable *lua.LTable, vm *lua.LState) error
    GetNamespace() string
}
```

### Configuration Loading Architecture (D案)

gmacs implements a sophisticated package loading system:

1. **Parse**: Extract `use_package` declarations from config file
2. **Download**: Fetch packages using `go get`
3. **Load**: Initialize packages and API extensions
4. **Execute**: Run configuration with full API available

This allows natural configuration syntax where package APIs are immediately available.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## 📄 License

[License information to be added]

## 🚧 Current Status

gmacs is under active development. Current features:

- ✅ Basic text editing
- ✅ Lua configuration system
- ✅ Package declaration and downloading
- ✅ Key binding system
- ✅ File I/O operations
- 🚧 Dynamic package loading (`.so` files)
- 🚧 Comprehensive package ecosystem
- 🚧 Advanced editing features

## 📚 Examples

See the `examples/` directory for:
- `init.lua.example` - Complete configuration example
- Package configuration patterns
- Custom command examples

---

**Note**: gmacs is inspired by GNU Emacs but is a separate implementation focused on simplicity and Go integration.