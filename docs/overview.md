# PinchTab

Welcome to PinchTab — browser control for AI agents, scripts, and automation workflows.

## What is PinchTab?

PinchTab is a **standalone HTTP server** that gives you direct control over a Chrome browser. Any AI agent can use the CLI or HTTP API.

**CLI example:**
```bash
# Navigate
pinchtab nav https://pinchtab.com

# Get interactive elements
pinchtab snap -i -c

# Click element by ref
pinchtab click e5
```

**HTTP example (realistic flow):**
```bash
# 1. Navigate to URL (returns tabId)
TAB=$(curl -s -X POST http://localhost:9867/tab \
  -d '{"action":"new","url":"https://pinchtab.com"}' | jq -r '.tabId')

# 2. Get page structure
curl -s "http://localhost:9867/snapshot?tabId=$TAB&filter=interactive" | jq

# 3. Click element using the tabId
curl -s -X POST http://localhost:9867/action \
  -d "{\"kind\":\"click\",\"ref\":\"e5\",\"tabId\":\"$TAB\"}"
```

---

## Characteristics

- **Tab-Centric** — Everything revolves around tabs, not URLs
- **Stateful** — Sessions persist between requests. Log in once, stay logged in across restarts
- **Token Inexpensive** — Text extraction at 800 tokens/page (5-13x cheaper than full snapshots)
- **Flexible Modes** — Run headless, headed, with browser profiles, or connect to external Chrome via CDP
- **Monitoring & Control** — Tab locking for multi-agent safety, stealth mode for bot detection bypass

---

## Features

- 🌲 **Accessibility Tree** — Structured DOM with stable refs (e0, e1...) for click, type, read. No coordinate guessing.
- 🎯 **Smart Filters** — `?filter=interactive` returns only buttons, links, inputs. Fewer tokens per snapshot.
- 🕵️ **Stealth Mode** — Patches `navigator.webdriver`, spoofs UA, hides automation flags for bot detection bypass.
- 📝 **Text Extraction** — Readability mode (clean) or raw (full HTML). Choose based on workflow.
- 🖱️ **Direct Actions** — Click, type, fill, press, focus, hover, select, scroll by ref or selector.
- ⚡ **JavaScript Execution** — Run arbitrary JS in any tab. Escape hatch for workflow gaps.
- 📸 **Screenshots** — JPEG output with quality control.
- 📄 **PDF Export** — Full pages to PDF with headers, footers, landscape mode.
- 🎭 **Multi-Tab** — Create, switch, close tabs. Work with multiple pages concurrently.

---

## Support & Community

- **GitHub Issues** — https://github.com/pinchtab/pinchtab/issues
- **Discussions** — https://github.com/pinchtab/pinchtab/discussions
- **Twitter/X** — [@pinchtabdev](https://x.com/pinchtabdev)

---

## License

[MIT](https://github.com/pinchtab/pinchtab?tab=MIT-1-ov-file#readme)
