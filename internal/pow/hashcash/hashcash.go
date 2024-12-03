package hashcash

import (
	"context"
	"crypto/rand"
	"crypto/sha256"

	"github.com/theoptz/pow-tcp/internal/pow"
)

var (
	_ pow.ProofOfWork = (*HashCash)(nil)
)

const (
	byteSize = 8
)

type HashCash struct{}

func New() *HashCash {
	return &HashCash{}
}

func (h *HashCash) GenerateChallenge(difficulty uint8) (pow.Challenge, error) {
	var buf pow.Challenge

	buf[0] = difficulty
	if _, err := rand.Read(buf[1:]); err != nil {
		return buf, err
	}

	return buf, nil
}

func (h *HashCash) VerifySolution(challenge pow.Challenge, solution []byte) bool {
	if len(solution) != 8 {
		return false
	}

	data := append(challenge[:], solution...)
	hash := sha256.Sum256(data)

	fullBytes := challenge[0] / byteSize
	leadingBits := challenge[0] % byteSize

	for i := uint8(0); i < fullBytes; i++ {
		if hash[i] != 0 {
			return false
		}
	}

	if leadingBits > 0 {
		mask := byte(0xff << (byteSize - leadingBits))
		if solution[fullBytes]&mask != 0 {
			return false
		}
	}

	return true
}

func (h *HashCash) Solve(ctx context.Context, challenge pow.Challenge) ([]byte, error) {
	solution := make([]byte, 8)

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		_, err := rand.Read(solution)
		if err != nil {
			return nil, err
		}

		if h.VerifySolution(challenge, solution) {
			return solution, nil
		}
	}
}
