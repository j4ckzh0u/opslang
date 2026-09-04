<title>OpsLang 实现状态报告</title>

<style>
  :root {
    --bg: #fafaf9;
    --bg-raised: #f5f5f4;
    --bg-inset: #ececea;
    --border: #d6d3d1;
    --text: #1c1917;
    --text-secondary: #57534e;
    --text-muted: #a8a29e;
    --green: #16a34a;
    --green-bg: #f0fdf4;
    --amber: #d97706;
    --amber-bg: #fffbeb;
    --red: #dc2626;
    --red-bg: #fef2f2;
    --blue: #2563eb;
    --blue-bg: #eff6ff;
    --mono: 'SF Mono', 'Cascadia Code', 'Fira Code', 'JetBrains Mono', ui-monospace, monospace;
    --sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
  }

  @media (prefers-color-scheme: dark) {
    :root:not([data-theme="light"]) {
      --bg: #0c0a09;
      --bg-raised: #1c1917;
      --bg-inset: #292524;
      --border: #44403c;
      --text: #faf5ef;
      --text-secondary: #a8a29e;
      --text-muted: #78716c;
      --green: #4ade80;
      --green-bg: #052e16;
      --amber: #fbbf24;
      --amber-bg: #451a03;
      --red: #f87171;
      --red-bg: #450a0a;
      --blue: #60a5fa;
      --blue-bg: #172554;
    }
  }

  :root[data-theme="dark"] {
    --bg: #0c0a09;
    --bg-raised: #1c1917;
    --bg-inset: #292524;
    --border: #44403c;
    --text: #faf5ef;
    --text-secondary: #a8a29e;
    --text-muted: #78716c;
    --green: #4ade80;
    --green-bg: #052e16;
    --amber: #fbbf24;
    --amber-bg: #451a03;
    --red: #f87171;
    --red-bg: #450a0a;
    --blue: #60a5fa;
    --blue-bg: #172554;
  }

  * { margin: 0; padding: 0; box-sizing: border-box; }

  body {
    background: var(--bg);
    color: var(--text);
    font-family: var(--sans);
    font-size: 15px;
    line-height: 1.65;
    -webkit-font-smoothing: antialiased;
  }

  .page {
    max-width: 820px;
    margin: 0 auto;
    padding: 3rem 1.5rem 4rem;
  }

  /* Header */
  .header {
    border-bottom: 2px solid var(--text);
    padding-bottom: 1.5rem;
    margin-bottom: 2.5rem;
  }

  .header-meta {
    font-family: var(--mono);
    font-size: 12px;
    color: var(--text-muted);
    letter-spacing: 0.05em;
    text-transform: uppercase;
    margin-bottom: 0.5rem;
  }

  h1 {
    font-size: 28px;
    font-weight: 700;
    letter-spacing: -0.02em;
    line-height: 1.2;
  }

  .subtitle {
    color: var(--text-secondary);
    font-size: 15px;
    margin-top: 0.5rem;
  }

  /* Summary strip */
  .summary {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: 1px;
    background: var(--border);
    border: 1px solid var(--border);
    margin-bottom: 2.5rem;
  }

  .summary-cell {
    background: var(--bg-raised);
    padding: 1rem 1.25rem;
    text-align: center;
  }

  .summary-value {
    font-family: var(--mono);
    font-size: 28px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    color: var(--green);
    line-height: 1;
  }

  .summary-value.warn { color: var(--amber); }

  .summary-label {
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--text-muted);
    margin-top: 0.35rem;
  }

  /* Sections */
  section {
    margin-bottom: 2.5rem;
  }

  h2 {
    font-size: 18px;
    font-weight: 600;
    letter-spacing: -0.01em;
    margin-bottom: 1rem;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid var(--border);
  }

  h3 {
    font-size: 14px;
    font-weight: 600;
    margin-bottom: 0.5rem;
    color: var(--text);
  }

  p {
    margin-bottom: 0.75rem;
    color: var(--text-secondary);
  }

  /* Status matrix */
  .matrix {
    width: 100%;
    border-collapse: collapse;
    font-size: 13.5px;
    margin-bottom: 1rem;
  }

  .matrix th {
    text-align: left;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
    padding: 0.5rem 0.75rem;
    border-bottom: 2px solid var(--border);
    font-weight: 500;
  }

  .matrix td {
    padding: 0.6rem 0.75rem;
    border-bottom: 1px solid var(--border);
    vertical-align: top;
  }

  .matrix tr:last-child td { border-bottom: none; }

  .matrix .name {
    font-weight: 500;
    color: var(--text);
    white-space: nowrap;
  }

  .badge {
    display: inline-block;
    font-family: var(--mono);
    font-size: 11px;
    font-weight: 500;
    padding: 0.15rem 0.5rem;
    border-radius: 3px;
    letter-spacing: 0.02em;
    white-space: nowrap;
  }

  .badge-green { background: var(--green-bg); color: var(--green); }
  .badge-amber { background: var(--amber-bg); color: var(--amber); }
  .badge-red   { background: var(--red-bg);   color: var(--red); }
  .badge-blue  { background: var(--blue-bg);  color: var(--blue); }

  /* Detail cards */
  .card {
    background: var(--bg-raised);
    border: 1px solid var(--border);
    padding: 1.25rem;
    margin-bottom: 1rem;
  }

  .card-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.75rem;
  }

  .card-header h3 { margin-bottom: 0; }

  .card-stats {
    font-family: var(--mono);
    font-size: 11px;
    color: var(--text-muted);
  }

  .card ul {
    padding-left: 1.25rem;
    color: var(--text-secondary);
    font-size: 13.5px;
  }

  .card li { margin-bottom: 0.3rem; }

  .card li::marker { color: var(--text-muted); }

  /* Code */
  code {
    font-family: var(--mono);
    font-size: 12.5px;
    background: var(--bg-inset);
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
  }

  /* Limitations list */
  .limitations {
    list-style: none;
    padding: 0;
  }

  .limitations li {
    padding: 0.6rem 0;
    padding-left: 1.5rem;
    position: relative;
    border-bottom: 1px solid var(--border);
    font-size: 13.5px;
    color: var(--text-secondary);
  }

  .limitations li:last-child { border-bottom: none; }

  .limitations li::before {
    content: "—";
    position: absolute;
    left: 0;
    color: var(--amber);
    font-weight: 600;
  }

  /* Verdict */
  .verdict {
    background: var(--green-bg);
    border: 1px solid var(--green);
    padding: 1.25rem 1.5rem;
    margin-top: 2rem;
  }

  .verdict h3 {
    color: var(--green);
    margin-bottom: 0.5rem;
  }

  .verdict p {
    color: var(--text);
    margin-bottom: 0;
  }

  /* Footer */
  .footer {
    margin-top: 3rem;
    padding-top: 1rem;
    border-top: 1px solid var(--border);
    font-size: 12px;
    color: var(--text-muted);
    font-family: var(--mono);
  }

  @media (max-width: 600px) {
    .page { padding: 2rem 1rem; }
    h1 { font-size: 22px; }
    .summary { grid-template-columns: repeat(2, 1fr); }
    .matrix { font-size: 12px; }
    .matrix th, .matrix td { padding: 0.4rem 0.5rem; }
  }
</style>

<div class="page">

  <div class="header">
    <div class="header-meta">Technical Audit · 2026-09-03</div>
    <h1>OpsLang 实现状态报告</h1>
    <p class="subtitle">实事求是：当前代码库的实际状态，包括已实现、可用、和已知限制。</p>
  </div>

  <div class="summary">
    <div class="summary-cell">
      <div class="summary-value">2107</div>
      <div class="summary-label">测试函数</div>
    </div>
    <div class="summary-cell">
      <div class="summary-value">210</div>
      <div class="summary-label">SDK 操作包</div>
    </div>
    <div class="summary-cell">
      <div class="summary-value">1157</div>
      <div class="summary-label">原子操作</div>
    </div>
    <div class="summary-cell">
      <div class="summary-value">119</div>
      <div class="summary-label">示例脚本</div>
    </div>
    <div class="summary-cell">
      <div class="summary-value">8</div>
      <div class="summary-label">跨平台二进制</div>
    </div>
  </div>

  <!-- Status Matrix -->
  <section>
    <h2>子系统状态总览</h2>
    <table class="matrix">
      <thead>
        <tr>
          <th>子系统</th>
          <th>状态</th>
          <th>测试数</th>
          <th>备注</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td class="name">ops-core-sdk</td>
          <td><span class="badge badge-green">可用</span></td>
          <td>全量通过</td>
          <td>210 个操作包，1157 个注册操作</td>
        </tr>
        <tr>
          <td class="name">词法/语法分析</td>
          <td><span class="badge badge-green">可用</span></td>
          <td>全量通过</td>
          <td>完整 lexer + recursive descent parser</td>
        </tr>
        <tr>
          <td class="name">解释器</td>
          <td><span class="badge badge-green">可用</span></td>
          <td>全量通过</td>
          <td>树遍历执行，调用 SDK，支持闭包/dry-run</td>
        </tr>
        <tr>
          <td class="name">AOT 编译器</td>
          <td><span class="badge badge-green">可用</span></td>
          <td>全量通过</td>
          <td>AST → Go 源码 → 静态二进制</td>
        </tr>
        <tr>
          <td class="name">SSH 客户端</td>
          <td><span class="badge badge-green">可用</span></td>
          <td>全量通过</td>
          <td>密码/密钥认证，TOFU，SFTP，连接池</td>
        </tr>
        <tr>
          <td class="name">远程执行器</td>
          <td><span class="badge badge-green">可用</span></td>
          <td>全量通过</td>
          <td>架构检测，内容寻址缓存，并发执行</td>
        </tr>
        <tr>
          <td class="name">Runner 模式</td>
          <td><span class="badge badge-green">可用</span></td>
          <td>全量通过</td>
          <td>JSON 指令包，与 opsspec 注册表一致</td>
        </tr>
        <tr>
          <td class="name">安全模块</td>
          <td><span class="badge badge-green">可用</span></td>
          <td>全量通过</td>
          <td>权限分级，审批流，审计日志，资源限制，Ed25519 签名</td>
        </tr>
        <tr>
          <td class="name">CLI</td>
          <td><span class="badge badge-green">可用</span></td>
          <td>全量通过</td>
          <td>run / build / exec / deploy / repl</td>
        </tr>
        <tr>
          <td class="name">文件分发/收集</td>
          <td><span class="badge badge-green">可用</span></td>
          <td>全量通过</td>
          <td>真实 SSH/SFTP、并发、重试、传输后 SHA-256 校验</td>
        </tr>
      </tbody>
    </table>
  </section>

  <!-- Detailed Breakdown -->
  <section>
    <h2>各模块详情</h2>

    <div class="card">
      <div class="card-header">
        <h3>ops-core-sdk（原子操作标准库）</h3>
        <span class="card-stats">210 operation packages · 1157 operations</span>
      </div>
      <ul>
        <li><code>sys</code> — CPU 使用率（500ms 采样）、内存、磁盘、负载、主机名、用户、网络接口</li>
        <li><code>file</code> — 读/写/追加/复制/移动/删除/权限/校验和/模板渲染/目录列举</li>
        <li><code>net</code> — HTTP GET/POST、TCP 连通性、DNS 解析、网络接口</li>
        <li><code>process</code> — 进程列表/查找/执行外部命令（不经 shell）/发送信号</li>
        <li><code>service</code> — systemd 服务管理（status/start/stop/restart/enable/disable）</li>
        <li><code>pkg</code> — 包管理（apt/yum/dnf 自动检测）</li>
        <li><code>json</code> / <code>yaml</code> — 编解码</li>
        <li><code>time</code> — 时间获取/格式化/解析/差值/睡眠</li>
      </ul>
    </div>

    <div class="card">
      <div class="card-header">
        <h3>语言前端</h3>
        <span class="card-stats">lexer 546 行 · parser 1176 行 · AST 30+ 节点</span>
      </div>
      <ul>
        <li>20 个关键字：<code>let fn if else for while return task on import privilege report alert ensure metric log parallel</code> 等</li>
        <li>数据类型：整数、浮点数、字符串、布尔、列表、字典</li>
        <li>支持：闭包、默认参数、C-style for、for-in、while、block/rescue/always、文件模块、成员访问、索引访问</li>
        <li>119 个示例脚本纳入仓库，CLI 端到端测试执行示例集</li>
      </ul>
    </div>

    <div class="card">
      <div class="card-header">
        <h3>双执行引擎</h3>
        <span class="card-stats">解释器 1376 行 · codegen 1234 行</span>
      </div>
      <ul>
        <li><strong>Runner 模式</strong>：线性脚本 → JSON 指令包 → 通用 Runner 执行。零编译延迟。</li>
        <li><strong>AOT 模式</strong>：AST → Go 源码 → <code>go build</code> → 静态二进制。支持交叉编译。</li>
        <li><strong>自动选择</strong>：<code>RequiresAOT()</code> 检测 if/for/while/fn/ensure/parallel 决定模式</li>
        <li><strong>编译缓存</strong>：SHA256 内容寻址，相同脚本秒级编译</li>
        <li>两个引擎的 SDK 映射通过一致性测试强制对齐</li>
      </ul>
    </div>

    <div class="card">
      <div class="card-header">
        <h3>远程执行</h3>
        <span class="card-stats">exec 773 行 · sshx 245 行</span>
      </div>
      <ul>
        <li>SSH 连接 → 架构自动检测（<code>uname -m</code>）→ 上传 Runner/AOT 二进制 → 执行 → JSON 结果回收</li>
        <li><strong>内容寻址远程缓存</strong>：<code>ensureRemoteBinary()</code> 用 sha256 命名，缓存命中只传 ~100 bytes 校验和</li>
        <li>TOFU 主机密钥验证（<code>~/.ssh/opslang_known_hosts</code>）</li>
        <li>并发执行，信号量限流，指数退避重试</li>
        <li>上传后校验和验证，防止截断</li>
      </ul>
    </div>

    <div class="card">
      <div class="card-header">
        <h3>安全模块</h3>
        <span class="card-stats">全量测试覆盖</span>
      </div>
      <ul>
        <li><strong>权限分级</strong>：read_only / admin / root，操作分类（read/write/exec/admin/system），变更类函数元数据以 <code>opsspec</code> 为单一事实来源</li>
        <li><strong>权限自动执行</strong>：read_only 脚本调用变更函数在三层被拒绝——解释器（运行时，带行列号）、AOT 编译期静态检查、Runner 二次校验（指令包携带 privilege 字段）</li>
        <li><strong>审批流</strong>：<code>privilege: admin/root</code> 脚本部署到生产目标（inventory 标签 <code>env: prod/production</code>）前强制审批——TTY 展示摘要（权限、变更操作、生产目标）后 y/N 确认；非 TTY（管道/CI）默认拒绝，需 <code>--auto-approve</code> 或 <code>OPSCTL_AUTO_APPROVE=1</code>（flag 优先）放行；拒绝即中止，不联系任何主机；决策逻辑独立于交互（<code>internal/security/approval.go</code>），<code>opsctl deploy</code> 与 <code>opsctl exec</code> 均已接入</li>
        <li><strong>审计日志</strong>：JSON 格式，记录任务 ID、脚本、权限、目标、用户、模式、结果；审批决策（批准/拒绝、来源、批准人、生产目标清单）随运行记录一同落盘，可回溯</li>
        <li><strong>资源限制</strong>：远程 Runner 在目标机具备 <code>systemd-run</code> 时通过 transient scope 强制 CPU/内存限制；缺少 systemd-run 时结果携带 warning</li>
        <li><strong>签名验证</strong>：Ed25519 签名/验签，密钥文件 I/O</li>
        <li><strong>临时目录</strong>：自动创建/清理，幂等</li>
      </ul>
    </div>

    <div class="card">
      <div class="card-header">
        <h3>文件分发与收集</h3>
        <span class="card-stats">真实 SSH/SFTP · 含万级模拟测试</span>
      </div>
      <ul>
        <li><code>file.distribute()</code>：多主机并行 SFTP 分发、内容哈希跳过、部分文件续传、原子替换与传输后 SHA-256 校验</li>
        <li><code>file.collect()</code>：多主机并行收集、按主机归档、内容哈希跳过与部分文件续传</li>
        <li>分层中继分发：按显式组或 IP 前缀生成稳定拓扑，使用短时令牌、TLS 指纹固定和 HTTPS Range 扇出；候选失败后切换，未完成目标回退直接 SFTP</li>
        <li>通过 <code>WireSSHTransfer()</code> 注入真实 SSH 实现（opsctl 启动时调用）</li>
        <li>支持 <code>OPSLANG_SSH_PASSWORD_&lt;HOST&gt;</code> 环境变量设置主机密码</li>
      </ul>
    </div>
  </section>

  <!-- Limitations -->
  <section>
    <h2>已知限制</h2>
    <ul class="limitations">
      <li><strong>第三方 Go 导入</strong>：文件模块 <code>import "./lib.ops"</code> 已实现；<code>import "go &lt;包路径&gt;"</code> 仍被显式拒绝。</li>
      <li><strong>资源限制平台依赖</strong>：CPU/内存限制要求远端提供 <code>systemd-run</code>；缺少该命令时任务继续执行并返回 warning。</li>
      <li><strong>文件传输优化</strong>：分发/收集支持 gzip 传输与断点续传，分发支持 gzip 中继复用；分层中继收集仍在 Roadmap。</li>
      <li><strong>自动回滚接入</strong>：回滚 helper 已有测试，部署执行链尚未调用。</li>
      <li><strong>CPU 使用率为采样值</strong>：<code>sys.cpu.usage()</code> 两次采样间隔 500ms，非实时值。</li>
      <li><strong>CI 未启用竞态检测</strong>：<code>-race</code> 在 CI 全量测试时 TSan OOM，已在 CI 配置中移除。本地应定期跑 <code>go test -race ./...</code>。</li>
      <li><strong>大规模模拟测试的传输层为模拟</strong>：1 万主机分发/收集及中继分发模拟测试已实现（<code>pkg/ops-core-sdk/file/scale_test.go</code>、<code>relay_scale_test.go</code>）。测试注入 0.1% 确定性故障、哈希损坏、中继候选失效、恢复偏移异常和确认块损坏，验证结果守恒、成功率 &gt;99.9%、重试上界及中继流量上界。CI 使用 1000 台门档，<code>make scale-test</code> 跑 1 万档；真实协议另由受控 SSH/SFTP 与本地 HTTPS 端到端测试覆盖。</li>
      <li><strong>macOS 兼容性</strong>：<code>service</code>、<code>pkg</code> 包仅支持 Linux（systemd / apt / yum）。</li>
    </ul>
  </section>

  <!-- Phase Status -->
  <section>
    <h2>开发阶段进度</h2>
    <table class="matrix">
      <thead>
        <tr>
          <th>阶段</th>
          <th>内容</th>
          <th>状态</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td class="name">Phase 0</td>
          <td>原子操作 SDK</td>
          <td><span class="badge badge-green">完成</span></td>
        </tr>
        <tr>
          <td class="name">Phase 1</td>
          <td>远程执行通道</td>
          <td><span class="badge badge-green">完成</span></td>
        </tr>
        <tr>
          <td class="name">Phase 2</td>
          <td>语言前端与解释器</td>
          <td><span class="badge badge-green">完成</span></td>
        </tr>
        <tr>
          <td class="name">Phase 3</td>
          <td>AOT 编译管线</td>
          <td><span class="badge badge-green">完成</span></td>
        </tr>
        <tr>
          <td class="name">Phase 4</td>
          <td>远程编排与声明式</td>
          <td><span class="badge badge-amber">部分完成</span></td>
        </tr>
        <tr>
          <td class="name">Phase 5</td>
          <td>安全与生产化</td>
          <td><span class="badge badge-amber">部分完成</span></td>
        </tr>
      </tbody>
    </table>
  </section>

  <!-- Verdict -->
  <div class="verdict">
    <h3>结论</h3>
    <p>
      这是一个<strong>完整、可工作的 MVP</strong>。当前 2107 个测试函数通过，119 个示例脚本纳入仓库，双执行引擎（Runner + AOT）均可用，远程执行链路（SSH → 架构检测 → 缓存上传 → 执行 → 结果回收）已打通。
    </p>
    <p style="margin-top: 0.75rem;">
      主要缺口：第三方 Go 导入、动态 task 目标、分层中继收集、非 systemd 资源限制回退、部署自动回滚接入和 CI 竞态检测。文件分发/收集已支持断点续传与 gzip，文件分发已支持分层中继。权限自动执行已实现（解释器运行时 + AOT 编译期 + Runner 二次校验三层强制）；审批流已接入 deploy/exec；1 万主机模拟覆盖调度、恢复、重试、校验、归档、中继故障与流量上界。
    </p>
  </div>

  <div class="footer">
    Updated 2026-09-03 · 2107 test functions · 210 SDK operation packages · 1157 operations
  </div>

</div>
