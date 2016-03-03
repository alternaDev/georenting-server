package geomodel

import (
  "reflect"
  "math"
  "strings"
)

func deleteRecords(data []string, remove []string) []string {
    w := 0 // write index

loop:
    for _, x := range data {
        for _, id := range remove {
            if id == x {
                continue loop
            }
        }
        data[w] = x
        w++
    }
    return data[:w]
}

func contains(data []LocationComparableTuple, e LocationComparableTuple) bool {
	for _, a := range data {
		if reflect.DeepEqual(a, e) {
			return true
		}
	}
	return false
}

func DegToRad(val float64) float64 {
	return (math.Pi / 180) * val
}

func Adjacent(cell string, dir []int) string {
	var dx int = dir[0]
	var dy int = dir[1]
	var i  int = len(cell) - 1

	for i >= 1 && (dx != 0 || dy != 0) {
		var l []int = SubdivXY(rune(cell[i]))
		var x int = l[0]
		var y int = l[1]

		// Horizontal
		if dx == -1 {
			if x == 0 {
				x = GEOCELL_GRID_SIZE - 1
			} else {
				x--
				dx = 0
			}
		} else if dx == 1 {
			if x == GEOCELL_GRID_SIZE - 1 {
				x = 0
			} else {
				x++
				dx = 0
			}
		}

		// Vertical
		if dy == 1 {
			if y == GEOCELL_GRID_SIZE - 1 {
				y = 0
			} else {
				y++
				dy = 0
			}
		} else if dy == -1 {
			if y == 0 {
				y = GEOCELL_GRID_SIZE - 1
			} else {
				y--
				dy = 0
			}
		}

		var l2 []int = []int{x, y}
		cell = string(append([]byte(cell[:i - 1]), SubdivChar(l2)))
		if i < len(cell) {
			cell = string(append([]byte(cell), []byte(cell[i + 1:])...))
		}
		i--
	}

	if dy != 0 {
		return ""
	}

	return cell
}

func SubdivXY(char_ rune) []int {
	var charI int = strings.IndexRune(GEOCELL_ALPHABET, char_)
	return []int{(charI & 4) >> 1 | (charI & 1) >> 0, (charI & 8) >> 2 | (charI & 2) >> 1}
}

func SubdivChar(pos []int) uint8 {
	return GEOCELL_ALPHABET[(pos[1] & 2) << 2 | (pos[0] & 2) << 1 | (pos[1] & 1) << 1 | (pos[0] & 1) << 0]
}
