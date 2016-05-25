package maths

func Max(vals []int) int {
	var max = vals[0]
	for _, v := range vals {
		if v > max {
			max = v
		}
	}
	return max
}

func Min(vals []int) int {
	var min = vals[0]
	for _, v := range vals {
		if v < min {
			min = v
		}
	}
	return min
}
