package main

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func worker(ctx context.Context, jobs <-chan [2]string, results chan<- error) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			err := Center(job[0], job[1])
			results <- err
		}
	}
}

func producer(ctx context.Context, jobs chan<- [2]string, srcDir, destDir string) error {
	defer close(jobs)

	matches, err := filepath.Glob(fmt.Sprintf("%s/*.jpg", srcDir))
	if err != nil {
		return err
	}

	for _, src := range matches {
		dest := fmt.Sprintf("%s/%s", destDir, filepath.Base(src))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case jobs <- [2]string{src, dest}:
		}
	}

	return nil
}

// Center creates destFile which is the center of image encode in data.
func Center(srcFile, destFile string) error {
	file, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer file.Close()

	src, err := jpeg.Decode(file)
	if err != nil {
		return err
	}

	x, y := src.Bounds().Max.X, src.Bounds().Max.Y
	r := image.Rect(0, 0, x/2, y/2)
	dest := image.NewRGBA(r)
	draw.Draw(dest, dest.Bounds(), src, image.Point{x / 4, y / 4}, draw.Over)

	out, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer out.Close()

	return jpeg.Encode(out, dest, nil)
}

// CenterDir calls Center on every image in srcDir. n is the maximal number of goroutines.
func CenterDir(ctx context.Context, srcDir, destDir string, n int) error {
	if err := os.Mkdir(destDir, 0750); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}

	matches, err := filepath.Glob(fmt.Sprintf("%s/*.jpg", srcDir))
	if err != nil {
		return err
	}

	for _, src := range matches {
		dest := fmt.Sprintf("%s/%s", destDir, filepath.Base(src))
		if err := Center(src, dest); err != nil {
			return err
		}

	}

	return nil
}

func main() {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	n := runtime.GOMAXPROCS(0) // number of cores

	// Define source and destination directories
	srcDir := "input"
	destDir := "output"

	err := CenterDir(ctx, srcDir, destDir, n)

	duration := time.Since(start)
	log.Printf("info: finished in %v (err=%v)", duration, err)
}
