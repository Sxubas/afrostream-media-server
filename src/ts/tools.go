package ts

func Min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func Min32(a, b uint32) uint32 {
	if a <= b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a <= b {
		return b
	}
	return a
}

func Max32(a, b uint32) uint32 {
	if a <= b {
		return b
	}
	return a
}
