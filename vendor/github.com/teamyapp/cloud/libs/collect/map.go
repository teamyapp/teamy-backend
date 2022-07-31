package collect

func Map[From any, To any](items []From, transform func(fromItem From, index int) To) []To {
	newItems := make([]To, 0)
	for index, item := range items {
		newItems = append(newItems, transform(item, index))
	}

	return newItems
}
