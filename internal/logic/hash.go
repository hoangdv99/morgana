package logic

import (
	"context"
	"errors"

	"github.com/hoangdv99/morgana/internal/configs"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Hash interface {
	Hash(ctx context.Context, data string) (string, error)
	IsHashEqual(ctx context.Context, data string, hash string) (bool, error)
}

type hash struct {
	authConfig configs.Auth
}

func NewHash(authConfig configs.Auth) Hash {
	return &hash{
		authConfig: authConfig,
	}
}

func (h hash) Hash(_ context.Context, data string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(data), h.authConfig.Hash.Cost)
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to hash data: %+v", err)
	}
	return string(hashed), nil
}

func (h hash) IsHashEqual(_ context.Context, data string, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(data))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, status.Errorf(codes.Internal, "failed to check if data equal hash: %+v", err)
	}
	return true, nil
}
