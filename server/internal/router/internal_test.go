package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gsession"
)

func TestAiCustomerServiceAuthRemovesAuthorization(t *testing.T) {
	originalAdapter := g.Cfg().GetAdapter()
	t.Cleanup(func() { g.Cfg().SetAdapter(originalAdapter) })
	s := ghttp.GetServer(t.Name())
	s.SetAddr("127.0.0.1:0")
	s.SetDumpRouterMap(false)
	s.SetLogStdout(false)
	s.SetAccessLogEnabled(false)
	s.SetErrorLogEnabled(false)
	s.SetSessionStorage(gsession.NewStorageMemory())
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(aiCustomerServiceAuth)
		group.GET("/test", func(r *ghttp.Request) {
			if r.Header.Get("Authorization") != "" {
				t.Error("downstream handler received service token")
			}
			r.Response.Write("authorized")
		})
	})
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	for _, tt := range []struct {
		name, header, environment, config string
		status                            int
	}{
		{"environment accepted", "Bearer env-test-token", "env-test-token", "yaml-test-token", http.StatusOK},
		{"environment overrides yaml", "Bearer yaml-test-token", "env-test-token", "yaml-test-token", http.StatusUnauthorized},
		{"unset environment uses yaml", "Bearer yaml-test-token", "<unset>", "yaml-test-token", http.StatusOK},
		{"empty environment uses yaml", "Bearer yaml-test-token", "", "yaml-test-token", http.StatusOK},
		{"blank environment uses yaml", "Bearer yaml-test-token", "  ", "yaml-test-token", http.StatusOK},
		{"wrong yaml token rejected", "Bearer incorrect-test-token", "", "yaml-test-token", http.StatusUnauthorized},
		{"both empty rejected", "Bearer incorrect-test-token", "", "", http.StatusUnauthorized},
		{"missing header rejected", "", "", "yaml-test-token", http.StatusUnauthorized},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AI_CUSTOMER_SERVICE_TOKEN", tt.environment)
			if tt.environment == "<unset>" {
				if err := os.Unsetenv("AI_CUSTOMER_SERVICE_TOKEN"); err != nil {
					t.Fatal(err)
				}
			}
			adapter, err := gcfg.NewAdapterContent("aiCustomerService:\n  serviceToken: '" + tt.config + "'\n")
			if err != nil {
				t.Fatal(err)
			}
			g.Cfg().SetAdapter(adapter)
			req := httptest.NewRequest(http.MethodGet, "http://localhost/test", nil)
			req.Header.Set("Authorization", tt.header)
			res := httptest.NewRecorder()
			s.ServeHTTP(res, req)
			if res.Code != tt.status {
				t.Fatalf("status=%d, want=%d", res.Code, tt.status)
			}
			if req.Header.Get("Authorization") != "" {
				t.Error("request retained service token for logging")
			}
		})
	}
}

func TestAiCustomerServiceAuthorization(t *testing.T) {
	for _, tt := range []struct {
		name, header, token string
		want                bool
	}{
		{"valid", "Bearer private-test-token", "private-test-token", true},
		{"case insensitive scheme", "bearer private-test-token", "private-test-token", true},
		{"missing header", "", "private-test-token", false},
		{"missing environment", "Bearer private-test-token", "", false},
		{"blank environment", "Bearer   ", "   ", false},
		{"wrong token", "Bearer incorrect", "private-test-token", false},
		{"empty token", "Bearer ", "private-test-token", false},
		{"basic scheme", "Basic private-test-token", "private-test-token", false},
		{"extra token", "Bearer private-test-token extra", "private-test-token", false},
		{"prefix attack", "Bearer private-test-token-suffix", "private-test-token", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := validAiCustomerServiceAuthorization(tt.header, tt.token); got != tt.want {
				t.Fatalf("authorized=%v, want=%v", got, tt.want)
			}
		})
	}
}
