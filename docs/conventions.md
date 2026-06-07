# 开发规范

- HTTP 契约放在 `idl/api` 中
- RPC 契约放在 `idl/rpc` 中
- SQL 放在 `internal/repository` 中
- 编排规则放在 `internal/domain/service` 中
- 传输适配放在 `internal/logic` 中
- 对非平凡的模型、工作流、缓存键和消费者逻辑保留详细注释