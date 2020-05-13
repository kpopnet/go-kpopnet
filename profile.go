package kpopnet

import (
	"github.com/Kagami/go-face"
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

// TrainData contains information about all recognized idols.
type TrainData struct {
	Samples []face.Descriptor
	Cats    []int32
	Labels  map[int]string
}
