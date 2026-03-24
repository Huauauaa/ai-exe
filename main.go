package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	dataFile := dataFilePath()
	store, err := loadStore(dataFile)
	if err != nil {
		return err
	}

	switch args[0] {
	case "add":
		if len(args) < 2 {
			return errors.New("add command requires a task title")
		}
		title := strings.TrimSpace(strings.Join(args[1:], " "))
		if title == "" {
			return errors.New("task title cannot be empty")
		}
		task := Task{
			ID:        store.NextID,
			Title:     title,
			Done:      false,
			CreatedAt: time.Now(),
		}
		store.NextID++
		store.Tasks = append(store.Tasks, task)
		if err := saveStore(dataFile, store); err != nil {
			return err
		}
		fmt.Printf("Added task #%d: %s\n", task.ID, task.Title)
		return nil

	case "list":
		if len(store.Tasks) == 0 {
			fmt.Println("No tasks found.")
			return nil
		}
		printTasks(store.Tasks)
		return nil

	case "done":
		id, err := parseIDArg(args)
		if err != nil {
			return err
		}
		updated, err := markTask(store.Tasks, id, true)
		if err != nil {
			return err
		}
		store.Tasks = updated
		if err := saveStore(dataFile, store); err != nil {
			return err
		}
		fmt.Printf("Marked task #%d as done.\n", id)
		return nil

	case "undone":
		id, err := parseIDArg(args)
		if err != nil {
			return err
		}
		updated, err := markTask(store.Tasks, id, false)
		if err != nil {
			return err
		}
		store.Tasks = updated
		if err := saveStore(dataFile, store); err != nil {
			return err
		}
		fmt.Printf("Marked task #%d as not done.\n", id)
		return nil

	case "delete":
		id, err := parseIDArg(args)
		if err != nil {
			return err
		}
		updated, deleted := deleteTask(store.Tasks, id)
		if !deleted {
			return fmt.Errorf("task #%d not found", id)
		}
		store.Tasks = updated
		if err := saveStore(dataFile, store); err != nil {
			return err
		}
		fmt.Printf("Deleted task #%d.\n", id)
		return nil

	case "clear-done":
		var pending []Task
		for _, task := range store.Tasks {
			if !task.Done {
				pending = append(pending, task)
			}
		}
		store.Tasks = pending
		if err := saveStore(dataFile, store); err != nil {
			return err
		}
		fmt.Println("Removed all completed tasks.")
		return nil

	case "help", "-h", "--help":
		printUsage()
		return nil

	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", args[0])
	}
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

func parseIDArg(args []string) (int, error) {
	if len(args) < 2 {
		return 0, errors.New("command requires task ID")
	}
	id, err := strconv.Atoi(args[1])
	if err != nil || id <= 0 {
		return 0, errors.New("task ID must be a positive integer")
	}
	return id, nil
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

func printTasks(tasks []Task) {
	fmt.Printf("%-5s %-8s %-50s\n", "ID", "STATUS", "TITLE")
	for _, task := range tasks {
		status := "TODO"
		if task.Done {
			status = "DONE"
		}
		fmt.Printf("%-5d %-8s %-50s\n", task.ID, status, task.Title)
	}
}

func printUsage() {
	fmt.Println(`TODO List CLI

Usage:
  todo add <task title>        Add a new task
  todo list                    List all tasks
  todo done <id>               Mark a task as done
  todo undone <id>             Mark a task as not done
  todo delete <id>             Delete a task
  todo clear-done              Delete all completed tasks
  todo help                    Show this help

Data file:
  Default: ./todo.json
  Custom: set TODO_FILE environment variable`)
}
