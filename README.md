# ai-exe

一个使用 Go + Fyne 编写的桌面版 TODO List 应用。

## 功能

- 图形界面新增任务
- 勾选/取消任务完成状态
- 删除单个任务
- 一键清理已完成任务
- 显示任务统计（总数、已完成、未完成）

任务默认保存在当前目录的 `todo.json` 中；也可以通过环境变量 `TODO_FILE` 指定路径。

## 本地运行

```bash
go run .
```

程序会启动桌面窗口，直接在界面中进行任务管理。

## 本地构建（Linux）

Fyne 在 Linux 下依赖图形库开发包，首次构建前请安装：

```bash
sudo apt-get update
sudo apt-get install -y libgl1-mesa-dev xorg-dev
```

然后构建：

```bash
mkdir -p dist
go build -o dist/todo-linux-amd64 .
```

## GitHub Release 自动发布

仓库已配置 GitHub Actions 工作流：`.github/workflows/release.yml`

- 触发条件：推送 `v*` 格式标签（例如 `v1.0.0`）
- 自动构建产物：
  - `todo-linux-amd64`
- 自动创建 GitHub Release 并上传该文件和 `checksums.txt`

示例发布命令：

```bash
git tag v1.0.0
git push origin v1.0.0
```
