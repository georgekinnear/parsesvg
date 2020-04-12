package parsesvg

import (
	"encoding/xml"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
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

	if openBracket < 0 || comma < 0 || closeBracket < 0 {
		return 0.0, 0.0
	}

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

func scanUnitStringToPP(str string) (float64, error) {

	str = strings.TrimSpace(str)
	length := len(str)
	units := str[length-2 : length]
	value, err := strconv.ParseFloat(str[0:length-2], 64)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Couldn't parse  %s when split into value %s with units %s", str, str[0:length-2], units))
	}

	switch units {
	case "mm":
		return value * geo.PPMM, nil
	case "px":
		return value * geo.PPPX, nil
	case "pt":
		return value, nil //TODO check pt doesn't somehow default to not present
	case "in":
		return value * geo.PPIN, nil
	}

	return 0, errors.New(fmt.Sprintf("didn't understand the units %s", units))

}

func getLadderDim(svg *Csvg__svg) (geo.Dim, error) {
	dim := geo.Dim{}

	if svg == nil {
		return dim, errors.New("nil pointer to svg")
	}

	w, err := scanUnitStringToPP(svg.Width)
	if err != nil {
		return dim, err
	}
	h, err := scanUnitStringToPP(svg.Height)
	if err != nil {
		return dim, err
	}

	return geo.Dim{W: w, H: h}, nil

}

func DefineLayoutFromSVG(input []byte) (*Layout, error) {

	var svg Csvg__svg
	layout := &Layout{}

	err := xml.Unmarshal(input, &svg)

	if err != nil {
		return nil, err
	}

	// get title
	if svg.Cmetadata__svg.CRDF__rdf != nil {
		if svg.Cmetadata__svg.CRDF__rdf.CWork__cc != nil {
			if svg.Cmetadata__svg.CRDF__rdf.CWork__cc.Ctitle__dc != nil {
				layout.ID = svg.Cmetadata__svg.CRDF__rdf.CWork__cc.Ctitle__dc.String
			}
		}
	}

	layout.Anchor = geo.Point{X: 0, Y: 0}

	layoutDim, err := getLadderDim(&svg)
	if err != nil {
		return nil, err
	}

	layout.Dim = layoutDim

	var dx, dy float64

	// look for reference anchor position
	for _, g := range svg.Cg__svg {
		// get transform applied to layer, if any
		if g.AttrInkscapeSpacelabel == geo.AnchorsLayer {
			dx, dy = getTranslate(g.Transform)

			layout.Anchors = make(map[string]geo.Point)

			for _, r := range g.Cpath__svg {
				fmt.Printf("anchors %s %s\n", r.Cx, r.Cy)
				x, err := strconv.ParseFloat(r.Cx, 64)
				if err != nil {
					fmt.Printf("Anchors %v", r)
					return nil, err
				}
				y, err := strconv.ParseFloat(r.Cy, 64)
				if err != nil {
					fmt.Printf("Anchors %v", r)
					return nil, err
				}

				ddx, ddy := getTranslate(r.Transform)

				newX := x + dx + ddx
				newY := y + dy + ddy

				if r.Title != nil {
					if r.Title.String == geo.AnchorReference {

						fmt.Printf("%s %s %v\n", r.Title.String, geo.AnchorReference, r.Title.String == geo.AnchorReference)

						layout.Anchor = geo.Point{X: newX, Y: newY}
					} else {
						layout.Anchors[r.Title.String] = geo.Point{X: newX, Y: newY}
					}

				} else {
					log.Errorf("Anchor at (%f,%f) has no title, so ignoring\n", newX, newY)
				}
			}
		}
	}

	// look for pageDims
	layout.PageDims = make(map[string]geo.Dim)
	for _, g := range svg.Cg__svg {
		if g.AttrInkscapeSpacelabel == geo.PagesLayer {
			fmt.Printf("pages group\n")
			for _, r := range g.Crect__svg {
				fmt.Printf("pages %s %s\n", r.Width, r.Height)
				w, err := strconv.ParseFloat(r.Width, 64)
				if err != nil {
					fmt.Printf("PageDims %v", r)
					return nil, err
				}
				h, err := strconv.ParseFloat(r.Height, 64)
				if err != nil {
					fmt.Printf("PageDims %v", r)
					return nil, err
				}

				if r.Title != nil { //avoid seg fault, obvs

					layout.PageDims[r.Title.String] = geo.Dim{W: w, H: h}
					fmt.Printf("PageDims: %v", r.Title.String)

				} else {
					log.Errorf("Page at with size (%f,%f) has no title, so ignoring\n", w, h)
				}
			}
		}

	}
	// sort textfields based on tab order
	err = ApplyDocumentUnitsScaleLayout(&svg, layout)
	if err != nil {
		return nil, err
	}

	return layout, nil
}

func DefineLadderFromSVG(input []byte) (*Ladder, error) {

	var svg Csvg__svg
	ladder := &Ladder{}

	err := xml.Unmarshal(input, &svg)

	if err != nil {
		return nil, err
	}

	if svg.Cmetadata__svg.CRDF__rdf != nil {
		if svg.Cmetadata__svg.CRDF__rdf.CWork__cc != nil {
			if svg.Cmetadata__svg.CRDF__rdf.CWork__cc.Ctitle__dc != nil {
				ladder.ID = svg.Cmetadata__svg.CRDF__rdf.CWork__cc.Ctitle__dc.String
			}
		}
	}

	ladder.Anchor = geo.Point{X: 0, Y: 0}

	ladderDim, err := getLadderDim(&svg)
	if err != nil {
		return nil, err
	}

	ladder.Dim = ladderDim

	var dx, dy float64

	// look for reference anchor position
	for _, g := range svg.Cg__svg {
		// get transform applied to layer, if any
		if g.AttrInkscapeSpacelabel == geo.AnchorsLayer {
			dx, dy = getTranslate(g.Transform)

		}
		for _, r := range g.Cpath__svg {
			if r.Title != nil {
				if true { //r.Title.String == geo.AnchorReference {
					fmt.Printf("%s %s %v\n", r.Title.String, geo.AnchorReference, r.Title.String == geo.AnchorReference)
					x, err := strconv.ParseFloat(r.Cx, 64)
					if err != nil {
						return nil, err
					}
					y, err := strconv.ParseFloat(r.Cy, 64)
					if err != nil {
						return nil, err
					}

					ddx, ddy := getTranslate(r.Transform)

					newX := x + dx + ddx
					newY := y + dy + ddy
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

				tf.TabSequence = getTabSequence(r)

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
	// sort textfields based on tab order

	sort.Slice(ladder.TextFields, func(i, j int) bool {
		return ladder.TextFields[i].TabSequence < ladder.TextFields[j].TabSequence
	})

	err = ApplyDocumentUnits(&svg, ladder)
	if err != nil {
		return nil, err
	}
	err = convertToPDFYScale(ladder)
	if err != nil {
		return nil, err
	}
	return ladder, nil
}

func ApplyDocumentUnits(svg *Csvg__svg, ladder *Ladder) error {

	// iterate through the structure applying the conversion from
	// document units to points

	//note we do NOT apply the modification to ladder.DIM because this has its own
	//units in it and has already been handled.

	units := svg.Cnamedview__sodipodi.AttrInkscapeSpacedocument_dash_units

	sf := float64(1)

	switch units {
	case "mm":
		sf = geo.PPMM
	case "px":
		sf = geo.PPPX
	case "pt":
		sf = 1
	case "in":
		sf = geo.PPIN
	}

	ladder.Anchor.X = sf * ladder.Anchor.X
	ladder.Anchor.Y = sf * ladder.Anchor.Y

	for idx, tf := range ladder.TextFields {
		err := scaleTextFieldUnits(&tf, sf)
		if err != nil {
			return err
		}
		ladder.TextFields[idx] = tf
	}

	return nil
}

func ApplyDocumentUnitsScaleLayout(svg *Csvg__svg, layout *Layout) error {

	// iterate through the structure applying the conversion from
	// document units to points

	//note we do NOT apply the modification to ladder.DIM because this has its own
	//units in it and has already been handled.

	units := svg.Cnamedview__sodipodi.AttrInkscapeSpacedocument_dash_units

	sf := float64(1)

	switch units {
	case "mm":
		sf = geo.PPMM
	case "px":
		sf = geo.PPPX
	case "pt":
		sf = 1
	case "in":
		sf = geo.PPIN
	}

	layout.Anchor.X = sf * layout.Anchor.X
	layout.Anchor.Y = sf * layout.Anchor.Y

	Ytop := layout.Dim.H - layout.Anchor.Y //TODO triple check this sign!

	for k, v := range layout.Anchors {
		v.X = sf * v.X
		v.Y = Ytop - (sf * v.Y)
		layout.Anchors[k] = v
	}
	for k, v := range layout.PageDims {
		v.W = sf * v.W
		v.H = sf * v.H
		layout.PageDims[k] = v

	}

	return nil
}

func scaleTextFieldUnits(tf *TextField, sf float64) error {
	if tf == nil {
		return errors.New("nil pointer to TextField")
	}

	tf.Rect.Corner.X = sf * tf.Rect.Corner.X
	tf.Rect.Corner.Y = sf * tf.Rect.Corner.Y
	tf.Rect.Dim.W = sf * tf.Rect.Dim.W
	tf.Rect.Dim.H = sf * tf.Rect.Dim.H

	return nil
}

func convertToPDFYScale(ladder *Ladder) error {
	if ladder == nil {
		return errors.New("nil pointer to ladder")
	}

	Ytop := ladder.Dim.H - ladder.Anchor.Y //TODO triple check this sign!

	for idx, tf := range ladder.TextFields {

		tf.Rect.Corner.Y = Ytop - tf.Rect.Corner.Y
		ladder.TextFields[idx] = tf
	}
	return nil

}

func formRect(tf TextField) []float64 {

	return []float64{tf.Rect.Corner.X, tf.Rect.Corner.Y - tf.Rect.Dim.H, tf.Rect.Corner.X + tf.Rect.Dim.W, tf.Rect.Corner.Y}

}

func getTabSequence(r *Crect__svg) int64 {
	var TabSequence = regexp.MustCompile(`(?i:(tab|tab-))([0-9]+)`)
	var SequenceNumber = regexp.MustCompile(`([0-9]+)`)
	//fmt.Printf("Id %s Tab order %s\n", r.Id, SequenceNumber.FindString(TabSequence.FindString(r.Id)))
	//TODO - combine regexp into one
	var n int64
	n, err := strconv.ParseInt(SequenceNumber.FindString(TabSequence.FindString(r.Id)), 10, 64)
	if err != nil {
		return int64(0)
	}
	return n
}
