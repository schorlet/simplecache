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

	hashes := cache.Hashes()
	if len(hashes) == 0 {
		t.Fatal("empty cache hashes")
	}

	urls := cache.URLs()
	if len(urls) == 0 {
		t.Fatal("empty cache urls")
	}

	if len(hashes) != len(urls) {
		t.Fatal("mismatch len between hashes and urls")
	}

	_, err = cache.OpenURL("http://foo.com")
	if err != simplecache.ErrNotFound {
		t.Fatalf("got:%v, want:%v", err, simplecache.ErrNotFound)
	}

	for _, url := range urls {
		entry, err := cache.OpenURL(url)
		if err != nil {
			t.Fatal(err)
		}

		if entry.URL != url {
			t.Fatalf("bad url:%s, want:%s", entry.URL, url)
		}

		header, err := entry.Header()
		if err != nil {
			t.Fatal(err)
		}

		if len(header) == 0 {
			t.Fatal("got:empty header")
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
			t.Fatalf("bad stream-length:%d, want:%d", n, nlength)
		}
	}
}

func TestEntry(t *testing.T) {
	url := "https://golang.org/doc/gopher/pkg.png"
	hash := simplecache.EntryHash(url)

	entry, err := simplecache.OpenEntry(hash, "testdata")
	if err != nil {
		t.Fatal(err)
	}

	if entry.URL != url {
		t.Fatalf("bad url:%s, want:%s", entry.URL, url)
	}

	header, err := entry.Header()
	if err != nil {
		t.Fatal(err)
	}

	cl := header.Get("Content-Length")
	if cl != "5409" {
		t.Fatalf("bad content-length:%s, want:5409", cl)
	}

	ct := header.Get("Content-Type")
	if ct != "image/png" {
		t.Fatalf("bad content-type:%s, want:image/png", ct)
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
		t.Fatalf("bad stream length:%d, want:5409", n)
	}

	err = body.Close()
	if err != nil {
		t.Fatal(err)
	}
}
