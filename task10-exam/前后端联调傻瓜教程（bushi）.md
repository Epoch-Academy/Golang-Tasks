# Linkdesk 前后端互调傻瓜式教程

这份教程对应：

- 前端代码：[`frontend/`](./frontend/)
- 接口合同：[`API_REQUIREMENTS.md`](./API_REQUIREMENTS.md)

前端使用原生 HTML、CSS 和 JavaScript，不需要安装 React、Vue 或其他构建工具。只要 Go 后端按照接口合同实现，前端就可以直接运行。

## 1. 准备环境

需要准备：

- 可以正常运行的 Go 后端；
- Python 3 或 Node.js 二选一；
- Chrome 或 Edge 浏览器。

检查环境：

```powershell
go version
python --version
node --version
```

项目目录：

```text
task10-exam/
├─ README.md
├─ API_REQUIREMENTS.md
├─ AI_COLLAB_TEMPLATE.md
├─ FRONTEND_INTEGRATION_GUIDE.md
└─ frontend/
   ├─ index.html
   ├─ styles.css
   ├─ app.js
   └─ config.js
```

## 2. 启动 Go 后端

打开第一个 PowerShell，进入 Go 后端目录：

```powershell
cd C:\path\to\your-go-backend
```

后端至少要配置：

```text
SERVER_PORT=8080
DB_DSN=./data/app.db
DAILY_LINK_QUOTA=5
PUBLIC_BASE_URL=http://localhost:8080
CORS_ORIGINS=http://localhost:5173
TIMEZONE=Asia/Shanghai
```

如果后端使用 JWT，还需要配置：

```text
JWT_SECRET=请替换成你自己的随机密钥
```

启动后端：

```powershell
go run .
```

如果后端入口在 `cmd/server`：

```powershell
go run ./cmd/server
```

不要关闭这个窗口。

### 先检查后端

打开第二个 PowerShell，执行：

```powershell
Invoke-RestMethod http://localhost:8080/health
```

应该能看到 `status: ok`。如果这里不通，先解决后端启动、端口或数据库问题，不要先排查前端。

## 3. 配置前端 API 地址

打开：

```text
frontend/config.js
```

默认内容：

```javascript
window.LINKDESK_CONFIG = Object.freeze({
  API_BASE_URL: "http://localhost:8080",
  PAGE_SIZE: 100
});
```

如果后端端口是 `9000`，只改成：

```javascript
API_BASE_URL: "http://localhost:9000"
```

注意：

- 地址最后不要写 `/`；
- 不要把 `/api/v1` 写进 `API_BASE_URL`；
- 正确：`http://localhost:8080`；
- 错误：`http://localhost:8080/`；
- 错误：`http://localhost:8080/api/v1`。

前端会自动请求：

```text
http://localhost:8080/api/v1/auth/login
http://localhost:8080/api/v1/links
http://localhost:8080/r/a1B2c3
```

## 4. 启动前端

不要直接双击 `frontend/index.html`。使用静态服务器启动，否则浏览器可能把页面当作 `file://`，导致接口请求或 CORS 出错。

### 方式 A：Python

在项目根目录打开 PowerShell：

```powershell
cd C:\Users\24409\Desktop\goTask\task10-exam\frontend
python -m http.server 5173
```

浏览器打开：

```text
http://localhost:5173
```

### 方式 B：Node.js

如果没有 Python：

```powershell
cd C:\Users\24409\Desktop\goTask\task10-exam\frontend
npx serve -l 5173
```

如果 npx 询问是否安装 `serve`，输入 `y` 并回车。然后打开：

```text
http://localhost:5173
```

## 5. 按顺序完成联调

### 5.1 注册

点击“注册”，填写：

```text
用户名：tom
密码：123456
```

点击“创建账号”。如果提示用户名已存在，换成 `tom2026`。

对应接口：

```http
POST /api/v1/auth/register
```

### 5.2 登录

输入用户名和密码，点击“登录工作台”。前端会调用：

```http
POST /api/v1/auth/login
```

登录成功后，token 会存到浏览器 `localStorage`，后续管理接口会自动带上：

```http
Authorization: Bearer <token>
```

前端不会解析 token，也不关心后端使用 JWT 还是数据库 token。

### 5.3 创建短链接

输入：

```text
https://go.dev
```

点击“生成短链接”。请求是：

```http
POST /api/v1/links
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{
  "url": "https://go.dev"
}
```

成功后，短链接会出现在下方列表。

### 5.4 访问短链接

点击列表中的“访问”。后端应该返回：

```http
HTTP/1.1 302 Found
Location: https://go.dev
```

前端使用浏览器导航打开短链接，不使用 `fetch()` 请求跳转地址。这样浏览器才会正常完成跳转。

### 5.5 检查点击数

返回工作台，点击“刷新列表”。对应短链接的点击数应该增加。

### 5.6 复制和删除

点击“复制”可以复制 `short_url`。点击“删除”并确认后：

- 列表中不再显示该短链接；
- 原短码再次访问返回 `404`；
- 删除不会恢复当天已使用的配额。

### 5.7 测试每日配额

停止后端，将配置改成：

```text
DAILY_LINK_QUOTA=2
```

重启后端。同一个账号创建第 3 条短链接时，后端必须返回：

```http
429 Too Many Requests
```

```json
{
  "code": "DAILY_QUOTA_EXCEEDED",
  "message": "今日创建次数已达上限"
}
```

前端会显示“今日创建次数已达上限，请明天再试”。

## 6. 前端使用的接口清单

| 前端操作 | 后端接口 | token |
|---|---|---:|
| 检查服务 | `GET /health` | 否 |
| 注册 | `POST /api/v1/auth/register` | 否 |
| 登录 | `POST /api/v1/auth/login` | 否 |
| 恢复登录状态 | `GET /api/v1/me` | 是 |
| 创建短链接 | `POST /api/v1/links` | 是 |
| 读取列表 | `GET /api/v1/links?page=1&page_size=100` | 是 |
| 删除短链接 | `DELETE /api/v1/links/:id` | 是 |
| 访问短链接 | `GET /r/:code` | 否 |

字段、状态码和完整业务规则以 [`API_REQUIREMENTS.md`](./API_REQUIREMENTS.md) 为准。

## 7. 常见问题

### 页面提示“无法连接后端”

按顺序检查：

1. 后端是否仍在运行；
2. 浏览器打开 `http://localhost:8080/health` 是否返回成功；
3. `frontend/config.js` 中端口是否正确；
4. `API_BASE_URL` 是否误写了 `/api/v1`；
5. 8080 端口是否被其他程序占用。

### 浏览器控制台出现 CORS 错误

后端必须配置：

```text
CORS_ORIGINS=http://localhost:5173
```

并允许以下请求头和方法：

```text
请求头：Content-Type, Authorization
方法：GET, POST, DELETE, OPTIONS
```

改完 CORS 后需要重启后端。

### 页面不断回到登录页

这通常代表 token 过期或无效。打开浏览器开发者工具：

1. 按 `F12`；
2. 打开 `Application`；
3. 找到 `Local Storage`；
4. 删除 `linkdesk_token` 和 `linkdesk_user`；
5. 刷新页面后重新登录。

### 列表显示失败

列表接口成功响应至少应该是：

```json
{
  "items": [],
  "page": 1,
  "page_size": 100,
  "total": 0
}
```

没有数据时，`items` 必须是空数组 `[]`，不能是 `null`。

### 点击访问后不跳转

在浏览器 `F12` 的 Network 面板检查：

- 地址是否是 `/r/:code`；
- 状态码是否为 `302`；
- 响应头是否有 `Location`；
- `Location` 是否为 `http://` 或 `https://` 地址。

### 修改 `config.js` 后没有效果

使用：

```text
Ctrl + F5
```

强制刷新浏览器。

## 8. 最终验收清单

在浏览器 Network 面板中确认：

```text
POST /api/v1/auth/register   201
POST /api/v1/auth/login      200
GET  /health                 200
GET  /api/v1/me              200
GET  /api/v1/links           200
POST /api/v1/links           201
GET  /r/:code                302
DELETE /api/v1/links/:id     200
```

还要验证：

- 不带 token 创建短链接返回 `401`；
- A 用户看不到 B 用户的短链接；
- 创建第 3 条时配额为 2 的用户收到 `429`；
- 后端重启后已创建的数据仍然存在；
- 数据库中密码不是明文。

## 9. 最短启动版

只记住下面 6 步：

1. 后端目录执行 `go run .`。
2. 确认 `http://localhost:8080/health` 可访问。
3. 检查 `frontend/config.js` 的端口。
4. `frontend` 目录执行 `python -m http.server 5173`。
5. 浏览器打开 `http://localhost:5173`。
6. 注册、登录、创建、访问短链接。

