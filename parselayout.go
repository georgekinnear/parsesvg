package parsesvg

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/mattetti/filebuffer"
	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/geo"
	"github.com/unidoc/unipdf/v3/annotator"
	"github.com/unidoc/unipdf/v3/creator"
	"github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/model/optimize"
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

				if r.Title != nil {
					if r.Title.String == geo.AnchorReference {

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
	layout.PageDims = make(map[string]geo.Dim)
	for _, g := range svg.Cg__svg {
		if g.AttrInkscapeSpacelabel == geo.PagesLayer {
			for _, r := range g.Crect__svg {
				w, err := strconv.ParseFloat(r.Width, 64)
				if err != nil {
					return nil, err
				}
				h, err := strconv.ParseFloat(r.Height, 64)
				if err != nil {
					return nil, err
				}

				if r.Title != nil { //avoid seg fault, obvs

					fullname := r.Title.String
					name := ""
					isDynamic := false

					switch {
					case strings.HasPrefix(fullname, "page-dynamic-"):
						name = strings.TrimPrefix(fullname, "page-dynamic-")
						isDynamic = true
					case strings.HasPrefix(fullname, "page-static-"):
						name = strings.TrimPrefix(fullname, "page-static-")
					default:
						// unadorned pages are considered static
						// because this is the least surprising behaviour
						name = strings.TrimPrefix(fullname, "page-")
					}

					if name != "" { //reject anonymous pages
						layout.PageDims[name] = geo.Dim{Width: w, Height: h, DynamicWidth: isDynamic}
					}

				} else {
					log.Errorf("Page at with size (%f,%f) has no title, so ignoring\n", w, h)
				}
			}
		}
	}
	// look for previousImageDims
	layout.ImageDims = make(map[string]geo.Dim)
	for _, g := range svg.Cg__svg {
		if g.AttrInkscapeSpacelabel == geo.ImagesLayer {
			for _, r := range g.Crect__svg {
				w, err := strconv.ParseFloat(r.Width, 64)
				if err != nil {
					return nil, err
				}
				h, err := strconv.ParseFloat(r.Height, 64)
				if err != nil {
					return nil, err
				}

				if r.Title != nil { //avoid seg fault, obvs

					fullname := r.Title.String
					name := ""
					isDynamic := false

					switch {
					case strings.HasPrefix(fullname, "image-dynamic-"):
						name = strings.TrimPrefix(fullname, "image-dynamic-")
						name = strings.TrimPrefix(name, "width-")  //we may want this later, so leave in API
						name = strings.TrimPrefix(name, "height-") //getting info from box size for now
						isDynamic = true
					case strings.HasPrefix(fullname, "image-static-"):
						name = strings.TrimPrefix(fullname, "image-static-")
					default:
						// we're just trying to strip off prefixes,
						// not prevent underadorned names from working
						name = strings.TrimPrefix(fullname, "image-")
					}

					if name != "" { //reject anonymous images - can't place them
						layout.ImageDims[name] = geo.Dim{Width: w, Height: h, DynamicWidth: isDynamic}
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

	Ytop := layout.Dim.Height - layout.Anchor.Y //TODO triple check this sign!

	layout.Anchor.X = sf * layout.Anchor.X
	layout.Anchor.Y = Ytop - (sf * layout.Anchor.Y)

	for k, v := range layout.Anchors {
		v.X = sf * v.X
		v.Y = Ytop - (sf * v.Y)
		layout.Anchors[k] = v

	}
	for k, v := range layout.PageDims {
		v.Width = sf * v.Width
		v.Height = sf * v.Height
		layout.PageDims[k] = v

	}

	for k, v := range layout.ImageDims {
		v.Width = sf * v.Width
		v.Height = sf * v.Height
		layout.ImageDims[k] = v
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

func PrintLayout(layout *Layout) error {

	json, err := json.Marshal(layout)
	if err != nil {
		return err
	}

	fmt.Println(string(json))
	return nil
}

func PrettyPrintStruct(layout interface{}) error {

	json, err := json.MarshalIndent(layout, "", "\t")
	if err != nil {
		return err
	}

	fmt.Println(string(json))
	return nil
}

func RenderSpread(svgLayoutPath string, spreadName string, previousImagePath string, pageNumber int, pdfOutputPath string) error {

	svgBytes, err := ioutil.ReadFile(svgLayoutPath)

	if err != nil {
		return errors.New(fmt.Sprintf("Error opening layout file %s: %v\n", svgLayoutPath, err))
	}

	layout, err := DefineLayoutFromSVG(svgBytes)
	if err != nil {
		return errors.New(fmt.Sprintf("Error obtaining layout from svg %s\n", svgLayoutPath))
	}

	spread := Spread{}

	spread.Name = spreadName

	foundPage := false
	for k, v := range layout.PageDims {
		if strings.Contains(k, spread.Name) {
			spread.Dim = v
			foundPage = true
		}
	}

	if !foundPage {
		return errors.New(fmt.Sprintf("No page size info for spread %s\n", spread.Name))
	}

	// find svg & img elements for this name
	var svgFilenames, imgFilenames []string

	for k, _ := range layout.Filenames {
		if strings.Contains(k, spread.Name) {

			// assume jpg- or no prefix is image; svg- is ladder (image plus acroforms)
			if strings.HasPrefix(k, geo.SVGElement) {
				svgFilenames = append(svgFilenames, k) //we'll get the contents later
			} else {
				imgFilenames = append(imgFilenames, k)
			}
		}
	}

	// get all the textfields (and put image of associated chrome into images list)
	// note that if page dynamic, textfields are ALL dynamically shifting wrt to dynamic page edge,
	// no matter what side of the previous image edge they are. This means we only need one set of dims
	// the layout engine will just add the amount of the previous image's size in the dynamic dimension
	// We need to add the anchor position to the textfield positions (which are relative to that anchor)

	for _, svgname := range svgFilenames {

		offset := geo.Point{}

		if thisAnchor, ok := layout.Anchors[svgname]; !ok {
			//default to layout anchor if not in the list - keeps layout drawing cleaner
			offset = geo.Point{X: 0, Y: 0}
		} else {
			offset = DiffPosition(layout.Anchor, thisAnchor)
		}

		svgfilename := fmt.Sprintf("%s.svg", layout.Filenames[svgname])
		imgfilename := fmt.Sprintf("%s.jpg", layout.Filenames[svgname]) //TODO check again library is jpg-only?

		svgBytes, err := ioutil.ReadFile(svgfilename)
		if err != nil {
			return errors.New(fmt.Sprintf("Entity %s: error opening svg file %s", svgname, svgfilename))
		}

		ladder, err := DefineLadderFromSVG(svgBytes)
		if err != nil {
			return errors.New(fmt.Sprintf("Ladder %s: Error defining ladder from svg because %v", svgname, err))
		}

		if ladder == nil {
			continue //throw error?
		}
		spread.Ladders = append(spread.Ladders, *ladder)

		// append chrome image to the images list
		image := ImageInsert{
			Filename: imgfilename,
			Corner:   TranslatePosition(ladder.Anchor, offset),
			Dim:      ladder.Dim,
		}

		spread.Images = append(spread.Images, image) //add chrome to list of images to include

		//append TextFields to the Textfield list
		for _, tf := range ladder.TextFields {

			//shift the text field and add it to the list
			//let engine take care of mangling name to suit page
			tf.Rect.Corner = TranslatePosition(tf.Rect.Corner, offset)
			spread.TextFields = append(spread.TextFields, tf)
		}
	}

	// get all the static images that decorate this page, but not the special script "previous-image"

	for _, imgname := range imgFilenames {

		if _, ok := layout.ImageDims[imgname]; !ok {
			return errors.New(fmt.Sprintf("No size for image %s (must be provided in layout - check you have a correctly named box on the images layer in Inkscape)\n", imgname))
		}

		offset := geo.Point{}

		if thisAnchor, ok := layout.Anchors[imgname]; !ok {
			//default to layout anchor if not in the list
			offset = layout.Anchor
		} else {
			offset = DiffPosition(layout.Anchor, thisAnchor)
		}

		imgfilename := imgname //in case not specified, e.g. previous image

		if filename, ok := layout.Filenames[imgname]; ok {
			imgfilename = fmt.Sprintf("%s.jpg", filename)
		}
		// append chrome image to the images list

		image := ImageInsert{
			Filename: imgfilename,
			Corner:   offset,
			Dim:      layout.ImageDims[imgname],
		}

		spread.Images = append(spread.Images, image) //add chrome to list of images to include
	}

	// Obtain the special "previous-image" which is flattened/rendered to image version of this page at the last step

	previousImageAnchorName := fmt.Sprintf("img-previous-%s", spread.Name)

	previousImageDimName := fmt.Sprintf("previous-%s", spread.Name)

	offset := DiffPosition(layout.Anchors[previousImageAnchorName], layout.Anchor)

	previousImage := ImageInsert{
		Filename: previousImagePath,
		Corner:   offset,
		Dim:      layout.ImageDims[previousImageDimName],
	}

	// We do NOT add the previousImage to spread.Images because we treat it differently

	// We do things in a funny order here so that we can load the previous-image
	// and set the dynamic page size if needed

	c := creator.New()
	c.SetPageMargins(0, 0, 0, 0) // we're not printing so use the whole page

	img, err := c.NewImageFromFile(previousImage.Filename)

	if err != nil {
		return errors.New(fmt.Sprintf("Error opening spread %s previous-image file %s: %v", spread.Name, previousImage.Filename, err))
	}

	// Now we do the scaling to fit the page - see timdrysdale/pagescale for a demo
	if spread.Dim.DynamicWidth {
		img.ScaleToHeight(spread.Dim.Height)
		spread.ExtraWidth = img.Width() //we'll increase the page size by the image size
	} else {
		imgScaledWidth := img.Width() * previousImage.Dim.Height / img.Height()

		if imgScaledWidth > previousImage.Dim.Width {
			// oops, we're too big, so scale using width instead
			img.ScaleToWidth(previousImage.Dim.Width)
		} else {
			img.ScaleToHeight(previousImage.Dim.Height)
		}

	}
	fmt.Printf("PreviousImage Dims (%f,%f)\n", previousImage.Dim.Width, previousImage.Dim.Height)
	img.SetPos(previousImage.Corner.X, previousImage.Corner.Y)
	// we use GetWidth() so value includes fixed width plus extra width
	c.SetPageSize(creator.PageSize{spread.GetWidth(), spread.Dim.Height})

	c.NewPage()

	c.Draw(img) //draw previous image

	for _, v := range spread.Images {
		img, err := c.NewImageFromFile(v.Filename)

		if err != nil {
			return errors.New(fmt.Sprintf("Error opening image file %s: %s", v.Filename, err))
		}
		// all these images are static so we set dims directly
		// user needs to spot if they did their artwork to the wrong spec
		// or maybe they want it that way - we'd never know...
		// TODO consider logging a warning here for GUI etc
		img.SetWidth(v.Dim.Width)
		img.SetHeight(v.Dim.Height)
		if spread.Dim.DynamicWidth {
			img.SetPos(v.Corner.X+spread.ExtraWidth, v.Corner.Y)
		} else {
			img.SetPos(v.Corner.X, v.Corner.Y) //TODO check this has correct sense for non-zero offsets
		}
		c.Draw(img)
	}

	// This is the bit where we cross an internal boundary in the underlying library that has
	// strong opinions about where it gets it bytes from
	// So as to avoid making mods to the library, and for speed, we write to a memory file

	// write to memory
	var buf bytes.Buffer

	err = c.Write(&buf)
	if err != nil {
		return errors.New(fmt.Sprintf("Error: %v\n", err))
	}

	// convert buffer to readseeker
	var bufslice []byte
	fbuf := filebuffer.New(bufslice)
	fbuf.Write(buf.Bytes())

	// read in from memory
	pdfReader, err := model.NewPdfReader(fbuf)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading opening our internal page buffer %v\n", err))
	}

	pdfWriter := model.NewPdfWriter()

	page, err := pdfReader.GetPage(1)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading page from our internal page buffer %v\n", err))
	}

	/*******************************************************************************
	  Note that multipage acroforms are a wriggly issue!
	  This code is intended for single-page demos - check gradex-overlay for the
	  multipage method
	  ******************************************************************************/
	form := model.NewPdfAcroForm()

	for _, tf := range spread.TextFields {

		tfopt := annotator.TextFieldOptions{Value: tf.Prefill} //TODO - MaxLen?!
		// TODO consider allowing a more templated mangling of the ID number
		// For multi-student entries (although, OTH, there will be per-page ID data etc embedded too
		// which may be more useful in this regard, rather than overloading the textfield id)

		if spread.Dim.DynamicWidth {
			tf.Rect.Corner.X = tf.Rect.Corner.X + spread.ExtraWidth
		}
		textf, err := annotator.NewTextField(page, name, formRect(tf), tfopt)
		if err != nil {
			panic(err)
		}
		*form.Fields = append(*form.Fields, textf.PdfField)
		page.AddAnnotation(textf.Annotations[0].PdfAnnotation)
	}

	err = pdfWriter.SetForms(form)
	if err != nil {
		return errors.New(fmt.Sprintf("Error: %v\n", err))
	}

	err = pdfWriter.AddPage(page)
	if err != nil {
		return errors.New(fmt.Sprintf("Error: %v\n", err))
	}

	of, err := os.Create(pdfOutputPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Error: %v\n", err))
	}

	defer of.Close()

	pdfWriter.SetOptimizer(optimize.New(optimize.Options{
		CombineDuplicateDirectObjects:   true,
		CombineIdenticalIndirectObjects: true,
		CombineDuplicateStreams:         true,
		CompressStreams:                 true,
		UseObjectStreams:                true,
		ImageQuality:                    80,
		ImageUpperPPI:                   100,
	}))

	pdfWriter.Write(of)

	return nil
}
