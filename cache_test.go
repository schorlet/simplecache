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
		t.Fatal("urls is empty")
	}

	for i := range urls {
		testEntry(t, urls[i], "testdata")
	}
}

func testEntry(t *testing.T, url, path string) {
	entry, err := simplecache.Get(url, path)
	if err != nil {
		t.Fatalf("get entry: %v", err)
	}
	if entry.URL != url {
		t.Fatalf("url: %s, want: %s", entry.URL, url)
	}

	header, err := entry.Header()
	if err != nil {
		t.Fatalf("header: %v", err)
	}
	if len(header) == 0 {
		t.Fatal("header is empty")
	}
	clength := header.Get("Content-Length")
	nlength, err := strconv.ParseInt(clength, 10, 64)
	if err != nil {
		t.Fatal(err)
	}

	body, err := entry.Body()
	if err != nil {
		t.Fatalf("body: %v", err)
	}
	n, err := io.Copy(ioutil.Discard, body)
	if err != nil {
		t.Fatalf("discard body: %v", err)
	}
	if err = body.Close(); err != nil {
		t.Fatalf("close body: %v", err)
	}
	if n != nlength {
		t.Fatalf("body stream-length: %d, want: %d", n, nlength)
	}
}

func TestBadURL(t *testing.T) {
	_, err := simplecache.Get("http://foo.com", "testdata")
	if err == nil {
		t.Fatalf("err is nil")
	}
}
