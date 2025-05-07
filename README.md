# qng-mcp
The mcp server for qng supports the Model Context Protocol（MCP）protocol 


## use stdio mode

```json
{
  "mcpServers": {
    "github": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "qng_mcp_server",
        "-rpc",
        "http://127.0.0.1:8545"
      ]
    }
  }
}
```

```json
{
    "mcpServers": {
        "qngserver": {
            "command": "./qng_server/qng_server.exe",
            "args": [
                "-rpc",
                "http://127.0.0.1:8545/"
            ]
        }
    }
}

```
```bash
.\mcphost.exe  -m ollama:qwen2.5:3b --config .\conf\stdio.json --debug
```
## use sse mode
```json
{
    "mcpServers": {
        "qngserver": {
            "url": "http://localhost:8080/sse"
        }
    }
}
```
```bash
# start server
.\qng_server -t sse
.\mcphost.exe  -m ollama:qwen2.5:3b --config .\conf\sse.json --debug
```
