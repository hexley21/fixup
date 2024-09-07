package mapper

import (
	"fmt"
	"strconv"

	"github.com/hexley21/handy/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/handy/internal/user/entity"
)

func MapUserToDto(entity *entity.User, cdnUrlFmt string) (*dto.User, error) {
	return &dto.User{
		ID:          strconv.FormatInt(entity.ID, 10),
		FirstName:   entity.FirstName,
		LastName:    entity.LastName,
		PhoneNumber: entity.PhoneNumber,
		Email:       entity.Email,
		PictureUrl:  fmt.Sprintf(cdnUrlFmt, entity.PictureName),
		Role:        string(entity.Role),
		UserStatus:  entity.UserStatus.Bool,
		CreatedAt:   entity.CreatedAt.Time,
	}, nil
}
