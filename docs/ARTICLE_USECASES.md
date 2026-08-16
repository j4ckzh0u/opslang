# 用 OpsLang 替代你的 Shell 脚本

> 10 个真实运维场景，对比 Shell 和 OpsLang 的写法。看完你自己判断。

## 1. 采集系统信息

**Shell：**

```bash
#!/bin/bash
HOSTNAME=$(hostname)
CPU=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)
MEM_TOTAL=$(free -m | awk '/Mem:/{print $2}')
MEM_USED=$(free -m | awk '/Mem:/{print $3}')
MEM_PCT=$(echo "scale=1; $MEM_USED * 100 / $MEM_TOTAL" | bc)
DISK_PCT=$(df -h / | awk 'NR==2{print $5}' | tr -d '%')

echo "host=$HOSTNAME cpu=$CPU mem=${MEM_PCT}% disk=${DISK_PCT}%"
```

问题：
- `top -bn1` 的 CPU 值是自启动以来的平均值，不是当前值
- `free -m` 的 "used" 包含 buffer/cache，不准确
- 输出是字符串，需要正则解析
- 错误处理？不存在的

**OpsLang：**

```ops
let cpu = sys.cpu.usage()
let mem = sys.memory.info()
let disk = sys.disk.usage("/")

report {
    host: sys.hostname(),
    cpu: cpu.percent,
    mem: mem.used_percent,
    disk: disk.used_percent
}
```

输出：

```json
{
  "host": "web-01",
  "cpu": 12.5,
  "mem": 68.3,
  "disk": 45.2
}
```

`sys.cpu.usage()` 两次采样间隔 500ms，反映当前值。返回值是结构体，直接访问字段。

## 2. 检查磁盘空间并报警

**Shell：**

```bash
#!/bin/bash
THRESHOLD=90
df -h | awk 'NR>1 {print $5, $6}' | while read pct mount; do
    pct_num=${pct%\%}
    if [ "$pct_num" -gt "$THRESHOLD" ]; then
        echo "WARNING: $mount is ${pct_num}% full"
        # 发邮件？调用告警 API？
        # curl -X POST https://alerts.example.com/webhook \
        #   -d "{\"text\": \"$mount is ${pct_num}% full\"}"
    fi
done
```

问题：
- 字符串比较转数字，容易出错
- 告警逻辑混在循环里
- 多主机？再套一层 SSH 循环

**OpsLang：**

```ops
let threshold = 90
let disks = sys.disk.partitions()

for disk in disks {
    let usage = sys.disk.usage(disk.mount_point)
    if usage.used_percent > threshold {
        alert("磁盘空间不足: " + disk.mount_point + " " + str(usage.used_percent) + "%")
    }
}

report {
    host: sys.hostname(),
    disks: disks
}
```

`alert()` 是内置函数，自动对接告警通道。结构化输出可以被监控系统直接解析。

## 3. 文件部署

**Shell：**

```bash
#!/bin/bash
SOURCE="/data/releases/app-v2.1.0.tar.gz"
DEST="/opt/app/releases/"
HOSTS="host1 host2 host3"

for host in $HOSTS; do
    scp "$SOURCE" "$host:$DEST"
    ssh "$host" "cd /opt/app && tar xzf releases/app-v2.1.0.tar.gz && systemctl restart myapp"
    if [ $? -ne 0 ]; then
        echo "FAILED: $host"
    else
        echo "OK: $host"
    fi
done
```

问题：
- 串行执行，3 台主机就要等 3 倍时间
- 没有校验和验证，传输中断不知道
- 没有回滚机制
- 没有幂等性

**OpsLang：**

```ops
task "deploy" on ["host1", "host2", "host3"] {
    file.distribute(
        source: "/data/releases/app-v2.1.0.tar.gz",
        dest: "/opt/app/releases/",
        compress: true,
        checksum: true,
        on_changed: {
            process.exec("tar", "xzf", "/opt/app/releases/app-v2.1.0.tar.gz", "-C", "/opt/app")
            service.restart("myapp")
        }
    )
}
```

- 并行分发，3 台和 1 台耗时差不多
- SHA256 校验和，传输完整性保证
- `on_changed` 只在文件变化时触发操作（幂等）
- 失败自动重试，临时文件自动清理

## 4. 服务健康检查

**Shell：**

```bash
#!/bin/bash
SERVICES="nginx mysql redis"

for svc in $SERVICES; do
    status=$(systemctl is-active $svc)
    if [ "$status" != "active" ]; then
        echo "$svc is $status"
        systemctl restart $svc
    fi
done
```

问题：
- 只检查了 active 状态，没检查是否响应请求
- 重启失败不知道
- 没有结构化输出

**OpsLang：**

```ops
let services = ["nginx", "mysql", "redis"]

for name in services {
    let status = service.status(name)
    if status.active != true {
        alert("服务异常: " + name + " state=" + status.sub_state)
        let result = service.restart(name)
        if result.success != true {
            alert("重启失败: " + name + " error=" + result.error)
        }
    }
}

report {
    host: sys.hostname(),
    services: services
}
```

`service.status()` 返回完整状态（active/inactive/failed + sub_state），`service.restart()` 返回成功/失败和错误信息。

## 5. 日志收集

**Shell：**

```bash
#!/bin/bash
HOSTS="host1 host2 host3"
LOG="/var/log/nginx/access.log"
DEST="/data/logs/"

for host in $HOSTS; do
    mkdir -p "$DEST/$host"
    scp "$host:$LOG" "$DEST/$host/access_$(date +%Y%m%d).log"
done
```

问题：
- 串行 scp，大日志很慢
- 没有压缩，浪费带宽
- 断点续传？不存在的
- 日志按日期命名，但同名文件会覆盖

**OpsLang：**

```ops
task "collect_logs" on ["host1", "host2", "host3"] {
    file.collect(
        source: "/var/log/nginx/access.log",
        dest: "/data/logs/{host}/access_{date}.log",
        compress: true,
        resume: true
    )
}
```

- 并行收集
- 自动压缩传输
- 断点续传（`resume: true`）
- `{host}` 和 `{date}` 自动替换，不会覆盖

## 6. 进程管理

**Shell：**

```bash
#!/bin/bash
# 查找占用 8080 端口的进程
PID=$(lsof -ti :8080)
if [ -n "$PID" ]; then
    echo "Process $PID is using port 8080"
    kill -TERM $PID
    sleep 2
    if kill -0 $PID 2>/dev/null; then
        kill -KILL $PID
    fi
fi
```

问题：
- `lsof` 不是所有系统都有
- 信号处理靠 sleep + kill -0，不优雅
- 错误处理缺失

**OpsLang：**

```ops
let procs = process.find_by_port(8080)

for p in procs {
    print("Found: pid=" + str(p.pid) + " name=" + p.name)
    let result = process.kill(p.pid, "TERM")
    if result.success != true {
        print("TERM failed, sending KILL")
        process.kill(p.pid, "KILL")
    }
}
```

`process.find_by_port()` 直接读取 `/proc` 或调用系统 API，不依赖 `lsof`。`process.kill()` 返回成功/失败，不需要手动检查进程是否还活着。

## 7. 包管理

**Shell：**

```bash
#!/bin/bash
# 安装 nginx，区分 apt/yum
if command -v apt-get >/dev/null; then
    apt-get update && apt-get install -y nginx
elif command -v yum >/dev/null; then
    yum install -y nginx
elif command -v dnf >/dev/null; then
    dnf install -y nginx
else
    echo "Unsupported package manager"
    exit 1
fi
```

问题：
- 每次都要写 if-elif 判断包管理器
- 返回值不明确
- 幂等性？apt-get install 会重复执行

**OpsLang：**

```ops
let result = pkg.install("nginx")

if result.success {
    if result.changed {
        print("nginx installed")
    } else {
        print("nginx already installed")
    }
} else {
    alert("安装失败: " + result.error)
}
```

`pkg.install()` 自动检测 apt/yum/dnf，返回是否安装成功、是否有变更、错误信息。幂等——已安装就跳过。

## 8. 网络检查

**Shell：**

```bash
#!/bin/bash
# 检查 HTTP 端点
STATUS=$(curl -s -o /dev/null -w "%{http_code}" https://api.example.com/health)
if [ "$STATUS" != "200" ]; then
    echo "API health check failed: $STATUS"
fi

# 检查 DNS
IP=$(dig +short api.example.com)
if [ -z "$IP" ]; then
    echo "DNS resolution failed"
fi

# 检查 TCP 端口
nc -z db.example.com 3306
if [ $? -ne 0 ]; then
    echo "Database port unreachable"
fi
```

问题：
- 三个工具：curl、dig、nc
- 输出格式各不相同
- 超时处理？重试？

**OpsLang：**

```ops
let http = net.http_get("https://api.example.com/health")
if http.status_code != 200 {
    alert("API 异常: status=" + str(http.status_code))
}

let dns = net.dns_lookup("api.example.com")
if len(dns.records) == 0 {
    alert("DNS 解析失败")
}

let tcp = net.tcp_check("db.example.com", 3306)
if tcp.reachable != true {
    alert("数据库不可达: " + tcp.error)
}

report {
    http: http,
    dns: dns,
    tcp: tcp
}
```

统一接口，结构化返回。`net.http_get()` 返回状态码、响应体、耗时。`net.tcp_check()` 返回是否可达、延迟、错误信息。

## 9. 配置渲染

**Shell：**

```bash
#!/bin/bash
# 模板渲染，用 envsubst 或 sed
export DB_HOST="db.example.com"
export DB_PORT="3306"
export APP_NAME="myapp"

envsubst < config.template > config.yaml
```

问题：
- 需要安装 `envsubst`
- 只能替换 `${VAR}` 格式
- 复杂逻辑（条件、循环）做不到

**OpsLang：**

```ops
let vars = {
    "db_host": "db.example.com",
    "db_port": "3306",
    "app_name": "myapp",
    "debug": false
}

let result = file.template("config.template", vars)

if result.success {
    file.write("config.yaml", result.content)
}
```

`file.template()` 支持 `{{key}}` 占位符，不修改源文件，返回渲染后的内容。可以在 vars 里放任意数据结构。

## 10. 多主机编排

**Shell：**

```bash
#!/bin/bash
HOSTS="host1 host2 host3"
SCRIPT="check.sh"

for host in $HOSTS; do
    scp "$SCRIPT" "$host:/tmp/"
    ssh "$host" "bash /tmp/check.sh" > "result_${host}.txt" &
done

wait
echo "All done"
```

问题：
- 并发控制？没有限制
- 结果汇总？要自己解析多个文件
- 错误处理？一个失败全部继续
- 架构不同？x86 和 ARM 不能跑同一个二进制

**OpsLang：**

```ops
task "check_all" on ["host1", "host2", "host3"] parallel {
    let cpu = sys.cpu.usage()
    let mem = sys.memory.info()

    report {
        host: sys.hostname(),
        cpu: cpu.percent,
        mem: mem.used_percent
    }
}
```

- 自动并行执行，信号量限流（默认 10 并发）
- 结果自动汇总为 JSON
- 自动检测目标架构，上传对应二进制
- 失败主机单独报告，不影响其他

输出：

```json
{
  "task_id": "abc123",
  "results": {
    "host1": { "status": "success", "data": { "cpu": 12.5, "mem": 68.3 } },
    "host2": { "status": "success", "data": { "cpu": 45.2, "mem": 72.1 } },
    "host3": { "status": "failed", "error": "timeout" }
  }
}
```

## 总结

| 场景 | Shell | OpsLang |
|------|-------|---------|
| 系统信息采集 | 字符串解析，不准确 | 结构化返回，精确 |
| 磁盘报警 | 字符串比较，易错 | 数值比较，清晰 |
| 文件部署 | 串行，无校验 | 并行，校验和，幂等 |
| 服务检查 | 只查状态，不查响应 | 完整状态，自动重启 |
| 日志收集 | 串行，无压缩 | 并行，压缩，断点续传 |
| 进程管理 | 依赖 lsof，信号处理粗糙 | 直接读 /proc，信号封装 |
| 包管理 | if-elif 判断包管理器 | 自动检测，幂等 |
| 网络检查 | 三个工具，格式各异 | 统一接口，结构化 |
| 配置渲染 | 需要 envsubst | 内置模板引擎 |
| 多主机编排 | 手动并发控制 | 自动并行，结果汇总 |

**Shell 不是不好，而是不适合做这些事。** 它适合交互式操作和胶水脚本，但不适合作为运维自动化的主力语言。

OpsLang 不是要替代 Shell 的所有用途。它解决的是：**结构化的、可预测的、大规模的运维自动化。**

试试看：https://github.com/j4ckzh0u/opslang
