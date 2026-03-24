# ai-exe

一个使用 Go 编写的命令行 TODO List，并可编译为 Windows 可执行文件（`.exe`）。

## 功能

- `add <标题>`：新增任务
- `list`：查看任务列表
- `done <id>`：标记完成
- `undone <id>`：取消完成
- `delete <id>`：删除任务
- `clear-done`：清空已完成任务

任务默认保存在当前目录的 `todo.json` 中；也可以通过环境变量 `TODO_FILE` 指定路径。

## 本地运行

```bash
# 无参数启动交互模式（适合双击 exe 后使用）
go run .

go run . add "买牛奶"
go run . list
go run . done 1
go run . list
```

交互模式中输入 `help` 查看命令，输入 `exit` 或 `quit` 退出程序。

## 构建 Windows EXE

在 Linux/macOS 下交叉编译：

```bash
mkdir -p dist
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o dist/todo.exe .
```

生成文件：`dist/todo.exe`

## GitHub Release 自动发布

仓库已配置 GitHub Actions 工作流：`.github/workflows/release.yml`

- 触发条件：推送 `v*` 格式标签（例如 `v1.0.0`）
- 自动构建产物：
  - `todo-windows-amd64.exe`
  - `todo-linux-amd64`
  - `todo-linux-arm64`
  - `todo-darwin-amd64`
  - `todo-darwin-arm64`
- 自动创建 GitHub Release 并上传以上文件

示例发布命令：

```bash
git tag v1.0.0
git push origin v1.0.0
```
