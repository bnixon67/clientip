package main

import (
	"net/http"

	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/webhandler"
)

func RegisterRoutes(mux *http.ServeMux, app *webapp.WebApp) {
	mux.HandleFunc("GET /", ClientIPGetHandler)
}

func AddMiddleware(h http.Handler) http.Handler {
	// Functions are executed in reverse, so last added is called first.
	h = webhandler.LogRequest(h)
	h = webhandler.MiddlewareLogger(h)
	h = webhandler.NewRequestIDMiddleware(h)

	return h
}
