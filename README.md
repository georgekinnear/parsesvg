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
- [```dropdowns```]
- [```checkboxes```]
- [```comboboxes```]
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

When combining multiple ladders into a workflow, we'll want to use the individual ladders as leaf cells, and arrange their respective positions on the final page by placing ```position anchors```. The ```reference anchor``` for a particular ladder is just mapped onto the ```position anchor```. We'll do some fu with naming schemes to sort this out.


## Acroforms

Acroforms supports several types of field. I'm ignoring signature boxes for now because we can do [opticalcheckboxes](https://github.com/timdrysdale/opticalcheckbox) which play better with the idea of freely annotating anywhere.

- [```textfields```]
- [```dropdowns```]
- [```checkboxes```]
- [```comboboxes```]

### Labelling and annotating

#### Ladders

In the document properties tab, ```Ctrl-Shift-D``` set the name of the ladder in the Title field of the metadata. For example,

![alt text][metadata-title]

```svg
 inkscape:version="0.92.4 (5da689c313, 2019-01-14)"
   sodipodi:docname="rect3.svg">
  <title
     id="title50415">ladder-rect3</title>
  <defs
     id="defs47238" />
	 ```

#### Anchors

- Give the reference anchor the title ```ref-anchor```. Behaviour is undefined if you add more than one reference anchor.
- Give any position anchors the title ```pos-anchor-<ladder_name>```, where <ladder_name> is a text tag that the parser will know to associate with the ladder

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



[anchor]: ./img/inkscape-anchor-alignment.png "circle on corner of page and snap settings bar"
[example]: ./img/example.png "example of three textfields with pretty surrounds in red"
[example-chrome]: ./img/example-chrome.png "just showing the pretty surrounds, not the anchor or textfields"
[example-nonchrome]: ./img/example-nonchrome.png "just showing the anchor or textfields"
[metadata-title]: ./img/metadata-title.png "inkscape metadata tab in document properties, showing title is ladder3-rect"
[status]: https://img.shields.io/badge/alpha-in%20development-orange "Alpha status, in development" 