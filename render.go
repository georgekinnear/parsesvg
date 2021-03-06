package parsesvg

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/mattetti/filebuffer"
	"github.com/timdrysdale/geo"
	"github.com/timdrysdale/pdfcomment"
	"github.com/timdrysdale/pdfpagedata"
	"github.com/timdrysdale/unipdf/v3/annotator"
	"github.com/timdrysdale/unipdf/v3/creator"
	"github.com/timdrysdale/unipdf/v3/model"
	"github.com/timdrysdale/unipdf/v3/model/optimize"
)

func RenderSpread(svgLayoutPath string, spreadName string, previousImagePath string, pageNumber int, pdfOutputPath string) error {

	contents := SpreadContents{
		SvgLayoutPath:     svgLayoutPath,
		SpreadName:        spreadName,
		PreviousImagePath: previousImagePath,
		PageNumber:        pageNumber,
		PdfOutputPath:     pdfOutputPath,
	}
	return RenderSpreadExtra(contents, []*PaperStructure{})

}

func RenderSpreadExtra(contents SpreadContents, parts_and_marks []*PaperStructure) error {

	svgLayoutPath := contents.SvgLayoutPath
	spreadName := contents.SpreadName
	previousImagePath := contents.PreviousImagePath
	prefillImagePaths := contents.PrefillImagePaths
	comments := contents.Comments
	pageNumber := contents.PageNumber
	pdfOutputPath := contents.PdfOutputPath
		
	svgBytes, err := ioutil.ReadFile(svgLayoutPath)

	if err != nil {
		return errors.New(fmt.Sprintf("Error opening layout file %s: %v\n", svgLayoutPath, err))
	}

	layout, err := DefineLayoutFromSVG(svgBytes)
	if err != nil {
		return errors.New(fmt.Sprintf("Error obtaining layout from svg %s\n", svgLayoutPath))
	}
	
	//fmt.Println(layout)
	
	//fmt.Println("\n\n", layout.Filenames)

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

	//fmt.Println(svgname)
	
		corner := geo.Point{X: 0, Y: 0} //default to layout anchor if not in the list - keeps layout drawing cleaner

		if thisAnchor, ok := layout.Anchors[svgname]; ok {
			corner = thisAnchor
		}

		svgfilename := fmt.Sprintf("%s.svg", layout.Filenames[svgname])
		imgfilename := fmt.Sprintf("%s.png", layout.Filenames[svgname]) //TODO check again library is jpg-only?

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
			Corner:   corner,
			Dim:      ladder.Dim,
		}

		spread.Images = append(spread.Images, image) //add chrome to list of images to include

		//append TextFields to the Textfield list
		for _, tf := range ladder.TextFields {

			//shift the text field and add it to the list
			//let engine take care of mangling name to suit page
			tf.Rect.Corner = TranslatePosition(corner, tf.Rect.Corner)
			spread.TextFields = append(spread.TextFields, tf)
		}
		//append TextPrefills to the TextPrefill list
		for _, tp := range ladder.TextPrefills {

			//shift the text field and add it to the list
			//let engine take care of mangling name to suit page
			tp.Rect.Corner = TranslatePosition(corner, tp.Rect.Corner)
			spread.TextPrefills = append(spread.TextPrefills, tp)
		}

		
		// Add the script info to the placeholders
		if svgname == "svg-mark-header" || svgname == "svg-moderate-header" {
		
			for _, tp := range ladder.Placeholders {
				new_rect := tp.Rect
				new_rect.Corner = TranslatePosition(corner, new_rect.Corner)
				new_text_field := TextPrefill{Rect: new_rect, ID: tp.ID,
											  Text:     Paragraph{Text: "",
																  TextSize: 20,
																  Alignment: "left"}} // TODO - get alignment to work
				switch box_type := tp.ID; box_type {
					case "script-info-course":
						new_text_field.Text.Text = contents.CourseCode
					case "script-info-diet":
						new_text_field.Text.Text = contents.ExamDiet
					case "script-info-student":
						new_text_field.Text.Text = contents.Candidate
					default:
						// do nothing
				}
				spread.TextPrefills = append(spread.TextPrefills, new_text_field)
			
				if tp.ID == "script-info-course" {
				}
			}
		}
		
		// Add the Marker ID to the placeholder
		if spread.Name == "mark" && contents.Marker != "" {
		
			for _, tp := range ladder.Placeholders {
				if tp.ID == "marker-id" {
					new_rect := tp.Rect
					new_rect.Corner = TranslatePosition(corner, new_rect.Corner)
					new_text_field := TextPrefill{Rect: new_rect,
												  ID:       "marker-id",
												  Text:     Paragraph{Text: contents.Marker,
																	  TextSize: 20,
																	  Alignment: "center"}} // TODO - get alignment to work
					spread.TextPrefills = append(spread.TextPrefills, new_text_field)
				}
			}
		}
		
		// Write any existing fields into the file
		if len(contents.PreviousFields) > 0 {
			//fmt.Println(contents.PreviousFields)
			
			base_rect := geo.Rect{}
			for _, tp := range ladder.Placeholders {
				if tp.ID == "prev-fields" {
					base_rect = tp.Rect
				}
			}
			
			fcount := 0
			for fname, fvalue := range contents.PreviousFields {
				//fmt.Println("Field",fcount,fname,fvalue)
									
				new_rect := base_rect
				new_rect.Corner = TranslatePosition(corner, new_rect.Corner)
				new_rect.Corner = TranslatePosition(geo.Point{X:0, Y:new_rect.Dim.Height * float64(fcount) * -1.2}, new_rect.Corner)
				
				new_text_field := TextField{Rect: new_rect,
											ID:         "prevfield-"+fname,
											Prefill:	fvalue}
				spread.TextFields = append(spread.TextFields, new_text_field)
				fcount++
			}
		}
		
	
		// Try out adding extra stuff
		if (len(parts_and_marks) > 0) {
			for pnum, part := range parts_and_marks {
			//	fmt.Println("Part: ",part.Part, pnum)
			//	fmt.Println("   ",part.Marks, " marks")
				// a hack - if you leave a blank row in the csv you get an empty row.
				// TODO - see if this can work with "|| part.Marks is empty", and move it to the qn-part-mark case below,
				// so that only the form field is omitted if one of parts/marks is not specified - this could allow (part,marks) = (Q1,) to give a "question heading" effect
				if part.Part == "" { 
					continue
				}
				for _, tp := range ladder.Placeholders {
					new_rect := tp.Rect
					new_rect.Corner = TranslatePosition(corner, new_rect.Corner)
					new_rect.Corner = TranslatePosition(geo.Point{X:0, Y:new_rect.Dim.Height * float64(pnum) * 1.2}, new_rect.Corner)
					switch box_type := tp.ID; box_type {
					case "qn-part-name":
						new_text_field := TextPrefill{Rect: new_rect,
													  ID:       "qn-part-name-"+strconv.Itoa(pnum),
													  Text:     Paragraph{Text: part.Part,
																		  TextSize: 14,
																		  Alignment: "right"}} // TODO - get alignment to work
						spread.TextPrefills = append(spread.TextPrefills, new_text_field)
					case "qn-part-total":
						// nudge it down a little
						new_rect.Corner = TranslatePosition(geo.Point{X:0, Y:5}, new_rect.Corner)
						new_text_field := TextPrefill{Rect: new_rect,
													  ID:         "qn-part-total-"+strconv.Itoa(pnum),
													  Text:     Paragraph{Text:"/"+strconv.Itoa(part.Marks), TextSize: 12}}
						spread.TextPrefills = append(spread.TextPrefills, new_text_field)
					case "qn-part-mark":
						fallthrough
					case "qn-part-moderate":
						new_text_field := TextField{Rect: new_rect,
													ID:         box_type+"-"+strconv.Itoa(pnum)}
						spread.TextFields = append(spread.TextFields, new_text_field)
						
						// append chrome image to the images list
						image := ImageInsert{
							Filename: "som/markbox.png",  // TODO - make this customisable, e.g. using JSON in the placeholder's description
							Corner:   new_rect.Corner,
							Dim:      new_rect.Dim,
						}

						spread.Images = append(spread.Images, image) //add chrome to list of images to include
						
						//fmt.Println(new_text_field)
						//fmt.Println("size of textfields: ", len(spread.TextFields))
						//fmt.Println("\n")
					
						
					default:
						// do nothing
					}
				}
			}
		}
		
	//fmt.Println("\nend of "+svgname)	
	//fmt.Println("size of prefills: ", len(spread.TextPrefills))
	//fmt.Println("size of textfields: ", len(spread.TextFields))
	//fmt.Println(spread.TextFields)
	//fmt.Println("\n")
	
	}
	
	
	// get all the static images that decorate this page, but not the special script "previous-image"

	//fmt.Println(prefillImagePaths)
	//fmt.Println(imgFilenames)
	for _, imgname := range imgFilenames {

		if _, ok := layout.ImageDims[imgname]; !ok {
			return errors.New(fmt.Sprintf("No size for image %s (must be provided in layout - check you have a correctly named box on the images layer in Inkscape)\n", imgname))
		}

		imgfilename := imgname //in case not specified, e.g. previous image

		if filename, ok := layout.Filenames[imgname]; ok {
			imgfilename = fmt.Sprintf("%s.jpg", filename)
		}

		// overwrite filename with dynamically supplied one, if supplied
		if filename, ok := prefillImagePaths[imgname]; ok {

			imgfilename = fmt.Sprintf("%s.jpg", filename)
		}

		corner := layout.Anchor

		if thisAnchor, ok := layout.Anchors[imgname]; ok {
			corner = thisAnchor
		}

		// append chrome image to the images list
		image := ImageInsert{
			Filename: imgfilename,
			Corner:   corner,
			Dim:      layout.ImageDims[imgname],
		}

		spread.Images = append(spread.Images, image) //add chrome to list of images to include
	}

	// Obtain the special "previous-image" which is flattened/rendered to image version of this page at the last step

	previousImageAnchorName := fmt.Sprintf("img-previous-%s", spread.Name)

	previousImageDimName := fmt.Sprintf("previous-%s", spread.Name)

	corner := layout.Anchors[previousImageAnchorName] //DiffPosition(layout.Anchors[previousImageAnchorName], layout.Anchor)

	previousImage := ImageInsert{
		Filename: previousImagePath,
		Corner:   corner,
		Dim:      layout.ImageDims[previousImageDimName],
	}

	// We do NOT add the previousImage to spread.Images because we treat it differently

	// We do things in a funny order here so that we can load the previous-image
	// and set the dynamic page size if needed

	c := creator.New()
	c.SetPageMargins(0, 0, 0, 0) // we're not printing so use the whole page

	if strings.Compare(previousImage.Filename, "") != 0 {

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
		img.SetPos(previousImage.Corner.X, previousImage.Corner.Y)
		// we use GetWidth() so value includes fixed width plus extra width
		c.SetPageSize(creator.PageSize{spread.GetWidth(), spread.Dim.Height})

		c.NewPage()

		c.Draw(img) //draw previous image
	} else {
		c.SetPageSize(creator.PageSize{spread.GetWidth(), spread.Dim.Height})

		c.NewPage()

	}

	// put our pagedata in first

	if !reflect.DeepEqual(contents.PageData, pdfpagedata.PageData{}) {
		pdfpagedata.MarshalPageData(c, &contents.PageData)
	}

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

	// Draw in our flattened comments
	rowHeight := 12.0
	numComments := float64(len(comments.GetByPage(pageNumber)))
	x := 0.3 * rowHeight
	y := c.Height() - ((0.3 + numComments) * rowHeight)
	for i, cmt := range comments.GetByPage(pageNumber) {

		pdfcomment.DrawComment(c, cmt, strconv.Itoa(i), x, y)
		y = y + rowHeight
	}

	for _, tp := range spread.TextPrefills {
		//update prefill contents from info given
		if val, ok := contents.Prefills[pageNumber][tp.ID]; ok {
			tp.Text.Text = val
		}
		// update our prefill text
		p := c.NewParagraph(tp.Text.Text)
		//fmt.Printf("Font size: %f", tp.Text.TextSize)
		p.SetFontSize(tp.Text.TextSize)
		p.SetPos(tp.Rect.Corner.X, tp.Rect.Corner.Y)
		//fmt.Printf("prefill %f,%f\n", tp.Rect.Corner.X, tp.Rect.Corner.Y)
		c.Draw(p)
		//fmt.Println(tp)

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
		//fmt.Printf("Textfie %f %f\n", tf.Rect.Corner.X, tf.Rect.Corner.Y)
		//fmt.Printf("formRe %v\n", formRect(tf, layout.Dim))
		
	//	new_bb := formRect(tf, layout.Dim) // returns [bbox.Llx, bbox.Lly, bbox.Urx, bbox.Ury]
		// Add chrome for mark boxes
		if strings.HasPrefix(tf.ID, "qn-part-mark-") {
				
		/* Tried drawing this is a rectangle, but it caused problems as it was selectable in some viewers, and got in the way of the form field
			rectDef := annotator.RectangleAnnotationDef{}
			rectDef.X = new_bb[0]-1
			rectDef.Y = new_bb[1]+1
			rectDef.Width = new_bb[2]-new_bb[0]+2
			rectDef.Height = new_bb[3]-new_bb[1]
			rectDef.Opacity = 1 // Semi transparent.
			rectDef.FillEnabled = true
			rectDef.FillColor = model.NewPdfColorDeviceRGB(1, 1, 1) // White fill
			rectDef.BorderEnabled = true
			rectDef.BorderWidth = 1
			rectDef.BorderColor = model.NewPdfColorDeviceRGB(0.93, 0.10, 0.12) // Red border

			rectAnnotation, err := annotator.CreateRectangleAnnotation(rectDef)
			if err != nil {
				return err
			}

			// Add to the page annotations.
			page.AddAnnotation(rectAnnotation)
		*/
		
		/* A possible alternative way, not yet tried but do some code like this?
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
		*/
		}
		
		textf, err := annotator.NewTextField(page, name, formRect(tf, layout.Dim), tfopt)
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
		ImageQuality:                    90,
		ImageUpperPPI:                   150,
	}))

	pdfWriter.Write(of)

	return nil
}
