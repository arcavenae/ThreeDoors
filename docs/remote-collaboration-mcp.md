# MCP Bridge Integration Guide

Connect a remote Claude Code session to a multiclaude system using the MCP (Model Context Protocol) bridge server. Once configured, multiclaude tools appear natively in the remote Claude Code session's tool list.

This guide complements [Remote Collaboration Guide](remote-collaboration.md), which covers SSH-based CLI access. The MCP bridge provides a structured, programmatic interface instead of raw CLI commands.

## Quick Start

The minimum steps to get a remote Claude Code session connected to multiclaude:

1. Build the bridge binary on the multiclaude host:
   ```bash
   cd /path/to/ThreeDoors
   just build-mcp-bridge
   sudo cp bin/multiclaude-mcp-bridge /usr/local/bin/
   ```

2. Add this to your remote Claude Code session's `.claude/settings.json`:
   ```json
   {
     "mcpServers": {
       "multiclaude-remote": {
         "command": "ssh",
         "args": ["user@multiclaude-host", "/usr/local/bin/multiclaude-mcp-bridge"]
       }
     }
   }
   ```

3. Restart Claude Code. The multiclaude tools (`multiclaude_status`, `multiclaude_worker_list`, etc.) should appear in the tool list.

4. Verify on the host:
   ```bash
   ./scripts/verify-mcp-bridge.sh
   ```

## Prerequisites

- **SSH access** to the machine running multiclaude (key-based auth required — password auth blocks MCP stdio transport)
- **multiclaude** installed and daemon running on the host
- **Go toolchain** on the host to build the bridge binary (or a pre-built binary)
- **Claude Code** on the remote machine (the machine you're working from)

## SSH Transport Setup

MCP's SSH stdio transport is a native Claude Code capability. Claude Code spawns an SSH process, and the remote binary communicates via stdin/stdout using JSON-RPC.

### 1. Build and Install the Bridge Binary

On the multiclaude host:

```bash
cd /path/to/ThreeDoors
just build-mcp-bridge
```

This produces `bin/multiclaude-mcp-bridge`. Install it somewhere in PATH:

```bash
sudo cp bin/multiclaude-mcp-bridge /usr/local/bin/
```

Or install directly via Go:

```bash
go install github.com/arcavenae/ThreeDoors/cmd/multiclaude-mcp-bridge@latest
```

### 2. Verify SSH Key Auth Works

The MCP transport uses SSH in non-interactive mode — password prompts will hang the connection. Verify key-based auth works:

```bash
# From your remote machine:
ssh -o BatchMode=yes user@multiclaude-host "echo ok"
```

If this fails, set up SSH keys before proceeding:

```bash
ssh-copy-id user@multiclaude-host
```

### 3. Verify the Binary Is Accessible Over SSH

SSH sessions may not load your full shell profile. Verify the binary is reachable:

```bash
ssh user@multiclaude-host "multiclaude-mcp-bridge --help"
```

If `command not found`, use the full path in the MCP configuration:

```bash
ssh user@multiclaude-host "/usr/local/bin/multiclaude-mcp-bridge --help"
```

## Settings Configuration

### Project-Level Settings (Recommended)

Add to `.claude/settings.json` in your project directory. This scopes the MCP server to the project:

```json
{
  "mcpServers": {
    "multiclaude-remote": {
      "command": "ssh",
      "args": ["user@multiclaude-host", "/usr/local/bin/multiclaude-mcp-bridge"]
    }
  }
}
```

### User-Level Settings

Add to `~/.claude/settings.json` to make the MCP bridge available across all projects:

```json
{
  "mcpServers": {
    "multiclaude-remote": {
      "command": "ssh",
      "args": ["user@multiclaude-host", "/usr/local/bin/multiclaude-mcp-bridge"]
    }
  }
}
```

### Using an SSH Config Alias

If you have an SSH config entry for the multiclaude host:

```
# ~/.ssh/config
Host mc-host
    HostName 192.168.1.100
    User skippy
    ServerAliveInterval 60
    ServerAliveCountMax 3
```

Use the alias in the MCP config:

```json
{
  "mcpServers": {
    "multiclaude-remote": {
      "command": "ssh",
      "args": ["mc-host", "/usr/local/bin/multiclaude-mcp-bridge"]
    }
  }
}
```

### Available Tools

Once configured, these tools appear in your Claude Code session:

| Tool | Description |
|------|-------------|
| `multiclaude_status` | System status: agents, states, worker count |
| `multiclaude_worker_list` | Active workers with names, tasks, creation times |
| `multiclaude_message_list` | Pending inter-agent messages |
| `multiclaude_message_read` | Full content of a specific message (requires `message_id`) |
| `multiclaude_repo_history` | Recent task history (optional `count` and `status` filters) |

## Verification

### Automated Verification

Run the verification script on the multiclaude host:

```bash
./scripts/verify-mcp-bridge.sh
```

This checks:
1. The bridge binary exists and is executable
2. The multiclaude daemon is running
3. The MCP `initialize` handshake succeeds
4. All expected tools are listed

Specify a custom binary path:

```bash
./scripts/verify-mcp-bridge.sh --binary /opt/multiclaude/mcp-bridge
```

### Manual Verification

Test the MCP handshake by piping a JSON-RPC request to the binary:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | multiclaude-mcp-bridge
```

Expected response (formatted):

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": { "tools": {} },
    "serverInfo": {
      "name": "multiclaude-mcp-bridge",
      "version": "dev"
    }
  }
}
```

### Verify Tools Over SSH

From your remote machine, test the full SSH transport:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | ssh user@multiclaude-host multiclaude-mcp-bridge
```

The second line of output should list all five tools.

## Cloudflare Tunnel (Optional)

For NAT-traversal scenarios where direct SSH is not possible (e.g., the multiclaude host is behind a NAT or firewall without port forwarding), Cloudflare Tunnel provides an alternative transport.

### How It Works

Cloudflare Tunnel (`cloudflared`) creates an outbound-only connection from your host to Cloudflare's edge network. Remote clients connect through Cloudflare without needing direct network access to the host.

### Setup Steps

1. **Install cloudflared** on the multiclaude host:
   ```bash
   brew install cloudflared
   ```

2. **Authenticate with Cloudflare** (one-time):
   ```bash
   cloudflared tunnel login
   ```

3. **Create a tunnel**:
   ```bash
   cloudflared tunnel create multiclaude-mcp
   ```

4. **Configure the tunnel** to proxy SSH traffic. Create `~/.cloudflared/config.yml`:
   ```yaml
   tunnel: multiclaude-mcp
   credentials-file: /Users/you/.cloudflared/<tunnel-id>.json

   ingress:
     - hostname: mcp.yourdomain.com
       service: ssh://localhost:22
     - service: http_status:404
   ```

5. **Route DNS**:
   ```bash
   cloudflared tunnel route dns multiclaude-mcp mcp.yourdomain.com
   ```

6. **Start the tunnel**:
   ```bash
   cloudflared tunnel run multiclaude-mcp
   ```

7. **Configure the remote machine's SSH** to use cloudflared as a proxy. Add to `~/.ssh/config`:
   ```
   Host mcp-tunnel
       HostName mcp.yourdomain.com
       User your-user
       ProxyCommand cloudflared access ssh --hostname %h
   ```

8. **Update the MCP settings** to use the tunnel hostname:
   ```json
   {
     "mcpServers": {
       "multiclaude-remote": {
         "command": "ssh",
         "args": ["mcp-tunnel", "/usr/local/bin/multiclaude-mcp-bridge"]
       }
     }
   }
   ```

For full documentation, see the [Cloudflare Tunnel docs](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/).

## Tailscale (Optional)

Tailscale's mesh VPN provides a simpler alternative when both machines can run the Tailscale client. It assigns each machine a stable IP on a private WireGuard mesh network, removing the need for port forwarding, dynamic DNS, or tunnel proxies.

### How It Works

Both machines join the same Tailscale network (tailnet). Each gets a stable `100.x.y.z` IP address. SSH works directly between the Tailscale IPs with no NAT or firewall configuration needed.

### Setup Steps

1. **Install Tailscale** on both machines:
   ```bash
   # macOS
   brew install tailscale

   # Linux
   curl -fsSL https://tailscale.com/install.sh | sh
   ```

2. **Authenticate both machines**:
   ```bash
   sudo tailscale up
   ```

3. **Find the multiclaude host's Tailscale IP**:
   ```bash
   tailscale ip -4
   # Example output: 100.64.0.5
   ```

4. **Enable Tailscale SSH** (optional — uses Tailscale's built-in SSH server, no OpenSSH needed):
   ```bash
   sudo tailscale up --ssh
   ```

5. **Configure MCP settings** with the Tailscale IP or hostname:
   ```json
   {
     "mcpServers": {
       "multiclaude-remote": {
         "command": "ssh",
         "args": ["100.64.0.5", "/usr/local/bin/multiclaude-mcp-bridge"]
       }
     }
   }
   ```

   Or using Tailscale's MagicDNS hostname:
   ```json
   {
     "mcpServers": {
       "multiclaude-remote": {
         "command": "ssh",
         "args": ["mc-host.tail12345.ts.net", "/usr/local/bin/multiclaude-mcp-bridge"]
       }
     }
   }
   ```

### Advantages Over Direct SSH

- No port forwarding or firewall rules needed
- Works across NATs, cellular networks, and changing IPs
- Built-in key management via Tailscale's identity system
- MagicDNS provides stable hostnames

For full documentation, see the [Tailscale docs](https://tailscale.com/kb/).

## Troubleshooting

### SSH Key Problems

**Symptom:** Claude Code hangs when starting, or the MCP server never connects.

**Cause:** SSH is prompting for a password, which blocks the stdio transport.

**Fix:**
```bash
# Verify key-based auth works in batch mode (no prompts)
ssh -o BatchMode=yes user@host "echo ok"

# If it fails, set up key auth:
ssh-copy-id user@host

# Verify the key is loaded in your agent:
ssh-add -l
```

### multiclaude Not in PATH Over SSH

**Symptom:** MCP bridge returns errors like `command multiclaude failed: exec: "multiclaude": executable file not found in $PATH`.

**Cause:** SSH non-interactive sessions load a minimal environment. The shell profile that adds multiclaude to PATH may not be sourced.

**Fix:** Use the full path to `multiclaude` in the bridge binary's environment, or create a symlink:
```bash
sudo ln -s /path/to/multiclaude /usr/local/bin/multiclaude
```

Alternatively, set the PATH in `.ssh/environment` on the host (requires `PermitUserEnvironment yes` in `sshd_config`):
```
PATH=/usr/local/bin:/usr/bin:/bin:/path/to/multiclaude/parent
```

### MCP Handshake Failures

**Symptom:** Claude Code reports "MCP server failed to initialize" or similar.

**Diagnosis:**
```bash
# Test the handshake manually:
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | ssh user@host multiclaude-mcp-bridge

# Check for stderr output (errors go to stderr, responses to stdout):
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ssh user@host multiclaude-mcp-bridge 2>/tmp/mcp-stderr.log
cat /tmp/mcp-stderr.log
```

**Common causes:**
- Binary not found (see PATH issue above)
- Binary built for wrong architecture (verify with `file multiclaude-mcp-bridge`)
- Permission denied (verify with `ls -la /usr/local/bin/multiclaude-mcp-bridge`)

### Firewall Blocking SSH

**Symptom:** `ssh: connect to host X port 22: Connection refused` or timeouts.

**Diagnosis:**
```bash
# Check if port 22 is reachable:
nc -zv multiclaude-host 22

# Check if sshd is running on the host:
ssh user@host "systemctl status sshd"  # Linux
ssh user@host "launchctl list com.openssh.sshd"  # macOS
```

**Fixes:**
- Open port 22 in the firewall/security group
- Use Cloudflare Tunnel (see above) to bypass firewall restrictions
- Use Tailscale (see above) to create a mesh VPN that bypasses NAT/firewalls

### Bridge Binary Crashes on Startup

**Symptom:** The SSH connection closes immediately with no output.

**Diagnosis:**
```bash
# Run the binary directly on the host to see errors:
echo '' | multiclaude-mcp-bridge
# If it crashes, the error goes to stderr

# Check if it was built for the right OS/architecture:
file /usr/local/bin/multiclaude-mcp-bridge
```

### Stale SSH Connections

**Symptom:** MCP tools suddenly stop responding, then work again after restarting Claude Code.

**Cause:** The underlying SSH connection dropped but Claude Code hasn't detected it yet.

**Fix:** Add keepalive settings to your SSH config:
```
Host mc-host
    ServerAliveInterval 15
    ServerAliveCountMax 3
```

This makes SSH send a keepalive every 15 seconds and drop the connection after 3 missed responses (45 seconds), allowing Claude Code to detect the failure and reconnect.
