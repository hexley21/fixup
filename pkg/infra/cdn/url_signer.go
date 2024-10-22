package cdn

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign"
	"github.com/hexley21/fixup/pkg/config"
)

type URLSigner interface {
	SignURL(fileName string) (string, error)
}

type CloudFrontURLSigner struct {
	cfg config.CDN
}

func NewCloudFrontURLSigner(cfg config.CDN) *CloudFrontURLSigner {
	return &CloudFrontURLSigner{cfg: cfg}
}

// SignURL generates a signed URL for the given file name using CloudFront URL signing.
// It creates a new URL signer with the configured KeyPairId and PrivateKey, then signs
// the URL formatted with the file name and an expiry time. It returns the signed URL
// or an error if the signing process fails.
func (s *CloudFrontURLSigner) SignURL(fileName string) (string, error) {
	signer := sign.NewURLSigner(s.cfg.KeyPairId, s.cfg.PrivateKey)

	signedURL, err := signer.Sign(
		fmt.Sprintf(s.cfg.UrlFmt, fileName),
		time.Now().Add(s.cfg.Expiry),
	)
	if err != nil {
		return "", err
	}

	return signedURL, nil
}
