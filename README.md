# go-pager

[![GoDoc](https://godoc.org/snai.pe/go-pager?status.svg)](https://godoc.org/snai.pe/go-pager)  

```
go get snai.pe/go-pager
```

A Go library to help programs call and write to the user's pager.

The library introduces a `Pager` type that abstracts a pager program as an
`io.WriteCloser`. It understands `$PAGER` in a way that is compliant with
POSIX's `man 1 man`.
