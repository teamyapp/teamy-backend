package stream

func Map[InItem any, OutItem any](input <-chan InItem, transform func(item InItem) OutItem) <-chan OutItem {
	output := make(chan OutItem)
	go func() {
		for item := range input {
			output <- transform(item)
		}
		close(output)
	}()
	return output
}
