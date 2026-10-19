# 契约尚未冻结

2026-10-19，用户已授权 [ADR-0001](../docs/decisions/ADR-0001-v1-stack.md) / Plan §2.4 条件施工路线；K01 READY，随后 K02→K03→K04 严格串行，G2完整通过才启动业务线。P05/G1 的 CONDITIONAL 不是实测通过或契约冻结。

`tests/pocs/` 的消息格式仍用于隔离实验；K01 须按[完整性清单](../docs/implementation/proposals/K01-gap-review.md)补齐、验证和审查，再由集成负责人迁入正式 IPC/MCP/Directory 契约。当前没有新增生产 schema 或冻结兼容承诺，未经评审不得直接发布 POC 类型。
