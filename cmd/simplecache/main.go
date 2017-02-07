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

	var cmd, url, hash, cachedir string
	parseArgs(&cmd, &url, &hash, &cachedir)

	if cmd == "list" {
		cache, err := simplecache.Open(cachedir)
		if err != nil {
			log.Fatal(err)
		}

		for _, url := range cache.URLs() {
			hash := simplecache.EntryHash(url)
			fmt.Printf("%016x %s\n", hash, url)
		}

	} else if cmd == "header" {
		entry := openEntry(url, hash, cachedir)
		printHeader(entry)

	} else if cmd == "body" {
		entry := openEntry(url, hash, cachedir)
		printBody(entry)

	} else {
		log.Fatalf("unknown command: %s", cmd)
	}
}

func parseArgs(cmd, url, hash, cachedir *string) {
	if len(os.Args) == 1 {
		printUsage()
	}

	// cmd
	*cmd = os.Args[1]

	// flags
	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.Usage = printUsage

	flags.StringVar(url, "url", "", "entry url")
	flags.StringVar(hash, "hash", "", "entry hash")

	err := flags.Parse(os.Args[2:])
	if err != nil {
		log.Fatal(err)
	}

	if *cmd != "list" && flags.NFlag() != 1 {
		printUsage()
	}

	if flags.NArg() != 1 {
		printUsage()
	}

	*cachedir = flags.Arg(0)
}

func openEntry(url, hash, dir string) *simplecache.Entry {
	var id uint64
	var err error

	if hash != "" {
		id, err = strconv.ParseUint(hash, 16, 64)
	} else {
		id = simplecache.EntryHash(url)
	}

	if err != nil {
		log.Fatal(err)
	}

	entry, err := simplecache.OpenEntry(id, dir)
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

	err = body.Close()
	if err != nil {
		log.Fatal(err)
	}
}
