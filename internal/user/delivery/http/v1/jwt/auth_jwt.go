package jwt

import (
	"time"

	"github.com/hexley21/handy/pkg/config"
	"github.com/hexley21/handy/pkg/jwt"
)

type AuthJwt interface {
	GenerateAccessKey(id string, role string) (string, error)
	GenerateRefreshKey(id string, role string) (string, error)
	VerifyAccessToken(tokenString string) (*AuthClaims, error)
}

type authJwtImpl struct {
	cfg *config.JWT
}

func NewAuthJwtImpl(cfg *config.JWT) *authJwtImpl {
	return &authJwtImpl{cfg: cfg}
}

func (j *authJwtImpl) GenerateAccessKey(id string, role string) (string, error) {
	return jwt.GenerateJWT(newAuthClaims(id, role, j.cfg.AccessTTL), j.cfg.AccessSecret)
}

func (j *authJwtImpl) GenerateRefreshKey(id string, role string) (string, error) {
	return jwt.GenerateJWT(newAuthClaims(id, role, j.cfg.RefreshTTL), j.cfg.RefreshSecret)
}

func (j *authJwtImpl) VerifyAccessToken(tokenString string) (*AuthClaims, error) {
	mapClaims, err := jwt.VerifyJWT(tokenString, j.cfg.AccessSecret)
	if err != nil {
		return nil, err
	}

	return newAuthClaims(mapClaims["id"].(string), mapClaims["role"].(string), time.Duration(mapClaims["exp"].(float64))), nil
}
