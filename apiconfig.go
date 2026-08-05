package main

import "sync/atomic"

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (c *apiConfig) increment() {
	c.fileserverHits.Add(1)
}

func (c *apiConfig) resetHits() {
	c.fileserverHits.Store(0)
}

// func (c *apiConfig) printHits() string {
// 	counter := c.fileserverHits.Load()
// 	Text := "Hits: " + strconv.Itoa(int(counter))
// 	return Text
// }

func (c *apiConfig) hitsCount() int {
	counter := c.fileserverHits.Load()
	return int(counter)
}
