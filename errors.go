package kpopnet

import "errors"

var (
	// ErrInternal is returned in case of something went wrong on server side.
	ErrInternal = errors.New("internal error")
	// ErrParseForm is returned on malformed HTTP POST form.
	ErrParseForm = errors.New("error parsing form")
	// ErrParseFile is returned on input reading error.
	ErrParseFile = errors.New("error parsing form file")
	// ErrBadImage is returned on malformed/unsupported input image.
	ErrBadImage = errors.New("invalid image")
	// ErrNoSingleFace is returned when input image doesn't contain a single face
	// (0 or several).
	ErrNoSingleFace = errors.New("not a single face")
	// ErrNoIdol is returned when face wasn't recognized.
	ErrNoIdol = errors.New("cannot find idol")
)
