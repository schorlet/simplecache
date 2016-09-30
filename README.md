simplecache [![GoDoc](https://godoc.org/github.com/schorlet/simplecache?status.png)](https://godoc.org/github.com/schorlet/simplecache)
===

Package simplecache provides support for reading Chromium simple cache v6 or v7.

Usage
-----

See the [example_test.go](example_test.go) for a basic example.


Documentation
-------------

The simple cache contains different files:

 Filename                 | Description
 ------------------------ | --------------------
 index                    | Fake index file
 #####_0                  | Data stream file
 #####_1                  | Data stream file
 #####_s                  | Data stream file
 index-dir/the-real-index | The real index file



### Fake index file format (index)


 offset | size | value              | description
 ------ | ---- | ------------------ | -----------
 0      | 8    | 0x656e74657220796f | Magic
 8      | 4    | 7                  | Version
 12     | 8    | 0x0                | Padding



### Real index file format (the-real-index)


Overview:

- File header
- Last write reason (starting v7)
- Index table
- Last modified


#### File header


The index file header (struct indexHeader) is 36 bytes in size and consists of:

 offset | size | value              | description
 ------ | ---- | ------------------ | -----------
 0      | 4    |                    | Payload size
 4      | 4    |                    | Payload CRC32
 8      | 8    | 0x656e74657220796f | Magic
 16     | 4    | 7                  | Version
 20     | 8    |                    | Number of entries
 28     | 8    |                    | Cache size



#### Index table


The index table is an array of entries. An entry (struct indexEntry) is 24 bytes in size and consists of:

 offset | size | value              | description
 ------ | ---- | ------------------ | -----------
 0      | 8    |                    | Hash
 8      | 8    |                    | Last used
 16     | 8    |                    | Size





### Data stream file format (#####_0)


Overview:
- File header
- URL
- Data stream (stream 1)
- Stream EOF
- HTTP headers (stream 0)
- (optionally) the SHA256 of the URL
- Stream EOF



#### File header


The index file header (struct entryHeader) is 20 bytes in size and consists of:

 offset | size | value              | description
 ------ | ---- | ------------------ | -----------
 0      | 8    | 0xfcfb6d1ba7725c30 | Magic
 8      | 4    | 5                  | Version
 12     | 4    |                    | URL length
 16     | 4    |                    | URL MD5



#### Stream EOF

The separator (struct entryEOF) contains information about the stream it succeeds. It is 20 bytes in size and consists of:

 offset | size | value              | description
 ------ | ---- | ------------------ | -----------
 0      | 8    | 0xf4fa6f45970d41d8 | Magic
 8      | 4    |                    | Flag
 12     | 4    |                    | stream CRC32
 16     | 4    |                    | stream size



##### Flag


 value | description
 ----- | ----------------------
 0     |
 1     | the stream has CRC32
 2     | the URL has SHA256
 3     | 1 + 2




### Data stream file format (#####_1)

Overview:
- File header
- URL
- Data stream (stream 2)
- Stream EOF




### Data stream file format (#####_s)

Overview:
- File header
- many of the following:
	- Range header
	- Range data stream



#### File header


The index file header (struct entryHeader) is 20 bytes in size and consists of:

 offset | size | value              | description
 ------ | ---- | ------------------ | -----------
 0      | 8    | 0xfcfb6d1ba7725c30 | Magic
 8      | 4    | 7                  | Version
 12     | 4    |                    | URL length
 16     | 4    |                    | URL MD5



#### Range header

The range header (struct sparseRangeHeader) contains information about the range stream it points to. It is 28 bytes in size and consists of:

 offset | size | value              | description
 ------ | ---- | ------------------ | -----------
 0      | 8    | 0xeb97bf016553676b | Magic
 8      | 8    |                    | stream offset
 16     | 8    |                    | stream size
 24     | 4    |                    | stream CRC32

