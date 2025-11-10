# Vinegar Multi-Desktop Fork

A fork of [Vinegar](https://github.com/vinegarhq/vinegar) that adds support for multiple wine desktops, allowing Roblox Studio instances to run in isolated environments.

## 🚀 New Features

### Multi-Desktop Support
- **Multiple Wine Prefixes**: Each desktop runs in its own isolated wine prefix
- **Desktop Assignment**: Automatic or manual desktop assignment for Roblox instances
- **Environment Isolation**: Separate environment variables per desktop
- **Backward Compatibility**: Fully compatible with existing single-desktop setups

### Command Line Interface
- `--desktop <name>` or `-d <name>`: Specify which desktop to use
- `--list-desktops`: List all available desktops
- `--create-desktop <name>`: Create a new desktop (future feature)

## 📋 Configuration

### Enabling Multi-Desktop

Add to your `config.toml`:

```toml
[desktop_manager]
enabled = true
auto_assign = true
max_desktops = 10
default_desktop = "desktop-1"
desktop_prefix = "desktop-"
isolation_level = "basic"
```

### Desktop-Specific Configuration

```toml
[[desktop_manager.desktops]]
name = "development"
env = { "WINEDEBUG" = "+all", "VINEGAR_DESKTOP_TYPE" = "development" }

[[desktop_manager.desktops]]
name = "testing"
env = { "WINEDEBUG" = "warn+err", "VINEGAR_DESKTOP_TYPE" = "testing" }
```

## 🎯 Usage Examples

### Basic Usage (Backward Compatible)
```bash
# Uses default desktop or single prefix mode
vinegar-multi-desktop
```

### Specific Desktop
```bash
# Launch on development desktop
vinegar-multi-desktop --desktop development

# Launch on testing desktop
vinegar-multi-desktop -d testing
```

### List Available Desktops
```bash
vinegar-multi-desktop --list-desktops
```

### Roblox Studio Arguments
```bash
# Launch with specific desktop and Studio arguments
vinegar-multi-desktop --desktop development -protocolString "roblox-studio:1/launch"
```

## 🏗️ Architecture Changes

### Configuration System
- Added `DesktopManager` struct to `internal/config/config.go`
- New methods: `PrefixForDesktop()`, `GetAvailableDesktops()`, `AssignDesktop()`
- Desktop-specific environment variable support

### Application Structure
- Modified `app` struct to support multiple prefixes (`prefixes map[string]*wine.Prefix`)
- Added `GetPrefixForDesktop()` and `AssignDesktopForInstance()` methods
- Enhanced command line parsing for desktop selection

### Bootstrapper Updates
- Added `desktopName` and `currentPfx` fields to `bootstrapper` struct
- New `runOnDesktop()` method for desktop-specific execution
- Updated all prefix operations to use desktop-specific prefixes

## 📁 Directory Structure

Multi-desktop mode creates separate wine prefixes:

```
~/.local/share/vinegar/prefixes/
├── studio/           # Default/legacy prefix
├── desktop-1/        # Auto-assigned desktop
├── desktop-2/        # Auto-assigned desktop
├── development/      # Custom desktop
└── testing/          # Custom desktop
```

## 🔧 Configuration Options

### Desktop Manager Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `enabled` | bool | `false` | Enable multi-desktop support |
| `auto_assign` | bool | `true` | Automatically assign desktops |
| `max_desktops` | int | `10` | Maximum number of desktops |
| `default_desktop` | string | `"desktop-1"` | Default desktop name |
| `desktop_prefix` | string | `"desktop-"` | Prefix for auto-generated names |
| `isolation_level` | string | `"basic"` | Isolation level (full/basic/minimal) |

### Desktop Configuration

Each desktop can have:
- `name`: Desktop identifier
- `path`: Custom prefix path (optional)
- `env`: Desktop-specific environment variables

## 🔄 Migration from Original Vinegar

1. **No Breaking Changes**: Existing configurations work unchanged
2. **Enable Multi-Desktop**: Add `desktop_manager.enabled = true` to config
3. **Configure Desktops**: Add desktop definitions as needed
4. **Update Launch Commands**: Use `--desktop` flag for specific desktops

## 🐛 Troubleshooting

### Common Issues

1. **Desktop Not Found**: Ensure desktop is defined in configuration
2. **Permission Errors**: Check wine prefix directory permissions
3. **Environment Conflicts**: Verify desktop-specific environment variables

### Debug Mode

Enable debug logging:
```toml
debug = true
```

### Environment Variables

Multi-desktop mode sets these automatically:
- `VINEGAR_DESKTOP`: Current desktop name
- `VINEGAR_DESKTOP_ID`: Desktop ID (number without prefix)

## 🚧 Future Enhancements

- [ ] Desktop creation/deletion commands
- [ ] GUI desktop manager
- [ ] Load balancing across desktops
- [ ] Desktop templates
- [ ] Process monitoring per desktop

## 📝 Development Notes

### Building
```bash
go build -o vinegar-multi-desktop ./cmd/vinegar
```

### Testing
```bash
# Test basic functionality
./vinegar-multi-desktop --list-desktops

# Test desktop assignment
./vinegar-multi-desktop --desktop development
```

## 🤝 Contributing

This fork maintains compatibility with the original Vinegar project while adding multi-desktop functionality. Contributions are welcome!

## 📄 License

Same license as the original Vinegar project.

---

**Note**: This is a fork of Vinegar with additional multi-desktop features. For the original project, visit [vinegarhq/vinegar](https://github.com/vinegarhq/vinegar).
