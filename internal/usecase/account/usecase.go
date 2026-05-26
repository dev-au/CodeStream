package account

import (
	"context"

	"github.com/dev-au/CodeStream/internal/domain/account"
)

type UseCase struct {
	repo account.Repository
}

// func CreateUser()

func (uc *UseCase) GetUser(ctx context.Context, userID int) (*account.User, error) {
	return uc.repo.GetUser(userID)
}
