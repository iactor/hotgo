// Package sysin
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2023 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
package sysin

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"hotgo/internal/consts"
)

// UpdateConfigInp 更新指定分组的配置
type UpdateConfigInp struct {
	Group string `json:"group"`
	List  g.Map  `json:"list"`
}

func (in *UpdateConfigInp) Filter(_ context.Context) error {
	if in.Group != consts.ConfigGroupAiCustomerService {
		return nil
	}
	conf, err := FilterAiCustomerServiceConfig(in.List)
	if err != nil {
		return err
	}
	in.List = conf.Map()
	return nil
}

// GetConfigInp 获取指定分组的配置
type GetConfigInp struct {
	Group string `json:"group"`
}

type GetConfigModel struct {
	List g.Map `json:"list"`
}
