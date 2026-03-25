package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Task struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Done        bool       `json:"done"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type Store struct {
	NextID int    `json:"next_id"`
	Tasks  []Task `json:"tasks"`
}

func main() {
	dataFile := dataFilePath()
	store, err := loadStore(dataFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	runGUI(dataFile, &store)
}

func runGUI(dataFile string, store *Store) {
	a := app.NewWithID("ai-exe.todo")
	w := a.NewWindow("TODO List")
	w.Resize(fyne.NewSize(860, 580))

	statusLabel := widget.NewLabel("就绪")
	statusLabel.Wrapping = fyne.TextWrapWord
	summaryLabel := widget.NewLabel("")
	summaryLabel.TextStyle = fyne.TextStyle{Italic: true}

	input := widget.NewEntry()
	input.SetPlaceHolder("输入任务标题，例如：买牛奶")
	input.Wrapping = fyne.TextWrapWord

	var list *widget.List

	updateSummary := func() {
		total := len(store.Tasks)
		done := 0
		for _, task := range store.Tasks {
			if task.Done {
				done++
			}
		}
		summaryLabel.SetText(fmt.Sprintf("总任务: %d   已完成: %d   未完成: %d", total, done, total-done))
	}

	saveAndRefresh := func(successMessage string) bool {
		if err := saveStore(dataFile, *store); err != nil {
			dialog.ShowError(err, w)
			return false
		}
		sortTasks(store.Tasks)
		list.Refresh()
		updateSummary()
		statusLabel.SetText(successMessage)
		return true
	}

	addTask := func() {
		title := strings.TrimSpace(input.Text)
		if title == "" {
			statusLabel.SetText("任务标题不能为空。")
			return
		}
		task := Task{
			ID:        store.NextID,
			Title:     title,
			Done:      false,
			CreatedAt: time.Now(),
		}
		store.NextID++
		store.Tasks = append(store.Tasks, task)
		if saveAndRefresh(fmt.Sprintf("已添加任务 #%d。", task.ID)) {
			input.SetText("")
			w.Canvas().Focus(input)
		}
	}

	input.OnSubmitted = func(string) {
		addTask()
	}

	list = widget.NewList(
		func() int {
			return len(store.Tasks)
		},
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			title := widget.NewLabel(" ")
			title.Wrapping = fyne.TextWrapWord
			deleteBtn := widget.NewButton("删除", nil)
			rowTop := container.NewHBox(check, title, layout.NewSpacer(), deleteBtn)

			meta := widget.NewLabel(" ")
			meta.Wrapping = fyne.TextWrapWord
			meta.TextStyle = fyne.TextStyle{Italic: true}
			return container.NewVBox(rowTop, meta, widget.NewSeparator())
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(store.Tasks) {
				return
			}
			task := store.Tasks[id]
			box := obj.(*fyne.Container)

			rowTop := box.Objects[0].(*fyne.Container)
			check := rowTop.Objects[0].(*widget.Check)
			title := rowTop.Objects[1].(*widget.Label)
			deleteBtn := rowTop.Objects[3].(*widget.Button)
			meta := box.Objects[1].(*widget.Label)

			check.OnChanged = nil
			check.SetChecked(task.Done)
			check.OnChanged = func(done bool) {
				updated, err := markTask(store.Tasks, task.ID, done)
				if err != nil {
					dialog.ShowError(err, w)
					list.Refresh()
					return
				}
				store.Tasks = updated
				if !saveAndRefresh(fmt.Sprintf("已更新任务 #%d。", task.ID)) {
					list.Refresh()
				}
			}

			state := "TODO"
			if task.Done {
				state = "DONE"
			}
			title.SetText(fmt.Sprintf("#%d [%s] %s", task.ID, state, task.Title))

			completedAt := "-"
			if task.CompletedAt != nil {
				completedAt = task.CompletedAt.Local().Format("2006-01-02 15:04")
			}
			meta.SetText(fmt.Sprintf("创建: %s    完成: %s", task.CreatedAt.Local().Format("2006-01-02 15:04"), completedAt))

			deleteBtn.OnTapped = func() {
				dialog.NewConfirm(
					"删除任务",
					fmt.Sprintf("确认删除任务 #%d 吗？", task.ID),
					func(confirm bool) {
						if !confirm {
							return
						}
						updated, deleted := deleteTask(store.Tasks, task.ID)
						if !deleted {
							dialog.ShowError(fmt.Errorf("task #%d not found", task.ID), w)
							return
						}
						store.Tasks = updated
						saveAndRefresh(fmt.Sprintf("已删除任务 #%d。", task.ID))
					},
					w,
				).Show()
			}
		},
	)

	addButton := widget.NewButton("新增任务", addTask)
	clearDoneButton := widget.NewButton("清理已完成", func() {
		dialog.NewConfirm("清理已完成", "确认删除全部已完成任务吗？", func(confirm bool) {
			if !confirm {
				return
			}
			pending := make([]Task, 0, len(store.Tasks))
			removed := 0
			for _, task := range store.Tasks {
				if task.Done {
					removed++
					continue
				}
				pending = append(pending, task)
			}
			if removed == 0 {
				statusLabel.SetText("没有已完成任务可清理。")
				return
			}
			store.Tasks = pending
			saveAndRefresh(fmt.Sprintf("已清理 %d 个已完成任务。", removed))
		}, w).Show()
	})

	header := container.NewVBox(
		widget.NewLabelWithStyle("TODO List (Fyne)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		summaryLabel,
		container.NewBorder(nil, nil, nil, addButton, input),
		container.NewHBox(clearDoneButton, layout.NewSpacer(), statusLabel),
		widget.NewSeparator(),
	)

	mainContent := container.NewBorder(header, nil, nil, nil, list)
	w.SetContent(mainContent)

	sortTasks(store.Tasks)
	updateSummary()
	w.ShowAndRun()
}

func dataFilePath() string {
	if custom := strings.TrimSpace(os.Getenv("TODO_FILE")); custom != "" {
		return custom
	}
	return "todo.json"
}

func loadStore(path string) (Store, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Store{NextID: 1}, nil
		}
		return Store{}, fmt.Errorf("read data file: %w", err)
	}

	var store Store
	if err := json.Unmarshal(file, &store); err != nil {
		return Store{}, fmt.Errorf("parse data file: %w", err)
	}
	if store.NextID < 1 {
		store.NextID = nextID(store.Tasks)
	}
	return store, nil
}

func saveStore(path string, store Store) error {
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create data directory: %w", err)
		}
	}

	content, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("serialize data file: %w", err)
	}
	content = append(content, '\n')

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, content, 0o644); err != nil {
		return fmt.Errorf("write temp data file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace data file: %w", err)
	}
	return nil
}

func markTask(tasks []Task, id int, done bool) ([]Task, error) {
	for i := range tasks {
		if tasks[i].ID != id {
			continue
		}
		tasks[i].Done = done
		if done {
			now := time.Now()
			tasks[i].CompletedAt = &now
		} else {
			tasks[i].CompletedAt = nil
		}
		return tasks, nil
	}
	return nil, fmt.Errorf("task #%d not found", id)
}

func deleteTask(tasks []Task, id int) ([]Task, bool) {
	for i := range tasks {
		if tasks[i].ID == id {
			return append(tasks[:i], tasks[i+1:]...), true
		}
	}
	return tasks, false
}

func nextID(tasks []Task) int {
	maxID := 0
	for _, task := range tasks {
		if task.ID > maxID {
			maxID = task.ID
		}
	}
	return maxID + 1
}

func sortTasks(tasks []Task) {
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
}
