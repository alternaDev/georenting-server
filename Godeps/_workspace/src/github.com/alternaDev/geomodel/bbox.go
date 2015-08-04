package geomodel

type BoundingBox struct {
	latNE float64
	lonNE float64
	latSW float64
	lonSW float64
}

func NewBoundingBox(north, east, south, west float64) BoundingBox {
	var north_, south_ float64
	if south > north {
		south_ = north
		north_ = south
	} else {
		south_ = south
		north_ = north
	}

	return BoundingBox{north_, east, south_, west}
}
