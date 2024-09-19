package user

import (
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/labstack/echo/v4"
)

func (h *userHandler) MapRoutes(
	e *echo.Group,
	jwtAccessMiddleware echo.MiddlewareFunc,
	onlyVerifiedMiddleware echo.MiddlewareFunc,
) *echo.Group {
	usersGroup := e.Group("/users")

	usersGroup.Use(
		jwtAccessMiddleware,
		middleware.AllowSelfOrRole(enum.UserRoleADMIN, enum.UserRoleMODERATOR),
	)

	usersGroup.GET("/:id", h.FindUserById, onlyVerifiedMiddleware)
	usersGroup.PATCH("/:id", h.UpdateUserData, onlyVerifiedMiddleware)
	usersGroup.DELETE("/:id", h.DeleteUser, onlyVerifiedMiddleware)

	usersGroup.PATCH("/:id/pfp", h.UploadProfilePicture,
		middleware.AllowFilesAmount("image", 1),
		middleware.AllowContentType("image", "image/jpeg", "image/png"),
	)

	e.PATCH("/me/change-password", h.ChangePassword, jwtAccessMiddleware)

	return usersGroup
}
