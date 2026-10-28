package sysin

import (
	"encoding/json"
	"fmt"
	"hotgo/internal/consts"
	"hotgo/internal/library/dict"
	"math"
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// AiCustomerServiceConfigModel 仅包含允许提供给客服服务的非敏感配置。
type AiCustomerServiceConfigModel struct {
	ModelBaseURL            string  `json:"model_base_url"`
	ModelName               string  `json:"model_name"`
	ModelReasoningEffort    string  `json:"model_reasoning_effort"`
	ModelTimeoutSeconds     float64 `json:"model_timeout_seconds"`
	RetrievalLimit          int     `json:"retrieval_limit"`
	EmbeddingEnabled        bool    `json:"embedding_enabled"`
	EmbeddingBaseURL        string  `json:"embedding_base_url"`
	EmbeddingModel          string  `json:"embedding_model"`
	EmbeddingTimeoutSeconds float64 `json:"embedding_timeout_seconds"`
	RerankEnabled           bool    `json:"rerank_enabled"`
	RerankURL               string  `json:"rerank_url"`
	RerankModel             string  `json:"rerank_model"`
	RerankTimeoutSeconds    float64 `json:"rerank_timeout_seconds"`
	RetrievalCandidates     int     `json:"retrieval_candidates"`
	VectorMinScore          float64 `json:"vector_min_score"`
	RerankMinScore          float64 `json:"rerank_min_score"`
}

// FilterAiCustomerServiceConfig 在转换数据库字符串前校验，防止非法数值被静默转换。
func FilterAiCustomerServiceConfig(values g.Map) (*AiCustomerServiceConfigModel, error) {
	allowed := map[string]bool{"model_base_url": true, "model_name": true, "model_timeout_seconds": true, "retrieval_limit": true}
	for key := range allowed {
		if _, exists := values[key]; !exists {
			return nil, gerror.New("智能客服配置缺少聊天模型必填参数")
		}
	}
	for _, key := range []string{"embedding_enabled", "embedding_base_url", "embedding_model", "embedding_timeout_seconds", "rerank_enabled", "rerank_url", "rerank_model", "rerank_timeout_seconds", "retrieval_candidates", "vector_min_score", "rerank_min_score"} {
		allowed[key] = true
	}
	allowed["model_reasoning_effort"] = true
	for key := range values {
		if !allowed[key] {
			return nil, gerror.New("智能客服配置包含不支持的字段")
		}
	}
	baseURL, ok := values["model_base_url"].(string)
	if !ok {
		return nil, gerror.New("模型地址必须是 HTTP(S) 地址")
	}
	baseURL = strings.TrimSpace(baseURL)
	if strings.IndexFunc(baseURL, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0 {
		return nil, gerror.New("模型地址不能包含空白或控制字符")
	}
	u, err := url.Parse(baseURL)
	if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.ForceQuery || strings.Contains(baseURL, "#") {
		return nil, gerror.New("模型地址必须是完整的 HTTP(S) 地址，且不能包含凭据、查询参数或片段")
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return nil, gerror.New("模型地址端口不合法")
		}
	}
	modelName, ok := values["model_name"].(string)
	if !ok {
		return nil, gerror.New("模型名称必须是文本")
	}
	modelName = strings.TrimSpace(modelName)
	if utf8.RuneCountInString(modelName) < 1 || utf8.RuneCountInString(modelName) > 200 {
		return nil, gerror.New("模型名称需为 1 至 200 个字符")
	}
	timeout, ok := aiCustomerServiceNumber(values["model_timeout_seconds"])
	if !ok || timeout <= 0 || timeout > 120 {
		return nil, gerror.New("请求超时必须大于 0 且不超过 120 秒")
	}
	limit, ok := aiCustomerServiceNumber(values["retrieval_limit"])
	if !ok || limit < 1 || limit > 10 || math.Trunc(limit) != limit {
		return nil, gerror.New("检索片段数必须是 1 至 10 的整数")
	}
	conf := &AiCustomerServiceConfigModel{
		ModelBaseURL: strings.TrimRight(baseURL, "/"), ModelName: modelName,
		ModelTimeoutSeconds: timeout, RetrievalLimit: int(limit),
		EmbeddingTimeoutSeconds: 30, RerankTimeoutSeconds: 30, RetrievalCandidates: 12, VectorMinScore: 0.3, RerankMinScore: 0.1,
	}
	if value, exists := values["model_reasoning_effort"]; exists {
		level, ok := value.(string)
		if !ok || !dict.HasOptionKey(consts.AiReasoningEffortOptions, level) {
			return nil, gerror.New("思考等级不合法，请选择模型支持的等级或默认")
		}
		conf.ModelReasoningEffort = level
	}
	if err := conf.filterRetrieval(values); err != nil {
		return nil, err
	}
	return conf, nil
}

func (conf *AiCustomerServiceConfigModel) filterRetrieval(values g.Map) error {
	for key, target := range map[string]*bool{"embedding_enabled": &conf.EmbeddingEnabled, "rerank_enabled": &conf.RerankEnabled} {
		if value, exists := values[key]; exists {
			switch v := value.(type) {
			case bool:
				*target = v
			case string:
				if v != "true" && v != "false" && v != "1" && v != "0" {
					return gerror.New("模型开关必须为布尔值")
				}
				*target = v == "true" || v == "1"
			default:
				return gerror.New("模型开关必须为布尔值")
			}
		}
	}
	for key, target := range map[string]*string{"embedding_base_url": &conf.EmbeddingBaseURL, "embedding_model": &conf.EmbeddingModel, "rerank_url": &conf.RerankURL, "rerank_model": &conf.RerankModel} {
		if value, exists := values[key]; exists {
			text, ok := value.(string)
			if !ok {
				return gerror.New("检索模型名称和地址必须为文本")
			}
			*target = strings.TrimSpace(text)
		}
	}
	for _, target := range []*string{&conf.EmbeddingBaseURL, &conf.RerankURL} {
		if *target == "" {
			continue
		}
		u, err := url.Parse(*target)
		if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || strings.ContainsAny(*target, "?#") || strings.IndexFunc(*target, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0 {
			return gerror.New("检索模型地址必须为无凭据、查询参数和片段的 HTTP(S) 地址")
		}
		if u.Port() != "" {
			port, err := strconv.Atoi(u.Port())
			if err != nil || port < 1 || port > 65535 {
				return gerror.New("检索模型地址端口不合法")
			}
		}
		if target == &conf.EmbeddingBaseURL {
			*target = strings.TrimRight(*target, "/")
		}
	}
	if utf8.RuneCountInString(conf.EmbeddingModel) > 200 || utf8.RuneCountInString(conf.RerankModel) > 200 {
		return gerror.New("检索模型名称不能超过 200 个字符")
	}
	if conf.EmbeddingEnabled && (conf.EmbeddingBaseURL == "" || conf.EmbeddingModel == "") {
		return gerror.New("启用向量检索需要 Embedding 地址和模型名称")
	}
	if conf.RerankEnabled && (conf.RerankURL == "" || conf.RerankModel == "") {
		return gerror.New("启用重排序需要 Rerank 完整端点和模型名称")
	}
	for key, target := range map[string]*float64{"embedding_timeout_seconds": &conf.EmbeddingTimeoutSeconds, "rerank_timeout_seconds": &conf.RerankTimeoutSeconds, "vector_min_score": &conf.VectorMinScore, "rerank_min_score": &conf.RerankMinScore} {
		if value, exists := values[key]; exists {
			n, ok := aiCustomerServiceNumber(value)
			if !ok {
				return gerror.New("检索超时和阈值必须为有效数字")
			}
			*target = n
		}
	}
	if conf.EmbeddingTimeoutSeconds <= 0 || conf.EmbeddingTimeoutSeconds > 120 || conf.RerankTimeoutSeconds <= 0 || conf.RerankTimeoutSeconds > 120 {
		return gerror.New("检索模型超时必须大于 0 且不超过 120 秒")
	}
	if conf.VectorMinScore < -1 || conf.VectorMinScore > 1 || conf.RerankMinScore < 0 || conf.RerankMinScore > 1 {
		return gerror.New("向量阈值应在 -1 至 1，重排序阈值应在 0 至 1")
	}
	if value, exists := values["retrieval_candidates"]; exists {
		n, ok := aiCustomerServiceNumber(value)
		if !ok || n < 1 || n > 50 || math.Trunc(n) != n {
			return gerror.New("候选片段数必须是 1 至 50 的整数")
		}
		conf.RetrievalCandidates = int(n)
	}
	if conf.RetrievalCandidates < conf.RetrievalLimit {
		return gerror.New("候选片段数不能小于最终片段数")
	}
	return nil
}

func aiCustomerServiceNumber(value any) (float64, bool) {
	var text string
	switch v := value.(type) {
	case string:
		text = strings.TrimSpace(v)
	case json.Number:
		text = string(v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		text = fmt.Sprint(v)
	default:
		return 0, false
	}
	number, err := strconv.ParseFloat(text, 64)
	return number, err == nil && !math.IsNaN(number) && !math.IsInf(number, 0)
}

func (conf *AiCustomerServiceConfigModel) Map() g.Map {
	return g.Map{
		"model_reasoning_effort": conf.ModelReasoningEffort,
		"model_base_url":         conf.ModelBaseURL, "model_name": conf.ModelName, "model_timeout_seconds": conf.ModelTimeoutSeconds, "retrieval_limit": conf.RetrievalLimit,
		"embedding_enabled": conf.EmbeddingEnabled, "embedding_base_url": conf.EmbeddingBaseURL, "embedding_model": conf.EmbeddingModel, "embedding_timeout_seconds": conf.EmbeddingTimeoutSeconds,
		"rerank_enabled": conf.RerankEnabled, "rerank_url": conf.RerankURL, "rerank_model": conf.RerankModel, "rerank_timeout_seconds": conf.RerankTimeoutSeconds,
		"retrieval_candidates": conf.RetrievalCandidates, "vector_min_score": conf.VectorMinScore, "rerank_min_score": conf.RerankMinScore,
	}
}
