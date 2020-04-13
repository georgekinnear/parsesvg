package parsesvg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/timdrysdale/geo"
)

const expectedLayoutJSON = `{"anchor":{"x":1.2588355559055121e-15,"y":-0.0003496930299212599},"dim":{"w":901.4173228346458,"h":884.4094488188978},"id":"a4-portrait-layout","anchors":{"image-mark":{"x":0,"y":841.8902863859433},"mark-header":{"x":6.294177637795276e-16,"y":883.3468064709828},"svg-check-flow":{"x":7.086614173228347,"y":883.3468064709828},"svg-mark-flow":{"x":655.9848283464568,"y":883.346975189093},"svg-mark-ladder":{"x":600.4855842519686,"y":883.346975189093},"svg-moderate-active":{"x":762.7586173228348,"y":883.346975189093},"svg-moderate-inactive":{"x":763.2376157480315,"y":883.3468934095655}},"pageDimStatic":{"check":{"w":111.55415811023623,"h":883.3464566929134},"mark":{"w":763.2376157480315,"h":883.3464566929134},"moderate-active":{"w":899.7675590551182,"h":883.3464566929134},"moderate-inactive":{"w":786.7112314960631,"h":883.3464566929134}},"pageDimDynamic":{"moderate":{"dim":{"w":1.417039398425197,"h":881.5748031496064},"widthIsDynamic":true,"heightIsDynamic":false}},"filenames":{"mark-header":"ladders-a4-portrait-header","svg-check-flow":"./test/sidebar-312pt-check-flow","svg-mark-flow":"./test/sidebar-312pt-mark-flow","svg-mark-ladder":"./test/sidebar-312pt-mark-ladder","svg-moderate-active":"./test/sidebar-312pt-moderate-flow-alt-active","svg-moderate-inactive":"./test/sidebar-312pt-moderate-inactive"},"ImageDimStatic":{"mark-header":{"w":592.4409448818898,"h":39.68503937007874},"previous-mark":{"w":595.2755905511812,"h":839.0551181102363},"previous-moderate":{"w":763.2376157480315,"h":881.5748031496064}},"ImageDimDynamic":{"previous-check":{"dim":{"w":1.417039398425197,"h":881.5748031496064},"widthIsDynamic":true,"heightIsDynamic":false}}}`

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

	if !reflect.DeepEqual(want.PageDimStatic, got.PageDimStatic) {
		t.Errorf("PageDimStatic are different\n%v\n%v", want.PageDimStatic, got.PageDimStatic)
	}
	if !reflect.DeepEqual(want.PageDimDynamic, got.PageDimDynamic) {
		t.Errorf("PageDimDynamic are different\n%v\n%v", want.PageDimDynamic, got.PageDimDynamic)
	}

	if !reflect.DeepEqual(want.ImageDimStatic, got.ImageDimStatic) {
		t.Errorf("ImageDimStatic are different\n%v\n%v", want.ImageDimStatic, got.ImageDimStatic)
	}
	if !reflect.DeepEqual(want.ImageDimDynamic, got.ImageDimDynamic) {
		t.Errorf("ImageDimDynamic are different\n%v\n%v", want.ImageDimDynamic, got.ImageDimDynamic)
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

//TODO move this to types.go; add json tags
type Spread struct {
	Name            string
	PageDim         geo.Dim
	PageDimDelta    geo.Dim
	WidthIsDynamic  bool
	HeightIsDynamic bool
	Images          []ImageInsert
	Ladders         []Ladder
	TextFields      []TextField
}

type ImageInsert struct {
	Filename              string
	Corner                geo.Point
	Dim                   geo.Dim
	ScaleImage            bool
	ScaleByHeightNotWidth bool
}

func TestPrintSpreadsFromLayout(t *testing.T) {
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

	// find pages for this name
	pageStaticDim := geo.Dim{}
	pageDynamicDim := geo.DynamicDim{}
	pageIsDynamic := false

	for k, v := range layout.PageDimDynamic {
		if strings.Contains(k, spread.Name) {
			pageDynamicDim = v
			pageIsDynamic = true
		}
	}
	for k, v := range layout.PageDimStatic {
		if strings.Contains(k, spread.Name) {
			pageStaticDim = v
		}
	}

	if reflect.DeepEqual(pageStaticDim, geo.Dim{}) &&
		reflect.DeepEqual(pageDynamicDim, geo.DynamicDim{}) {
		// no page info - throw an error
		fmt.Printf("no page info for this spread")
		return
	}

	previousImageStaticDim := geo.Dim{}
	previousImageDynamicDim := geo.DynamicDim{}
	previousImageIsDynamic := false

	for k, v := range layout.ImageDimDynamic {
		if strings.Contains(k, spread.Name) {
			previousImageDynamicDim = v
			previousImageIsDynamic = true
		}
	}
	for k, v := range layout.ImageDimStatic {
		if strings.Contains(k, spread.Name) {
			previousImageStaticDim = v
		}
	}

	spread.WidthIsDynamic = false
	spread.HeightIsDynamic = false

	switch {

	case !pageIsDynamic && !previousImageIsDynamic:
		//fully static - quite likely
		spread.PageDim.H = pageStaticDim.H
		spread.PageDim.W = pageStaticDim.W

	case pageIsDynamic && !previousImageIsDynamic:
		// 'misconfigured', probably, but in a 'safe' way
		// good trick to save some space on the layout doc!
		// we know image dim so have page dims finalised
		if pageDynamicDim.WidthIsDynamic {
			spread.PageDim.H = pageDynamicDim.Dim.H
			spread.PageDim.W = pageStaticDim.W + previousImageStaticDim.W
		} else { //prevent both being true at once (TODO: use case for both true?)
			spread.PageDim.W = pageDynamicDim.Dim.W
			spread.PageDim.H = pageStaticDim.H + previousImageStaticDim.H
		}

	case pageIsDynamic && previousImageIsDynamic:
		// properly dynamic - but treat only one dimension as dynamic
		// we prepare a single spread up front, then apply to multiple
		// pages so we can't just measure the first image we get -
		// some pages will differ if they have (not) had extra stages of
		// processing before (greater/fewer # of sidebars)

		// check image description and page agree on which dimension is dynamic
		if pageDynamicDim.WidthIsDynamic != previousImageDynamicDim.WidthIsDynamic {
			fmt.Printf("Error: Page and previous image dynamic on different dimensions")
			return
		}
		if pageDynamicDim.HeightIsDynamic != previousImageDynamicDim.HeightIsDynamic {
			fmt.Printf("Error: Page and previous image dynamic on different dimensions")
			return
		}

		// default to letting width be dynamic if both are set dynamic
		if pageDynamicDim.WidthIsDynamic {
			// only put final Dims in PageDim - when complete, can render
			spread.PageDim.H = pageDynamicDim.Dim.H
			spread.PageDim.W = 0.0
			spread.PageDimDelta.W = pageStaticDim.W
			spread.PageDimDelta.H = 0.0 //is fixed
			spread.WidthIsDynamic = true
			spread.HeightIsDynamic = false
			// later, get width from image, add the delta width to get page width
		} else { //prevent both being true at once (TODO: use case for both true?)
			spread.PageDim.W = pageDynamicDim.Dim.W
			spread.PageDim.H = 0.0
			spread.PageDimDelta.H = pageStaticDim.H
			spread.PageDimDelta.W = 0.0 //is fixed
			spread.WidthIsDynamic = false
			spread.HeightIsDynamic = true
			// later, get height from image, add the delta height to get page height
		}

	default:
		// misconfigured - throw an error and stop
		fmt.Printf("Error: Page and previous image somehow disagree on dimensions")
		return

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

	//	TranslatePosition()

	for _, svgname := range svgFilenames {

		offset := geo.Point{}

		if thisOffset, ok := layout.Anchors[svgname]; !ok {
			//default to layout anchor if not in the list
			offset = layout.Anchor
		} else {

			offset = thisOffset
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
			Filename:   imgfilename,
			Corner:     TranslatePosition(ladder.Anchor, offset),
			Dim:        ladder.Dim,
			ScaleImage: false, //don't scale chrome (TODO- only previous-image gets scaled)
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

	//get all images, other than previous image
	//Since these are the images for the textfield chrome, it's the same story - page layout engine will sort.
	//note that we haven't got previous image, so just send filename as 'previous-image' and let engine work it out
	//note that _all_ non-svg images need an image dims box....else their size will depend on their quality (dpi)

	for _, imgname := range imgFilenames {

		if _, ok := layout.ImageDimStatic[imgname]; !ok {
			t.Errorf("No size for image %s (must be static)\n", imgname)
		}

		offset := geo.Point{}

		if thisOffset, ok := layout.Anchors[imgname]; !ok {
			//default to layout anchor if not in the list
			offset = layout.Anchor
		} else {

			offset = thisOffset
		}

		imgfilename := imgname //in case not specified, e.g. previous image

		if filename, ok := layout.Filenames[imgname]; ok {
			imgfilename = fmt.Sprintf("%s.jpg", filename)
		}
		// append chrome image to the images list
		image := ImageInsert{
			Filename:   imgfilename,
			Corner:     offset,
			Dim:        layout.ImageDimStatic[imgname], //TODO need to get dim from images layer
			ScaleImage: false,                          // (TODO- previous-image gets scaled)
		}
		spread.Images = append(spread.Images, image) //add chrome to list of images to include

	}

	// add the previous image...
	// do some fu with teh static/dynamic dims
	/*
			"previousImageDimStatic": {
				"mark": {
					"w": 595.2755905511812,
					"h": 839.0551181102363
				},

		image := ImageInsert{
			Filename:   "previous-image",
			Corner:     layout.Anchor,
			Dim:        layout.ImageDimStatic[imgname], //
			ScaleImage: true,                           // (TODO- previous-image gets scaled)
		}
		spread.Images = append(spread.Images, image) //add chrome to list of images to include
	*/
	/*
		// scale and position image
		img.ScaleToHeight(ladder.Dim.H)
		img.SetPos(ladder.Anchor.X, ladder.Anchor.Y) //TODO check this has correct sense for non-zero offsets

		// create new page with image
		c.SetPageSize(creator.PageSize{ladder.Dim.W, ladder.Dim.H})
		c.NewPage()
		c.Draw(img)
	*/

}

/*
type Spread struct {
	Name            string
	PageDim         geo.Dim
	PageDimDelta    geo.Dim
	WidthIsDynamic  bool
	HeightIsDynamic bool
	Images          []ImageInsert
	Ladders         []Ladder
	TextFields      []TextField
}

type ImageInsert struct {
	Filename              string
	Corner                geo.Point
	Dim                   geo.Dim
ScaleImage bool
	ScaleByHeightNotWidth bool
}*/
/*

	c := creator.New()

	c.SetPageMargins(0, 0, 0, 0) // we're not printing

	svgFilename := "./test/ladders-a4-portrait-mark.svg"
	jpegFilename := "./test/ladders-a4-portrait-mark.jpg"
	pageFilename := "./test/ladders-a4-portrait-mark.pdf"

	svgBytes, err := ioutil.ReadFile(svgFilename)
	if err != nil {
		t.Error(err)
	}

	img, err := c.NewImageFromFile(jpegFilename)

	if err != nil {
		t.Errorf("Error opening image file: %s", err)
	}

writeParsedGeometry(svgBytes, img, pageFilename, c, t)


*/
