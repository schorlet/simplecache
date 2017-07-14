// Command simplecache helps reading chromium simple cache v6 or v7.
//
//  Usage:
//	simplecache command [url] path
//
//	The commands are:
//		list        print cache urls
//		header      print url header
//		body        print url body
//
//	path is the path to the chromium cache directory.
package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/schorlet/simplecache"
)

const usage = `simplecache helps reading chromium simple cache v6 or v7.

Usage:
    simplecache command [url] path

The commands are:
    list        print cache urls
    header      print url header
    body        print url body

path is the path to the chromium cache directory.
`

func main() {
	log.SetFlags(0)

	var cmd, url, path string
	parseArgs(&cmd, &url, &path)

	if cmd == "list" {
		printList(path)

	} else if cmd == "header" {
		printHeader(url, path)

	} else if cmd == "body" {
		printBody(url, path)

	} else {
		log.Fatalf("Unknown command: %s", cmd)
	}
}

func parseArgs(cmd, url, path *string) {
	if len(os.Args) == 1 {
		log.Fatal(usage)
	}

	*cmd = os.Args[1]

	if *cmd == "list" {
		*path = os.Args[2]
	} else {
		*url = os.Args[2]
		*path = os.Args[3]
	}
}

func printList(path string) {
	urls, err := simplecache.URLs(path)
	if err != nil {
		log.Fatalf("Unable to open cache: %v", err)
	}

	for i := 0; i < len(urls); i++ {
		fmt.Println(urls[i])
	}
}

func printHeader(url, path string) {
	entry, err := simplecache.Get(url, path)
	if err != nil {
		log.Fatalf("Unable to open entry: %v", err)
	}

	header, err := entry.Header()
	if err != nil {
		log.Fatalf("Unable to read header: %v", err)
	}

	for key := range header {
		fmt.Printf("%s: %s\n", key, header.Get(key))
	}
}

func printBody(url, path string) {
	entry, err := simplecache.Get(url, path)
	if err != nil {
		log.Fatalf("Unable to open entry: %v", err)
	}

	body, err := entry.Body()
	if err != nil {
		log.Fatalf("Unable to read body: %v", err)
	}
	defer body.Close()

	_, err = io.Copy(os.Stdout, body)
	if err != nil {
		log.Fatalf("Unable to copy body to stdout: %v", err)
	}
}
