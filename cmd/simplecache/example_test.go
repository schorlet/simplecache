package main_test

import (
	"bufio"
	"bytes"
	"fmt"
	"image/png"
	"io"
	"log"
	"os/exec"
	"sort"
)

func init() {
	cmd := exec.Command("go", "build")

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func Example_list() {
	cmd := exec.Command("./simplecache", "list", "../../testdata")

	output := new(bytes.Buffer)
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	lines := read(output)
	for _, line := range lines {
		fmt.Println(line)
	}

	// Output:
	// 329f9c2d34eb0523	https://golang.org/pkg/os/
	// 36d2a4716c77194d	https://golang.org/lib/godoc/jquery.treeview.js
	// 4d522977a9d92736	https://golang.org/pkg/
	// 4e5b177c943d8ee0	https://golang.org/pkg/io/ioutil/
	// 51d54ab35ce343ea	https://golang.org/lib/godoc/godocs.js
	// 55782b6621f25a58	https://golang.org/pkg/io/
	// 6a5a092a607295ea	https://golang.org/pkg/bufio/
	// 8e8dcd288a0d7920	https://ssl.google-analytics.com/ga.js
	// 9d38f3624ed4a85c	https://golang.org/favicon.ico
	// a0a6f47a8175a75e	https://golang.org/pkg/strconv/
	// a95a6bc37488af73	https://ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js
	// bb9d1cda868d278c	https://golang.org/doc/gopher/pkg.png
	// c2e0a9bb2e5b256d	https://golang.org/lib/godoc/jquery.treeview.css
	// c47ff3921af67e65	https://golang.org/pkg/bytes/
	// c6b1c75ee113a942	https://golang.org/lib/godoc/style.css
	// df635ac21e8a65f1	https://golang.org/pkg/strings/
	// eab39d1ceb121cdc	https://golang.org/lib/godoc/playground.js
	// f21a90b578066ccc	https://golang.org/lib/godoc/jquery.treeview.edit.js
	// fb4ae632c995772d	https://golang.org/pkg/builtin/
}

func Example_header() {
	cmd := exec.Command("./simplecache", "header", "-hash", "bb9d1cda868d278c", "../../testdata")

	output := new(bytes.Buffer)
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	lines := read(output)
	for _, line := range lines {
		fmt.Println(line)
	}

	// Output:
	// Accept-Ranges: bytes
	// Alt-Svc: quic=":443"; ma=2592000; v="36,35,34,33,32,31,30,29,28,27,26,25"
	// Alternate-Protocol: 443:quic
	// Content-Length: 5409
	// Content-Type: image/png
	// Date: Sun, 17 Jul 2016 18:30:09 GMT
	// Last-Modified: Thu, 19 May 2016 18:04:32 GMT
	// Server: Google Frontend
	// Status: 200
	// X-Cloud-Trace-Context: b75923ae8631de089fbc3f00e79cc992
}

func Example_body() {
	cmd := exec.Command("./simplecache", "body", "-hash", "bb9d1cda868d278c", "../../testdata")

	output := new(bytes.Buffer)
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	config, err := png.DecodeConfig(output)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("PNG image data, %d x %d\n", config.Width, config.Height)

	// Output:
	// PNG image data, 83 x 120
}

func read(r io.Reader) []string {
	lines := make([]string, 0)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	sort.Strings(lines)
	return lines
}
