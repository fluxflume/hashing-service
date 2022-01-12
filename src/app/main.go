package main

import (
	"time"
)

func main() {
	// A max value to limit the number of outstanding hash requests.
	// If the buffer/channel becomes full, POST requests to create new hashed items will fail.
	maxPendingItems := 100_000

	// The amount of time to wait before allowing a hashed item to be available
	hashCompletionDelay := 5 * time.Second

	// Fixed num of workers for concurrent processing of hashed items
	numWorkers := 1000

	// The port for the server to bind to
	port := 3000

	StartLocalApp(port, maxPendingItems, numWorkers, hashCompletionDelay)
}
