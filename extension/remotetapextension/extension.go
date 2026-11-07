// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package remotetapextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/remotetapextension"

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"net"
	"net/http"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componentstatus"
	"go.opentelemetry.io/collector/extension"
)

//go:embed http/*
var httpFS embed.FS

type remoteObserverExtension struct {
	config   *Config
	settings extension.Settings
	server   *http.Server
}

func (s *remoteObserverExtension) Start(ctx context.Context, host component.Host) error {
	htmlContent, err := fs.Sub(httpFS, "http")
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(htmlContent)))
	s.server, err = s.config.ServerConfig.ToServer(ctx, host.GetExtensions(), s.settings.TelemetrySettings, mux)
	if err != nil {
		return err
	}

	// ToServer does not set the server address; the endpoint is applied through
	// ToListener. Serve on that listener so the server binds the configured
	// endpoint instead of defaulting to :http.
	var ln net.Listener
	if ln, err = s.config.ServerConfig.ToListener(ctx); err != nil {
		return err
	}
	go func() {
		if err := s.server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			componentstatus.ReportStatus(host, componentstatus.NewFatalErrorEvent(err))
		}
	}()
	return nil
}

func (s *remoteObserverExtension) Shutdown(_ context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Close()
}
