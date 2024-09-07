package cdn

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign"
	"github.com/hexley21/fixup/pkg/config"
)

type URLSigner interface {
	SignURL(pictureName string) (string, error)
}

type CloudFrontURLSigner struct {
	cfg config.CDN
}

func NewCloudFrontURLSigner(cfg config.CDN) *CloudFrontURLSigner {
	return &CloudFrontURLSigner{cfg: cfg}
}

func (s *CloudFrontURLSigner) SignURL(pictureName string) (string, error) {
	signer := sign.NewURLSigner(s.cfg.KeyPairId, s.cfg.PrivateKey)

	signedURL, err := signer.Sign(
		fmt.Sprintf(s.cfg.UrlFmt, pictureName),
		time.Now().Add(s.cfg.Expiry),
	)
	if err != nil {
		return "", err
	}

	return signedURL, nil
}
