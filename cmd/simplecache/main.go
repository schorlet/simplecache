// Command simplecache helps reading Chromium simple cache on command line.
//
//  Usage:
//	simplecache command [flag] CACHEDIR
//
//	The commands are:
//		list        list entries
//		header      print entry header
//		body        print entry body
//
//	The flags are:
//		-url string        entry url
//		-hash string       entry hash
//
//	CACHEDIR is the path to the chromium cache directory.
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

const usage = `simplecache is a tool for reading Chromium simple cache v6 or v7.

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

func main() {
	log.SetFlags(0)

	var cmd, url, hash, cachedir string
	parseArgs(&cmd, &url, &hash, &cachedir)

	if cmd == "list" {
		cache, err := simplecache.Open(cachedir)
		if err != nil {
			log.Fatalf("Unable to open cache: %v", err)
		}

		for _, url := range cache.URLs() {
			hash := simplecache.Hash(url)
			fmt.Printf("%016x\t%s\n", hash, url)
		}

	} else if cmd == "header" {
		entry := openEntry(url, hash, cachedir)
		printHeader(entry)

	} else if cmd == "body" {
		entry := openEntry(url, hash, cachedir)
		printBody(entry)

	} else {
		log.Fatalf("Unknown command: %s", cmd)
	}
}

func parseArgs(cmd, url, hash, cachedir *string) {
	if len(os.Args) == 1 {
		log.Fatal(usage)
	}

	// cmd
	*cmd = os.Args[1]

	// flags
	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.Usage = func() { log.Println(usage) }

	flags.StringVar(url, "url", "", "entry url")
	flags.StringVar(hash, "hash", "", "entry hash")

	err := flags.Parse(os.Args[2:])
	if err != nil {
		log.Fatalf("Unable to parse args: %v", err)
	}

	if *cmd != "list" && flags.NFlag() != 1 {
		log.Fatal(usage)
	}

	if flags.NArg() != 1 {
		log.Fatal(usage)
	}

	*cachedir = flags.Arg(0)
}

func openEntry(url, hash, dir string) *simplecache.Entry {
	var id uint64
	var err error

	if hash != "" {
		id, err = strconv.ParseUint(hash, 16, 64)
	} else {
		id = simplecache.Hash(url)
	}

	if err != nil {
		log.Fatalf("Unable to parse hash: %v", err)
	}

	entry, err := simplecache.OpenEntry(id, dir)
	if err != nil {
		log.Fatalf("Unable to open entry: %v", err)
	}

	return entry
}

func printHeader(entry *simplecache.Entry) {
	header, err := entry.Header()
	if err != nil {
		log.Fatalf("Unable to read header: %v", err)
	}
	for key := range header {
		fmt.Printf("%s: %s\n", key, header.Get(key))
	}
}

func printBody(entry *simplecache.Entry) {
	body, err := entry.Body()
	if err != nil {
		log.Fatalf("Unable to read body: %v", err)
	}
	defer body.Close()

	_, err = io.Copy(os.Stdout, body)
	if err != nil {
		log.Fatalf("Unable to copy body to stdout: %v", err)
	}

	err = body.Close()
	if err != nil {
		log.Fatalf("Unable to close body: %v", err)
	}
}
