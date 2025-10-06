# Go practice concurrency

## Timing HTTP calls

```go
package main

import (
 "io"
 "log"
 "net/http"
 "sync"
 "time"
)

// MutliURLTimes calls URLTime for every URL in URLs.
func MultiURLTime(urls []string) {
 wg := sync.WaitGroup{}
 wg.Add(len(urls))

 for _, url := range urls {
  go func(url string) {
   defer wg.Done()
   URLTime(url)
  }(url)
 }
}

// URLTime checks how much time it takes url to respond.
func URLTime(url string) {
 start := time.Now()

 resp, err := http.Get(url)
 if err != nil {
  log.Printf("error: %q - %s", url, err)
  return
 }
 if resp.StatusCode != http.StatusOK {
  log.Printf("error: %q - bad status - %s", url, resp.Status)
  return
 }
 // Read body
 _, err = io.Copy(io.Discard, resp.Body)
 if err != nil {
  log.Printf("error: %q - %s", url, err)
  return
 }

 duration := time.Since(start)
 log.Printf("info: %q - %v", url, duration)
}

func main() {
 start := time.Now()

 urls := []string{
  "http://localhost:8080/200",
  "http://localhost:8080/100",
  "http://localhost:8080/50",
 }

 MultiURLTime(urls)

 duration := time.Since(start)
 log.Printf("%d URLs in %v", len(urls), duration)
}
```

## Getting results from goroutines

```go
package main

import (
 "bytes"
 "crypto/sha1"
 "fmt"
 "io"
 "log"
 "time"
)

// sha1sig return SHA1 signature in the format "35aabcd5a32e01d18a5ef688111624f3c547e13b"
func sha1Sig(data []byte) (string, error) {
 w := sha1.New()
 r := bytes.NewReader(data)
 if _, err := io.Copy(w, r); err != nil {
  return "", err
 }

 sig := fmt.Sprintf("%x", w.Sum(nil))
 return sig, nil
}

type File struct {
 Name      string
 Content   []byte
 Signature string
}

type Reply struct {
 filename string
 match    bool
 err      error
}

func signWorker(file File, ch chan<- Reply) {
 sig, err := sha1Sig(file.Content)
 r := Reply{filename: file.Name, match: sig == file.Signature, err: err}
 ch <- r
}

// ValidateSigs return slice of OK files and slice of mismatched files
func ValidateSigs(files []File) ([]string, []string, error) {
 var okFiles []string
 var badFiles []string
 ch := make(chan Reply)

 for _, file := range files {
  go signWorker(file, ch)
 }

 for range files {
  r := <-ch
  if !r.match || r.err != nil {
   badFiles = append(badFiles, r.filename)
  } else {
   okFiles = append(okFiles, r.filename)
  }
 }
 return okFiles, badFiles, nil
}

func main() {
 start := time.Now()

 files := []File{
  {"file1.txt", []byte("Hello, World!"), "65a8e27d8879283831b664bd8b7f0ad4e5d5a1bd"},
  {"file2.txt", []byte("Go is awesome!"), "3c01bdbb26f358bab27f267924aa2c9a03fcfdb8"},
  {"file3.txt", []byte("Concurrency in Go"), "d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d"},
  {"file4.txt", []byte("Goroutines are lightweight threads"), "4e07408562bedb8b60ce05c1decfe3ad16b722309"},
  {"file5.txt", []byte("Channels for communication"), "3a7bd3e2360a3d80c2a4f1b5f1e6e6e6e6e6e6e"},
 }

 ok, bad, err := ValidateSigs(files)
 if err != nil {
  log.Fatalf("error: %v", err)
 }

 duration := time.Since(start)
 log.Printf("info: %d files in %v\n", len(ok)+len(bad), duration)
 log.Printf("ok: %v", ok)
 log.Printf("bad: %v", bad)
}
```

## Timeout and cancellation

```go
package main

import (
 "context"
 "log"
 "time"
)

var (
 // Everybody loves "The Princess Bride"
 defaultMovie = Movie{
  ID:    "tt0093779",
  Title: "The Princess Bride",
 }
 // Time it takes for BestNextMovie to finish
 bmvTime = 50 * time.Millisecond
)

// Movie is a movie recommendation
type Movie struct {
 ID    string
 Title string
}

// BestNextMovie return the best move recommendation for a user
func BestNextMovie(user string) Movie {
 time.Sleep(bmvTime) // Simulate work

 // Don't change this, otherwise the test will fail
 return Movie{
  ID:    "tt0083658",
  Title: "Blade Runner",
 }
}

// NextMovie return BestNextMovie result if it finished before ctx expires, otherwise defaultMovie
func NextMovie(ctx context.Context, user string) Movie {
 ch := make(chan Movie, 1)

 go func() {
  ch <- BestNextMovie(user)
 }()

 select {
 case m := <-ch:
  return m
 case <-ctx.Done():
  log.Printf("warn: context expired: %v", ctx.Err())
  return defaultMovie
 }
}

func main() {
 log.Printf("info: checking timeout")
 ctx, cancel := context.WithTimeout(context.Background(), bmvTime/2)
 defer cancel()

 mTimeout := NextMovie(ctx, "ridley")
 log.Printf("info: got %+v", mTimeout)
}
```

## 