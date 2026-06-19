# systemd 部署与回滚

这份说明适用于把 `gofurry-rag` 作为 Linux 服务长期运行。

## 部署前准备

1. 准备 PostgreSQL 和 pgvector。
2. 准备 Ollama，并确保 `rag.ollama_base_url` 可访问。
3. 把生产配置放到 `/etc/gofurry-rag/server.yaml`。
4. 确认控制台口令、JWT Secret 和数据库密码都不是默认值。

## 安装服务

最简单的方式是使用项目自带命令：

```bash
go run . --config /etc/gofurry-rag/server.yaml install
systemctl enable --now gofurry-rag
```

如果你需要手动管理 systemd，可以参考下面的单元文件：

```ini
[Unit]
Description=gofurry-rag
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/gofurry-rag
ExecStart=/opt/gofurry-rag/gofurry-rag serve --config /etc/gofurry-rag/server.yaml
Restart=always
RestartSec=30
Environment=APP_SERVER_MODE=release

[Install]
WantedBy=multi-user.target
```

## 日志与健康检查

- `journalctl -u gofurry-rag -f` 查看结构化日志。
- `curl http://127.0.0.1:8080/api/v1/health` 查看数据库、Ollama 和 worker 状态。
- `curl http://127.0.0.1:8080/readyz` 查看可用性。

## 回滚

1. 停止服务：

```bash
sudo systemctl stop gofurry-rag
```

2. 恢复上一个二进制或配置版本。
3. 如果需要回退数据库变更，优先使用备份恢复，而不是直接改表。
4. 重启服务：

```bash
sudo systemctl start gofurry-rag
```

## 常见排查

- 如果 `readyz` 失败，优先检查 PostgreSQL 连接。
- 如果 query 接口超时，检查 `rag.query_timeout_seconds`、Ollama 延迟和 worker 队列积压。
- 如果 worker 长时间停留在 `failed`，查看 `worker_recent_error` 和 `worker_recent_error_at`。
