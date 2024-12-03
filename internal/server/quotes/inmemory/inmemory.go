package inmemory

import (
	"context"
	_ "embed"
	"math/rand/v2"
	"strings"
)

//go:embed quotes.txt
var rawQuotes string

type Quotes struct {
	quotes [][]byte
}

func (q *Quotes) Get(_ context.Context) ([]byte, error) {
	return q.quotes[rand.IntN(len(q.quotes))], nil
}

func New() *Quotes {
	raw := strings.Split(rawQuotes, "\n")
	quotes := make([][]byte, len(raw))
	for i := range raw {
		quotes[i] = []byte(raw[i])
	}

	return &Quotes{
		quotes: quotes,
	}
}
