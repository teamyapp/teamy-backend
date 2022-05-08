package collect

func Filter[Item any](items []Item, match func(item Item) bool) []Item {
	matchedItems := make([]Item, 0)
	for _, item := range items {
		if match(item) {
			matchedItems = append(matchedItems, item)
		}
	}

	return matchedItems
}
