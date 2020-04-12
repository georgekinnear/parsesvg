package parsesvg

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
)

const expectedLayoutJSON = `{"anchor":{"x":1.2588355559055121e-15,"y":-0.0003496930299212599},"dim":{"w":901.4173228346458,"h":884.4094488188978},"id":"a4-portrait-layout","anchors":{"image-mark":{"x":0,"y":841.8902863859433},"mark-header":{"x":6.294177637795276e-16,"y":883.3468064709828},"svg-check-flow":{"x":7.086614173228347,"y":883.3468064709828},"svg-mark-flow":{"x":655.9848283464568,"y":883.346975189093},"svg-mark-ladder":{"x":600.4855842519686,"y":883.346975189093},"svg-moderate-active":{"x":762.7586173228348,"y":883.346975189093},"svg-moderate-inactive":{"x":763.2376157480315,"y":883.3468934095655}},"pageDimStatic":{"check":{"w":111.55415811023623,"h":883.3464566929134},"mark":{"w":763.2376157480315,"h":883.3464566929134},"moderate-active":{"w":899.7675590551182,"h":883.3464566929134},"moderate-inactive":{"w":786.7112314960631,"h":883.3464566929134}},"pageDimDynamic":{"width-moderate":{"dim":{"w":1.417039398425197,"h":881.5748031496064},"widthIsDynamic":true,"heightIsDynamic":false}},"filenames":{"mark-header":"ladders-a4-portrait-header","svg-check-flow":"sidebar-312pt-check-flow","svg-mark-flow":"sidebar-312pt-mark-flow","svg-mark-ladder":"sidebar-312pt-mark-ladder","svg-moderate-active":"sidebar-312pt-moerate-flow-alt-active","svg-moderate-inactive":"sidebar-312pt-moderate-inactive"},"previousImageDimStatic":{"mark":{"w":595.2755905511812,"h":839.0551181102363},"moderate":{"w":763.2376157480315,"h":881.5748031496064}},"previousImageDimDynamic":{"width-check":{"dim":{"w":1.417039398425197,"h":881.5748031496064},"widthIsDynamic":true,"heightIsDynamic":false}}}`

func TestDefineLayoutFromSvg(t *testing.T) {
	svgFilename := "./test/layout-312pt-static-mark-dynamic-moderate-static-check.svg"
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

	if !reflect.DeepEqual(want.PreviousImageStatic, got.PreviousImageStatic) {
		t.Errorf("PreviousImageStatic are different\n%v\n%v", want.PreviousImageStatic, got.PreviousImageStatic)
	}
	if !reflect.DeepEqual(want.PreviousImageDynamic, got.PreviousImageDynamic) {
		t.Errorf("PreviousImageDynamic are different\n%v\n%v", want.PreviousImageDynamic, got.PreviousImageDynamic)
	}
	if !reflect.DeepEqual(want.Filenames, got.Filenames) {
		t.Errorf("Filenames are different\n%v\n%v", want.Filenames, got.Filenames)
	}

}
