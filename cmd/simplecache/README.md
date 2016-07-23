simplecache client
==================

### List all entries

```sh
$ simplecache list ../../testcache/
a0a6f47a8175a75e https://golang.org/pkg/strconv/
df635ac21e8a65f1 https://golang.org/pkg/strings/
4e5b177c943d8ee0 https://golang.org/pkg/io/ioutil/
55782b6621f25a58 https://golang.org/pkg/io/
c47ff3921af67e65 https://golang.org/pkg/bytes/
6a5a092a607295ea https://golang.org/pkg/bufio/
9d38f3624ed4a85c https://golang.org/favicon.ico
c6b1c75ee113a942 https://golang.org/lib/godoc/style.css
bb9d1cda868d278c https://golang.org/doc/gopher/pkg.png
36d2a4716c77194d https://golang.org/lib/godoc/jquery.treeview.js
8e8dcd288a0d7920 https://ssl.google-analytics.com/ga.js
329f9c2d34eb0523 https://golang.org/pkg/os/
c2e0a9bb2e5b256d https://golang.org/lib/godoc/jquery.treeview.css
f21a90b578066ccc https://golang.org/lib/godoc/jquery.treeview.edit.js
eab39d1ceb121cdc https://golang.org/lib/godoc/playground.js
a95a6bc37488af73 https://ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js
fb4ae632c995772d https://golang.org/pkg/builtin/
4d522977a9d92736 https://golang.org/pkg/
51d54ab35ce343ea https://golang.org/lib/godoc/godocs.js
```

### Print entry header

```sh
$ simplecache header -hash bb9d1cda868d278c ../../testcache/
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
$ simplecache body -hash bb9d1cda868d278c ../../testcache/ | file -
/dev/stdin: PNG image data, 83 x 120, 8-bit grayscale, non-interlaced
```

```sh
$ simplecache body -hash bb9d1cda868d278c ../../testcache/ | hexdump -C -n 32
00000000  89 50 4e 47 0d 0a 1a 0a  00 00 00 0d 49 48 44 52  |.PNG........IHDR|
00000010  00 00 00 53 00 00 00 78  08 00 00 00 00 ab b2 91  |...S...x........|
00000020
```

