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
