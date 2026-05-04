# gofurry-rag Roadmap

这份路线图基于当前已经实现的 `gofurry-rag` 版本编写。它不追求一次性把 RAG 做成完整聊天机器人，而是先把“入库、切分、向量化、检索、管理、评估”这条链路打磨稳定。

当前原则：

```text
先让它找得准，再让它说得好。
先让问题可观测，再谈复杂优化。
先完善管理闭环，再扩展文件格式和生成模型。
```

## 当前版本现状

当前版本已经具备最小可用 RAG 闭环：

```text
文本或文件入库
  -> chunk 切分
  -> Ollama embedding
  -> pgvector 存储
  -> question embedding
  -> top_k 相似度检索
  -> 返回 sources
```

已经完成的基础能力：

- Go 服务骨架对齐 `gofurry-admin`，使用 Cobra + Viper，运行配置以 `config/server.yaml` 为主。
- 控制台使用唯一口令登录，服务端签发 HttpOnly JWT Cookie；管理接口不再使用 Admin Token 或 Bearer Header。
- `POST /api/v1/chat/query` 保持公开，用于前台或外部系统检索。
- 控制台使用 Vue + Tailwind，提供暗色左侧菜单与右侧工作区布局。
- 整体态势页每 5 秒自动刷新，并展示文档、chunks、数据库和 Ollama 状态。
- 文档管理支持手动文本入库、文件拖拽导入、批量提交入库、状态过滤、分页和删除确认。
- 文件导入已限制单文件最大 10 MiB，并限制为 `.txt`、`.md`、`.csv`、`.json`、`.yaml`、`.yml`、`.log`、`.html`、`.htm` 等文本类文件。
- 文档列表每页 6 条，并在文档 tab 打开时每 3 秒自动刷新。
- 单文档支持重新索引：删除旧 chunks，将文档设为 `pending`，由现有 worker 重新切分和向量化。
- Chunks tab 支持按文档标题或 ID 搜索，左侧文档列表每页 7 条。
- Chunk 支持查看、编辑和删除；编辑保存时会重新生成 embedding 并写回 pgvector。
- 查询页展示 top_k sources、rank、score、document/chunk 标识、来源字段和 chunk 内容。
- 控制台支持 chunk 参数切分预览，用于快速比较不同 `chunk_size/chunk_overlap` 的切分结果。
- embedding 输入模板已加入标题、来源类型、来源 ID、URL 和 chunk 正文，新入库、reindex、chunk 编辑保存都会使用。
- 已补充中文 usage、smoke-test、导入规范和 roadmap 入口。

当前暂未实现：

- 生成模型回答。
- 全量 reindex 和失败 retry。
- metadata filter。
- PDF、DOCX、OCR 等复杂文件解析。
- 检索质量测试集、评估记录模板和批量评估工具。

## v0.1.x：MVP 打磨与控制台完善

目标：让当前闭环更稳定、更容易手动运营。

当前阶段已完成：

- 完善 Chunks 管理体验：编辑失败提示、删除后刷新计数、空状态、长文本阅读、保存中反馈。
- 增加单文档 reindex 入口：`POST /api/v1/admin/documents/:id/reindex`，用于重新切分和向量化单个文档。
- 明确文件导入限制：单文件最大 10 MiB，限制文本类扩展名，不符合规则的文件会显示拒绝原因。
- 补充控制台端到端冒烟文档，覆盖登录、文件入库、状态刷新、reindex、chunk 编辑重向量化、检索命中。
- 将 `source_type`、`source_id`、`url` 的推荐使用方式写入导入规范。
- 修复控制台中文文案乱码，保持中文 UI 可读。

这个阶段仍然不接入生成模型。现在更重要的是确认 sources 是否稳定、准确、可解释。

## v0.2.0：检索质量与评估体系

目标：用真实问题证明“找得准”。

当前阶段已完成：

- 控制台强化查询调试体验：展示命中的 rank、score、document、chunk、source、URL、token_count 和原始 chunk 内容。
- `POST /api/v1/chat/query` 的 sources 增加 `source_type`、`source_id`、`chunk_index`、`token_count`，保持原有响应结构兼容。
- 新增登录态切分预览接口 `POST /api/v1/admin/debug/chunk-preview`，可用已有文档正文或临时文本对比 `500/80`、`700/120`、`900/150` 等参数。
- 切分预览只做 splitter 结果预览，不调用 Ollama、不写数据库、不改变现有文档。
- 优化 embedding 输入模板：向量化时加入标题、来源类型、来源 ID、URL 和 chunk 正文；数据库保存的 chunk 内容仍保持原文片段。
- 新入库、单文档 reindex、chunk 编辑保存都会使用新的 embedding 输入模板；旧数据需要手动重新索引后生效。

本阶段暂缓：

- 真实问题测试集暂不建立。
- 评估记录模板暂不新增。

阶段验收建议：

```text
top1 命中率 >= 60%
top3 命中率 >= 80%
top5 命中率 >= 90%
```

如果 top5 经常找不到正确资料，不建议继续接入生成模型。

## v0.3.0：Reindex、Retry 与 Metadata Filter

目标：让知识库可以反复调整、失败可恢复、检索可按范围约束。

建议新增：

- 全量 reindex：`POST /api/v1/admin/documents/reindex`
- 按来源 reindex：支持通过 `source_type` 或 metadata 选择目标文档。
- 完善状态机：`pending`、`processing`、`ready`、`failed`、`reindexing`。
- 增加 retry 计数、失败时间、处理完成时间和更清晰的错误信息。
- 在 query 请求中支持 metadata filters，例如 category、language、source_type。
- 给文档和 chunk 建立更规范的 metadata 结构。

## v0.4.0：安全、限流与运维观测

目标：让服务可以更放心地部署和长期运行。

建议新增：

- 对公开 query 接口增加限流，优先按 IP 限制。
- 限制 `top_k` 最大值和 question 最大长度。
- 给 query、embedding、ingest worker 增加明确超时。
- 增加结构化日志，记录 document_id、chunk_id、耗时、错误原因和 Ollama 调用状态。
- 在整体态势页增加 worker 状态、最近失败原因和处理耗时统计。
- 补充 systemd 部署和回滚文档。

这一阶段仍然不需要引入复杂账号体系。唯一口令 + JWT Cookie 对当前内嵌控制台已经足够。

## v0.5.0：生成模型回答与引用

目标：在检索稳定后接入 LLM，把 sources 组织成自然语言答案。

建议前置条件：

- 查询 sources 稳定，top_k 命中率达标。
- 控制台能清晰看到每次命中的 chunks 和 score。
- 支持 reindex 和失败 retry。
- query 有超时、长度限制和最大 top_k 限制。

建议能力：

- 增加可配置生成模型，例如本地 Ollama 的 `qwen3:0.6b` 或 `qwen3:1.7b`。
- 默认关闭生成能力，通过 `server.yaml` 显式启用。
- Prompt 约束模型只能基于 sources 回答；sources 没有答案时明确返回资料不足。
- API 返回结构继续保留 `sources`，不要只返回 `answer`。
- 控制台展示答案、引用来源、score 和原始 chunks。

回答约束示例：

```text
你是 GoFurry 网站的知识库助手。
请只根据提供的资料回答用户问题。
如果资料中没有答案，请回答“当前资料中没有找到相关信息”。
不要编造信息。
```

## 暂不做事项

以下能力暂不进入近期路线：

- 数据库账号表、用户管理、角色权限和复杂审计。
- PDF、DOCX、OCR、多媒体解析。
- 分布式任务队列和多节点调度。
- 复杂爬虫系统。
- 以聊天机器人体验为中心的大型对话记忆。
- rerank 模型，除非基础向量检索已经达到瓶颈。

## 近期任务清单

建议按这个顺序推进后续版本：

```text
1. 准备真实问题测试集
2. 增强查询调试和评估记录
3. 对比 chunk_size / chunk_overlap
4. 给 embedding 输入增加标题和来源上下文
5. 增加失败 retry 和更完整状态机
6. 增加 metadata filter
7. 增加 query 限流、超时和长度限制
8. 再接入生成模型和引用答案
```

最重要的一句话：

```text
不要急着让 RAG 会说话，先让它找得准、找得稳、找得可解释。
```
