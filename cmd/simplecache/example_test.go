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

	var output bytes.Buffer
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	lines := read(&output)
	for i := range lines {
		fmt.Println(lines[i])
	}

	// Output:
	// https://ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js
	// https://golang.org/doc/gopher/pkg.png
	// https://golang.org/favicon.ico
	// https://golang.org/lib/godoc/godocs.js
	// https://golang.org/lib/godoc/jquery.treeview.css
	// https://golang.org/lib/godoc/jquery.treeview.edit.js
	// https://golang.org/lib/godoc/jquery.treeview.js
	// https://golang.org/lib/godoc/playground.js
	// https://golang.org/lib/godoc/style.css
	// https://golang.org/pkg/
	// https://golang.org/pkg/bufio/
	// https://golang.org/pkg/builtin/
	// https://golang.org/pkg/bytes/
	// https://golang.org/pkg/io/
	// https://golang.org/pkg/io/ioutil/
	// https://golang.org/pkg/os/
	// https://golang.org/pkg/strconv/
	// https://golang.org/pkg/strings/
	// https://ssl.google-analytics.com/ga.js
}

func Example_header() {
	cmd := exec.Command("./simplecache", "header", "https://golang.org/doc/gopher/pkg.png", "../../testdata")

	var output bytes.Buffer
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	lines := read(&output)
	for i := range lines {
		fmt.Println(lines[i])
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
	cmd := exec.Command("./simplecache", "body", "https://golang.org/doc/gopher/pkg.png", "../../testdata")

	var output bytes.Buffer
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	config, err := png.DecodeConfig(&output)
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
