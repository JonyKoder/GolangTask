package main

import (
	"fmt"
	"time"
)

type Task struct {
	ID         int
	CreatedAt  string
	FinishedAt string
	Result     []byte
}

func main() {
	taskCreator := func(tasks chan Task) {
		go func() {
			for {
				createdAt := time.Now().Format(time.RFC3339)
				if time.Now().Nanosecond()%2 > 0 {
					createdAt = "Some error occurred"
				}
				tasks <- Task{ID: int(time.Now().Unix()), CreatedAt: createdAt}
			}
		}()
	}

	superChan := make(chan Task, 10)
	go taskCreator(superChan)

	taskWorker := func(task Task) Task {
		createdAt, _ := time.Parse(time.RFC3339, task.CreatedAt)
		if time.Since(createdAt) < 20*time.Second {
			task.Result = []byte("Task has been succeeded")
		} else {
			task.Result = []byte("Something went wrong")
		}
		task.FinishedAt = time.Now().Format(time.RFC3339Nano)
		time.Sleep(time.Millisecond * 150)
		return task
	}

	doneTasks := make(chan Task)
	undoneTasks := make(chan Task)

	taskSorter := func(task Task) {
		if string(task.Result) == "Task has been succeeded" {
			doneTasks <- task
		} else {
			undoneTasks <- task
		}
	}

	go func() {
		for task := range superChan {
			go func(t Task) {
				t = taskWorker(t)
				taskSorter(t)
			}(task)
		}
		close(superChan)
	}()

	result := map[int]Task{}
	err := []error{}

	go func() {
		for r := range doneTasks {
			result[r.ID] = r
		}
	}()

	go func() {
		for r := range undoneTasks {
			err = append(err, fmt.Errorf("Task id %d time %s, error %s", r.ID, r.CreatedAt, r.Result))
		}
	}()

	time.Sleep(time.Second * 3)

	fmt.Println("Errors:")
	for _, e := range err {
		fmt.Println(e)
	}

	fmt.Println("Done tasks:")
	for id := range result {
		fmt.Println(id)
	}
}
