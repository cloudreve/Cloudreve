package queue

import "sync"

type (
	// TaskRegistry is used in slave node to track in-memory stateful tasks.
	TaskRegistry interface {
		// NextID returns the next available Task ID.
		NextID() int
		// Get returns the Task by ID.
		Get(id int) (Task, bool)
		// Set sets the Task by ID.
		Set(id int, t Task)
		// Delete deletes the Task by ID.
		Delete(id int)
	}

	taskRegistry struct {
		tasks   map[int]Task
		current int
		mu      sync.Mutex
	}
)

// NewTaskRegistry creates a new TaskRegistry.
func NewTaskRegistry() TaskRegistry {
	return &taskRegistry{
		tasks: make(map[int]Task),
	}
}

func (r *taskRegistry) NextID() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.current++
	return r.current
}

func (r *taskRegistry) Get(id int) (Task, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tasks[id]
	return t, ok
}

func (r *taskRegistry) Set(id int, t Task) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tasks[id] = t
}

func (r *taskRegistry) Delete(id int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.tasks, id)
}
