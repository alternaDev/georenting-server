package maths

// Max calculates the largest value in an int slice.
func Max(vals []int) int {
	var max = vals[0]
	for _, v := range vals {
		if v > max {
			max = v
		}
	}
	return max
}

// Min calculates the minimum value in an int slice.
func Min(vals []int) int {
	var min = vals[0]
	for _, v := range vals {
		if v < min {
			min = v
		}
	}
	return min
}
