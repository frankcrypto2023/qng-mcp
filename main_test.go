package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestGetMethods(t *testing.T) {
	re, err := handleQngWeb3Rpc(context.Background(), mcp.CallToolRequest{
		Params: struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
			Meta      *struct {
				// If specified, the caller is requesting out-of-band progress
				// notifications for this request (as represented by
				// notifications/progress). The value of this parameter is an
				// opaque token that will be attached to any subsequent
				// notifications. The receiver is not obligated to provide these
				// notifications.
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Name: "get_peer_info",
			Arguments: map[string]interface{}{
				"parameterNum": 1,
				"parameter0":   true,
			},
		},
	})
	if err != nil {
		t.Errorf("Error getting methods: %v", err)
	}
	fmt.Println(re.Content[0])
}
