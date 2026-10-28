package aicustomerservice

import (
	"github.com/gogf/gf/v2/frame/g"
	"hotgo/internal/model/input/sysin"
)

type ConfigReq struct {
	g.Meta `path:"/ai-customer-service/config" method:"get" tags:"内部服务" summary:"获取智能客服运行配置"`
}

type ConfigRes struct {
	*sysin.AiCustomerServiceConfigModel
}
