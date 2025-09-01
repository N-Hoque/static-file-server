package server

import (
	"fmt"
	"net/http"

	"github.com/N-Hoque/static-file-server/pkg/config"
	"github.com/N-Hoque/static-file-server/pkg/handle"
)

var (
	// Values to be overridden to simplify unit testing.
	selectHandler  = handlerSelector
	selectListener = listenerSelector
)

// Run server.
func Run() error {
	if config.Target.Debug {
		config.Log()
	}
	// Choose and set the appropriate, optimized static file serving function.
	handler := selectHandler()

	// Serve files over HTTP or HTTPS based on paths to TLS files being
	// provided.
	listener := selectListener()

	binding := fmt.Sprintf("%s:%d", config.Target.Host, config.Target.Port)
	return listener(binding, handler)
}

// handlerSelector returns the appropriate request handler based on
// configuration.
func handlerSelector() http.HandlerFunc {
	var (
		handler          http.HandlerFunc
		serveFileHandler handle.FileServerFunc
	)

	serveFileHandler = http.ServeFile
	if config.Target.Debug {
		serveFileHandler = handle.WithLogging(serveFileHandler)
	}

	if len(config.Target.Referrers) > 0 {
		serveFileHandler = handle.WithReferrers(
			serveFileHandler, config.Target.Referrers...,
		)
	}

	// Choose and set the appropriate, optimized static file serving function.
	if len(config.Target.URLPrefix) == 0 {
		handler = handle.Basic(serveFileHandler, config.Target.Folder)
	} else {
		handler = handle.Prefix(
			serveFileHandler,
			config.Target.Folder,
			config.Target.URLPrefix,
		)
	}

	// Determine whether index files should hidden.
	if !config.Target.ShowListing {
		if config.Target.AllowIndex {
			handler = handle.PreventListings(handler, config.Target.Folder, config.Target.URLPrefix)
		} else {
			handler = handle.IgnoreIndex(handler)
		}
	}

	// If configured, apply wildcard CORS support.
	if config.Target.Cors {
		handler = handle.AddCorsWildcardHeaders(handler)
	}

	// If configured, apply key code access control.
	if len(config.Target.AccessKey) > 0 {
		handler = handle.AddAccessKey(handler, config.Target.AccessKey)
	}

	return handler
}

// listenerSelector serves files over HTTP or HTTPS
// based on paths to TLS files being provided
func listenerSelector() handle.ListenerFunc {
	if len(config.Target.TLSCert) == 0 {
		return handle.Listening()
	}

	handle.SetMinimumTLSVersion(config.Target.TLSMinVers)
	return handle.TLSListening(
		config.Target.TLSCert,
		config.Target.TLSKey,
	)
}
