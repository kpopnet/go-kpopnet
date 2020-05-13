package kpopnet

import (
	"net/http"

	"github.com/dimfeld/httptreemux/v5"
)

// ServerOptions is server's config.
type ServerOptions struct {
	Address string
}

// StartServer starts HTTP server with specified config.
func StartServer(o ServerOptions) (err error) {
	router := createRouter()
	return http.ListenAndServe(o.Address, router)
}

func createRouter() http.Handler {
	r := httptreemux.New()

	r.GET("/api/idols/profiles", ServeProfiles)
	r.POST("/api/idols/recognize", ServeRecognize)
	r.GET("/api/idols/by-image/:id", ServeImageInfo)

	return http.Handler(r)
}
