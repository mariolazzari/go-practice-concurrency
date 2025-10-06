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
