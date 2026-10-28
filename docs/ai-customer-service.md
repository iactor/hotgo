# 智能客服配置

系统设置中的“智能客服”页复用 `/config/get` 和 `/config/update` 管理员权限，维护聊天、Embedding、Rerank 三类非敏感设置。密钥仍由 Python 客服后端 `backend/.env` 的 `MODEL_API_KEY`、`EMBEDDING_API_KEY`、`RERANK_API_KEY` 管理。

现有系统配置表（逻辑表名 `sys_config`，物理表名沿用当前数据库前缀）需预先配置 `group = ai_customer_service` 的四条必填记录及可选扩展项：

| key | type | 校验 |
| --- | --- | --- |
| model_base_url | string | 完整 HTTP(S) 地址，不含凭据、查询、片段、空白或控制字符 |
| model_name | string | 去除首尾空白后 1–200 个字符 |
| model_reasoning_effort（可选） | string | 默认空字符串；none/minimal/low/medium/high/xhigh/max，取决于模型支持情况 |
| model_timeout_seconds | float64 | 大于 0 且不超过 120 秒 |
| retrieval_limit | int | 1–10 的整数 |

检索扩展项：

| key | type | 默认值/校验 |
| --- | --- | --- |
| embedding_enabled / rerank_enabled | bool | false；启用时地址和名称必须完整 |
| embedding_base_url | string | 空；完整 HTTP(S) 基础地址，Python 自动追加 /embeddings |
| embedding_model / rerank_model | string | 空；最长 200 字符 |
| rerank_url | string | 空；完整 Rerank 端点 |
| embedding_timeout_seconds / rerank_timeout_seconds | float64 | 30；大于 0，不超过 120 |
| retrieval_candidates | int | 12；1–50，且不小于 retrieval_limit |
| vector_min_score | float64 | 0.3；-1 至 1 |
| rerank_min_score | float64 | 0.1；0 至 1 |

已有四项配置可正常读取，新字段使用默认值。保存前需补齐数据库配置记录：在相邻 `ai-customer-service` 目录运行 `backend/.venv/Scripts/python.exe scripts/local_rollout.py provision-retrieval`，只插入缺失项，不覆盖管理员修改。升级时先构建新版、停止旧服务，再迁移和重启新版 Go/Python。完整接入说明见 `ai-customer-service/docs/guided-retrieval.md`。

初始化时复制当前客服已验证的实际参数，不使用演示地址覆盖运行设置。初始化记录由部署操作完成；页面保存不会自行插入记录，也不会接受白名单之外的配置字段。保存时聊天模型四个字段必须一并提交，页面同时提交检索字段。读取时同样验证数据库原始值，不会把小数检索数量自动截断。

## 思考等级升级

在相邻 `ai-customer-service` 目录运行 `backend/.venv/Scripts/python.exe scripts/local_rollout.py provision-reasoning`，只补充缺失的思考等级配置记录，保留管理员已有值。先构建新版本、停止旧 Go/Python 服务，再执行迁移并启动新版服务；旧版 Go 白名单及 Python 配置会拒绝新增字段。

后台“聊天模型”区域可选择思考等级。默认空字符串表示不向模型发送参数；`none` 明确请求关闭思考，二者不同。不同模型支持的等级不同，不支持此参数的服务请使用默认。较高等级通常增加耗时和用量，参数同时用于工具选择和回答生成，不影响 Embedding/Rerank。

Python 本地配置模式使用 `MODEL_REASONING_EFFORT`；HotGo 模式使用后台值。旧数据库缺少此项时可正常读取，保存前须执行上面的迁移。保存后最多 30 秒应用于新请求。

## 服务间读取

`GET /internal/ai-customer-service/config` 使用 `Authorization: Bearer <service token>`。HotGo 优先读取非空的 `AI_CUSTOMER_SERVICE_TOKEN` 环境变量；未设置、空字符串或仅空白时，读取 `server/manifest/config/config.yaml` 的 `aiCustomerService.serviceToken`。两处均未配置、配置读取失败或凭据错误时返回 HTTP 401。环境变量有值但不匹配时不会再尝试 YAML 中的令牌。

```yaml
aiCustomerService:
  serviceToken: "填写与 Python HOTGO_SERVICE_TOKEN 相同的值"
```

将同一值配置到 Python 后端的 `HOTGO_SERVICE_TOKEN`，将完整接口地址配置到 `HOTGO_CONFIG_URL`，并设置 `CONFIG_SOURCE=hotgo`。HotGo 不自动加载 `.env`；本地填好 YAML 后可直接使用 `gf run` 启动。更改环境变量后需重启 HotGo。令牌只保存在服务端私有配置/进程环境中，不放入系统配置表、浏览器参数或代码仓库；示例配置中的令牌保持为空。

成功响应沿用 HotGo 标准格式，`data` 包含上述非敏感配置字段。Python 最多缓存 30 秒，所以后台保存后新的问答请求在缓存到期后生效；正在进行的问答使用原有配置快照。

## 本地验证

后端在 `server` 目录执行：

```text
go test ./internal/model/input/sysin ./internal/router ./internal/controller/integration ./internal/logic/sys -run AiCustomerService -count=1
go build -o temp/ai-customer-service.exe .
```

前端生产构建沿用 `web/package.json` 的 `build` 脚本（Vite 构建后运行 `build/script/postBuild.ts`）。当前仓库全量类型检查存在既有类型配置及公共组件错误，新增页面需同时检查 Vue 模板类型，并与未修改的公共依赖基线比较，不能将生产构建成功视为全量类型检查通过。
