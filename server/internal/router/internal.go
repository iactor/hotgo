package router

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	internalcontroller "hotgo/internal/controller/integration"
	"net/http"
	"os"
	"strings"
)

// Internal 服务间接口独立鉴权，不复用管理员登录凭据。
func Internal(_ context.Context, group *ghttp.RouterGroup) {
	group.Group("/internal", func(group *ghttp.RouterGroup) {
		group.Middleware(aiCustomerServiceAuth)
		group.Bind(internalcontroller.AiCustomerService)
	})
}

func aiCustomerServiceAuth(r *ghttp.Request) {
	authorization := r.Header.Get("Authorization")
	// 请求结束后的访问日志可能读取 Header；无论鉴权结果如何都不保留服务令牌。
	r.Header.Del("Authorization")
	expected := os.Getenv("AI_CUSTOMER_SERVICE_TOKEN")
	if strings.TrimSpace(expected) == "" {
		// 本地开发可使用私有 YAML 配置；读取失败时保持拒绝访问。
		if value, err := g.Cfg().Get(r.Context(), "aiCustomerService.serviceToken"); err == nil {
			expected = value.String()
		}
	}
	if !validAiCustomerServiceAuthorization(authorization, expected) {
		r.Response.Status = http.StatusUnauthorized
		r.SetError(gerror.New("内部服务身份验证失败"))
		return
	}
	r.Middleware.Next()
}

func validAiCustomerServiceAuthorization(header, expected string) bool {
	parts := strings.Fields(header)
	if strings.TrimSpace(expected) == "" || len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return false
	}
	actualHash, expectedHash := sha256.Sum256([]byte(parts[1])), sha256.Sum256([]byte(expected))
	return subtle.ConstantTimeCompare(actualHash[:], expectedHash[:]) == 1
}
