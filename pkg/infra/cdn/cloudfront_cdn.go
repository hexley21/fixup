package cdn

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hexley21/handy/pkg/config"
)

type cloudfrontCdn struct {
	cdnClient      *cloudfront.Client
	distributionId string
}

func NewClient(awsCfg config.AWSCfg, cdnCfg config.CDN) (*cloudfrontCdn, error) {
	clientCfg, err := awsCfg.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	cdn := &cloudfrontCdn{
		cdnClient:      cloudfront.NewFromConfig(clientCfg),
		distributionId: cdnCfg.DistributionId,
	}

	return cdn, nil
}

func (c *cloudfrontCdn) InvalidateFile(ctx context.Context, fileName string) error {

	var x int32 = 1
	if _, err := c.cdnClient.CreateInvalidation(ctx, &cloudfront.CreateInvalidationInput{
		DistributionId: &c.distributionId,
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: &fileName,
			Paths: &types.Paths{
				Quantity: &x,
				Items:    []string{ "/" + fileName },
			},
		},
	}); err != nil {
		return err
	}

	return nil
}
