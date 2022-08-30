# `filestore`

The `filestore` package provides a facade that encapsulates
the common operations of a readable/writable file system. This could
be the actual underlying disk, an in-memory store, S3, whatever.

Currently, this package only ships with an implementation that
uses the underlying disk file system. Over time, I might offer
more options through plugins, but disk is all I needed when I wrote
it, so that's what's available :)

### WARNING

This is absolutely a work in progress, and I am VERY likely
to completely change the interfaces as I use it in "real" projects
and realize my abstractions could use some work. Should you
stumble upon this and want to use it, go for it. I would, however,
suggest that you version lock to a specific version as I'm quite
likely to check in breaking changes to meet my own needs until
I feel good about the API.

## Getting Started

```bash
go get -u github.com/monadicstack/filestore
```

## Basic Usage

You can create a `filestore.FS` that utilizes your local
disk, and all of the operations are implicitly based in that location.
From there, you have most of the standard file system operations
you would have on the Unix command line:

```go
fs := filestore.Disk("data")
```

### List Files in a Directory

Your `fs` is already tied to the data/ directory, so if
you wanted to list its contents, you would just do the following:
```go
files, err := fs.List(".")
if err != nil {
    // handle error
}
for _, file := range files {
    fmt.Printf("%s [Dir=%v]\n", file.Name(), file.IsDir())
}
```

Additionally, you could list the contents of any child directory
within data/ by specifying the path. For instance, to list the
contents of data/images/logos/ it would look like this:

```go
files, err := fs.List("images/logos")
```

Lastly, you can filter results down to just those that meet
specific criteria. The `filestore` package ships with a few
common filters for things like file name pattern or extension, but
you can provide any function that matches the filter signature:

```go
pngFiles, err := fs.List("images/logos", filestore.WithExt("png"))
```

### Reading/Writing Files

The `filestore` package makes it easy to read/write files in
the underlying file system. Let's say you wanted to read the
file data/images/logos/splash-256.png, here's what that would
look like. Notice, that the `filestore.ReaderFile` returned by `Read()`
implements `io.Reader`, so you can use it with any of Go's standard
stream processing operations:

```go
file, err := fs.Read("images/logos/splash-256.png")
if err != nil {
	// handle error
}
defer file.Close()

img, err := png.Decode(file)
```

Writing files is just as idiomatic. The `Write()` operation returns
a `filestore.WriterFile` which implements `io.Writer`, so you can
hook into your favorite stream processing code.

```go
// If the data/conf/ directory doesn't exist, Write() will
// lazily create it on the fly for you!!!
file, err := fs.Write("conf/config.json")
if err != nil {
    // handle error
}
defer file.Close()

file.Write([]byte(`{"timeout":"10s"}`))
```

### Other Handy Operations

The idea this package is to give you most of the same tools/behaviors
that you'd have on the command line. 

Delete a single file or even a whole directory and everything in it...
```go
// Single file
err := fs.Remove("conf/config.json")
// Whole directory
err := fs.Remove("tmp/uploads")
```

Move a file or directory 

```go
// Single file
err := fs.Move("uploads/tmp/upload.png", "images/logos/splash-256.png")
// Whole directory
err := fs.Move("uploads/processing", "uploads/done")
```

Get a `filestore.FS` that is scoped to a subdirectory...

```go
logos := fs.ChangeDirectory("images/logos")
logoFiles, err := logos.List(".", filestore.WithExts("png", "jpg"))

// Prints "data/images/logos"
fmt.Println(logos.WorkingDirectory())
```
Check if a file exists or not...

```go
if fs.Exists("conf/config.json") {
    // Do something interesting
}
```

Get name/size/type/etc. details about a file w/o reading it.

```go
info, err := fs.Stat("conf/config.json")
fmt.Printf("%s [Size=%d][Dir=%v]\n", info.Name(), info.Size(), info.IsDir())
```
