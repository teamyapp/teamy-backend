package stream

func Filter[Item any](input <-chan Item, match func(item Item) bool) <-chan Item {
	output := make(chan Item)
	go func() {
		for item := range input {
			if match(item) {
				output <- item
			}
		}
		close(output)
	}()
	return output
}
