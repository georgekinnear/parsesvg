package parsesvg

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/geo"
)

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

	// look for reference & header/ladder anchor positions
	// these also contain the base filename in the description
	for _, g := range svg.Cg__svg {
		// get transform applied to layer, if any
		if g.AttrInkscapeSpacelabel == geo.AnchorsLayer {
			dx, dy = getTranslate(g.Transform)

			layout.Anchors = make(map[string]geo.Point)
			layout.Filenames = make(map[string]string)

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

						if r.Desc != nil {
							layout.Filenames[r.Title.String] = r.Desc.String
						}
					}
				} else {
					log.Errorf("Anchor at (%f,%f) has no title, so ignoring\n", newX, newY)
				}
			}
		}
	}

	// look for pageDims
	layout.PageDimStatic = make(map[string]geo.Dim)
	layout.PageDimDynamic = make(map[string]geo.Dim)
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

					fullname := r.Title.String
					name := ""
					isDynamic := false
					fmt.Printf("PageDims: %v", r.Title.String)

					switch {
					case strings.HasPrefix(fullname, "page-dynamic-"):
						name = strings.TrimPrefix(fullname, "page-dynamic-")
						isDynamic = true
					case strings.HasPrefix(fullname, "page-static-"):
						name = strings.TrimPrefix(fullname, "page-static-")
					default:
						// we're just trying to strip off prefixes,
						// not prevent underadorned names from working
						name = strings.TrimPrefix(fullname, "page-")
					}

					if name != "" {
						if isDynamic {
							layout.PageDimDynamic[name] = geo.Dim{W: w, H: h}
						} else {
							layout.PageDimStatic[name] = geo.Dim{W: w, H: h}
						}
					}

				} else {
					log.Errorf("Page at with size (%f,%f) has no title, so ignoring\n", w, h)
				}
			}
		}

	}

	err = ApplyDocumentUnitsScaleLayout(&svg, layout)
	if err != nil {
		return nil, err
	}

	return layout, nil
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
	for k, v := range layout.PageDimStatic {
		v.W = sf * v.W
		v.H = sf * v.H
		layout.PageDimStatic[k] = v

	}

	return nil
}

func PrettyPrintLayout(layout *Layout) error {

	json, err := json.MarshalIndent(layout, "", "\t")
	if err != nil {
		return err
	}

	fmt.Println(string(json))
	return nil
}
