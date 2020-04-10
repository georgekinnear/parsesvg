package parsesvg

import (
	"encoding/xml"
	"strconv"
	"strings"

	"github.com/timdrysdale/geo"
)

func ParseSvg(input []byte) *Csvg__svg {

	var svg Csvg__svg

	xml.Unmarshal(input, &svg)

	return &svg
}

/*
type TextField struct {
	Rect    geo.Rect //Corner.X/Y, Dim.W/H
	ID      string
	Prefill string
}

type Ladder struct {
	Anchor     geo.Point //X,Y
	Dim        geo.Dim   //W,H
	ID         string
	TextFields []TextField
}*/

func getTranslate(transform string) (float64, float64) {

	if len(transform) <= 0 {
		return 0.0, 0.0
	}

	if !strings.Contains(transform, geo.Translate) {
		return 0.0, 0.0
	}

	openBracket := strings.Index(transform, "(")
	comma := strings.Index(transform, ",")
	closeBracket := strings.Index(transform, ")")

	if openBracket == comma || comma == closeBracket {
		return 0.0, 0.0
	}

	dx, err := strconv.ParseFloat(transform[openBracket+1:comma], 64)
	if err != nil {
		return 0.0, 0.0
	}
	dy, err := strconv.ParseFloat(transform[comma+1:closeBracket], 64)
	if err != nil {
		return 0.0, 0.0
	}

	return dx, dy

}

/*
	w, err := strconv.ParseFloat(svg.Width, 64)
	if err != nil {
		return nil, err
	}
	h, err := strconv.ParseFloat(svg.Width, 64)
	if err != nil {
		return nil, err
	}

	ladder.Dim = geo.Dim{W: strconv.ParseFloat(svg.Width, 64),
		H: strconv.ParseFloat(svg.Height, 64)}
*/

func DefineLadderFromSVG(input []byte) (*Ladder, error) {

	var svg Csvg__svg
	ladder := &Ladder{}

	err := xml.Unmarshal(input, &svg)

	if err != nil {
		return nil, err
	}

	ladder.Anchor = geo.Point{X: 0, Y: 0}

	var dx, dy float64

	// look for reference anchor position
	for _, g := range svg.Cg__svg {
		// get transform applied to layer, if any
		if g.AttrInkscapeSpacelabel == geo.AnchorsLayer {
			dx, dy = getTranslate(g.Transform)
		}
		for _, r := range g.Cpath__svg {
			if r.Title != nil {
				if r.Title.String == geo.AnchorReference {
					x, err := strconv.ParseFloat(r.Cx, 64)
					if err != nil {
						return nil, err
					}
					y, err := strconv.ParseFloat(r.Cy, 64)
					if err != nil {
						return nil, err
					}

					newX := x + dx
					newY := y + dy
					ladder.Anchor = geo.Point{X: newX, Y: newY}
				}
			}
		}
	}

	// look for textFields
	for _, g := range svg.Cg__svg {
		if g.AttrInkscapeSpacelabel == geo.TextFieldsLayer {
			for _, r := range g.Crect__svg {
				tf := TextField{}
				if r.Title != nil { //avoid seg fault, obvs
					tf.ID = r.Title.String
				}
				if r.Desc != nil {
					tf.Prefill = r.Desc.String
				}
				w, err := strconv.ParseFloat(r.Width, 64)
				if err != nil {
					return nil, err
				}
				h, err := strconv.ParseFloat(r.Height, 64)
				if err != nil {
					return nil, err
				}

				tf.Rect.Dim.W = w
				tf.Rect.Dim.H = h
				dx, dy := getTranslate(r.Transform) //check if rotate will cause box to be out of place
				x, err := strconv.ParseFloat(r.Rx, 64)
				if err != nil {
					return nil, err
				}
				y, err := strconv.ParseFloat(r.Ry, 64)
				if err != nil {
					return nil, err
				}

				tf.Rect.Corner.X = x + dx
				tf.Rect.Corner.Y = y + dy
				ladder.TextFields = append(ladder.TextFields, tf)
			}
		}

	}

	err = ApplyDocumentUnits(&ladder)
	if err != nil {
		return nil, err
	}

	return ladder, nil
}

func ApplyDocumentUnits(ladder *Ladder) error {

	// iterate through the structure applying the conversion from
	// document units to points
	return error.Error("Not implemented!")
}
