// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package remotetapextension

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/extension/extensiontest"
)

func availableEndpoint(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())
	return addr
}

// The server must listen on the configured endpoint and serve the embedded UI
// at "/". This is a regression test for the endpoint being ignored (the server
// bound ":http" instead) and for the file server reading a non-existent "html"
// sub-tree, which made every path return 404.
func TestExtensionServesUIOnConfiguredEndpoint(t *testing.T) {
	endpoint := availableEndpoint(t)
	cfg := createDefaultConfig().(*Config)
	cfg.ServerConfig.NetAddr.Endpoint = endpoint

	ext, err := NewFactory().Create(context.Background(), extensiontest.NewNopSettings(typ), cfg)
	require.NoError(t, err)
	require.NoError(t, ext.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { require.NoError(t, ext.Shutdown(context.Background())) })

	url := "http://" + endpoint + "/"
	require.Eventually(t, func() bool {
		resp, err := http.Get(url)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 10*time.Second, 50*time.Millisecond)
}
