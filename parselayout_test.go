package parsesvg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/mattetti/filebuffer"
	"github.com/timdrysdale/geo"
	"github.com/unidoc/unipdf/v3/annotator"
	"github.com/unidoc/unipdf/v3/creator"
	"github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/model/optimize"
)

const expectedLayoutJSON = `{"anchor":{"x":3.568352756897515e-15,"y":884.4107897677605},"dim":{"width":901.4173228346458,"height":884.4094488188978,"dynamicWidth":false},"id":"a4-portrait-layout","anchors":{"img-previous-mark":{"x":0,"y":841.8902863859433},"mark-header":{"x":6.294177637795276e-16,"y":883.3468064709828},"svg-check-flow":{"x":7.086614173228347,"y":883.3468064709828},"svg-mark-flow":{"x":655.9848283464568,"y":883.346975189093},"svg-mark-ladder":{"x":600.4855842519686,"y":883.346975189093},"svg-moderate-active":{"x":762.7586173228348,"y":883.346975189093},"svg-moderate-inactive":{"x":763.2376157480315,"y":883.3468934095655}},"pageDims":{"check":{"width":111.55415811023623,"height":883.3464566929134,"dynamicWidth":false},"mark":{"width":763.2376157480315,"height":883.3464566929134,"dynamicWidth":false},"moderate-active":{"width":899.7675590551182,"height":883.3464566929134,"dynamicWidth":false},"moderate-inactive":{"width":786.7112314960631,"height":883.3464566929134,"dynamicWidth":false},"width-moderate":{"width":1.417039398425197,"height":881.5748031496064,"dynamicWidth":true}},"filenames":{"mark-header":"./test/ladders-a4-portrait-header","svg-check-flow":"./test/sidebar-312pt-check-flow","svg-mark-flow":"./test/sidebar-312pt-mark-flow","svg-mark-ladder":"./test/sidebar-312pt-mark-ladder","svg-moderate-active":"./test/sidebar-312pt-moderate-flow-comment-active","svg-moderate-inactive":"./test/sidebar-312pt-moderate-inactive"},"ImageDims":{"mark-header":{"width":592.4409448818898,"height":39.68503937007874,"dynamicWidth":false},"previous-check":{"width":1.417039398425197,"height":881.5748031496064,"dynamicWidth":true},"previous-mark":{"width":595.2755905511812,"height":839.0551181102363,"dynamicWidth":false},"previous-moderate":{"width":763.2376157480315,"height":881.5748031496064,"dynamicWidth":false}}}`

func TestDefineLayoutFromSvg(t *testing.T) {
	svgFilename := "./test/layout-312pt-static-mark-dynamic-moderate-static-check-v2.svg"
	svgBytes, err := ioutil.ReadFile(svgFilename)

	if err != nil {
		t.Error(err)
	}

	got, err := DefineLayoutFromSVG(svgBytes)
	if err != nil {
		t.Errorf("Error defining layout %v", err)
	}

	want := Layout{}

	_ = json.Unmarshal([]byte(expectedLayoutJSON), &want)

	if !reflect.DeepEqual(want.Anchor, got.Anchor) {
		t.Errorf("Anchor is different\n%v\n%v", want.Anchor, got.Anchor)
	}
	if !reflect.DeepEqual(want.Dim, got.Dim) {
		t.Errorf("Dim is different\n%v\n%v", want.Dim, got.Dim)
	}
	if !reflect.DeepEqual(want.ID, got.ID) {
		t.Errorf("ID is different\n%v\n%v", want.ID, got.ID)
	}
	if !reflect.DeepEqual(want.Anchors, got.Anchors) {
		t.Errorf("Anchors are different\n%v\n%v", want.Anchors, got.Anchors)
	}

	if !reflect.DeepEqual(want.PageDims, got.PageDims) {
		t.Errorf("PageDims are different\n%v\n%v", want.PageDims, got.PageDims)
	}
	if !reflect.DeepEqual(want.ImageDims, got.ImageDims) {
		t.Errorf("ImageDims are different\n%v\n%v", want.ImageDims, got.ImageDims)
	}
	if !reflect.DeepEqual(want.Filenames, got.Filenames) {
		t.Errorf("Filenames are different\n%v\n%v", want.Filenames, got.Filenames)
	}

}

func testPrettyPrintLayout(t *testing.T) {
	// helper for writing the tests on this file - not actually a test
	svgFilename := "./test/layout-312pt-static-mark-dynamic-moderate-static-check-v2.svg"
	svgBytes, err := ioutil.ReadFile(svgFilename)

	if err != nil {
		t.Error(err)
	}

	got, err := DefineLayoutFromSVG(svgBytes)
	if err != nil {
		t.Errorf("Error defining layout %v", err)
	}
	PrettyPrintLayout(got)

}
func testPrintLayout(t *testing.T) {
	// helper for writing the tests on this file - not actually a test
	svgFilename := "./test/layout-312pt-static-mark-dynamic-moderate-static-check-v2.svg"
	svgBytes, err := ioutil.ReadFile(svgFilename)

	if err != nil {
		t.Error(err)
	}

	got, err := DefineLayoutFromSVG(svgBytes)
	if err != nil {
		t.Errorf("Error defining layout %v", err)
	}
	PrintLayout(got)

}

func testPrintSpreadsFromLayout(t *testing.T) {
	svgFilename := "./test/layout-312pt-static-mark-dynamic-moderate-static-check-v2.svg"
	svgBytes, err := ioutil.ReadFile(svgFilename)

	if err != nil {
		t.Error(err)
	}

	layout, err := DefineLayoutFromSVG(svgBytes)
	if err != nil {
		t.Errorf("Error defining layout %v", err)
	}

	spread := Spread{}

	spread.Name = "mark"

	foundPage := false
	for k, v := range layout.PageDims {
		if strings.Contains(k, spread.Name) {
			spread.Dim = v
			foundPage = true
		}
	}

	if !foundPage {
		fmt.Printf("no page info for this spread %s\n", spread.Name)
		return
	}

	/* TODO - CUT THIS STALE CODE?

	 ImageDims := geo.Dim{}

		for k, v := range layout.ImageDims {
			if strings.Contains(k, spread.Name) {
				ImageDims = v
			}
		}*/

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

	//	TranslatePosition()

	for _, svgname := range svgFilenames {

		offset := geo.Point{}

		if thisAnchor, ok := layout.Anchors[svgname]; !ok {
			//default to layout anchor if not in the list
			offset = geo.Point{X: 0, Y: 0}
			fmt.Printf("didn't find anchor for %s\n", svgname)
		} else {

			fmt.Printf("%s@%v ref@%v\n", svgname, thisAnchor, layout.Anchor)
			offset = DiffPosition(layout.Anchor, thisAnchor)
			//fmt.Printf("Offset %s %v\n", svgname, offset)
		}

		svgfilename := fmt.Sprintf("%s.svg", layout.Filenames[svgname])
		imgfilename := fmt.Sprintf("%s.jpg", layout.Filenames[svgname]) //fixed by pdf library (I think)

		svgBytes, err := ioutil.ReadFile(svgfilename)
		if err != nil {
			t.Errorf("Error opening svg file %s", svgfilename)
		}

		ladder, err := DefineLadderFromSVG(svgBytes)
		if err != nil {
			t.Errorf("Error defining ladder %v", err)
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

	//get all images, other than image-previous, that comes separately via own arg
	//Since these are the images for the textfield chrome, it's the same story - page layout engine will sort.
	//note that we haven't got previous image, so just send filename as 'previous-image' and let engine work it out
	//note that _all_ non-svg images need an image dims box....else their size will depend on their quality (dpi)

	for _, imgname := range imgFilenames {

		if _, ok := layout.ImageDims[imgname]; !ok {
			t.Errorf("No size for image %s (must be provided in layout)\n", imgname)
		}

		offset := geo.Point{}

		if thisAnchor, ok := layout.Anchors[imgname]; !ok {
			//default to layout anchor if not in the list
			offset = layout.Anchor
		} else {

			offset = DiffPosition(layout.Anchor, thisAnchor)
			fmt.Printf("Previous: %v %v\n", layout.Anchor, offset)
		}

		imgfilename := imgname //in case not specified, e.g. previous image

		if filename, ok := layout.Filenames[imgname]; ok {
			imgfilename = fmt.Sprintf("%s.jpg", filename)
		}
		// append chrome image to the images list

		image := ImageInsert{
			Filename: imgfilename,
			Corner:   offset,
			Dim:      layout.ImageDims[imgname], //TODO need to get dim from images layer
		}
		spread.Images = append(spread.Images, image) //add chrome to list of images to include

	}

	offset := DiffPosition(layout.Anchors["img-previous-mark"], layout.Anchor) //TODO change example layout & parserto image-previous-mark
	fmt.Printf("image-mark: %v %v\n", layout.Anchor, offset)
	previousImage := ImageInsert{
		Filename: "./test/script.jpg",
		Corner:   offset,                            //geo.Point{X: 0, Y: 61.5},               //offset,
		Dim:      layout.ImageDims["previous-mark"], //
	}

	// draw images
	c := creator.New()
	c.SetPageMargins(0, 0, 0, 0) // we're not printing

	img, err := c.NewImageFromFile(previousImage.Filename)

	if err != nil {
		t.Errorf("Error opening image file: %s", err)
	}

	//see timdrysdale/pagescale if confused
	if spread.Dim.DynamicWidth {
		img.ScaleToHeight(spread.Dim.Height)
		spread.Dim.Width = img.Width()
	} else {
		imgScaledWidth := img.Width() * spread.Dim.Height / img.Height()

		if imgScaledWidth > spread.Dim.Width {
			// oops, we're too big, so scale using width instead
			img.ScaleToWidth(spread.Dim.Width)
		} else {
			img.ScaleToHeight(spread.Dim.Height)
		}

	}

	img.SetPos(previousImage.Corner.X, previousImage.Corner.Y)

	c.SetPageSize(creator.PageSize{spread.Dim.Width, spread.Dim.Height})
	c.NewPage()
	c.Draw(img)

	for _, v := range spread.Images {
		fmt.Printf("Printing image %s to pdf\n", v.Filename)
		img, err := c.NewImageFromFile(v.Filename)

		if err != nil {
			t.Errorf("Error opening image file: %s", err)
		}
		// all these images are static
		img.SetWidth(v.Dim.Width)
		img.SetHeight(v.Dim.Height)
		img.SetPos(v.Corner.X, v.Corner.Y) //TODO check this has correct sense for non-zero offsets
		fmt.Printf("Setting position to (%f, %f)\n------------------\n", v.Corner.X, v.Corner.Y)
		// create new page with image

		c.Draw(img)
	}

	// write to memory
	var buf bytes.Buffer

	err = c.Write(&buf)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// convert buffer to readseeker
	var bufslice []byte
	fbuf := filebuffer.New(bufslice)
	fbuf.Write(buf.Bytes())

	// read in from memory
	pdfReader, err := model.NewPdfReader(fbuf)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	pdfWriter := model.NewPdfWriter()

	page, err := pdfReader.GetPage(1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	form := model.NewPdfAcroForm()

	for _, tf := range spread.TextFields {

		tfopt := annotator.TextFieldOptions{Value: tf.Prefill} //TODO - MaxLen?!
		name := fmt.Sprintf("Page-00-%s", tf.ID)
		textf, err := annotator.NewTextField(page, name, formRect(tf), tfopt)
		if err != nil {
			panic(err)
		}
		*form.Fields = append(*form.Fields, textf.PdfField)
		page.AddAnnotation(textf.Annotations[0].PdfAnnotation)
	}

	err = pdfWriter.SetForms(form)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	err = pdfWriter.AddPage(page)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	of, err := os.Create("./test/mark-spread.pdf")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
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
}

// gs -dNOPAUSE -sDEVICE=jpeg -sOutputFile=mark-spread-gs.jpg -dJPEGQ=95 -r300 -q mark-spread.pdf -c quit
func TestRenderSpreadMark(t *testing.T) {

	svgLayoutPath := "./test/layout-312pt-static-mark-dynamic-moderate-comment-static-check.svg"

	pdfOutputPath := "./test/render-mark-spread.pdf"

	previousImagePath := "./test/script.jpg"

	spreadName := "mark"

	pageNumber := int(16)

	err := renderSpread(svgLayoutPath, spreadName, previousImagePath, pageNumber, pdfOutputPath)

	if err != nil {
		t.Error(err)
	}

}

func TestRenderSpreadModerate(t *testing.T) {

	svgLayoutPath := "./test/layout-312pt-static-mark-dynamic-moderate-comment-static-check.svg"

	pdfOutputPath := "./test/render-moderate-active-spread.pdf"

	previousImagePath := "./test/mark-spread-gs.jpg"

	spreadName := "moderate-active"

	pageNumber := int(16)

	err := renderSpread(svgLayoutPath, spreadName, previousImagePath, pageNumber, pdfOutputPath)

	if err != nil {
		t.Error(err)
	}

}

func TestRenderSpreadCheck(t *testing.T) {

	svgLayoutPath := "./test/layout-312pt-static-mark-dynamic-moderate-comment-static-check.svg"

	pdfOutputPath := "./test/render-check-spread.pdf"

	previousImagePath := "./test/moderate-active-gs.jpg"

	spreadName := "check"

	pageNumber := int(16)

	err := renderSpread(svgLayoutPath, spreadName, previousImagePath, pageNumber, pdfOutputPath)

	if err != nil {
		t.Error(err)
	}

}

func TestRenderSpreadCheckAfterInactive(t *testing.T) {

	svgLayoutPath := "./test/layout-312pt-static-mark-dynamic-moderate-comment-static-check.svg"

	pdfOutputPath := "./test/render-check-after-inactive-spread.pdf"

	previousImagePath := "./test/moderate-inactive-spread-gs.jpg"

	spreadName := "check"

	pageNumber := int(16)

	err := renderSpread(svgLayoutPath, spreadName, previousImagePath, pageNumber, pdfOutputPath)

	if err != nil {
		t.Error(err)
	}

}

func renderSpread(svgLayoutPath string, spreadName string, previousImagePath string, pageNumber int, pdfOutputPath string) error {

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
		imgScaledWidth := img.Width() * spread.Dim.Height / img.Height()

		if imgScaledWidth > spread.Dim.Width {
			// oops, we're too big, so scale using width instead
			img.ScaleToWidth(spread.Dim.Width)
		} else {
			img.ScaleToHeight(spread.Dim.Height)
		}

	}
	fmt.Printf("Spread is Dynamic? %v Width is static %f extra %f image %f getter %f\n", spread.Dim.DynamicWidth, spread.Dim.Width, spread.ExtraWidth, img.Width(), spread.GetWidth())
	img.SetPos(previousImage.Corner.X, previousImage.Corner.Y)
	fmt.Printf("setpos previous image: %f, %f\n", previousImage.Corner.X, previousImage.Corner.Y)
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
		name := fmt.Sprintf("page-%03d-%s", pageNumber, tf.ID)
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

func TestPrettyPrintLayout(t *testing.T) {
	// helper for writing the tests on this file - not actually a test
	svgFilename := "./test/layout-312pt-static-mark-dynamic-moderate-comment-static-check.svg"
	svgBytes, err := ioutil.ReadFile(svgFilename)

	if err != nil {
		t.Error(err)
	}

	got, err := DefineLayoutFromSVG(svgBytes)
	if err != nil {
		t.Errorf("Error defining layout %v", err)
	}
	PrettyPrintLayout(got)

}
