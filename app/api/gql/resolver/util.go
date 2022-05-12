package resolver

func contains(arr []uint64, element uint64) bool {
	for _, e := range arr {
		if e == element {
			return true
		}
	}
	return false
}
