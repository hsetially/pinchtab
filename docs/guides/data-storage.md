# Data Storage Guide

Pinchtab stores configuration, state, and browser profiles on your local filesystem. This guide explains what files are saved, why, where they're located, and how storage locations changed in recent versions.

## What Files Does Pinchtab Store?

| File/Directory | Purpose | Configurable Via |
|----------------|---------|------------------|
| `profiles/<name>/` | Chrome browser profile (cookies, cache, localStorage, etc.) | `profiles.baseDir` + `profiles.defaultProfile` in config |
| `config.json` | Runtime configuration (port, token, headless mode, etc.) | `PINCHTAB_CONFIG` env var |
| `action_logs.json` | Profile action history and analytics | *(not currently configurable)* |
| `*.state.json` | Bridge state files (orchestrator state, etc.) | `server.stateDir` in config |

### Chrome Profile Directory

The largest and most important directory. Contains:
- **Cookies & Session Storage** — Login sessions, auth tokens
- **LocalStorage & IndexedDB** — Web app data
- **Cache** — Images, scripts, and other cached resources
- **Extensions** — If you've installed Chrome extensions via the Bridge API

**Why it exists:** Chrome requires a profile directory to run. Without it, you'd lose all session state between restarts.

**Size:** Can grow to 100MB+ depending on usage (cache, cookies, etc.).

### Configuration File

`config.json` stores runtime settings:
```json
{
  "server": {
    "port": "9867",
    "token": "your-secret-token",
    "stateDir": "/custom/path/state"
  },
  "instanceDefaults": {
    "mode": "headless"
  },
  "profiles": {
    "baseDir": "/custom/path/profiles",
    "defaultProfile": "default"
  }
}
```

**Why it exists:** Allows persistent configuration without environment variables.

### Action Logs

`action_logs.json` tracks browser actions for analytics:
- URL visits per profile
- Common hosts accessed
- Action timestamps

**Why it exists:** Provides usage analytics via the `/profiles/analytics` API endpoint.

### State Files

Internal bridge state (orchestrator, instance tracking, etc.).

**Why they exist:** Restore browser instances after restart (if `noRestore` is false).

## Storage Locations

### Current (v1.x+, After XDG Migration)

Pinchtab now uses **OS-native application data directories**:

| OS | Default Location |
|----|------------------|
| **Linux** | `~/.config/pinchtab/` (or `$XDG_CONFIG_HOME/pinchtab/`) |
| **macOS** | `~/Library/Application Support/pinchtab/` |
| **Windows** | `%APPDATA%\pinchtab\` (`C:\Users\YourName\AppData\Roaming\pinchtab\`) |

Inside that directory:
```
pinchtab/
├── profiles/               # Browser profiles
│   └── default/
├── config.json             # Configuration
├── action_logs.json        # Action history
└── *.state.json            # Bridge state files
```

**Why the change?** 
- **Linux Snap/AppArmor compatibility** — Security policies allow standard XDG directories but block arbitrary dotfolders under `$HOME`
- **OS conventions** — Follows platform-specific best practices
- **Better Windows support** — Uses proper `%APPDATA%` instead of a dotfolder

### Legacy (Pre-XDG Migration)

Previously, everything lived in `~/.pinchtab/`:
```
~/.pinchtab/
├── chrome-profile/
├── config.json
├── action_logs.json
└── *.state.json
```

**Backwards compatibility:** If you have an existing `~/.pinchtab/` directory with data, pinchtab will continue using it automatically (migration logic checks for the old location first).

## Customizing Storage Locations

### Via Environment Variables

Only `PINCHTAB_CONFIG` remains for storage-related path selection:

```bash
# Custom config file location
export PINCHTAB_CONFIG=/etc/pinchtab/config.json
pinchtab
```

### Via Configuration File

Set custom paths in `config.json`:
```json
{
  "server": {
    "stateDir": "/var/lib/pinchtab/state"
  },
  "profiles": {
    "baseDir": "/mnt/data/profiles",
    "defaultProfile": "default"
  }
}
```

**Priority order:** `PINCHTAB_CONFIG` chooses which config file to load. Storage paths inside that file control the actual directories.

## Migration from Legacy Location

If you're upgrading from an older version that used `~/.pinchtab/`, you have three options:

### Option 1: Automatic (Recommended)

Do nothing. Pinchtab will detect the old location and continue using it if:
- `~/.pinchtab/` exists
- The new OS-native location doesn't exist yet

### Option 2: Manual Migration

Move your data to the new location:

**Linux:**
```bash
mkdir -p ~/.config/pinchtab
mv ~/.pinchtab/* ~/.config/pinchtab/
rmdir ~/.pinchtab
```

**macOS:**
```bash
mkdir -p ~/Library/Application\ Support/pinchtab
mv ~/.pinchtab/* ~/Library/Application\ Support/pinchtab/
rmdir ~/.pinchtab
```

**Windows (PowerShell):**
```powershell
mkdir $env:APPDATA\pinchtab
mv ~/.pinchtab/* $env:APPDATA\pinchtab\
rmdir ~/.pinchtab
```

### Option 3: Stay on Legacy Location

Set your config file to point to the legacy location:
```json
{
  "server": {
    "stateDir": "~/.pinchtab"
  },
  "profiles": {
    "baseDir": "~/.pinchtab/profiles",
    "defaultProfile": "default"
  }
}
```

## Container Deployments

In Docker/containerized environments:

1. **Mount a volume** for persistence:
   ```bash
   docker run -v /host/pinchtab-data:/data \
              -e PINCHTAB_CONFIG=/data/config.json \
              pinchtab/pinchtab
   ```

2. **Set `HOME` environment variable** if needed:
   ```dockerfile
   ENV HOME=/app
   ```

3. **Store explicit paths in the mounted config file**.

## Security Considerations

### Profile Directory Contains Sensitive Data

- **Cookies & Sessions** — Can be used to impersonate logged-in users
- **LocalStorage** — May contain auth tokens, API keys
- **History & Cache** — Reveals browsing activity

**Recommendations:**
- Set restrictive permissions on profile directories, for example `chmod 700 ~/.config/pinchtab/profiles/default`
- Don't commit profile directories to version control
- Use separate profiles for different security contexts
- Consider encrypting the filesystem or using encrypted volumes

### Configuration File

`config.json` may contain:
- **`token`** — Used to authenticate API requests

**Recommendations:**
- Set restrictive permissions: `chmod 600 config.json`
- Use environment variables for tokens in production (don't hardcode in config files)

## Cleanup

To completely remove all pinchtab data:

**Linux/macOS:**
```bash
rm -rf ~/.config/pinchtab          # New location
rm -rf ~/Library/Application\ Support/pinchtab  # macOS
rm -rf ~/.pinchtab                 # Legacy location (if still exists)
```

**Windows:**
```powershell
rmdir /s $env:APPDATA\pinchtab
```

This will delete:
- All browser profiles and sessions
- Configuration
- State files
- Action logs

**Warning:** You'll lose all saved sessions, cookies, and browser state. Back up your `chrome-profile/` directory if you want to preserve login sessions.

## Troubleshooting

### "Permission denied" on Linux (Snap/AppArmor)

**Symptom:** Chrome fails to start with:
```
Failed to create SingletonLock: Permission denied (13)
```

**Cause:** Using an old pinchtab version that stores profiles in `~/.pinchtab` (blocked by Snap AppArmor).

**Solution:** Upgrade to the latest version (uses `~/.config/pinchtab` by default) or move profile storage under your configured `profiles.baseDir`.

See [Issue #98](https://github.com/pinchtab/pinchtab/issues/98) for details.

### "Config file not found" after upgrade

**Cause:** Upgraded from legacy `~/.pinchtab/` but pinchtab is now looking in the new location.

**Solution:** Copy your config:
```bash
cp ~/.pinchtab/config.json ~/.config/pinchtab/config.json  # Linux
```

Or use `PINCHTAB_CONFIG`:
```bash
export PINCHTAB_CONFIG=~/.pinchtab/config.json
```

### Profile directory size is huge

**Cause:** Chrome cache grows over time.

**Solution:** Clear the cache periodically:
```bash
rm -rf ~/.config/pinchtab/profiles/default/Default/Cache
rm -rf ~/.config/pinchtab/profiles/default/Default/Code\ Cache
```

Or point `profiles.baseDir` at a temporary directory for short-lived sessions.

## Further Reading

- [Configuration Reference](../references/configuration.md) — Full list of config options
- [CLI Quick Reference](../references/cli-quick-reference.md) — Command-line usage
- [Issue #98](https://github.com/pinchtab/pinchtab/issues/98) — XDG directory migration discussion
