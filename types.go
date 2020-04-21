package parsesvg

import (
	"github.com/timdrysdale/geo"
)

type TextField struct {
	Rect        geo.Rect
	ID          string
	Prefill     string
	TabSequence int64
	Multiline   bool
}

type TextPrefill struct {
	Rect       geo.Rect
	ID         string
	Properties string
	Text       Paragraph
}

// we read the properties from a JSON object in the Description field
// and then apply when writing the text field - these are private fields
// in the Paragraph struct in unipdf
type Paragraph struct {
	Text                string    `json:"text"`
	TextFont            string    `json:"textFont"`
	TextSize            float64   `json:"textSize"`
	LineHeight          float64   `json:"lineHeight"`
	Alignment           string    `json:"alignment"`
	EnableWrap          bool      `json:"enableWrap"`
	WrapWidth           float64   `json:"wrapWidth"`
	Angle               float64   `json:"angle"`
	AbsolutePositioning bool      `json:"absolutePositioning"`
	Margins             []float64 `json:"margins"`
	XPos                float64   `json:"xpos"`
	YPos                float64   `json:"ypos"`
	ColorHex            string    `json:"colorHex"`
}

type Ladder struct {
	Anchor       geo.Point
	Dim          geo.Dim
	ID           string
	TextFields   []TextField
	TextPrefills []TextPrefill
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
	Name       string
	Dim        geo.Dim
	ExtraWidth float64
	Images     []ImageInsert
	Ladders    []Ladder
	TextFields []TextField
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
	if s.Dim.DynamicWidth {
		return s.Dim.Width + s.ExtraWidth
	} else {
		return s.Dim.Width
	}
}

//unipdf fonts - see unipdf/model/font/
//Courier
//CourierBold
//CourierOblique
//CourierBoldOblique
//Helvetica
//HelveticaBold
//HelveticaOblique
//HelveticaBoldOblique
//Symbol
//ZapfDingbats
//TimesRoman
//TimesBold
//TimesItalic
//TimesBoldItalic
