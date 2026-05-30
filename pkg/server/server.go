// Package server is the core of the static file server
package server

import (
	"net"
	"net/http"
	"strconv"

	"github.com/N-Hoque/static-file-server/pkg/config"
	"github.com/N-Hoque/static-file-server/pkg/handle"
)

var (
	// Values to be overridden to simplify unit testing.
	selectHandler  = handlerSelector
	selectListener = listenerSelector
)

// Run server.
func Run(cfg *config.Config) error {
	if cfg.Debug {
		config.Log(cfg)
	}
	// Choose and set the appropriate, optimized static file serving function.
	handler := selectHandler(cfg)

	// Serve files over HTTP or HTTPS based on paths to TLS files being
	// provided.
	listener := selectListener(cfg)

	binding := net.JoinHostPort(cfg.Host, strconv.FormatUint(uint64(cfg.Port), 10))
	return listener(binding, handler)
}

// handlerSelector returns the appropriate request handler based on
// configuration.
func handlerSelector(cfg *config.Config) http.HandlerFunc {
	var (
		handler          http.HandlerFunc
		serveFileHandler handle.FileServerFunc
	)

	serveFileHandler = http.ServeFile
	if cfg.Debug {
		serveFileHandler = handle.WithLogging(serveFileHandler)
	}

	if len(cfg.Referrers) > 0 {
		serveFileHandler = handle.WithReferrers(
			serveFileHandler, cfg.Referrers...,
		)
	}

	// Choose and set the appropriate, optimized static file serving function.
	if len(cfg.URLPrefix) == 0 {
		handler = handle.Basic(serveFileHandler, cfg.Folder)
	} else {
		handler = handle.Prefix(
			serveFileHandler,
			cfg.Folder,
			cfg.URLPrefix,
		)
	}

	// Determine whether index files should hidden.
	if !cfg.ShowListing {
		if cfg.AllowIndex {
			handler = handle.PreventListings(handler, cfg.Folder, cfg.URLPrefix)
		} else {
			handler = handle.IgnoreIndex(handler)
		}
	}

	// If configured, apply wildcard CORS support.
	if cfg.Cors {
		handler = handle.AddCorsWildcardHeaders(handler)
	}

	// If configured, apply key code access control.
	if len(cfg.AccessKey) > 0 {
		handler = handle.AddAccessKey(handler, cfg.AccessKey)
	}

	return handler
}

// listenerSelector serves files over HTTP or HTTPS
// based on paths to TLS files being provided
func listenerSelector(cfg *config.Config) handle.ListenerFunc {
	if len(cfg.TLSCert) == 0 {
		return handle.Listening()
	}

	handle.SetMinimumTLSVersion(cfg.TLSMinVers)
	return handle.TLSListening(cfg.TLSCert, cfg.TLSKey)
}
