# simplecache

This is the simplecache tool to read the Chromium simple cache from command line.


## Usage

```
simplecache command [url] path

The commands are:
	list        print cache urls
	header      print url header
	body        print url body

path is the path to the chromium cache directory.
```

## Examples

```sh
$ URL=https://golang.org/doc/gopher/pkg.png
$ CHROME_CACHE=../../testdata/
```

### List all entries

```sh
$ simplecache list $CHROME_CACHE
https://golang.org/pkg/strconv/
https://golang.org/pkg/strings/
https://golang.org/pkg/io/ioutil/
https://golang.org/pkg/io/
https://golang.org/pkg/bytes/
https://golang.org/pkg/bufio/
https://golang.org/favicon.ico
https://golang.org/lib/godoc/style.css
https://golang.org/doc/gopher/pkg.png
https://golang.org/lib/godoc/jquery.treeview.js
https://ssl.google-analytics.com/ga.js
https://golang.org/pkg/os/
https://golang.org/lib/godoc/jquery.treeview.css
https://golang.org/lib/godoc/jquery.treeview.edit.js
https://golang.org/lib/godoc/playground.js
https://ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js
https://golang.org/pkg/builtin/
https://golang.org/pkg/
https://golang.org/lib/godoc/godocs.js
```

### Print entry header

```sh
$ simplecache header $URL $CHROME_CACHE
Status: 200
Content-Type: image/png
Content-Length: 5409
Last-Modified: Thu, 19 May 2016 18:04:32 GMT
Date: Sun, 17 Jul 2016 18:30:09 GMT
Accept-Ranges: bytes
Server: Google Frontend
X-Cloud-Trace-Context: b75923ae8631de089fbc3f00e79cc992
Alternate-Protocol: 443:quic
Alt-Svc: quic=":443"; ma=2592000; v="36,35,34,33,32,31,30,29,28,27,26,25"
```

### Print entry body

```sh
$ simplecache body $URL $CHROME_CACHE | file -
/dev/stdin: PNG image data, 83 x 120, 8-bit grayscale, non-interlaced
```

```sh
$ simplecache body $URL $CHROME_CACHE | hexdump -C -n 32
00000000  89 50 4e 47 0d 0a 1a 0a  00 00 00 0d 49 48 44 52  |.PNG........IHDR|
00000010  00 00 00 53 00 00 00 78  08 00 00 00 00 ab b2 91  |...S...x........|
00000020
```


### Watch webm videos:

```sh
$ CHROME_CACHE=~/.cache/chromium/Default/Media\ Cache/
$ simplecache list "$CHROME_CACHE" | grep 'webm$' | \
while read URL; do
	simplecache body $URL "$CHROME_CACHE" | vlc -q --play-and-exit -
done
```
