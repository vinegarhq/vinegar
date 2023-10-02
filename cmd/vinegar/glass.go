package main

func (b *Binary) EmptyGlass(exit <-chan bool) {
	for {
		select {
		case <-b.log:
		case <-b.progress:
		case <-exit:
			return
		}
	}
}
