package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Qitmeer/qng/log"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// qng rpc url
var rpcUrl = "http://127.0.0.1:8545/"

// log level control
var logLevel = "info"

// current mcp server
var currentMcpServer = "localhost:8080"

// RPC timeout configuration
var rpcTimeout = 60 * time.Second

// HTTP client with timeout and connection pooling
var httpClient = &http.Client{
	Timeout: 60 * time.Second, // 增加全局超时到60秒
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		// 增加连接超时设置
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second, // 增加响应头超时到60秒
		// 添加连接保活设置
		DisableKeepAlives:  false,
		MaxConnsPerHost:    20,
		DisableCompression: false,
	},
}

// Mutex for protecting concurrent RPC calls
var rpcMutex sync.RWMutex

// Simple logging wrapper that can be replaced with QNG logging later
// To use the actual QNG logging library, replace this with:
// import "github.com/Qitmeer/qng/log"
// and replace all log.* calls with log.* calls

// JSONRPCRequest struct is used to construct JSON-RPC requests
type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// JSONRPCResponse struct is used to parse JSON-RPC responses
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result"`
	Error   *RPCError   `json:"error"`
}

// RPCError struct is used to parse JSON-RPC errors
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JsonRpcResponse constructs and sends a JSON-RPC request and returns the response
// with retry mechanism and timeout control
func JsonRpcResponse(rpcurl, method string, params []interface{}) ([]byte, error) {
	const maxRetries = 1
	var retryDelay time.Duration

	// 根据方法类型设置不同的重试延迟
	if method == "qng_getStateRoot" {
		retryDelay = 3 * time.Second // 状态根查询使用更长的重试间隔
	} else {
		retryDelay = 1 * time.Second
	}

	// 根据不同的RPC方法设置不同的超时时间
	var requestTimeout time.Duration
	switch method {
	case "qng_getBlockByOrder", "qng_getBlockByID", "qng_getBlockByNum":
		requestTimeout = 60 * time.Second // 区块查询需要更长时间
	case "qng_getRawTransactions", "qng_getMempool":
		requestTimeout = 50 * time.Second // 交易查询需要较长时间
	case "qng_getStateRoot":
		requestTimeout = 90 * time.Second // 状态根查询需要更长时间，这是最耗时的操作
	default:
		requestTimeout = 40 * time.Second // 其他方法使用默认超时
	}

	// 构建 JSON-RPC 请求体
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	// 将请求体编码为 JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		log.Error("Error marshaling JSON request:", err)
		return nil, err
	}

	// 使用读写锁保护并发请求
	rpcMutex.RLock()
	defer rpcMutex.RUnlock()

	// 记录请求开始
	log.Debug("Starting RPC request", "method", method, "timeout", requestTimeout, "req", string(requestBody))

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		// 对于状态根查询，如果重试则增加超时时间
		currentTimeout := requestTimeout
		if method == "qng_getStateRoot" && attempt > 0 {
			currentTimeout = requestTimeout + time.Duration(attempt*30)*time.Second
			log.Debug("Increased timeout for state root query", "attempt", attempt+1, "newTimeout", currentTimeout)
		}

		// 创建带超时的上下文，使用动态超时时间
		ctx, cancel := context.WithTimeout(context.Background(), currentTimeout)

		// 创建HTTP请求
		req, err := http.NewRequestWithContext(ctx, "POST", rpcurl, bytes.NewBuffer(requestBody))
		if err != nil {
			cancel()
			log.Error("Error creating HTTP request:", err)
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		// 发送请求
		resp, err := httpClient.Do(req)
		cancel() // 释放上下文资源

		if err != nil {
			lastErr = err
			// 检查是否是超时错误
			if ctx.Err() == context.DeadlineExceeded {
				log.Warn("RPC request timeout", "method", method, "timeout", requestTimeout, "attempt", attempt+1, "of", maxRetries)
			} else {
				log.Warn("HTTP request failed", "method", method, "attempt", attempt+1, "of", maxRetries, "error", err)
			}
			if attempt < maxRetries-1 {
				time.Sleep(retryDelay * time.Duration(attempt+1)) // 指数退避
				continue
			}
			return nil, err
		}

		// 读取响应
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = err
			log.Warn("Error reading response body, attempt", attempt+1, "of", maxRetries, "error:", err)
			if attempt < maxRetries-1 {
				time.Sleep(retryDelay * time.Duration(attempt+1))
				continue
			}
			return nil, err
		}

		// 检查HTTP状态码
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP error: %d", resp.StatusCode)
			log.Warn("HTTP error status", resp.StatusCode, "attempt", attempt+1, "of", maxRetries)
			if attempt < maxRetries-1 {
				time.Sleep(retryDelay * time.Duration(attempt+1))
				continue
			}
			return nil, lastErr
		}

		// 成功返回
		log.Debug("RPC request successful", "method", method, "attempt", attempt+1)
		return body, nil
	}

	return nil, fmt.Errorf("RPC request failed after %d attempts, last error: %v", maxRetries, lastErr)
}

// authKey is a custom context key for storing the auth token.
type authKey struct{}

// withAuthKey adds an auth key to the context.
func withAuthKey(ctx context.Context, auth string) context.Context {
	return context.WithValue(ctx, authKey{}, auth)
}

// authFromRequest extracts the auth token from the request headers.
func authFromRequest(ctx context.Context, r *http.Request) context.Context {
	return withAuthKey(ctx, r.Header.Get("Authorization"))
}

// authFromEnv extracts the auth token from the environment
func authFromEnv(ctx context.Context) context.Context {
	return withAuthKey(ctx, os.Getenv("API_KEY"))
}

type MCPServer struct {
	server *server.MCPServer
}

func handleQngWeb3Rpc(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	pn, ok := request.Params.Arguments["parameterNum"]
	if !ok {
		pn = 0
	}
	// return nil, fmt.Errorf("missing or invalid " + fmt.Sprintf("parameter%d", 1))
	name := strings.ReplaceAll(request.Params.Name, "qngserver__", "")
	methods, err := GetMethods()
	if err != nil {
		return nil, err
	}
	method, err := methods.FindName(name)
	if err != nil {
		return nil, err
	}
	count, ok := pn.(int)
	if !ok {
		count, _ = strconv.Atoi(pn.(string))
	}
	if method.Params != count {
		if method.Params < count {
			count = method.Params
		} else {
			return nil, fmt.Errorf("missing or invalid " + fmt.Sprintf("parameter%d", method.Params))
		}
	}
	params := make([]interface{}, 0)
	for i := 0; i < count; i++ {
		p, ok := request.Params.Arguments[fmt.Sprintf("parameter%d", i)]
		if !ok {
			return nil, fmt.Errorf("missing or invalid " + fmt.Sprintf("parameter%d", i))
		}
		// Optimize parameter type handling using type switch
		switch v := p.(type) {
		case string:
			if v == "true" {
				params = append(params, true)
			} else if v == "false" {
				params = append(params, false)
			} else if num, err := strconv.Atoi(v); err == nil {
				params = append(params, num)
			} else if fnum, err := strconv.ParseFloat(v, 64); err == nil {
				params = append(params, fnum)
			} else {
				params = append(params, v)
			}
		default:
			params = append(params, v)
		}
	}
	log.Debug("handleQngWeb3Rpc", "method", method.Name, "params", params)
	body, err := JsonRpcResponse(rpcUrl, method.Name, params)
	if err != nil {
		return nil, err
	}
	log.Debug("handleQngWeb3Rpc", "result", string(body))
	return mcp.NewToolResultText(string(body)), nil
}
func NewMCPServer() *MCPServer {
	mcpServer := server.NewMCPServer(
		"qng-mcp-server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithToolCapabilities(true),
	)

	// Core QNG blockchain tools with enhanced descriptions for better AI model understanding

	mcpServer.AddTool(mcp.NewTool("qng_get_block_by_order",
		mcp.WithDescription("QNG BLOCK RETRIEVAL: Fetches complete block information by block order/height. Returns block header, transactions, timestamps, hash, and all blockchain metadata. Use this tool when you need detailed information about a specific block in the QNG blockchain."),
		mcp.WithString("rpc_url",
			mcp.Description("QNG RPC endpoint URL (required). Format: http://ip:port/ or https://ip:port/. Example: http://127.0.0.1:8545/"),
			mcp.Required(),
		),
		mcp.WithNumber("block_order",
			mcp.Description("Block order/height number (required). Non-negative integer representing block position in chain. Example: 1000 for block 1000"),
			mcp.Required(),
		),
	), handleGetBlockByOrderTool)

	mcpServer.AddTool(mcp.NewTool("qng_get_block_count",
		mcp.WithString("rpc_url",
			mcp.Description("QNG RPC endpoint URL (required). Format: http://ip:port/ or https://ip:port/. Example: http://127.0.0.1:8545/"),
			mcp.Required(),
		),
		mcp.WithDescription("QNG BLOCKCHAIN HEIGHT: Returns the total number of blocks in the QNG blockchain. This gives you the current blockchain height/length. Use this tool to check how many blocks have been mined since genesis, or to get the latest block number."),
	), handleGetBlockCount)

	mcpServer.AddTool(mcp.NewTool("qng_get_stateroot",
		mcp.WithDescription("QNG STATE ROOT: Retrieves the stateroot hash of a specific QNG block. The state root is a cryptographic hash representing the complete blockchain state at that block (all account balances, smart contract states, etc.). Use this tool for state verification and blockchain analysis."),
		mcp.WithNumber("block_order",
			mcp.Description("Block order/height number (required). Non-negative integer representing block position in chain. Example: 1000 for block 1000"),
			mcp.Required(),
		),
		mcp.WithString("rpc_url",
			mcp.Description("QNG RPC endpoint URL (required). Format: http://ip:port/ or https://ip:port/. Example: http://127.0.0.1:8545/"),
			mcp.Required(),
		),
	), handleGetStateRoot)

	return &MCPServer{
		server: mcpServer,
	}
}

func parseAndGenerateGoCode() []mcp.Tool {
	ret := []mcp.Tool{}
	methods, err := GetMethods()
	if err != nil {
		log.Info("Error:", err)
		return nil
	}
	for _, m := range methods {
		toolOpt := make([]mcp.ToolOption, 0)
		toolOpt = append(toolOpt, mcp.WithDescription(m.Desc))
		toolOpt = append(toolOpt, mcp.WithNumber("parameterNum",
			mcp.Title("Parameter Count"),
			mcp.Description("The total number of parameters required for this QNG RPC method call. This must match the expected parameter count for the specific method being called."), mcp.Required()))
		for i := 0; i < m.Params; i++ {
			toolOpt = append(toolOpt,
				mcp.WithString(fmt.Sprintf("parameter%d", i),
					mcp.Title(fmt.Sprintf("Parameter %d", i+1)),
					mcp.Description(fmt.Sprintf("The %dth parameter for the %s method call. Refer to QNG RPC documentation for specific parameter requirements and data types.", i+1, m.Name)),
					mcp.Required()))
		}
		ret = append(ret, mcp.NewTool(
			m.Call,
			toolOpt...,
		))
	}

	return ret
}

func (s *MCPServer) ServeSSE(addr string) *server.SSEServer {
	return server.NewSSEServer(s.server,
		server.WithBaseURL(fmt.Sprintf("http://%s", addr)),
		server.WithSSEContextFunc(authFromRequest),
	)
}

func (s *MCPServer) ServeStdio() error {
	return server.ServeStdio(s.server, server.WithStdioContextFunc(authFromEnv))
}

// handleGetBlockByOrderTool handles the qng_get_block_by_order tool request.
func handleGetBlockByOrderTool(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	order, ok := request.Params.Arguments["block_order"]
	if !ok {
		return nil, fmt.Errorf("missing or invalid block_order parameter")
	}
	rpc, ok := request.Params.Arguments["rpc_url"]
	if !ok {
		log.Debug("handleGetBlockByOrderTool", "rpc_url", rpc)
		return nil, fmt.Errorf("missing or invalid rpc_url parameter")
	}
	body, err := JsonRpcResponse(rpc.(string), "qng_getBlockByOrder", []interface{}{order, true})
	if err != nil {
		return nil, err
	}
	log.Debug("handleGetBlockByOrderTool", "result", string(body))
	return mcp.NewToolResultText(string(body)), nil
}

// handleGetBlockCount handles the qng_get_block_count tool request.
func handleGetBlockCount(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	rpc, ok := request.Params.Arguments["rpc_url"]
	if !ok {
		log.Debug("handleGetBlockCount", "rpc_url", rpc)
		return nil, fmt.Errorf("missing or invalid rpc_url parameter")
	}
	body, err := JsonRpcResponse(rpc.(string), "qng_getBlockCount", []interface{}{})
	if err != nil {
		return nil, err
	}
	log.Debug("handleGetBlockCount", "result", string(body))
	return mcp.NewToolResultText(string(body)), nil
}

// handleGetStateRoot handles the qng_get_block_stateroot tool request.
func handleGetStateRoot(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	order, ok := request.Params.Arguments["block_order"]
	if !ok {
		log.Debug("handleGetStateRoot", "block_order", order)
		return nil, fmt.Errorf("missing or invalid block_order parameter")
	}

	rpc, ok := request.Params.Arguments["rpc_url"]
	if !ok {
		log.Debug("handleGetStateRoot", "rpc_url", rpc)
		return nil, fmt.Errorf("missing or invalid rpc_url parameter")
	}
	log.Debug("handleGetStateRoot", "block_order", order, "rpc_url", rpc)
	orderNum := int64(0)
	switch order.(type) {
	case int64:
		orderNum = order.(int64)
	case string:
		orderNum, _ = strconv.ParseInt(order.(string), 10, 64)
	case float64:
		orderNum = int64(order.(float64))
	default:
		log.Debug("handleGetStateRoot", "block_order", order)
		return nil, fmt.Errorf("missing or invalid block_order parameter")
	}
	body, err := JsonRpcResponse(rpc.(string), "qng_getStateRoot", []interface{}{orderNum, true})
	if err != nil {
		log.Debug("JsonRpcResponse", "error", err)
		return nil, err
	}
	log.Debug("handleGetStateRoot", "result", string(body), "rpc_url", rpc)
	return mcp.NewToolResultText(string(body)), nil
}

func main() {
	var transport string
	var timeoutSeconds int
	flag.StringVar(&transport, "t", "stdio", "Transport type (stdio or sse)")
	flag.StringVar(&rpcUrl, "rpc", "http://127.0.0.1:8545/", "qng rpc url")
	flag.StringVar(&logLevel, "loglevel", "info", "Log level (debug, info, warn, error)")
	flag.StringVar(&currentMcpServer, "mcp", "localhost:8080", "mcp server url")
	flag.IntVar(&timeoutSeconds, "timeout", 60, "RPC request timeout in seconds (default: 60)")
	flag.StringVar(
		&transport,
		"transport",
		"stdio",
		"Transport type (stdio or sse)",
	)
	flag.Parse()

	// 设置RPC超时
	rpcTimeout = time.Duration(timeoutSeconds) * time.Second
	httpClient.Timeout = rpcTimeout
	// 设置日志等级为 DEBUG
	lvl, err := log.LvlFromString(logLevel)
	if err != nil {
		log.Error("Error: Invalid log level:", err)
		os.Exit(1)
	}
	log.Glogger().Verbosity(lvl)
	// Print usage instructions
	log.Info("\nUsage:")
	log.Info("  -t, --transport  Transport type (stdio or sse)")
	log.Info("  --rpc            QNG Web3 RPC URL")
	log.Info("  --loglevel       Log level (debug, info, warn, error)")
	log.Info("  --timeout        RPC request timeout in seconds (default: 60)")
	log.Info("\nExample:")
	log.Info("  ./qng-mcp -t stdio --rpc http://127.0.0.1:8545/ --loglevel debug --mcp localhost:8080 --timeout 90")

	// Print system status
	log.Info("Starting QNG MCP Server...", "logLevel", logLevel)
	log.Debug("Transport type", "Transport type", transport)
	log.Debug("QNG Node Web3 RPC URL", "QNG Node Web3 RPC URL", rpcUrl)

	// Check configuration
	if rpcUrl == "" {
		log.Error("Error: RPC URL is not configured. Please provide a valid RPC URL using the -rpc flag.")
		os.Exit(1)
	}

	s := NewMCPServer()

	switch transport {
	case "stdio":
		log.Info("Running in stdio mode...")
		if err := s.ServeStdio(); err != nil {
			log.Error("Server error: %v", err)
			os.Exit(1)
		}
	case "sse":
		log.Info("Running in SSE mode...")
		sseServer := s.ServeSSE(currentMcpServer)
		log.Info("SSE server listening on :8080")
		if err := sseServer.Start(":8080"); err != nil {
			log.Error("Server error: %v", err)
			os.Exit(1)
		}
	default:
		log.Error(
			"Invalid transport type: %s. Must be 'stdio' or 'sse'",
			transport,
		)
		os.Exit(1)
	}
}
