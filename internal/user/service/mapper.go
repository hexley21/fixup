package service

import (
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/user/domain"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

func MapUserModelToEntity(user repository.User) (*domain.User, error) {
	accountInfo, err := MapUserAccountInfo(user.Role, user.Verified)
	if err != nil {
		return nil, err
	}

	return domain.NewUser(
		user.ID,
		user.Picture.String,
		domain.NewUserPersonalInfo(
			user.Email,
			user.PhoneNumber,
			user.FirstName,
			user.LastName,
		),
		accountInfo,
		user.CreatedAt.Time,
	), nil
}

func MapUserAccountInfo(r string, verifier pgtype.Bool) (domain.UserAccountInfo, error) {
	role, err := enum.ParseRole(r)
	if err != nil {
		return domain.UserAccountInfo{}, err
	}

	return domain.NewUserAccountInfo(
		role,
		verifier.Bool,
	), nil
}

func MapUserIdentity(id int64, r string, verifier pgtype.Bool) (domain.UserIdentity, error) {
	accountInfo, err := MapUserAccountInfo(r, verifier)
	if err != nil {
		return domain.UserIdentity{}, err
	}

	return domain.NewUserIdentity(id, accountInfo), nil
}