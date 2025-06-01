// Copyright (c) 2025 voidint <voidint@126.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"log"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
	"github.com/pkg/errors"
)

func main() {
	transportServer := transport.NewStdioServerTransport()

	mcpServer, err := server.NewServer(transportServer, server.WithServerInfo(protocol.Implementation{
		Name:    "g-mcp-server",
		Version: "0.1.0",
	}))
	if err != nil {
		log.Fatal(errors.WithStack(err))
	}

	lsTool, err := protocol.NewTool("g_ls", "List installed go sdk versions", struct{}{})
	if err != nil {
		log.Fatal(errors.WithStack(err))
	}

	lsRemoteTool, err := protocol.NewTool("g_ls-remote", "List remote go sdk versions available for install", struct{}{})
	if err != nil {
		log.Fatal(errors.WithStack(err))
	}

	mcpServer.RegisterTool(lsTool, lsHandler)
	mcpServer.RegisterTool(lsRemoteTool, lsRemoteHandler)

	if err = mcpServer.Run(); err != nil {
		log.Fatal(errors.WithStack(err))
	}
}
