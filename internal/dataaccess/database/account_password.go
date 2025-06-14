package database

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/hoangdv99/morgana/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	TabNameAccountPasswords = goqu.T("account_passwords")
)

const (
	ColNameAccountPasswordsAccountID = "account_id"
	ColNameAccountPasswordsHash      = "hash"
)

type AccountPassword struct {
	AccountID uint64 `db:"account_id" goqu:"skipupdate"`
	Hash      string `db:"hash"`
}

type AccountPasswordDataAccessor interface {
	CreateAccountPassword(ctx context.Context, accountPassword AccountPassword) error
	GetAccountPassword(ctx context.Context, accountID uint64) (AccountPassword, error)
	UpdateAccountPassword(ctx context.Context, accountPassword AccountPassword) error
	WithDatabase(database Database) AccountPasswordDataAccessor
}

type accountPasswordDataAccessor struct {
	database Database
	logger   *zap.Logger
}

func NewAccountPasswordDataAccessor(database *goqu.Database, logger *zap.Logger) AccountPasswordDataAccessor {
	return &accountPasswordDataAccessor{
		database: database,
		logger:   logger,
	}
}

func (a accountPasswordDataAccessor) CreateAccountPassword(ctx context.Context, accountPassword AccountPassword) error {
	logger := utils.LoggerWithContext(ctx, a.logger)
	_, err := a.database.
		Insert(TabNameAccountPasswords).
		Rows(goqu.Record{
			ColNameAccountPasswordsAccountID: accountPassword.AccountID,
			ColNameAccountPasswordsHash:      accountPassword.Hash,
		}).
		Executor().
		ExecContext(ctx)

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create account password")
		return status.Error(codes.Internal, "failed to create account password")
	}

	return nil
}

func (a accountPasswordDataAccessor) GetAccountPassword(ctx context.Context, accountID uint64) (AccountPassword, error) {
	logger := utils.LoggerWithContext(ctx, a.logger)
	accountPassword := AccountPassword{}
	found, err := a.database.
		From(TabNameAccountPasswords).
		Where(goqu.Ex{ColNameAccountPasswordsAccountID: accountID}).
		ScanStructContext(ctx, &accountPassword)

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get account password by account id")
		return AccountPassword{}, status.Error(codes.Internal, "failed to get account password by account id")
	}

	if !found {
		logger.Warn("account password not found")
		return AccountPassword{}, sql.ErrNoRows
	}

	return accountPassword, nil
}

func (a accountPasswordDataAccessor) UpdateAccountPassword(ctx context.Context, accountPassword AccountPassword) error {
	logger := utils.LoggerWithContext(ctx, a.logger)
	_, err := a.database.
		Update(TabNameAccountPasswords).
		Set(goqu.Record{
			ColNameAccountPasswordsHash: accountPassword.Hash,
		}).
		Where(goqu.Ex{ColNameAccountPasswordsAccountID: accountPassword.AccountID}).
		Executor().
		ExecContext(ctx)

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to update account password")
		return status.Error(codes.Internal, "failed to update account password")
	}

	return nil
}

func (a accountPasswordDataAccessor) WithDatabase(database Database) AccountPasswordDataAccessor {
	return &accountPasswordDataAccessor{
		database: database,
		logger:   a.logger,
	}
}
