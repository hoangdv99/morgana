package logic

import (
	"context"
	"errors"

	"github.com/hoangdv99/morgana/internal/configs"
	"golang.org/x/crypto/bcrypt"
)

type Hash interface {
	Hash(ctx context.Context, data string) (string, error)
	IsHashEqual(ctx context.Context, data string, hash string) (bool, error)
}

type hash struct {
	accountConfig configs.Account
}

func NewHash(accountConfig configs.Account) Hash {
	return &hash{
		accountConfig: accountConfig,
	}
}

func (h hash) Hash(ctx context.Context, data string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(data), h.accountConfig.HashCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (h hash) IsHashEqual(ctx context.Context, data string, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(data))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
