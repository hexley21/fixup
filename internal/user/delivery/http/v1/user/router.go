package user

import (
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/labstack/echo/v4"
)

func (h *userHandler) MapRoutes(e *echo.Group, accessSecretKey string) *echo.Group {
	usersGroup := e.Group("/users")

	usersGroup.Use(
		middleware.JWT(accessSecretKey),
		middleware.AllowSelfOrRole(enum.UserRoleADMIN, enum.UserRoleMODERATOR),
	)

	usersGroup.GET("/:id", h.findUserById)
	usersGroup.PATCH("/:id", h.updateUserData)
	usersGroup.DELETE("/:id", h.deleteUser)

	usersGroup.PUT("/:id/pfp", h.uploadProfilePicture,
		middleware.AllowFilesAmount("image", 1),
		middleware.AllowContentType("image", "image/jpeg", "image/png"),
	)

	return usersGroup
}
