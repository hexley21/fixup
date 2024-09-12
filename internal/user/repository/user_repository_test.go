package repository_test

import (
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/hexley21/fixup/internal/user/repository"
)

var (
	userCreateArgs = repository.CreateParams{
		FirstName: "test",
		LastName: "test",
		PhoneNumber: "995555555555",
		Email: "test@email.com",
		Hash: "Ehx0DNg86zL6QCB8gMZxzkm0fPt3ObwhQzKAu22bnVYZvVe84GAAh8jFp5Cf47R5YncjKqQCyLakki78isy5899YTeVNjNjxK3N2EwdXGz4RB9YHkILLdfyT89DfAEtK",
		Role: enum.UserRoleCUSTOMER,
	}
)