package parsesvg

import (
	"github.com/timdrysdale/geo"
)

type TextField struct {
	Rect        geo.Rect
	ID          string
	Prefill     string
	TabSequence int64
}

type Ladder struct {
	Anchor     geo.Point
	Dim        geo.Dim
	ID         string
	TextFields []TextField
}

type Layout struct {
	Anchor    geo.Point            `json:"anchor"`
	Dim       geo.Dim              `json:"dim"`
	ID        string               `json:"id"`
	Anchors   map[string]geo.Point `json:"anchors"`
	PageDims  map[string]geo.Dim   `json:"pageDims"`
	Filenames map[string]string    `json:"filenames"`
	ImageDims map[string]geo.Dim   `json:"ImageDims"`
}

//TODO move this to types.go; add json tags
type Spread struct {
	Name            string
	Dim             geo.Dim
	ExtraWidth      float64
	HasDynamicWidth bool
	Images          []ImageInsert
	Ladders         []Ladder
	TextFields      []TextField
}

type ImageInsert struct {
	Filename string
	Corner   geo.Point
	Dim      geo.Dim
}

// how to understand dynamic width
// the fixed additional width is known at design time
// the unknown part is loaded into spread.Dim.Width when known
// the

func (s *Spread) GetWidth() float64 {
	if s.HasDynamicWidth {
		return s.Dim.Width + s.ExtraWidth
	} else {
		return s.Dim.Width
	}
}
