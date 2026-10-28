package sysin

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
)

func validAiCustomerServiceValues() g.Map {
	return g.Map{"model_base_url": "https://model.example/v1", "model_name": "test-model", "model_timeout_seconds": 30.5, "retrieval_limit": 3}
}

func TestAiCustomerServiceReasoningEffort(t *testing.T) {
	legacy, err := FilterAiCustomerServiceConfig(validAiCustomerServiceValues())
	if err != nil || legacy.Map()["model_reasoning_effort"] != "" {
		t.Fatal("legacy configuration must default to provider behavior")
	}
	for _, level := range []string{"", "none", "minimal", "low", "medium", "high", "xhigh", "max"} {
		t.Run("level="+level, func(t *testing.T) {
			values := validAiCustomerServiceValues()
			values["model_reasoning_effort"] = level
			in := &UpdateConfigInp{Group: "ai_customer_service", List: values}
			if err := filterConfig(t, in); err != nil {
				t.Fatal(err)
			}
			conf, err := FilterAiCustomerServiceConfig(in.List)
			if err != nil || conf.Map()["model_reasoning_effort"] != level {
				t.Fatal("reasoning level lost in storage round trip")
			}
			encoded, err := json.Marshal(conf)
			if err != nil {
				t.Fatal(err)
			}
			var response map[string]any
			if err := json.Unmarshal(encoded, &response); err != nil || response["model_reasoning_effort"] != level {
				t.Fatal("reasoning level missing from internal response")
			}
		})
	}
	for _, value := range []any{nil, true, 1, []string{"high"}, "auto", "HIGH", " high "} {
		values := validAiCustomerServiceValues()
		values["model_reasoning_effort"] = value
		if _, err := FilterAiCustomerServiceConfig(values); err == nil {
			t.Fatalf("invalid reasoning level accepted: %#v", value)
		}
	}
}

func filterConfig(t *testing.T, in *UpdateConfigInp) error {
	t.Helper()
	filter, ok := any(in).(interface{ Filter(context.Context) error })
	if !ok {
		t.Fatal("configuration input must validate the AI customer service group")
	}
	return filter.Filter(context.Background())
}

func TestAiCustomerServiceConfigValidation(t *testing.T) {
	tests := []struct {
		name, key string
		value     any
	}{
		{"blank URL", "model_base_url", " "},
		{"URL credentials", "model_base_url", "https://user:password@model.example/v1"},
		{"URL query", "model_base_url", "https://model.example/v1?key=hidden"},
		{"URL empty query", "model_base_url", "https://model.example/v1?"},
		{"URL fragment", "model_base_url", "https://model.example/v1#secret"},
		{"URL scheme", "model_base_url", "file:///tmp/model"},
		{"URL host", "model_base_url", "http:///v1"},
		{"URL path whitespace", "model_base_url", "https://model.example/v1/a b"},
		{"URL path unicode whitespace", "model_base_url", "https://model.example/v1/a\u00a0b"},
		{"URL path control character", "model_base_url", "https://model.example/v1/a\u007fb"},
		{"URL empty fragment", "model_base_url", "https://model.example/v1#"},
		{"URL wrong type", "model_base_url", 1},
		{"blank model", "model_name", " "},
		{"long model", "model_name", strings.Repeat("中", 201)},
		{"model wrong type", "model_name", []string{"model"}},
		{"zero timeout", "model_timeout_seconds", 0},
		{"large timeout", "model_timeout_seconds", 120.01},
		{"invalid timeout", "model_timeout_seconds", "oops"},
		{"NaN timeout", "model_timeout_seconds", math.NaN()},
		{"infinite timeout", "model_timeout_seconds", math.Inf(1)},
		{"boolean timeout", "model_timeout_seconds", true},
		{"zero retrieval", "retrieval_limit", 0},
		{"large retrieval", "retrieval_limit", 11},
		{"fractional retrieval", "retrieval_limit", 3.2},
		{"stored fractional retrieval", "retrieval_limit", "3.2"},
		{"boolean retrieval", "retrieval_limit", true},
		{"nil retrieval", "retrieval_limit", nil},
		{"secret key", "api_key", "must-never-be-stored"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := validAiCustomerServiceValues()
			values[tt.key] = tt.value
			if err := filterConfig(t, &UpdateConfigInp{Group: "ai_customer_service", List: values}); err == nil {
				t.Fatal("invalid setting accepted")
			} else if strings.Contains(err.Error(), "must-never-be-stored") || strings.Contains(err.Error(), "password") {
				t.Fatal("validation error exposed a supplied secret")
			}
		})
	}
}

func TestAiCustomerServiceConfigRequiresEveryField(t *testing.T) {
	for key := range validAiCustomerServiceValues() {
		values := validAiCustomerServiceValues()
		delete(values, key)
		if err := filterConfig(t, &UpdateConfigInp{Group: "ai_customer_service", List: values}); err == nil {
			t.Fatalf("missing %s was accepted", key)
		}
	}
}

func TestAiCustomerServiceConfigNormalizesStorageValues(t *testing.T) {
	in := &UpdateConfigInp{Group: "ai_customer_service", List: g.Map{
		"model_base_url": " https://model.example/v1/ ", "model_name": " test-model ",
		"model_timeout_seconds": "0.5", "retrieval_limit": "10",
	}}
	if err := filterConfig(t, in); err != nil {
		t.Fatal(err)
	}
	if in.List["model_base_url"] != "https://model.example/v1" || in.List["model_name"] != "test-model" || in.List["model_timeout_seconds"] != float64(0.5) || in.List["retrieval_limit"] != 10 {
		t.Fatalf("unexpected normalization: %#v", in.List)
	}
	if err := filterConfig(t, &UpdateConfigInp{Group: "basic", List: g.Map{"existing": "unchanged"}}); err != nil {
		t.Fatalf("unrelated config group changed: %v", err)
	}
}

func TestAiCustomerServiceRetrievalConfig(t *testing.T) {
	values := validAiCustomerServiceValues()
	values["embedding_enabled"] = "true"
	values["embedding_base_url"] = "https://embed.example/v1/"
	values["embedding_model"] = "embedding-model"
	values["rerank_enabled"] = true
	values["rerank_url"] = "https://rank.example/v1/rerank/"
	values["rerank_model"] = "reranker"
	conf, err := FilterAiCustomerServiceConfig(values)
	if err != nil {
		t.Fatal(err)
	}
	if conf.Map()["embedding_enabled"] != true || conf.Map()["embedding_base_url"] != "https://embed.example/v1" {
		t.Fatal("retrieval configuration lost")
	}
	if conf.RerankURL != "https://rank.example/v1/rerank/" {
		t.Fatal("full rerank endpoint must preserve trailing slash")
	}
	if _, err := FilterAiCustomerServiceConfig(conf.Map()); err != nil {
		t.Fatal(err)
	}
	invalidValues := g.Map{"embedding_base_url": "https://user:secret@example.com", "rerank_url": "https://example.com?secret=x", "vector_min_score": 1.1, "rerank_min_score": -0.1, "retrieval_candidates": 2, "embedding_timeout_seconds": 0, "rerank_timeout_seconds": 121, "embedding_enabled": "yes", "rerank_model": ""}
	for key, value := range invalidValues {
		invalid := conf.Map()
		invalid[key] = value
		if _, err := FilterAiCustomerServiceConfig(invalid); err == nil {
			t.Fatalf("accepted invalid %s", key)
		}
	}
	legacy, err := FilterAiCustomerServiceConfig(validAiCustomerServiceValues())
	if err != nil || legacy.Map()["embedding_enabled"] != false {
		t.Fatal("legacy configuration must disable embedding")
	}
}
