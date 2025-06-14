package logic

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/hoangdv99/morgana/internal/configs"
	"github.com/hoangdv99/morgana/internal/dataaccess/cache"
	"github.com/hoangdv99/morgana/internal/dataaccess/database"
	"github.com/hoangdv99/morgana/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	rs512KeyPairBitCount = 2048
)

var (
	errUnexpectedSigningMethod = status.Error(codes.Unauthenticated, "unexpected signing method")
	errCannotGetTokensClaims   = status.Error(codes.Unauthenticated, "cannot get token's claims")
	errCannotGetTokensKidClaim = status.Error(codes.Unauthenticated, "cannot get token's kid claim")
	errCannotGetTokensSubClaim = status.Error(codes.Unauthenticated, "cannot get token's sub claim")
	errCannotGetTokensExpClaim = status.Error(codes.Unauthenticated, "cannot get token's exp claim")
	errTokenPublicKeyNotFound  = status.Error(codes.Unauthenticated, "token public key not found")
	errInvalidToken            = status.Error(codes.Unauthenticated, "invalid token")
	errFailedToSignToken       = status.Error(codes.Internal, "failed to sign token")
)

type Token interface {
	GetToken(ctx context.Context, accountID uint64) (string, time.Time, error)
	GetAccountIDAndExpireTime(ctx context.Context, token string) (uint64, time.Time, error)
	WithDatabase(database database.Database) Token
}

type token struct {
	accountDataAccessor database.AccountDataAccessor
	tokenPublicKeyCache cache.TokenPublicKey
	tokenPublicKey      database.TokenPublicKeyDataAccessor
	expiresIn           time.Duration
	privateKey          *rsa.PrivateKey
	tokenPublicKeyID    uint64
	authConfig          configs.Auth
	logger              *zap.Logger
}

func generateRSAKeyPair(bits int) (*rsa.PrivateKey, error) {
	privateKeyPair, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	return privateKeyPair, nil
}

func pemEncodePrivateKey(pubKey *rsa.PublicKey) ([]byte, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return nil, err
	}

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}

	return pem.EncodeToMemory(block), nil
}

func NewToken(
	accountDataAccessor database.AccountDataAccessor,
	tokenPublicKeyCache cache.TokenPublicKey,
	tokenPublicKeyDataAccessor database.TokenPublicKeyDataAccessor,
	authConfig configs.Auth,
	logger *zap.Logger,
) (Token, error) {
	expiresIn, err := authConfig.Token.GetExpiresInDuration()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get token expires in duration")
		return nil, err
	}

	rsaKeyPair, err := generateRSAKeyPair(rs512KeyPairBitCount)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to generate RSA key pair")
		return nil, err
	}

	publicKeyBytes, err := pemEncodePrivateKey(&rsaKeyPair.PublicKey)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to encode public key to PEM format")
		return nil, err
	}

	tokenPublicKeyID, err := tokenPublicKeyDataAccessor.CreatePublicKey(
		context.Background(),
		database.TokenPublicKey{
			PublicKey: string(publicKeyBytes),
		},
	)

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create token public key")
		return nil, err
	}

	return &token{
		accountDataAccessor: accountDataAccessor,
		tokenPublicKeyCache: tokenPublicKeyCache,
		tokenPublicKey:      tokenPublicKeyDataAccessor,
		expiresIn:           expiresIn,
		privateKey:          rsaKeyPair,
		tokenPublicKeyID:    tokenPublicKeyID,
		authConfig:          authConfig,
		logger:              logger,
	}, nil
}

func (t token) getJWTPublicKey(ctx context.Context, id uint64) (*rsa.PublicKey, error) {
	logger := utils.LoggerWithContext(ctx, t.logger).With(zap.Uint64("id", id))

	cachedPublicKeyBytes, err := t.tokenPublicKeyCache.Get(ctx, id)
	if err == nil {
		return jwt.ParseRSAPublicKeyFromPEM([]byte(cachedPublicKeyBytes))
	}

	logger.With(zap.Error(err)).Warn("failed to get cached public key bytes, will fail back to database")

	tokenPublicKey, err := t.tokenPublicKey.GetPublicKey(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errTokenPublicKeyNotFound
		}

		logger.With(zap.Error(err)).Error("cannot get token's public key from database")
	}

	err = t.tokenPublicKeyCache.Set(ctx, id, tokenPublicKey.PublicKey)
	if err != nil {
		logger.With(zap.Error(err)).Warn("failed to set public key bytes into cache")
	}

	return jwt.ParseRSAPublicKeyFromPEM([]byte(tokenPublicKey.PublicKey))
}

func (t token) GetAccountIDAndExpireTime(ctx context.Context, token string) (uint64, time.Time, error) {
	logger := utils.LoggerWithContext(ctx, t.logger)

	parsedToken, err := jwt.Parse(token, func(parsedToken *jwt.Token) (interface{}, error) {
		if _, ok := parsedToken.Method.(*jwt.SigningMethodRSA); !ok {
			logger.Error("unexpected signing method", zap.String("method", parsedToken.Header["alg"].(string)))
			return nil, errUnexpectedSigningMethod
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			logger.Error("failed to parse token claims")
			return nil, errCannotGetTokensClaims
		}

		tokenPublicKeyID, ok := claims["kid"].(float64)
		if !ok {
			logger.Error("failed to get token public key id from claims")
			return nil, errCannotGetTokensKidClaim
		}

		return t.getJWTPublicKey(ctx, uint64(tokenPublicKeyID))
	})

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to parse token")
		return 0, time.Time{}, err
	}

	if !parsedToken.Valid {
		logger.Error("invalid token")
		return 0, time.Time{}, errInvalidToken
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		logger.Error("cannot get token's claims")
		return 0, time.Time{}, errCannotGetTokensClaims
	}

	accountID, ok := claims["sub"].(float64)
	if !ok {
		logger.Error("cannot get token's sub claim")
		return 0, time.Time{}, errCannotGetTokensSubClaim
	}

	expireTimeUnix, ok := claims["exp"].(float64)
	if !ok {
		logger.Error("cannot get token's exp claim")
		return 0, time.Time{}, errCannotGetTokensExpClaim
	}

	return uint64(accountID), time.Unix(int64(expireTimeUnix), 0), nil
}

func (t token) GetToken(ctx context.Context, accountID uint64) (string, time.Time, error) {
	logger := utils.LoggerWithContext(ctx, t.logger)

	expireTime := time.Now().Add(t.expiresIn)
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"sub": accountID,
		"exp": expireTime.Unix(),
		"kid": t.tokenPublicKeyID,
	})

	tokenString, err := token.SignedString(t.privateKey)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to sign token")
		return "", time.Time{}, errFailedToSignToken
	}

	return tokenString, expireTime, nil
}

func (t token) WithDatabase(database database.Database) Token {
	t.accountDataAccessor = t.accountDataAccessor.WithDatabase(database)
	return t
}
