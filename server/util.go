package server

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/kpopnet/go-kpopnet"
)

func logError(err error) {
	log.Printf("kpopnet: %s\n%s\n", err, debug.Stack())
}

func hashBytes(buf []byte) string {
	hash := md5.Sum(buf)
	return base64.RawStdEncoding.EncodeToString(hash[:])
}

func checkEtag(w http.ResponseWriter, r *http.Request, etag string) bool {
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(304)
		return true
	}
	return false
}

func setAPIHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func serveEncodedJSON(w http.ResponseWriter, r *http.Request, data []byte) {
	setAPIHeaders(w)
	etag := fmt.Sprintf("W/\"%s\"", hashBytes(data))
	w.Header().Set("ETag", etag)
	if checkEtag(w, r, etag) {
		return
	}
	w.Write(data)
}

func serveJSON(w http.ResponseWriter, r *http.Request, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		serve500(w, r, err)
		return
	}
	serveEncodedJSON(w, r, data)
}

func serveError(w http.ResponseWriter, r *http.Request, err error, code int) {
	setAPIHeaders(w)
	w.WriteHeader(code)
	io.WriteString(w, fmt.Sprintf("{\"error\": \"%v\"}", err))
}

func serve400(w http.ResponseWriter, r *http.Request, err error) {
	serveError(w, r, err, 400)
}

func serve500(w http.ResponseWriter, r *http.Request, err error) {
	logError(err)
	serveError(w, r, kpopnet.ErrInternal, 500)
}
