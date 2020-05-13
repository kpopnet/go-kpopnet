package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kpopnet/go-kpopnet"
	"github.com/kpopnet/go-kpopnet/cache"
	"github.com/kpopnet/go-kpopnet/db"
	"github.com/kpopnet/go-kpopnet/facerec"
)

var (
	maxOverheadSize = int64(10 * 1024)
	maxFileSize     = int64(5 * 1024 * 1024)
	maxBodySize     = maxFileSize + maxOverheadSize
)

func setMainHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache")
}

func setBodyHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func setAPIHeaders(w http.ResponseWriter) {
	setMainHeaders(w)
	setBodyHeaders(w)
}

func serveData(w http.ResponseWriter, r *http.Request, data []byte) {
	setMainHeaders(w)
	etag := fmt.Sprintf("W/\"%s\"", hashBytes(data))
	w.Header().Set("ETag", etag)
	if checkEtag(w, r, etag) {
		return
	}
	setBodyHeaders(w)
	w.Write(data)
}

func serveJSON(w http.ResponseWriter, r *http.Request, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		handle500(w, r, err)
		return
	}
	serveData(w, r, data)
}

func serveError(w http.ResponseWriter, r *http.Request, err error, code int) {
	setAPIHeaders(w)
	w.WriteHeader(code)
	io.WriteString(w, fmt.Sprintf("{\"error\": \"%v\"}", err))
}

func handle500(w http.ResponseWriter, r *http.Request, err error) {
	logError(err)
	serveError(w, r, kpopnet.ErrInternal, 500)
}

// ServeProfiles returns a JSON object with information about all profiles.
func ServeProfiles(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	// TODO(Kagami): For some reason cached request is not fast enough.
	// TODO(Kagami): Use some trigger to invalidate cache.
	v, err := cache.Cached(cache.ProfileCacheKey, func() (v interface{}, err error) {
		ps, err := db.GetProfiles()
		if err != nil {
			return
		}
		// Takes ~5ms so better to store encoded.
		return json.Marshal(ps)
	})
	if err != nil {
		handle500(w, r, err)
		return
	}
	serveData(w, r, v.([]byte))
}

// ServeRecognize recognizes image uploaded via HTTP.
func ServeRecognize(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	if err := r.ParseMultipartForm(0); err != nil {
		serveError(w, r, kpopnet.ErrParseForm, 400)
		return
	}
	fhs := r.MultipartForm.File["files[]"]
	if len(fhs) != 1 {
		serveError(w, r, kpopnet.ErrParseFile, 400)
		return
	}
	idolID, err := facerec.RequestRecognizeMultipart(fhs[0])
	switch err {
	case kpopnet.ErrParseFile:
		serveError(w, r, err, 400)
		return
	case kpopnet.ErrBadImage:
		serveError(w, r, err, 400)
		return
	case kpopnet.ErrNoIdol:
		serveError(w, r, err, 404)
		return
	case nil:
		// Do nothing.
	default:
		handle500(w, r, err)
		return
	}
	if idolID == nil {
		serveError(w, r, kpopnet.ErrNoSingleFace, 400)
		return
	}
	result := map[string]string{"id": *idolID}
	serveJSON(w, r, result)
}
