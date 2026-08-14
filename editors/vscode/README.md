# OpsLang for VS Code

OpsLang 语言支持扩展，提供智能编辑体验。

## 功能

- **语法高亮** — 关键字、字符串、数字、注释、插值变量
- **实时诊断** — 语法错误即时提示
- **智能补全** — 关键字、内置函数、标准库模块
- **悬停信息** — 函数文档与模块说明
- **跳转定义** — 跳转到函数定义位置
- **代码片段** — 常用结构快速插入

## 安装

1. 安装 OpsLang CLI：
   ```bash
   git clone https://github.com/opslang/opslang.git
   cd opslang && make install
   ```

2. 安装此扩展：
   - 在 VS Code 中搜索 "OpsLang"
   - 或从 `.vsix` 文件安装

3. 确保 `ops` 命令在 PATH 中：
   ```bash
   ops version
   ```

## 配置

| 设置 | 默认值 | 说明 |
|------|--------|------|
| `opslang.serverPath` | `ops` | OpsLang 可执行文件路径 |
| `opslang.enableLSP` | `true` | 启用 LSP 语言服务器 |

如果 `ops` 不在 PATH 中，可以配置绝对路径：

```json
{
  "opslang.serverPath": "/usr/local/bin/ops"
}
```

## 命令

- `OpsLang: Restart Server` — 重启语言服务器

## 截图

### 智能补全
![补全示例](docs/completion.png)

### 悬停信息
![悬停示例](docs/hover.png)

### 错误诊断
![诊断示例](docs/diagnostics.png)

## 开发

```bash
cd editors/vscode
npm install
npm run compile
```

按 F5 在扩展开发主机中运行。

## 许可证

MIT
