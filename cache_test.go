package simplecache_test

import (
	"io"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/schorlet/simplecache"
)

func TestCrawl(t *testing.T) {
	cache, err := simplecache.Open("testdata")
	if err != nil {
		t.Fatal(err)
	}

	_, err = cache.OpenURL("http://foo.com")
	if err == nil {
		t.Fatalf("got: nil, want: an error")
	}

	urls := cache.URLs()
	if len(urls) == 0 {
		t.Fatal("empty cache")
	}

	for _, url := range urls {
		openURL(t, cache, url)
	}
}

func openURL(t *testing.T, cache *simplecache.Cache, url string) {
	entry, err := cache.OpenURL(url)
	if err != nil {
		t.Fatal(err)
	}

	if entry.URL != url {
		t.Fatalf("bad url: %s, want: %s", entry.URL, url)
	}

	header, err := entry.Header()
	if err != nil {
		t.Fatal(err)
	}

	if len(header) == 0 {
		t.Fatal("got: empty header")
	}

	clength := header.Get("Content-Length")
	nlength, err := strconv.ParseInt(clength, 10, 64)
	if err != nil {
		t.Fatal(err)
	}

	body, err := entry.Body()
	if err != nil {
		t.Fatal(err)
	}

	n, err := io.Copy(ioutil.Discard, body)
	if err != nil {
		t.Fatal(err)
	}

	err = body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if n != nlength {
		t.Fatalf("bad stream-length: %d, want: %d", n, nlength)
	}
}

func TestEntry(t *testing.T) {
	url := "https://golang.org/doc/gopher/pkg.png"
	hash := simplecache.Hash(url)

	entry, err := simplecache.OpenEntry(hash, "testdata")
	if err != nil {
		t.Fatal(err)
	}

	if entry.URL != url {
		t.Fatalf("bad url: %s, want: %s", entry.URL, url)
	}

	header, err := entry.Header()
	if err != nil {
		t.Fatal(err)
	}

	cl := header.Get("Content-Length")
	if cl != "5409" {
		t.Fatalf("bad content-length: %s, want: 5409", cl)
	}

	ct := header.Get("Content-Type")
	if ct != "image/png" {
		t.Fatalf("bad content-type: %s, want: image/png", ct)
	}

	body, err := entry.Body()
	if err != nil {
		t.Fatal(err)
	}

	n, err := io.Copy(ioutil.Discard, body)
	if err != nil {
		t.Fatal(err)
	}

	if n != 5409 {
		t.Fatalf("bad stream length: %d, want: 5409", n)
	}

	err = body.Close()
	if err != nil {
		t.Fatal(err)
	}
}
