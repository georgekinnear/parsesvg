# parsesvg
parse svg to obtain markbox location and size information for **gradex**™

![alt text][status]

## Why?

It's a lot easier to do geometry entry using a GUI, particularly when you want it to look nice. So why not leverage the fabulous [inkscape](https://inkscape.org/) for a double win? Pretty forms AND output in a format that we can parse to the find the exact coordinates of the acroforms we want to embed. Goodbye editing golang structs by hand in one window, with inkscape's ruler hard at work in another.

## How?
SVG is XML, so parsing for the position and size of an object is straightforward. There are some domain-specific constraints:

- inkscape has no idea we are doing this, so can't guide you on the drawing/annotating process 
- it's still _way_ easier than editing structs by hand, but it doesn't avoid having to think
- the output boxes always have pointy corners, so beware of your design-frustration sky-rocketing when you do nice rounded borders and find the light blue sharp corners of the TextField ruining your design vision. You can always hand craft some structs to relax again.
- conventions are a moving target ... you'll be naming a bunch of objects in inkscape, then re-doing it again later, just saying.
- there are transformations, such as translate, that we need to account for in calculating the position (but it seems that transforms are flattened on each application, so we should only need to do this once per object). There seems to be a global translate in all the svg I have looked at so far ... (hence the use of the reference ```anchors```)


```svg
  <g
     inkscape:label="Layer 1"
     inkscape:groupmode="layer"
     id="layer1"
     transform="translate(0,-247)">
    <rect
       style="opacity:1;fill:#ffffff;fill-opacity:0.75117373;stroke:#000000;stroke-width:0.26499999;stroke-miterlimit:4;stroke-dasharray:none;stroke-dashoffset:9.44881821;stroke-opacity:1"
       id="rect47789"
       width="24.735001"
       height="24.735001"
       x="14.363094"
       y="259.41382">
      <title
         id="title47791">markbox-00-00</title>
    </rect>
  </g>

```

## Procedure

This is bit finicky, but we'll make it make sense. I was on the receiving end of a procedure a bit like this (in terms of finickyness) when [creating custom content](https://github.com/timdrysdale/rf3a) for ```$BRAND_NAME``` model flying simulator - it was set up that you could copy exactly what they did, and worked if you had the same versions of the CAD software,  but nothing else. It was painful but I was grateful they left the door open for customisation. So here's the crack in the door .... 

### Set up a page to exactly the size your marks ladder, and name the ladder

Assuming you already have a design you are modifying, or some mock designs, you will know the page size you want. I find it helpful to set the page size and export the page to ```png``` when making images, and will usually end up resizing the page at least once during any drawing. So this step is not essential, now. ```Ctrl-Shift-D``` when you are ready. While you are there, put the ladder name in the ```Title``` field of the metadata. Resist the urge to hit either of the strangely tempting buttons in the metadata dialogue as this title data is unique to the ladder - it seems you needn't explicitly save in this dialogue. 

### Set up layers

We want at least three layers - your pretty design (the ```chrome```), the reference and position ```anchors``` at least one acroforms layer (one layer per type of form element).

- [```textfields```]
- [```dropdowns```]  --not implemented
- [```checkboxes```] --not implemented
- [```comboboxes```] --not implemented
- ```anchors```
- ```chrome```

For exporting your pretty chrome that goes around the textfields, just make textfields and anchors invisible. They're on top of the chrome so you can check alignment. Their names are expected exactly in this lower case format, and pluralised for ```textfields``` and ```anchors```. 

### Set up the reference anchor

So we can make sense of things at import, we're going to put an anchor in the top left corner of the ladder. We'll use this as an aid to position the ladder and textFields in a future step. Meanwhile, how do we make an anchor? We need an unambiguous reference that is optically distinguishable from a rectangular box, hence a circle. 
```
<svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
  <circle cx="50" cy="50" r="50"/>
</svg>
```
If we used another box, we'd be forever unsure if we had the wrong corner, and if we made it small enough to avoid that, we'd never see it. Whereas, the ```x,y``` coordinates of an ```svg``` circle are its centre. You can make the radius, colour of your anchor(s) anything you like. You just have to get the centre on the top left page corner as you perceive it. Snapping a circle onto a corner of a page avoids errors creeping into the rest of your relevant measuresments - the exact point of the anchor circle is used. Turn on the snap to page boundaries and snap to object centres, as well as the default snap settings. These are marked with a star:

![alt text][anchor]


### Anchor types

We're supporting an iterative work flow (multiple people grading things), so we need to anticipate that ladders are added at each step, and that we may wish to make one of them a different size at some point, without having to re-process all the other ladder configurations. So we are going to draw each ladder in its own file, so once we've got all the cells named, we can just leave it alone...

When combining multiple ladders into a workflow, we'll want to use the individual ladders as leaf cells, and arrange their respective positions on the final page by placing ```position anchors```. The ```reference anchor``` for a particular ladder is just mapped onto the ```position anchor```. We'll do some fu with naming schemes to sort this out. To add position anchors to a layout, name the anchor avoiding the reserved ```ref-anchor``` and in the metadata description file, place the base of the filename that of the ```svg```  and ```jpg``` versions of that sidebar or header. Note that BOTH files are needed (we get the forms elements from the ```svg``` and the image of the chrome we get Inkscape to render. (TODO: Yes, that means the sidebars and headers are not vector graphics, but they get rendered to images during any later steps in the process anyway. Although in future if ```svg`` rendering or ```pdf``` insertion becomes an option, it would help differentiate the active area from the previously-edited, because when active it would presumably be in vector format.) Inkscape does not produce ```jpg``` and the ```pdf``` library doesn't speak ```png``` so we simply use ImageMagick to convert
``` convert sidebar-312pt-mark-flow.png sidebar-31pt-mark-flow.jpg```



## Acroforms

Acroforms supports several types of field. I'm ignoring signature boxes for now because we can do [opticalcheckboxes](https://github.com/timdrysdale/opticalcheckbox) which play better with the idea of freely annotating anywhere. (TODO: So far support is only provided for textfields, but dropboxes are needed for the checking workflow)

- [```textfields```]
- [```dropdowns```] --not implemented
- [```checkboxes```] --not implemented
- [```comboboxes```] --not implemented

### Labelling and annotating

#### Ladders

In the document properties tab, ```Ctrl-Shift-D``` set the name of the layout element in the Title field of the metadata. Make sure it matches the filename, and the exported image.

![alt text][element-name]
![alt text][element-filename]

```svg
inkscape:version="0.92.4 (5da689c313, 2019-01-14)"
   sodipodi:docname="sidebar-312pt-check-flow.svg"
   inkscape:export-filename="/home/<snip>/parsesvg/test/sidebar-312pt-check-flow.png"
   inkscape:export-xdpi="299.86111"
   inkscape:export-ydpi="299.86111">
  <title
     id="title11203">sidebar-312pt-check-flow</title>
  <defs
     id="defs2">
	 ```
#### Anchors

- Give the reference anchor the title ```ref-anchor```. Behaviour is undefined if you add more than one reference anchor.
- Give any position anchors the title ```pos-anchor-<element_name>```, where <element_name> is meaningful to you. The ```svg``` for the element will be found in the file mentioned in the document metadata title, so the name of the pos-anchor does not have to match, i.e. you don't have to name your element's svg pos-anchor-whatever, you can just call it ```whatever.svg``` and put ```whatever``` in the document metadata as the document title. 

#### Textfields

Textfields don't necessarily take prefilling (but they can), whereas constrained-choice selections must be pre-populated. Let's do that in ```inkscape``` for an easy life. You can label and describe SVG elements in ```inkscape``` by ```Ctrl-Shift-O``` (remember to hit the 'Set' button - I kept forgetting first time out, so do check when you go back to an object that the data has persisted.) We'll use these to pass extra information into the parser, e.g. ```choiceBox``` options, or format strings that might help with hydrating the ```id``` to include page numbers etc. This bit is going to move rapidly ... so consider any implied API to be experimental and subject to change from minute to minute.

```
  <g
     inkscape:label="Layer 1"
     inkscape:groupmode="layer"
     id="layer1"
     transform="translate(0,-247)">
    <rect
       style="opacity:1;fill:#ffffff;fill-opacity:0.75117373;stroke:#000000;stroke-width:0.26499999;stroke-miterlimit:4;stroke-dasharray:none;stroke-dashoffset:9.44881821;stroke-opacity:1"
       id="rect47789"
       width="24.735001"
       height="24.735001"
       x="14.363094"
       y="259.41382">
      <desc
         id="desc50373">This is a description field I wonder if it makes it into the svg file .....</desc>
      <title
         id="title47791">markbox-00-00</title>
    </rect>
  </g>
 ```

### Tab order of acroforms elements

The order in which elements are written into the ```pdf``` determines the tab order as experienced by the user (which box you go to next when you hit tab). This strongly affects the ease of use of the workflow so it needs to be set logically (e.g. running from top to bottom) to avoid causing extra work to markers and checkers using keyboards. Inkscape does not offer a way to manipulate the order of elements in the ```xml```, e.g.  modifying the ID does not cause a reordering (for obvious efficiency reasons). Therefore, a sorting provision is included in the parser, that re-orders based on the tab number appended to the id as follows ...


![alt text][taborder]


### Exporting your sidebar/header element chrome for usage in a layout

The chrome layer is the only layer you should export.
- insert a white layer (or whatever suits your documents) in the background, or else the unfilled areas become black when converted to jpg
- turn off the visibility of the anchor and and forms layers, e.g. text fields.

What you want is this:

![alt text][layer-visibility]

A trouble-shooting graphic shows you the main things you might get wrong:

![alt text][export-troubleshooting]


## Example

The example we'll work with has three ```textfields```, that are floating in fairly random places, and at random sizes. A great example of where this parsing comes in handy - else lining it up by hand will be fiddly and tedious.


 All layers visible: ![alt text][example]

Only the chrome: ![alt text][example-chrome]

```anchor``` & ```textfields```: ![alt text][example-nonchrome]

We don't need to examine the chrome, because we'll import that in a separate step as a background image, using the position ```anchor``` from the overall layout (TODO: create section on that). The ```textfields``` and ```anchor``` are grouped into their own layers:

```svg
  <g
     inkscape:groupmode="layer"
     id="layer9"
     inkscape:label="anchors"
     style="display:inline"
     sodipodi:insensitive="true">
    <path
       transform="translate(0,-247)"
       style="opacity:1;fill:#00c041;fill-opacity:0.55399062;stroke:none;stroke-width:2;stroke-miterlimit:4;stroke-dasharray:none;stroke-dashoffset:9.44881821;stroke-opacity:1"
       id="path50406"
       sodipodi:type="arc"
       sodipodi:cx="0"
       sodipodi:cy="247"
       sodipodi:rx="4.8191962"
       sodipodi:ry="4.8191962"
       sodipodi:start="1.586911"
       sodipodi:end="1.582216"
       sodipodi:open="true"
       d="m -0.07765641,251.81857 a 4.8191962,4.8191962 0 0 1 -4.74100189,-4.89057 4.8191962,4.8191962 0 0 1 4.88500291,-4.74674 4.8191962,4.8191962 0 0 1 4.75246949,4.87943 4.8191962,4.8191962 0 0 1 -4.87384655,4.75819">
      <title
         id="title50411">ref-anchor</title>
    </path>
  </g>
  <g
     inkscape:groupmode="layer"
     id="layer8"
     inkscape:label="textfields"
     style="display:inline">
    <rect
       style="display:inline;opacity:1;fill:#e2e4ff;fill-opacity:0.75117373;stroke:none;stroke-width:0.26499999;stroke-miterlimit:4;stroke-dasharray:none;stroke-dashoffset:9.44881821;stroke-opacity:1"
       id="rect47789"
       width="7.8166327"
       height="8.0839024"
       x="6.810286"
       y="18.934586">
      <title
         id="title47791">badfile</title>
    </rect>
    <rect
       y="13.149623"
       x="38.650368"
       height="6.3086619"
       width="6.3086619"
       id="rect50379"
       style="display:inline;opacity:1;fill:#e2e4ff;fill-opacity:0.75117373;stroke:none;stroke-width:0.26499999;stroke-miterlimit:4;stroke-dasharray:none;stroke-dashoffset:9.44881821;stroke-opacity:1">
      <title
         id="title50377">markok</title>
    </rect>
    <rect
       style="display:inline;opacity:1;fill:#e2e4ff;fill-opacity:0.75117373;stroke:none;stroke-width:0.26499999;stroke-miterlimit:4;stroke-dasharray:none;stroke-dashoffset:9.44881821;stroke-opacity:1"
       id="rect50385"
       width="23.535746"
       height="5.3614168"
       x="22.629934"
       y="38.408676">
      <desc
         id="desc51388">Enter your intials here</desc>
      <title
         id="title50383">initials</title>
    </rect>
  </g>
```

Now that we have an example, we can fire up the parser, and punt out a test pdf.

## Actual usage

Our actual marking sidebars will have many textfields. Easy! We're using a GUI, so you can copy and paste, making sure to

- append the tab order code to the ID (e.g. ```<original-id>-tab-012```
- add the name of the element in the Title field, so you can identify it in extracted data later
 
This example is part of the tests (exported here with textfields layer turned on, showing in a light blue):

![alt text][sidebar-example]

## Layouts

We've got one sidebar working. Great! What next? The exams are going to make visits to moderators, and checkers too. They have their own sidebars. We've got this far using the GUI - let's continue to use it to organise where the individual sidebars and headers go. Here we make a separate ```layout.svg``` which contains the layers

- pages
- anchors
[- chrome]

The chrome is optional, but you'll want to include it for helping get your anchors in the right place. Delete the white backgrounds in your chrome, so you can see the pages. These will be used to set the paper size

For example:

![]alt text][layout-example]

This example represents a three stage process where all the incoming scans have been scaled to A4-portrait. Handling landscape is straightforward, by setting a flag that triggers the use of an alternate layout set tuned to landscape - that flagging process is not part of ```timdrysdale/parsesvg``` - search the ```pdf.gradex``` ecosystem for more details.

### Dynamic sidebar selection

Each stage in the process knows to expect a certain size image - if you set the incoming image for that stage to have non-zero height and width. We can get some dynamic behaviour by allowing EITHER zero height or zero width specifications for that image, so that the sidebar or header is just added to relative to the edge of the image, whatever size it is.

#### Why do we do it this way?

There is an inter-relationship between workflow logic and page layout, that would be architecturally-suspicious if it made too much of a dent in the way that the layouts are represented. So we do not include any logic or grouping of alternative options together. The Inkscape GUI isn't the right tool to ask to handle the logic for us. So our job here, is to simply provide all the options and let the selection take place elsewhere.

So what if the page size changes before you get to the Nth stage, because steps were omitted or had different size sidebars depending on whether they were active or inactive, etc?
The key datums are the edges of the incoming page. These should be uniform in either height or width, e.g. having optional headers AND optional sidebars at each stage is a doable thing, but it makes my brain sore trying to think why you'd want to put yourself through that (if you need progress bar iconography, which implies variable width headers to cope with different stages, then just bake it into your sidebars by making them all a standard height).

The easiest thing to do is just say, if the incoming image width (for example) is not specified, and is zero width, then we just accept whatever the image is (scale it to height for safety, of course) and then we add our sidebars relative to that. There's nothing to stop you overlaying the incoming image if you want - it's just that the use cases for that are a bit more advanced (annotating specific parts of a student answer, for example, and that requires knowledge of where they are putting stuff - which only works if they use printed paper. That's a future use case we'll see more of when the immediate $PANDEMIC is over)

so ... for cases where you _know_ for sure where your page edge is, because all steps before were mandatory, and you wnat


we need to know two things

We don't want to worry about unexpected results in estimating the new page size for dynamic pages when multiple layout elements may be contributing to it, so we need you to work it out in advance and tell us in the layour document with a ```page-extension-<yourpage>```. If you think your graphical design will work ok, there is nothing to stop this being an extension in two-dimensions but there is no in-built support for reasoning about the size of the incoming pages' unknown axis and modifying a graphic to suit that. If you make the graphic overlong, and place it anyway, it will be cropped, which gives you enough scope for borders that have translational symmetry. If you are adding vital elements, then it is your responsbility to have worked out a flow that will yield a minimum extent in the direction you are anticipating is dynamic - and that is the size of the most recent antecedent page's fixed size on that dimensions. We _could_ check for it, but it's just another false positive error to ignore if you do want a full length border and expect some cropping. (TODO consider adding flag support this use case, when the elements have dynamic information content that must be guaranteed to fit on the page - and bear in mind that a simple test run will show this issue immediately to a human).

if you are using a dynamic page, how big is the extra horizontal and vertical space that you are adding to the incoming image. Note that the incoming image will be scaled to suit the one axis you do define, because there s  



[anchor]: ./img/inkscape-anchor-alignment.png "circle on corner of page and snap settings bar"
[element-name]: ./img/element-name.png "name of the layout element entered into metadata"
[element-filename]: ./img/element-filename.png "name of the layout element used for saving image of the chrome"
[example]: ./img/example.png "example of three textfields with pretty surrounds in red"
[example-chrome]: ./img/example-chrome.png "just showing the pretty surrounds, not the anchor or textfields"
[example-nonchrome]: ./img/example-nonchrome.png "just showing the anchor or textfields"
[export-troubleshooting]: ./img/export-troubleshooting.png "examples with black background, anchor and textfields showing"
[layout-example]: ./img/layout-example-with-layer-dialogue.png "drawing with a header and three side bars surrounding an original scan of work to be marked"
[metadata-title]: ./img/metadata-title.png "inkscape metadata tab in document properties, showing title is ladder3-rect"
[layer-visibility]: ./img/layer-visibility-at-export.png "Layers dialog with the chrome layer set to visible, and all others invisible"
[sidebar-example]: ./img/sidebar-mark.png "marking sidebar with nearly 30 textfields"
[status]: https://img.shields.io/badge/alpha-in%20development-orange "Alpha status, in development"
[taborder]: ./img/taborder.png "object properties dialogue showing tab-02 appended to ID, setting tab order"