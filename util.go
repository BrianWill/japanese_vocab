package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
	"unicode"
)

func PickRandomN[T any](items []T, n int) ([]*T, error) {
	if n < 0 || n > len(items) {
		return nil, fmt.Errorf("n=%d out of range", n)
	}

	const ratioThreshold = 10

	if (len(items) / n) > ratioThreshold {
		return SampleN(items, n)
	}

	// Copy so we don't modify the original
	tmp := make([]*T, len(items))
	for i := range items {
		tmp[i] = &items[i]
	}

	rand.Shuffle(len(tmp), func(i, j int) {
		tmp[i], tmp[j] = tmp[j], tmp[i]
	})

	return tmp[:n], nil
}

// warning: only call if n is greater than len(items) by an order of magnitude
// if n is too close to len(items), then very inefficient when have to guess what indexes are unused
func SampleN[T any](items []T, n int) ([]*T, error) {
	if n < 0 || n > len(items) {
		return nil, fmt.Errorf("n=%d out of range", n)
	}

	chosen := make(map[int]struct{}, n)
	result := make([]*T, 0, n)

	for len(result) < n {
		idx := rand.Intn(len(items))
		if _, ok := chosen[idx]; !ok {
			chosen[idx] = struct{}{}
			result = append(result, &items[idx])
		}
	}
	return result, nil
}

type Entry struct {
	Value     string
	Timestamp time.Time
}

func recencyWeight(ts time.Time) float64 {
	now := time.Now()
	ageHours := now.Sub(ts).Hours()

	lambda := 0.1                       // adjust; larger = more bias toward recent
	return math.Exp(-lambda * ageHours) // exponential decay
}

func WeightedRandom(entries []Entry) Entry {
	// compute weights
	weights := make([]float64, len(entries))
	var total float64

	for i, e := range entries {
		w := recencyWeight(e.Timestamp)
		weights[i] = w
		total += w
	}

	// random threshold
	r := rand.Float64() * total

	// walk to find selected item
	var cumulative float64
	for i, w := range weights {
		cumulative += w
		if r <= cumulative {
			return entries[i]
		}
	}

	// fallback (numerical issues)
	return entries[len(entries)-1]
}

func WeightedSample(entries []Entry, n int) []Entry {
	result := make([]Entry, 0, n)
	pool := append([]Entry(nil), entries...) // copy

	for len(result) < n && len(pool) > 0 {
		pick := WeightedRandom(pool)
		result = append(result, pick)

		// remove from pool
		for i := range pool {
			if pool[i].Value == pick.Value {
				pool = append(pool[:i], pool[i+1:]...)
				break
			}
		}
	}

	return result
}

func assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}

func ContainsJapanese(s string) bool {
	for _, r := range s {
		if unicode.In(r,
			unicode.Hiragana,
			unicode.Katakana,
			unicode.Han,
		) {
			return true
		}
	}
	return false
}
