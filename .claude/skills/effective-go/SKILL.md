---
name: effective-go
description: Idiomatic Go patterns and conventions from official Effective Go guide. Use when writing Go code to ensure proper naming, error handling, concurrency patterns, interface design, and Go-specific idioms. Triggers on Go file creation/editing, code review, or when user asks for idiomatic Go. If you cannot find any info visit to https://go.dev/doc/effective_go
---

# Effective Go

Guidelines for writing clear, idiomatic Go code.

## Formatting

Use `gofmt`. No debates.

- Tabs for indentation
- No line length limit (wrap long lines, indent with extra tab)
- Fewer parentheses than C/Java

## Names

### Visibility

First character uppercase = exported. This makes naming semantically critical.

### Packages

- Lowercase, single-word, no underscores/mixedCaps
- Short names (everyone types them)
- Use package name to avoid stutter:

```go
bufio.Reader     // not bufio.BufReader
ring.New         // not ring.NewRing
```

### Getters/Setters

No "Get" prefix:

```go
owner := obj.Owner()      // not GetOwner()
obj.SetOwner(user)
```

### Interfaces

One-method interfaces: method name + "-er":

```go
Reader, Writer, Formatter, Stringer
```

### MixedCaps

Always `MixedCaps` or `mixedCaps`, never underscores.

## Control Structures

### If

Initialization statements reduce scope:

```go
if err := file.Chmod(0664); err != nil {
    return err
}
```

Omit else when body ends with return/break/continue:

```go
f, err := os.Open(name)
if err != nil {
    return err
}
codeUsing(f)
```

### For

Three forms:

```go
for init; condition; post { }  // C-style
for condition { }               // while
for { }                         // infinite
```

Range over collections:

```go
for key, value := range m { }
for key := range m { }          // key only
for _, value := range m { }     // value only
```

### Switch

No automatic fallthrough. Expressions need not be constants:

```go
switch {
case '0' <= c && c <= '9':
    return c - '0'
case 'a' <= c && c <= 'f':
    return c - 'a' + 10
}
```

Multiple cases:

```go
case ' ', '?', '&', '=':
    return true
```

### Type Switch

```go
switch t := value.(type) {
case string:
    return t
case Stringer:
    return t.String()
}
```

## Functions

### Multiple Returns

```go
func (f *File) Write(b []byte) (n int, err error)
```

### Named Results

```go
func ReadFull(r Reader, buf []byte) (n int, err error) {
    for len(buf) > 0 && err == nil {
        var nr int
        nr, err = r.Read(buf)
        n += nr
        buf = buf[nr:]
    }
    return
}
```

### Defer

Cleanup near allocation, guaranteed execution:

```go
func Contents(filename string) (string, error) {
    f, err := os.Open(filename)
    if err != nil {
        return "", err
    }
    defer f.Close()
    // ... read file
}
```

Arguments evaluated at defer time, executed LIFO.

## Data

### new vs make

- `new(T)` - allocates zeroed memory, returns `*T`
- `make(T, args)` - slices/maps/channels only, returns initialized `T`

```go
p := new([]int)       // *p == nil
v := make([]int, 100) // v is 100-element slice
```

### Composite Literals

```go
return &File{fd: fd, name: name}
```

### Zero Values

Design structs so zero value is useful:

```go
var buf bytes.Buffer  // ready to use
var mu sync.Mutex     // unlocked
```

### Slices over Arrays

Arrays are values (copy on assign). Slices are references. Prefer slices.

```go
func process(data []byte) { }  // not [100]byte
```

### Maps

```go
m := map[string]int{"one": 1}

// Comma-ok idiom
if val, ok := m[key]; ok {
    // key exists
}

delete(m, key)  // safe if absent
```

## Methods

### Pointer vs Value Receivers

- Pointer: can modify receiver, avoids copy
- Value: safe for concurrent use

```go
func (p *ByteSlice) Write(data []byte) (int, error)
```

Rule: value methods callable on pointers/values; pointer methods only on pointers.

## Interfaces

### Design

Small, focused interfaces:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

Define interfaces where used, not where implemented.

### Export interfaces, not types

```go
func NewReader() io.Reader  // not *myReader
```

### Type Assertions

```go
str, ok := value.(string)
if !ok {
    // not a string
}
```

## Embedding

Composition over inheritance:

```go
type ReadWriter struct {
    *Reader
    *Writer
}
```

Methods promoted automatically. Receiver is inner type.

## Concurrency

### Core Principle

> Do not communicate by sharing memory; share memory by communicating.

### Goroutines

```go
go func() {
    // concurrent work
}()
```

### Channels

```go
ch := make(chan int)      // unbuffered
ch := make(chan int, 100) // buffered
```

Synchronization:

```go
done := make(chan bool)
go func() {
    work()
    done <- true
}()
<-done
```

### Select

```go
select {
case msg := <-ch1:
    handle(msg)
case ch2 <- val:
    // sent
default:
    // non-blocking
}
```

### Semaphore Pattern

```go
var sem = make(chan struct{}, MaxWorkers)

func handle(r *Request) {
    sem <- struct{}{}
    defer func() { <-sem }()
    process(r)
}
```

## Errors

### Explicit Handling

Always check and handle:

```go
f, err := os.Open(name)
if err != nil {
    return fmt.Errorf("opening %s: %w", name, err)
}
```

### Custom Errors

```go
type PathError struct {
    Op   string
    Path string
    Err  error
}

func (e *PathError) Error() string {
    return e.Op + " " + e.Path + ": " + e.Err.Error()
}
```

### Panic/Recover

Panic for unrecoverable errors only:

```go
func safelyDo(work func()) {
    defer func() {
        if err := recover(); err != nil {
            log.Println("work failed:", err)
        }
    }()
    work()
}
```

## Blank Identifier

```go
_, err := os.Stat(path)           // discard value
_ = unusedVar                     // silence compiler (temp)
import _ "net/http/pprof"         // side-effect import
var _ json.Marshaler = (*T)(nil)  // compile-time interface check
```

## Initialization

### Constants with iota

```go
const (
    _ = iota
    KB = 1 << (10 * iota)
    MB
    GB
)
```

### init Functions

Run after variable declarations, before main:

```go
func init() {
    if user == "" {
        log.Fatal("$USER not set")
    }
}
```

## Printing

```go
fmt.Printf("%v", val)   // default format
fmt.Printf("%+v", val)  // struct with field names
fmt.Printf("%#v", val)  // Go syntax
fmt.Printf("%T", val)   // type
```

Custom String method:

```go
func (t T) String() string {
    return fmt.Sprintf("T{%d}", t.val)
}
```
