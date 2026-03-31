# cc-switch

一个用于在多个 Claude API 提供商之间快捷切换的 TUI 工具。

## 功能

- **交互式 TUI**：基于 Bubbletea 的终端界面，体验流畅
- **Profile 管理**：支持新增、编辑、删除配置
- **鼠标支持**：双击切换，单击选择，点击按钮操作
- **自动适配**：终端窗口大小变化时界面自动调整
- **安全存储**：profile 文件权限 0600，API 密钥安全保存
- **自动备份**：修改 settings.json 前自动备份，保留最近 10 个

## 安装

### 从源码构建

```bash
git clone https://github.com/xyustc/cc-switch.git
cd cc-switch
go build
```

### 从 GitHub Releases 下载

```bash
# macOS arm64 (Apple Silicon)
curl -sL https://github.com/xyustc/cc-switch/releases/latest/download/cc-switch_darwin_arm64.tar.gz | tar xz

# macOS amd64 (Intel)
curl -sL https://github.com/xyustc/cc-switch/releases/latest/download/cc-switch_darwin_amd64.tar.gz | tar xz

# Linux amd64
curl -sL https://github.com/xyustc/cc-switch/releases/latest/download/cc-switch_linux_amd64.tar.gz | tar xz
```

将 `cc-switch` 移动到 PATH 中：

```bash
mv cc-switch /usr/local/bin/
```

## 使用

直接运行：

```bash
cc-switch
```

查看版本：

```bash
cc-switch --version
```

### 界面操作

| 操作 | 效果 |
|------|------|
| `↑/↓` | 导航选择 profile |
| `Enter` 或双击 | 切换到选中的 profile |
| `n` | 新增 profile |
| `e` | 编辑选中的 profile |
| `d` | 删除选中的 profile |
| `q` | 退出 |

### Profile 配置示例

Profile 中的 `settings` 字段采用 JSON 格式，定义的字段会覆盖 `~/.claude/settings.json` 中的对应顶级字段。

```json
{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "sk-ant-..."
  }
}
```

或者配置多个环境变量：

```json
{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "sk-ant-...",
    "ANTHROPIC_BASE_URL": "https://api.anthropic.com"
  }
}
```

### 数据存储位置

- **Profiles**：`~/.cc-switch/profiles.json`（权限 0600）
- **Claude Settings**：`~/.claude/settings.json`
- **备份**：`~/.claude/settings.json.bak.1` ~ `.bak.10`

## 工作原理

切换 profile 时，cc-switch 会：

1. 读取 profile 中定义的 `settings` JSON
2. 对 `~/.claude/settings.json` 执行**顶级字段替换**（不是深合并）
3. profile 中定义的字段覆盖原值，未定义的字段保留原样
4. 写入前自动备份原文件
5. 更新 `~/.cc-switch/profiles.json` 的 `active` 字段

## 开发

### 依赖

- Go 1.26+
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI 框架
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI 组件
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - 样式库

### 项目结构

```
cc-switch/
├── main.go           # 入口
├── config/
│   ├── profile.go    # Profile CRUD
│   └── settings.go   # Settings 读写、备份
└── ui/
    ├── app.go        # 顶层 Model
    ├── list.go       # Profile 列表视图
    ├── form.go       # 新增/编辑表单
    └── confirm.go    # 删除确认对话框
```

### 发布新版本

项目使用 goreleaser 进行自动化发布。创建新版本 tag：

```bash
git tag v0.X.Y
git push origin v0.X.Y
```

GitHub Actions 会自动构建并发布到 Releases。

## 文档

- [CHANGELOG.md](CHANGELOG.md) — 版本更新记录

## License

MIT