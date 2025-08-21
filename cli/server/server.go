package server

import (
	"fmt"
	"net/http"

	"github.com/halverneus/static-file-server/config"
	"github.com/halverneus/static-file-server/handle"
)

var (
	// Values to be overridden to simplify unit testing.
	selectHandler  = handlerSelector
	selectListener = listenerSelector
)

// Run server.
func Run() error {
	if config.Get.Debug {
		config.Log()
	}
	// Choose and set the appropriate, optimized static file serving function.
	handler := selectHandler()

	// Serve files over HTTP or HTTPS based on paths to TLS files being
	// provided.
	listener := selectListener()

	binding := fmt.Sprintf("%s:%d", config.Get.Host, config.Get.Port)
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
	if config.Get.Debug {
		serveFileHandler = handle.WithLogging(serveFileHandler)
	}

	if len(config.Get.Referrers) > 0 {
		serveFileHandler = handle.WithReferrers(
			serveFileHandler, config.Get.Referrers...,
		)
	}

	// Choose and set the appropriate, optimized static file serving function.
	if len(config.Get.URLPrefix) == 0 {
		handler = handle.Basic(serveFileHandler, config.Get.Folder)
	} else {
		handler = handle.Prefix(
			serveFileHandler,
			config.Get.Folder,
			config.Get.URLPrefix,
		)
	}

	// Determine whether index files should hidden.
	if !config.Get.ShowListing {
		if config.Get.AllowIndex {
			handler = handle.PreventListings(handler, config.Get.Folder, config.Get.URLPrefix)
		} else {
			handler = handle.IgnoreIndex(handler)
		}
	}

	// If configured, apply wildcard CORS support.
	if config.Get.Cors {
		handler = handle.AddCorsWildcardHeaders(handler)
	}

	// If configured, apply key code access control.
	if len(config.Get.AccessKey) > 0 {
		handler = handle.AddAccessKey(handler, config.Get.AccessKey)
	}

	return handler
}

// listenerSelector serves files over HTTP or HTTPS
// based on paths to TLS files being provided
func listenerSelector() handle.ListenerFunc {
	if len(config.Get.TLSCert) == 0 {
		return handle.Listening()
	}

	handle.SetMinimumTLSVersion(config.Get.TLSMinVers)
	return handle.TLSListening(
		config.Get.TLSCert,
		config.Get.TLSKey,
	)
}
