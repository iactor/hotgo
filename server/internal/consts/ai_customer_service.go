package consts

import (
	"hotgo/internal/library/dict"
	"hotgo/internal/model"
)

const (
	AiReasoningEffortDefault = ""
	AiReasoningEffortNone    = "none"
	AiReasoningEffortMinimal = "minimal"
	AiReasoningEffortLow     = "low"
	AiReasoningEffortMedium  = "medium"
	AiReasoningEffortHigh    = "high"
	AiReasoningEffortXHigh   = "xhigh"
	AiReasoningEffortMax     = "max"
)

// AiReasoningEffortOptions 具体可用等级由模型服务决定，默认不传递该参数。
var AiReasoningEffortOptions = []*model.Option{
	dict.GenDefaultOption(AiReasoningEffortDefault, "默认（跟随模型）"),
	dict.GenInfoOption(AiReasoningEffortNone, "关闭（none）"),
	dict.GenInfoOption(AiReasoningEffortMinimal, "最低（minimal）"),
	dict.GenInfoOption(AiReasoningEffortLow, "低（low）"),
	dict.GenInfoOption(AiReasoningEffortMedium, "中（medium）"),
	dict.GenInfoOption(AiReasoningEffortHigh, "高（high）"),
	dict.GenInfoOption(AiReasoningEffortXHigh, "超高（xhigh）"),
	dict.GenInfoOption(AiReasoningEffortMax, "最高（max）"),
}

func init() {
	dict.RegisterEnums("AiReasoningEffortOptions", "智能客服思考等级", AiReasoningEffortOptions)
}
