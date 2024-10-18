package mapper

import (
	"errors"
	"strconv"

	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/domain"
	"github.com/hexley21/fixup/pkg/infra/cdn"
)

var (
	ErrNilEntity error = errors.New("entity was nil")
)

func MapUserEntityToDTO(entity *domain.User, urlSigner cdn.URLSigner) (*dto.User, error) {
	if entity == nil {
		return nil, ErrNilEntity
	}

	var url string
	if entity.Picture != "" {
		signedUrl, err := urlSigner.SignURL(entity.Picture)
		if err != nil {
			return nil, err
		}
		url = signedUrl
	}

	return &dto.User{
		ID:          strconv.FormatInt(entity.ID, 10),
		FirstName:   entity.PersonalInfo.FirstName,
		LastName:    entity.PersonalInfo.LastName,
		PhoneNumber: entity.PersonalInfo.PhoneNumber,
		Email:       entity.PersonalInfo.Email,
		PictureUrl:  url,
		Role:        entity.AccountInfo.Role,
		Active:      entity.AccountInfo.Active,
		CreatedAt:   entity.CreatedAt,
	}, nil
}
