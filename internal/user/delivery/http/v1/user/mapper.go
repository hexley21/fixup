package user

import (
	"errors"
	"strconv"

	"github.com/hexley21/fixup/internal/user/domain"
	"github.com/hexley21/fixup/pkg/infra/cdn"
)

var ErrNilEntity = errors.New("entity was nil")

func MapUserToDTO(entity *domain.User, urlSigner cdn.URLSigner) (*User, error) {
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

	return &User{
		ID:           strconv.FormatInt(entity.ID, 10),
		UserPersonalInfo: MapPersonalInfoToDTO(entity.PersonalInfo),
		PictureUrl:   url,
		Role:         string(entity.AccountInfo.Role),
		Verified:     entity.AccountInfo.Verified,
		CreatedAt:    entity.CreatedAt,
	}, nil
}

func MapPersonalInfoToDTO(vo *domain.UserPersonalInfo) *UserPersonalInfo {
	return &UserPersonalInfo{
		FirstName:   vo.FirstName,
		LastName:    vo.LastName,
		Email:       vo.Email,
		PhoneNumber: vo.PhoneNumber,
	}
}
