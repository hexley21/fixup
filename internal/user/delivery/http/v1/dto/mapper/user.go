package mapper

import (
	"strconv"

	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/entity"
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/pkg/infra/cdn"
)

func MapUserToDto(entity entity.User, urlSigner cdn.URLSigner) (dto.User, error) {
	var url string
	if entity.PictureName.String != "" {
		signedUrl, err := urlSigner.SignURL(entity.PictureName.String)
		if err != nil {
			return dto.User{}, err
		}
		url = signedUrl
	}

	return dto.User{
		ID:          strconv.FormatInt(entity.ID, 10),
		FirstName:   entity.FirstName,
		LastName:    entity.LastName,
		PhoneNumber: entity.PhoneNumber,
		Email:       entity.Email,
		PictureUrl:  url,
		Role:        string(entity.Role),
		UserStatus:  entity.UserStatus.Bool,
		CreatedAt:   entity.CreatedAt.Time,
	}, nil
}

func MapUserToProfileDto(entity entity.User, urlSigner cdn.URLSigner) (dto.Profile, error) {
	var url string
	if entity.PictureName.String != "" {
		signedUrl, err := urlSigner.SignURL(entity.PictureName.String)
		if err != nil {
			return dto.Profile{}, err
		}
		url = signedUrl
	}

	profile := dto.Profile{
		ID:         strconv.FormatInt(entity.ID, 10),
		FirstName:  entity.FirstName,
		LastName:   entity.LastName,
		PictureUrl: url,
		Role:       string(entity.Role),
		UserStatus: entity.UserStatus.Bool,
		CreatedAt:  entity.CreatedAt.Time,
	}

	if entity.Role == enum.UserRolePROVIDER {
		profile.PhoneNumber = entity.PhoneNumber
		profile.Email = entity.Email
	}

	return profile, nil
}
