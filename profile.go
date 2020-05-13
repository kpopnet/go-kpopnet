package kpopnet

import (
	"fmt"
	"image"
)

const (
	indexName = "index"
)

// Band info.
type Band map[string]interface{}

// Idol info.
type Idol map[string]interface{}

// Profiles contains information about known bands and idols.
type Profiles struct {
	Bands []Band `json:"bands"`
	Idols []Idol `json:"idols"`
}

// ImageInfo contains information about recognized image.
type ImageInfo struct {
	Rectangle image.Rectangle
	// TODO(Kagami): Add few most probable matches to simplify confirmation.
	IdolID    string
	Confirmed bool
}

// MarshalJSON returns JSON representation of ImageInfo.
func (i ImageInfo) MarshalJSON() ([]byte, error) {
	r := i.Rectangle
	s := fmt.Sprintf(
		`{"rect":[%d,%d,%d,%d],"id":"%s","confirmed":"%v"}`,
		r.Min.X, r.Min.Y, r.Max.X, r.Max.Y, i.IdolID, i.Confirmed)
	return []byte(s), nil
}
