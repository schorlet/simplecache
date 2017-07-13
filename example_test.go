package simplecache_test

import (
	"fmt"
	"image/png"
	"log"

	"github.com/schorlet/simplecache"
)

// Example gets an entry from the cache and prints to stdout.
func Example() {
	cache, err := simplecache.Open("testdata")
	if err != nil {
		log.Fatal(err)
	}

	entry, err := cache.OpenURL("https://golang.org/doc/gopher/pkg.png")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(entry.URL)

	// header
	header, err := entry.Header()
	if err != nil {
		log.Fatal(err)
	}
	for _, key := range []string{"Status", "Content-Length", "Content-Type"} {
		fmt.Printf("%s: %s\n", key, header.Get(key))
	}

	// body
	body, err := entry.Body()
	if err != nil {
		log.Fatal(err)
	}
	config, err := png.DecodeConfig(body)
	if err != nil {
		log.Fatal(err)
	}
	defer body.Close()

	fmt.Printf("PNG image data, %d x %d\n", config.Width, config.Height)

	// Output:
	// https://golang.org/doc/gopher/pkg.png
	// Status: 200
	// Content-Length: 5409
	// Content-Type: image/png
	// PNG image data, 83 x 120
}
