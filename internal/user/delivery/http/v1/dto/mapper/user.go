package mapper

import (
	"strconv"

	"github.com/hexley21/handy/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/handy/internal/user/entity"
	"github.com/hexley21/handy/internal/user/enum"
	"github.com/jackc/pgx/v5/pgtype"
)

func MapUserToEntity(dto *dto.User) (*entity.User, error) {
	id, err := strconv.ParseInt(dto.ID, 10, 64)
	if err != nil {
		return nil, err
	}

	userStatus := pgtype.Bool{}
	createdAt := pgtype.Timestamp{}

	if err := userStatus.Scan(dto.UserStatus); err != nil {
		return nil, err
	}

	if err := createdAt.Scan(dto.CreatedAt); err != nil {
		return nil, err
	}

	return &entity.User{
		ID:          id,
		FirstName:   dto.FirstName,
		LastName:    dto.LastName,
		PhoneNumber: dto.PhoneNumber,
		Email:       dto.Email,
		Hash:        "",
		Role:        enum.UserRole(dto.Role),
		UserStatus:  userStatus,
		CreatedAt:   createdAt,
	}, nil

}

func MapUserToDto(entity *entity.User) (*dto.User, error) {
	return &dto.User{
		ID:          strconv.FormatInt(entity.ID, 10),
		FirstName:   entity.FirstName,
		LastName:    entity.LastName,
		PhoneNumber: entity.PhoneNumber,
		Email:       entity.Email,
		Role:        string(entity.Role),
		UserStatus:  entity.UserStatus.Bool,
		CreatedAt:   entity.CreatedAt.Time,
	}, nil

}
