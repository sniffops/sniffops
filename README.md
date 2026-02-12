# SniffOps

> **Self-hosted observability platform for AI-driven Kubernetes operations**

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?logo=go)](https://go.dev)
[![MCP](https://img.shields.io/badge/MCP-enabled-green)](https://modelcontextprotocol.io)

SniffOps tracks and analyzes every Kubernetes command executed by AI agents (like Claude Code) through the Model Context Protocol (MCP). It provides real-time trace collection, risk evaluation, and cost estimation for AI-driven infrastructure operations.

---

## 🎯 Features

- **🔍 Trace Collection**: Capture all K8s operations performed by AI agents
- **⚠️ Risk Evaluation**: Automatic risk-level tagging (low/medium/high/critical)
- **💰 Cost Tracking**: Estimate LLM token usage and costs per operation
- **🛡️ Security**: Sanitize sensitive data (API keys, secrets) from traces
- **📊 Web Dashboard**: Timeline view and detailed trace inspection
- **🚀 Single Binary**: Zero dependencies, embedded web UI

---

## 📦 Installation

### Prerequisites

- **Go 1.21+** (for building from source)
- **Kubernetes cluster** with valid `kubeconfig`
- **Claude Code** (or any MCP-compatible client)

### Install from source

```bash
# Clone the repository
git clone https://github.com/sniffops/sniffops.git
cd sniffops

# Build the binary
make build

# Install to $GOPATH/bin
make install
```

### Install with `go install`

```bash
go install github.com/sniffops/sniffops/cmd/sniffops@latest
```

---

## 🚀 Quick Start

### 1. Register SniffOps as MCP Server

Add SniffOps to Claude Code's MCP configuration:

```bash
claude mcp add sniffops -- sniffops serve
```

Or manually edit `~/.claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "sniffops": {
      "command": "sniffops",
      "args": ["serve"]
    }
  }
}
```

### 2. Start Web UI (Optional)

In a separate terminal, start the web dashboard:

```bash
sniffops web --port 3000
```

Open http://localhost:3000 in your browser to view traces.

### 3. Use AI with Kubernetes

Ask Claude Code to perform K8s operations:

```
"Show me all pods in the production namespace"
"Scale my nginx deployment to 3 replicas"
"Check logs of pod xyz-123"
```

All operations will be automatically traced and stored in `~/.sniffops/traces.db`.

---

## 🛠️ Available Tools

SniffOps provides the following MCP tools:

| Tool | Description | Risk Level |
|------|-------------|:----------:|
| `sniff_get` | Get K8s resources (pods, deployments, etc.) | 🟢 Low |
| `sniff_logs` | Retrieve pod logs | 🟢 Low |
| `sniff_apply` | Create/update resources | 🟡 Medium |
| `sniff_delete` | Delete resources | 🔴 High |
| `sniff_scale` | Scale deployments | 🔴 High |
| `sniff_exec` | Execute commands in pods | 🔴 High |
| `sniff_traces` | Query stored traces | 🟢 Low |
| `sniff_stats` | View usage statistics | 🟢 Low |

---

## 📊 Architecture

SniffOps acts as an MCP server that sits between AI clients (like Claude Code) and the Kubernetes API:

```
┌─────────────┐         ┌──────────────┐         ┌─────────┐
│ Claude Code │ ←MCP→   │  SniffOps    │ ←API→   │   K8s   │
│ (AI Agent)  │         │ MCP Server   │         │ Cluster │
└─────────────┘         └──────┬───────┘         └─────────┘
                               │
                               ↓
                        ┌──────────────┐
                        │ SQLite Store │
                        │  (traces.db) │
                        └──────────────┘
```

For detailed architecture, see [docs/architecture.md](docs/architecture.md).

---

## 🧪 Development

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Format Code

```bash
make fmt
```

### Clean Build Artifacts

```bash
make clean
```

---

## 📝 Configuration

SniffOps uses the following default paths:

- **Traces database**: `~/.sniffops/traces.db`
- **Kubeconfig**: `~/.kube/config` (or `$KUBECONFIG`)
- **Web UI port**: `3000` (configurable with `--port`)

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

---

## 📄 License

Apache License 2.0 - see [LICENSE](LICENSE) file for details.

---

## 🔗 Links

- **Documentation**: [docs/](docs/)
- **Issue Tracker**: [GitHub Issues](https://github.com/sniffops/sniffops/issues)
- **MCP Protocol**: https://modelcontextprotocol.io

---

**Built with ❤️ for AI-driven DevOps**
