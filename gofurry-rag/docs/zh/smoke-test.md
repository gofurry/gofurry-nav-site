# 冒烟测试

1. 构建管理控制台：

```bash
cd web
npm ci
npm run build
cd ..
```

2. 运行后端测试：

```bash
go test ./...
```

3. 在 `.env` 中配置有效 PostgreSQL 密码并启动服务：

```bash
go run ./cmd/server
```

4. 检查健康状态：

```bash
curl http://127.0.0.1:8080/api/v1/health
```

期望返回：

```json
{"code":1,"message":"success","data":{"status":"ok"}}
```

5. 登录、创建文本文档，等待几秒后查询：

```bash
curl -c cookies.txt -X POST http://127.0.0.1:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"password":"change-me"}'

curl -X POST http://127.0.0.1:8080/api/v1/admin/documents/text \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"title":"Smoke Test","content":"GoFurry 会存储可检索的知识片段。","source_type":"manual"}'

curl -X POST http://127.0.0.1:8080/api/v1/chat/query \
  -H "Content-Type: application/json" \
  -d '{"question":"GoFurry 会存储什么？","top_k":3}'
```

后台 worker 完成后，查询结果应该包含至少一个 source。
