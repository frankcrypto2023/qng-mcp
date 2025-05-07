package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// qng rpc url
var rpcUrl = "http://127.0.0.1:8545/"

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
func JsonRpcResponse(method string, params []interface{}) ([]byte, error) {

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
		// fmt.Println("Error marshaling JSON request:", err)
		return nil, err
	}

	// 发送 HTTP POST 请求
	resp, err := http.Post(rpcUrl, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		// fmt.Println("Error sending HTTP request:", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// fmt.Println("Error reading response body:", err)
		return nil, err
	}
	return body, nil

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
	body, err := JsonRpcResponse(method.Name, params)
	if err != nil {
		return nil, err
	}
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

	mcpServer.AddTool(mcp.NewTool("get_block_by_order",
		mcp.WithDescription("Retrieves a qng block by its order"),
		mcp.WithNumber("order",
			mcp.Description("Order of the qng block to retrieve"),
			mcp.Required(),
		),
	), handleGetBlockByOrderTool)

	mcpServer.AddTool(mcp.NewTool("get_block_count",
		mcp.WithDescription("Retrieves a qng block total count"),
	), handleGetBlockCount)

	mcpServer.AddTool(mcp.NewTool("get_block_stateroot",
		mcp.WithDescription("Retrieves a qng block stateroot by its order"),
		mcp.WithNumber("order",
			mcp.Description("Order of the qng block stateroot to retrieve"),
			mcp.Required(),
		),
	), handleGetStateRoot)
	tools := parseAndGenerateGoCode()
	for _, tool := range tools {
		mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return handleQngWeb3Rpc(ctx, request)
		})
	}
	return &MCPServer{
		server: mcpServer,
	}
}

func parseAndGenerateGoCode() []mcp.Tool {
	ret := []mcp.Tool{}
	methods, err := GetMethods()
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	for _, m := range methods {
		toolOpt := make([]mcp.ToolOption, 0)
		toolOpt = append(toolOpt, mcp.WithDescription(m.Desc))
		toolOpt = append(toolOpt, mcp.WithNumber("parameterNum",
			mcp.Title("how many params of this method"),
			mcp.Description("how many params of this method,the first param is the tool function call"), mcp.Required()))
		for i := 0; i < m.Params; i++ {
			toolOpt = append(toolOpt,
				mcp.WithString(fmt.Sprintf("parameter%d", i),
					mcp.Title(fmt.Sprintf("params %d", i)),
					mcp.Description(fmt.Sprintf("params %d of this method,the %d param of the method call", i, i)),
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

// handleGetBlockByOrderTool handles the qng_getBlockByOrder tool request.
func handleGetBlockByOrderTool(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	order, ok := request.Params.Arguments["order"]
	if !ok {
		return nil, fmt.Errorf("missing or invalid order")
	}
	body, err := JsonRpcResponse("qng_getBlockByOrder", []interface{}{order, true})
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(body)), nil
}

// handleGetBlockCount handles the qng_getBlockCount tool request.
func handleGetBlockCount(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	body, err := JsonRpcResponse("qng_getBlockCount", []interface{}{})
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(body)), nil
}

// handleGetStateRoot handles the qng_getStateRoot tool request.
func handleGetStateRoot(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	order, ok := request.Params.Arguments["order"]
	if !ok {
		return nil, fmt.Errorf("missing or invalid order")
	}
	body, err := JsonRpcResponse("qng_getStateRoot", []interface{}{order, true})
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(body)), nil
}

func main() {
	var transport string
	flag.StringVar(&transport, "t", "stdio", "Transport type (stdio or sse)")
	flag.StringVar(&rpcUrl, "rpc", "http://127.0.0.1:8545/", "qng rpc url")
	flag.StringVar(
		&transport,
		"transport",
		"stdio",
		"Transport type (stdio or sse)",
	)
	flag.Parse()

	s := NewMCPServer()

	switch transport {
	case "stdio":
		if err := s.ServeStdio(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	case "sse":
		sseServer := s.ServeSSE("localhost:8080")
		log.Printf("SSE server listening on :8080")
		if err := sseServer.Start(":8080"); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	default:
		log.Fatalf(
			"Invalid transport type: %s. Must be 'stdio' or 'sse'",
			transport,
		)
	}
}
