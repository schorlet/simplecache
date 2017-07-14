package simplecache_test

import (
	"io"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/schorlet/simplecache"
)

func TestURLs(t *testing.T) {
	urls, err := simplecache.URLs("testdata")
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) == 0 {
		t.Fatal("empty cache")
	}

	for i := range urls {
		getURL(t, urls[i], "testdata")
	}
}

func getURL(t *testing.T, url, path string) {
	entry, err := simplecache.Get(url, path)
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

func TestBadURL(t *testing.T) {
	_, err := simplecache.Get("http://foo.com", "testdata")
	if err == nil {
		t.Fatalf("got: nil, want: an error")
	}
}
