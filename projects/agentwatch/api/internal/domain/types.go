package domain

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Archived  bool      `json:"archived"`
	CreatedAt time.Time `json:"created_at"`
}

type AgentState string

const (
	AgentOffline       AgentState = "offline"
	AgentIdle          AgentState = "idle"
	AgentRunning       AgentState = "running"
	AgentWaiting       AgentState = "waiting"
	AgentNeedsApproval AgentState = "needs_approval"
	AgentFailed        AgentState = "failed"
	AgentCompleted     AgentState = "completed"
)

type Agent struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	ProjectID     string     `json:"project_id"`
	Type          string     `json:"type"`
	State         AgentState `json:"state"`
	CurrentTaskID string     `json:"current_task_id,omitempty"`
	StartedAt     time.Time  `json:"started_at"`
	LastHeartbeat time.Time  `json:"last_heartbeat"`
	LatestMessage string     `json:"latest_message"`
}

type TaskStatus string

const (
	TaskQueued    TaskStatus = "queued"
	TaskRunning   TaskStatus = "running"
	TaskWaiting   TaskStatus = "waiting"
	TaskFailed    TaskStatus = "failed"
	TaskCompleted TaskStatus = "completed"
	TaskCancelled TaskStatus = "cancelled"
)

type Task struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	AgentID     string     `json:"agent_id,omitempty"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Event struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	ProjectID string         `json:"project_id,omitempty"`
	AgentID   string         `json:"agent_id,omitempty"`
	TaskID    string         `json:"task_id,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type Approval struct {
	ID        string     `json:"id"`
	ProjectID string     `json:"project_id"`
	AgentID   string     `json:"agent_id"`
	TaskID    string     `json:"task_id,omitempty"`
	Action    string     `json:"action"`
	Reason    string     `json:"reason"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	DecidedAt *time.Time `json:"decided_at,omitempty"`
}
