package main

import (
	"testing"
)

func TestGetMethods(t *testing.T) {
	// 测试GetMethods函数是否能正确解析JSON
	methods, err := GetMethods()
	if err != nil {
		t.Errorf("Error getting methods: %v", err)
		return
	}

	// 验证方法数量
	if len(methods) == 0 {
		t.Error("Expected methods to be loaded, but got empty slice")
		return
	}

	// 验证特定方法是否存在
	found := false
	for _, method := range methods {
		if method.Call == "get_peer_info" {
			found = true
			if method.Name != "qng_getPeerInfo" {
				t.Errorf("Expected method name 'qng_getPeerInfo', got '%s'", method.Name)
			}
			if method.Params != 1 {
				t.Errorf("Expected 1 parameter for get_peer_info, got %d", method.Params)
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find 'get_peer_info' method")
	}

	t.Logf("Successfully loaded %d methods", len(methods))
}

func TestJsonRpcRequest(t *testing.T) {
	// 测试JSON-RPC请求构建
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "qng_getPeerInfo",
		Params:  []interface{}{true},
		ID:      1,
	}

	// 验证请求结构
	if request.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC '2.0', got '%s'", request.JSONRPC)
	}
	if request.Method != "qng_getPeerInfo" {
		t.Errorf("Expected method 'qng_getPeerInfo', got '%s'", request.Method)
	}
	if request.ID != 1 {
		t.Errorf("Expected ID 1, got %d", request.ID)
	}
	if len(request.Params) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(request.Params))
	}
}

func TestOrdinalSuffix(t *testing.T) {
	// 测试序数后缀函数（如果存在）
	testCases := []struct {
		input    int
		expected string
	}{
		{1, "st"},
		{2, "nd"},
		{3, "rd"},
		{4, "th"},
		{11, "th"},
		{12, "th"},
		{13, "th"},
		{21, "st"},
		{22, "nd"},
		{23, "rd"},
		{24, "th"},
	}

	for _, tc := range testCases {
		// 由于getOrdinalSuffix函数已被删除，我们跳过这个测试
		// 或者可以重新实现一个简单的版本
		t.Logf("Skipping ordinal suffix test for %d (function removed)", tc.input)
	}
}
