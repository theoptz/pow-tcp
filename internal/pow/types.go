package pow

import "context"

type Challenge [9]byte

type ProofOfWork interface {
	GenerateChallenge(uint8) (Challenge, error)
	VerifySolution(Challenge, []byte) bool
}

type Solver interface {
	Solve(context.Context, Challenge) ([]byte, error)
}
