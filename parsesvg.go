package parsesvg

import "encoding/xml"

func ParseSvg(input []byte) *Csvg__svg {

	var svg Csvg__svg

	xml.Unmarshal(input, &svg)

	return &svg
}
