package parsesvg

import "github.com/timdrysdale/geo"

type TextField struct {
	Rect        geo.Rect //Corner.X/Y, Dim.W/H
	ID          string
	Prefill     string
	TabSequence int64
}

type Ladder struct {
	Anchor     geo.Point //X,Y
	Dim        geo.Dim   //W,H
	ID         string
	TextFields []TextField
}

type Layout struct {
	Anchor               geo.Point            `json:"anchor"`
	Dim                  geo.Dim              `json:"dim"`
	ID                   string               `json:"id"`
	Anchors              map[string]geo.Point `json:"anchors"`
	PageDimStatic        map[string]geo.Dim   `json:"pageDimStatic"`
	PageDimDynamic       map[string]geo.Dim   `json:"pageDimDynamic"`
	Filenames            map[string]string    `json:"filenames"`
	PreviousImageStatic  map[string]string    `json:"imageDimStatic"`
	PreviousImageDynamic map[string]string    `json:"imageDimDynamic"`
}

//TODO - structs for page and image dims
