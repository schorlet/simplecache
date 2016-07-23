// Command simplecache helps reading Chromium simple cache on command line.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/schorlet/simplecache"
)

const usage = `simplecache is a tool for reading Chromium simple cache v6.

Usage:

    simplecache command [flag] CACHEDIR

The commands are:
    list        list entries
    header      print entry header
    body        print entry body

The flags are:
    -url string        entry url
    -hash string       entry hash

CACHEDIR is the path to the chromium cache directory.
`

func printUsage() {
	log.Println(usage)
	os.Exit(2)
}

func main() {
	log.SetFlags(0)

	if len(os.Args) == 1 {
		printUsage()
	}

	// command
	command := os.Args[1]

	// flags
	cmdline := flag.NewFlagSet("", flag.ExitOnError)
	cmdline.Usage = printUsage
	aURL := cmdline.String("url", "", "entry url")
	aHash := cmdline.String("hash", "", "entry hash")

	err := cmdline.Parse(os.Args[2:])
	if err != nil {
		log.Fatal("error:", err)
	}

	if cmdline.NArg() != 1 {
		printUsage()
	}
	cachedir := cmdline.Arg(0)

	// exec
	if command == "list" {
		cache, err := simplecache.Open(cachedir)
		if err != nil {
			log.Fatal(err)
		}

		for _, url := range cache.URLs() {
			hash := simplecache.EntryHash(url)
			fmt.Printf("%016x %s\n", hash, url)
		}

	} else if cmdline.NFlag() != 1 {
		printUsage()

	} else {
		entry := openEntry(*aURL, *aHash, cachedir)

		if command == "header" {
			printHeader(entry)

		} else if command == "body" {
			printBody(entry)

		} else {
			log.Fatalf("unknown command: %s", command)
		}
	}
}

func openEntry(aURL, aHash, dir string) *simplecache.Entry {
	var hash uint64
	var err error

	if aHash != "" {
		hash, err = strconv.ParseUint(aHash, 16, 64)
	} else {
		hash = simplecache.EntryHash(aURL)
	}

	if err != nil {
		log.Fatal(err)
	}

	entry, err := simplecache.OpenEntry(hash, dir)
	if err != nil {
		log.Fatal(err)
	}

	return entry
}

func printHeader(entry *simplecache.Entry) {
	header, err := entry.Header()
	if err != nil {
		log.Fatal(err)
	}
	for key := range header {
		fmt.Printf("%s: %s\n", key, header.Get(key))
	}
}

func printBody(entry *simplecache.Entry) {
	body, err := entry.Body()
	if err != nil {
		log.Fatal(err)
	}
	defer body.Close()

	_, err = io.Copy(os.Stdout, body)
	if err != nil {
		log.Fatal(err)
	}
}
