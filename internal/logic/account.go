package logic

import (
	"context"
	"errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/hoangdv99/morgana/internal/dataaccess/cache"
	"github.com/hoangdv99/morgana/internal/dataaccess/database"
	morgana "github.com/hoangdv99/morgana/internal/generated/morgana/v1"
	"github.com/hoangdv99/morgana/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateAccountParams struct {
	AccountName string
	Password    string
}

type CreateAccountOutput struct {
	ID          uint64
	AccountName string
}

type CreateSessionParams struct {
	AccountName string
	Password    string
}

type CreateSessionOutput struct {
	Account *morgana.Account
	Token   string
}

type Account interface {
	CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error)
	CreateSession(ctx context.Context, params CreateSessionParams) (CreateSessionOutput, error)
}

type account struct {
	goquDatabase                *goqu.Database
	takenAccountNameCache       cache.TakenAccountName
	accountDataAccessor         database.AccountDataAccessor
	accountPasswordDataAccessor database.AccountPasswordDataAccessor
	hashLogic                   Hash
	tokenLogic                  Token
	logger                      *zap.Logger
}

func NewAccount(
	goquDatabase *goqu.Database,
	takenAccountNameCache cache.TakenAccountName,
	accountDataAccessor database.AccountDataAccessor,
	accountPasswordDataAccessor database.AccountPasswordDataAccessor,
	hashLogic Hash,
	tokenLogic Token,
	logger *zap.Logger,
) Account {
	return &account{
		goquDatabase:                goquDatabase,
		takenAccountNameCache:       takenAccountNameCache,
		accountDataAccessor:         accountDataAccessor,
		accountPasswordDataAccessor: accountPasswordDataAccessor,
		hashLogic:                   hashLogic,
		tokenLogic:                  tokenLogic,
		logger:                      logger,
	}
}

func (a account) isAccountNameTaken(ctx context.Context, accountName string) (bool, error) {
	logger := utils.LoggerWithContext(ctx, a.logger).With(zap.String("account_name", accountName))

	accountNameTaken, err := a.takenAccountNameCache.Has(ctx, accountName)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to check if account name is taken in cache")
	} else if accountNameTaken {
		return true, nil
	}

	_, err = a.accountDataAccessor.GetAccountByAccountName(ctx, accountName)
	if err != nil {
		if errors.Is(err, database.ErrAccountNotFound) {
			return false, nil
		}
		return false, err
	}

	err = a.takenAccountNameCache.Add(ctx, accountName)
	if err != nil {
		logger.With(zap.Error(err)).Warn("failed to add account name to taken account names cache")
		return false, err
	}

	return true, nil
}

func (a account) CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error) {
	accountNameTaken, err := a.isAccountNameTaken(ctx, params.AccountName)
	if err != nil {
		return CreateAccountOutput{}, status.Error(codes.AlreadyExists, "account name already taken")
	}

	if accountNameTaken {
		return CreateAccountOutput{}, status.Error(codes.AlreadyExists, "account name is already taken")
	}

	var accountID uint64
	txErr := a.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		accountNameTaken, err := a.isAccountNameTaken(ctx, params.AccountName)
		if err != nil {
			return err
		}

		if accountNameTaken {
			return errors.New("account name already taken")
		}

		accountID, err = a.accountDataAccessor.WithDatabase(td).CreateAccount(ctx, database.Account{
			AccountName: params.AccountName,
		})

		if err != nil {
			return err
		}

		hashedPassword, err := a.hashLogic.Hash(ctx, params.Password)
		if err != nil {
			return err
		}

		err = a.accountPasswordDataAccessor.WithDatabase(td).CreateAccountPassword(ctx, database.AccountPassword{
			AccountID: accountID,
			Hash:      hashedPassword,
		})
		if err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		return CreateAccountOutput{}, txErr
	}

	return CreateAccountOutput{
		ID:          accountID,
		AccountName: params.AccountName,
	}, nil
}

func (a account) databaseAccountToProtoAccount(account database.Account) *morgana.Account {
	return &morgana.Account{
		Id:          account.ID,
		AccountName: account.AccountName,
	}
}

func (a account) CreateSession(ctx context.Context, params CreateSessionParams) (CreateSessionOutput, error) {
	existingAccount, err := a.accountDataAccessor.GetAccountByAccountName(ctx, params.AccountName)
	if err != nil {
		return CreateSessionOutput{}, err
	}

	existingAccountPassword, err := a.accountPasswordDataAccessor.GetAccountPassword(ctx, existingAccount.ID)
	if err != nil {
		return CreateSessionOutput{}, err
	}

	isHashEqual, err := a.hashLogic.IsHashEqual(ctx, params.Password, existingAccountPassword.Hash)
	if err != nil {
		return CreateSessionOutput{}, err
	}

	if !isHashEqual {
		return CreateSessionOutput{}, status.Error(codes.Unauthenticated, "incorrect password")
	}

	token, _, err := a.tokenLogic.GetToken(ctx, existingAccount.ID)
	if err != nil {
		return CreateSessionOutput{}, err
	}

	return CreateSessionOutput{
		Account: a.databaseAccountToProtoAccount(existingAccount),
		Token:   token,
	}, nil
}
