package bench

import (
	"math"
	mrand "math/rand"
)

// benchVectorDim is the standard vector dimension used by vector DB drivers.
const benchVectorDim = 128

// deterministicVector generates a normalized vector from a key string.
// Used by pgvector, qdrant, and other vector DB drivers for consistent benchmarking.
func deterministicVector(key string, dim int) []float32 {
	var seed int64
	for i, c := range key {
		seed += int64(c) * int64(i+1)
	}
	rng := mrand.New(mrand.NewSource(seed))
	vec := make([]float32, dim)
	var norm float64
	for i := range vec {
		vec[i] = float32(rng.NormFloat64())
		norm += float64(vec[i] * vec[i])
	}
	norm = math.Sqrt(norm)
	for i := range vec {
		vec[i] /= float32(norm)
	}
	return vec
}
