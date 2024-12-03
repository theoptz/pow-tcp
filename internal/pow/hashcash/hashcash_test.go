package hashcash

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theoptz/pow-tcp/internal/pow"
)

func TestHashCash_GenerateChallenge(t *testing.T) {
	type args struct {
		difficulty uint8
	}
	tests := []struct {
		name    string
		args    args
		want    pow.Challenge
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "generate challenge",
			args: args{
				difficulty: 10,
			},
			want: pow.Challenge{
				0x0a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New()
			got, err := h.GenerateChallenge(tt.args.difficulty)
			if !tt.wantErr(t, err, fmt.Sprintf("GenerateChallenge(%v)", tt.args.difficulty)) {
				return
			}
			// Checking the first byte, which represents the difficulty level
			assert.Equalf(t, tt.want[0], got[0], "GenerateChallenge(%v)", tt.args.difficulty)
		})
	}
}

func TestHashCash_VerifySolution(t *testing.T) {
	var challenge = pow.Challenge{
		10, 0, 0, 0, 0, 0, 0, 0,
	}

	type args struct {
		challenge pow.Challenge
		solution  []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "invalid length",
			args: args{
				challenge: challenge,
				solution:  make([]byte, 9),
			},
			want: false,
		},
		{
			name: "return on checking fullBytes",
			args: args{
				challenge: challenge,
				solution:  make([]byte, 8),
			},
			want: false,
		},
		{
			name: "return on checking leadingBits",
			args: args{
				challenge: challenge,
				solution: []byte{
					61, 89, 78, 214, 135, 212, 198, 15,
				},
			},
			want: false,
		},
		{
			name: "valid solution",
			args: args{
				challenge: challenge,
				solution: []byte{
					127, 38, 252, 231, 62, 204, 14, 134,
				},
			},
			want: true,
		},
		{
			name: "valid when difficulty is 0",
			args: args{
				challenge: pow.Challenge(make([]byte, 9)),
				solution:  make([]byte, 8),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HashCash{}
			assert.Equalf(t, tt.want, h.VerifySolution(tt.args.challenge, tt.args.solution), "VerifySolution(%v, %v)", tt.args.challenge, tt.args.solution)
		})
	}
}

func TestHashCash_Solve(t *testing.T) {
	type args struct {
		ctx       context.Context
		challenge pow.Challenge
	}
	tests := []struct {
		name    string
		args    args
		timeout time.Duration
		wantErr error
	}{
		{
			name: "canceled context",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				challenge: pow.Challenge{},
			},
			wantErr: context.Canceled,
		},
		{
			name: "solved",
			args: args{
				ctx: context.Background(),
				challenge: pow.Challenge{
					//9, 0, 1, 2, 3, 4, 5, 6, 7,
					10, 0, 0, 0, 0, 0, 0, 0, 0,
				},
			},
			timeout: 60 * time.Second,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New()

			ctx := tt.args.ctx
			if tt.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.timeout)
				defer cancel()
			}

			got, err := h.Solve(ctx, tt.args.challenge)
			require.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				assert.Len(t, got, 8, "solution should be 8 bytes length")
			}
		})
	}
}
