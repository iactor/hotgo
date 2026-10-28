package integration

import (
	"context"
	"hotgo/api/integration/aicustomerservice"
	"hotgo/internal/service"
)

var AiCustomerService = cAiCustomerService{}

type cAiCustomerService struct{}

func (c *cAiCustomerService) Config(ctx context.Context, _ *aicustomerservice.ConfigReq) (res *aicustomerservice.ConfigRes, err error) {
	res = new(aicustomerservice.ConfigRes)
	res.AiCustomerServiceConfigModel, err = service.SysConfig().GetAiCustomerService(ctx)
	return
}
