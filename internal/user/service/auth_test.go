package service_test

import (
	"time"

	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/enum"
)

var (
	userDto = dto.User{
		ID: "1",
		FirstName: "Larry",
		LastName: "Page",
		PhoneNumber: "995111222333",
		Email: "larry@page.com",
		PictureUrl: "larrypage.png",
		Role: string(enum.UserRoleADMIN),
		UserStatus: true,
		CreatedAt: time.Now(),
	}
)