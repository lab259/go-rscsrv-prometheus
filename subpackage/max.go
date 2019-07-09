package subpackage

// Max returns the maximum value between `i`s
func Max(i ...int) int {
	var m int
	for _, a := range i {
		if a > m {
			m = a
		}
	}
	return m
}
