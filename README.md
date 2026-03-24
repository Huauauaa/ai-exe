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
go run . add "买牛奶"
go run . list
go run . done 1
go run . list
```

## 构建 Windows EXE

在 Linux/macOS 下交叉编译：

```bash
mkdir -p dist
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o dist/todo.exe .
```

生成文件：`dist/todo.exe`
