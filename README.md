# nanoid-go
A simple, efficient and secure Nano ID generator implemented in Go.

## Features
- Easy to learn and use, efficient and secure.
- Generate custom-sized Nano ID.
- Generate Nano ID using custom alphabet.
- Generate Nano ID using custom random generator.

## Install
In general, this package is compatible with most modern versions of Go.
```bash
go get -u github.com/nobody-night/nanoid-go
```

## Getting Started
In general, we need to generate Nano ID strings with a default size. This is easy to do:
```go
id, err := nanoid.New()
```
This returns a Nano ID string of default size `21` and any errors encountered.

If you need to generate a Nano ID string of the specified size, you can use the `nanoid.NewWithSize` function:
```go
id, err := nanoid.NewWithSize(39)
```
This returns a Nano ID string of size `39` and any errors encountered.

### Using Reader
We can use reader to use more advanced features. First, we need to create a new reader:
```go
reader, err := nanoid.NewReader()
```

To be able to use memory efficiently, we need to allocate a buffer slice of the appropriate size:
```go
buf := make([]byte, 21)
```

Immediately after that, we need to read data from the reader into the buffer slice:
```go
nr, err := reader.Read(buf)
```
The `Read` function returns the actual number of bytes read and any errors encountered. In general, the actual number of bytes read is the same as the size of the buffer slice.

The actual data read is the generated Nano ID string:
```go
id := string(buf)
```
It is worth noting that you should not let the variable `id` leave the current context unless necessary. Otherwise, it may cause memory escapes and degrade application performance.

#### Custom Alphabet
The function `nanoid.NewReader` accepts one or more options. For example, we can use a custom alphabet:
```go
reader, err := nanoid.NewReader(nanoid.WithAlphabet(...))
```
The function `nanoid.WithAlphabet` takes an alphabetic string and returns a reader option.

It is important to note that the maximum length of a custom alphabet string cannot exceed `256` characters. The maximum acceptable length is determined by the constant `nanoid.MaxAlphabetSize`.

#### Custom Random Reader
In addition, we can also use a custom random number generator:
```go
reader, err := nanoid.NewReader(nanoid.WithRandReader(...))
```
The function `nanoid.WithRandReader` takes an reader for random number generator and returns a reader option. The reader for this random number generator must be compatible with `io.Reader` interface.

<hr>

Released under the [MIT License](LICENSE).
