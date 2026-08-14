# OpsLang 快速入门

## 安装

```bash
# 从源码编译
git clone https://github.com/opslang/opslang.git
cd opslang
make build

# 安装到 PATH
cp bin/ops /usr/local/bin/
```

## 第一个脚本

创建 `hello.ops`：

```ops
// 变量和输出
name = "运维工程师"
print("你好, {name}!")

// 数组和循环
hosts = ["web01", "web02", "web03"]
for host in hosts
    print("检查: {host}")

// 函数
fn check_disk(host, threshold)
    usage = 75  // 模拟值
    if usage > threshold
        print("警告: {host} 磁盘 {usage}%")
    else
        print("正常: {host} 磁盘 {usage}%")

for host in hosts
    check_disk(host, 80)
```

运行：

```bash
ops run hello.ops
```

## REPL 交互

```bash
$ ops repl
ops>>> name = "OpsLang"
ops>>> print("Hello, {name}!")
Hello, OpsLang!
ops>>> for i in range(5)
...        print(i)
0
1
2
3
4
ops>>> :quit
```

## 编译为单二进制

```bash
# 编译
ops build hello.ops hello

# 运行（不需要 ops 运行时）
./hello

# 交叉编译到 Linux
GOOS=linux GOARCH=amd64 ops build hello.ops hello_linux
```

## 运维实战

### 批量检查服务器

```ops
#!/usr/bin/env ops run
// batch_check.ops

hosts = ["web01", "web02", "web03", "db01"]

// 批量执行 uptime
results = fleet.parallel(hosts, fn(h) => ssh.run(h, "uptime"))

for r in results
    host = r["host"]
    ok = r["ok"]
    if ok
        stdout = r["result"]["stdout"]
        print("✓ {host}: {stdout}")
    else
        print("✗ {host}: 连接失败")

summary = fleet.summary(results)
print("\n总计: {summary.total}, 成功: {summary.ok}, 失败: {summary.fail}")
```

### 配置文件管理

```ops
// 用三引号嵌入 YAML 配置
nginxConfig = """
server {
    listen 80;
    server_name example.com;
    location / {
        proxy_pass http://backend:8080;
    }
}
"""

// 声明式保证文件内容
result = ensure.file("/etc/nginx/conf.d/app.conf", nginxConfig.trim())
if result.changed
    process.shell("nginx -t && nginx -s reload")
    print("配置已更新并重载")
else
    print("配置无变化")
```

### 从清单批量部署

```ops
#!/usr/bin/env ops run
// deploy.ops

// 加载主机清单
inv = inventory.load("/etc/ops/hosts.ini")
webs = inventory.group(inv, "web_servers")

// 读取配置
appConfig = yaml.load_file("config/app.yaml")
version = appConfig["app"]["version"]

print("部署 v{version} 到 {len(webs)} 台服务器")

// 批量部署
results = fleet.parallel(webs, fn(h) => h)

for r in results
    if r["ok"]
        print("✓ {r.host} 部署成功")
    else
        print("✗ {r.host} 部署失败")
```

## 下一步

- [语言参考](language-reference.md) — 完整语法和 API
- [示例代码](../examples/) — 更多实战脚本
- [标准库](../stdlib/) — 所有模块文档
