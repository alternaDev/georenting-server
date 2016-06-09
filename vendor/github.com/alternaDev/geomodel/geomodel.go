/*
   Copyright 2012 Alexander Yngling

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package geomodel

import "math"
import "sort"
import "log"

const (
	GEOCELL_GRID_SIZE      = 4
	GEOCELL_ALPHABET       = "0123456789bcdefghjkmnpqrstuvwxyz"
	MAX_GEOCELL_RESOLUTION = 13 // The maximum *practical* geocell resolution.
)

var (
	NORTHWEST = []int{-1, 1}
	NORTH     = []int{0, 1}
	NORTHEAST = []int{1, 1}
	EAST      = []int{1, 0}
	SOUTHEAST = []int{1, -1}
	SOUTH     = []int{0, -1}
	SOUTHWEST = []int{-1, -1}
	WEST      = []int{-1, 0}
)

type LocationCapable interface {
	Latitude() float64
	Longitude() float64
	Key() string
	Geocells() []string
}

type LocationComparableTuple struct {
	first  LocationCapable
	second float64
}

type IntArrayDoubleTuple struct {
	first  []int
	second float64
}

type ByDistanceIA []IntArrayDoubleTuple

func (a ByDistanceIA) Len() int           { return len(a) }
func (a ByDistanceIA) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDistanceIA) Less(i, j int) bool { return a[i].second < a[j].second }

type ByDistance []LocationComparableTuple

func (a ByDistance) Len() int           { return len(a) }
func (a ByDistance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDistance) Less(i, j int) bool { return a[i].second < a[j].second }

type RepositorySearch func([]string) []LocationCapable

func GeoHash(lat, lon float64, resolution int) string {
	return GeoCell(lat, lon, resolution)
}

func GeoCell(lat, lon float64, resolution int) string {
	resolution = resolution + 1

	north := 90.0
	south := -90.0
	east := 180.0
	west := -180.0
	isEven := true
	mid := 0.0
	ch := 0
	bit := 0
	bits := []int{16, 8, 4, 2, 1}
	cell := make([]byte, resolution, resolution)

	i := 0

	for i = 0; i < resolution; {
		if isEven {
			mid = (west + east) / 2
			if lon > mid {
				ch |= bits[bit]
				west = mid
			} else {
				east = mid
			}
		} else {
			mid = (south + north) / 2
			if lat > mid {
				ch |= bits[bit]
				south = mid
			} else {
				north = mid
			}
		}
		isEven = !isEven
		if bit < 4 {
			bit = bit + 1
		} else {
			cell[i] = GEOCELL_ALPHABET[ch]
			i = i + 1
			bit = 0
			ch = 0
		}
	}

	cell[i-1] = 0

	return string(cell)
}

func GeoCells(lat, lon float64, resolution int) []string {
	g := GeoCell(lat, lon, resolution)
	cells := make([]string, len(g), len(g))
	for i := 0; i < resolution; i++ {
		cells[i] = g[0 : i+1]
	}
	return cells
}

func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	var p1lat = DegToRad(lat1)
	var p1lon = DegToRad(lon1)
	var p2lat = DegToRad(lat2)
	var p2lon = DegToRad(lon2)
	return 6378135 * math.Acos(math.Sin(p1lat)*math.Sin(p2lat)+math.Cos(p1lat)*math.Cos(p2lat)*math.Cos(p2lon-p1lon))
}

func DistanceSortedEdges(cells []string, lat, lon float64) []IntArrayDoubleTuple {
	var boxes []BoundingBox = make([]BoundingBox, 0, len(cells))
	for _, cell := range cells {
		boxes = append(boxes, ComputeBox(cell))
	}

	var maxNorth float64 = -math.MaxFloat64
	var maxEast float64 = -math.MaxFloat64
	var maxSouth float64 = -math.MaxFloat64
	var maxWest float64 = -math.MaxFloat64

	for _, box := range boxes {
		maxNorth = math.Max(maxNorth, box.latNE)
		maxEast = math.Max(maxEast, box.lonNE)
		maxSouth = math.Max(maxSouth, box.latSW)
		maxWest = math.Max(maxWest, box.lonSW)
	}

	result := make([]IntArrayDoubleTuple, 4)
	result[0] = IntArrayDoubleTuple{SOUTH, Distance(maxSouth, lon, lat, lon)}
	result[1] = IntArrayDoubleTuple{NORTH, Distance(maxNorth, lon, lat, lon)}
	result[2] = IntArrayDoubleTuple{WEST, Distance(lat, maxWest, lat, lon)}
	result[3] = IntArrayDoubleTuple{EAST, Distance(maxSouth, maxEast, lat, lon)}

	sort.Sort(ByDistanceIA(result))

	return result
}

func ComputeBox(cell string) BoundingBox {
	var bbox BoundingBox
	if cell == "" {
		return bbox
	}

	bbox = NewBoundingBox(90.0, 180.0, -90.0, -180.0)
	for len(cell) > 0 {
		var subcellLonSpan float64 = (bbox.lonNE - bbox.lonSW) / GEOCELL_GRID_SIZE
		var subcellLatSpan float64 = (bbox.latNE - bbox.latSW) / GEOCELL_GRID_SIZE

		var l []int = SubdivXY(rune(cell[0]))
		var x int = l[0]
		var y int = l[1]

		bbox = NewBoundingBox(bbox.latSW+subcellLatSpan*(float64(y)+1),
			bbox.lonSW+subcellLonSpan*(float64(x)+1),
			bbox.latSW+subcellLatSpan*float64(y),
			bbox.lonSW+subcellLonSpan*float64(x))
		cell = cell[1:]
	}

	return bbox
}

func ProximityFetch(lat, lon float64, maxResults int, maxDistance float64, search RepositorySearch, maxResolution int) []LocationCapable {
	var results []LocationComparableTuple

	// The current search geocell containing the lat,lon.
	var curContainingGeocell string = GeoCell(lat, lon, maxResolution)

	var searchedCells []string = make([]string, 0)

	/*
	 * The currently-being-searched geocells.
	 * NOTES:
	 * Start with max possible.
	 * Must always be of the same resolution.
	 * Must always form a rectangular region.
	 * One of these must be equal to the cur_containing_geocell.
	 */
	var curGeocells []string = make([]string, 0)
	curGeocells = append(curGeocells, curContainingGeocell)
	var closestPossibleNextResultDist float64 = 0

	var noDirection = []int{0, 0}

	var sortedEdgeDistances []IntArrayDoubleTuple
	sortedEdgeDistances = append(sortedEdgeDistances, IntArrayDoubleTuple{noDirection, 0})

	for len(curGeocells) != 0 {
		closestPossibleNextResultDist = sortedEdgeDistances[0].second
		if maxDistance > 0 && closestPossibleNextResultDist > maxDistance {
			break
		}

		var curTempUnique = deleteRecords(curGeocells, searchedCells)

		var curGeocellsUnique = curTempUnique

		var newResultEntities = search(curGeocellsUnique)

		searchedCells = append(searchedCells, curGeocells...)

		// Begin storing distance from the search result entity to the
		// search center along with the search result itself, in a tuple.
		var newResults []LocationComparableTuple = make([]LocationComparableTuple, 0, len(newResultEntities))
		for _, entity := range newResultEntities {
			newResults = append(newResults, LocationComparableTuple{entity, Distance(lat, lon, entity.Latitude(), entity.Longitude())})
		}

		sort.Sort(ByDistance(newResults))
		newResults = newResults[0:int(math.Min(float64(maxResults), float64(len(newResults))))]

		// Merge new_results into results
		for _, tuple := range newResults {
			// contains method will check if entity in tuple have same key
			if !contains(results, tuple) {
				results = append(results, tuple)
			}
		}

		sort.Sort(ByDistance(results))
		results = results[0:int(math.Min(float64(maxResults), float64(len(results))))]

		sortedEdgeDistances = DistanceSortedEdges(curGeocells, lat, lon)

		if len(results) == 0 || len(curGeocells) == 4 {
			/* Either no results (in which case we optimize by not looking at
			   adjacents, go straight to the parent) or we've searched 4 adjacent
			   geocells, in which case we should now search the parents of those
			   geocells.*/
			curContainingGeocell = curContainingGeocell[:int(math.Max(float64(len(curContainingGeocell))-1, float64(0)))]

			if len(curContainingGeocell) == 0 {
				break
			}

			var oldCurGeocells []string = curGeocells
			curGeocells = make([]string, 0)

			for _, cell := range oldCurGeocells {
				if len(cell) > 0 {
					var newCell string = cell[:len(cell)-1]
					i := sort.SearchStrings(curGeocells, newCell)
					if !(i < len(curGeocells) && curGeocells[i] == newCell) {
						curGeocells = append(curGeocells, newCell)
					}
				}
			}

			if len(curGeocells) == 0 {
				break
			}
		} else if len(curGeocells) == 1 {
			var nearestEdge []int = sortedEdgeDistances[0].first
			curGeocells = append(curGeocells, Adjacent(curGeocells[0], nearestEdge))
		} else if len(curGeocells) == 2 {
			var nearestEdge []int = DistanceSortedEdges([]string{curContainingGeocell}, lat, lon)[0].first
			var perpendicularNearestEdge []int = []int{0, 0}

			if nearestEdge[0] == 0 {
				for _, edgeDistance := range sortedEdgeDistances {
					if edgeDistance.first[0] != 0 {
						perpendicularNearestEdge = edgeDistance.first
						break
					}
				}
			} else {
				for _, edgeDistance := range sortedEdgeDistances {
					if edgeDistance.first[0] == 0 {
						perpendicularNearestEdge = edgeDistance.first
						break
					}
				}
			}

			var tempCells []string = make([]string, 0)

			for _, cell := range curGeocells {
				tempCells = append(tempCells, Adjacent(cell, perpendicularNearestEdge))
			}

			curGeocells = append(curGeocells, tempCells...)
		}

		if len(results) < maxResults {
			// Keep Searchin!
			log.Printf("%d results found but want %d results, continuing...", len(results), maxResults)
			continue
		}

		// Found things!
		log.Printf("%d results found.", len(results))

		var currentFarthestReturnableResultDist float64 = Distance(lat, lon, results[maxResults-1].first.Latitude(), results[maxResults-1].first.Longitude())

		if closestPossibleNextResultDist >= currentFarthestReturnableResultDist {
			// Done
			log.Printf("DONE next result at least %d away, current farthest is %d dist.", closestPossibleNextResultDist, currentFarthestReturnableResultDist)
			break
		}

		log.Printf("next result at least %d away, current farthest is %d dist", closestPossibleNextResultDist, currentFarthestReturnableResultDist)

	}

	var result []LocationCapable = make([]LocationCapable, 0)

	for _, entry := range results[0:int(math.Min(float64(maxResults), float64(len(results))))] {
		if maxDistance == 0 || entry.second < maxDistance {
			result = append(result, entry.first)
		}
	}

	return result
}
