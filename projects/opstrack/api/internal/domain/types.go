package domain

import "time"

type Role string

const (
	RoleWorker     Role = "worker"
	RoleSupervisor Role = "supervisor"
	RoleAdmin      Role = "admin"
)

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         Role   `json:"role"`
}

type Status string

const (
	StatusNew        Status = "new"
	StatusAssigned   Status = "assigned"
	StatusAccepted   Status = "accepted"
	StatusInProgress Status = "in_progress"
	StatusBlocked    Status = "blocked"
	StatusCompleted  Status = "completed"
	StatusVerified   Status = "verified"
	StatusCancelled  Status = "cancelled"
)

type WorkOrder struct {
	ID                 string    `json:"id"`
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	Location           string    `json:"location"`
	Priority           string    `json:"priority"`
	AssignedEmployee   string    `json:"assigned_employee"`
	DueDate            time.Time `json:"due_date"`
	Status             Status    `json:"status"`
	Instructions       string    `json:"instructions"`
	EquipmentTag       string    `json:"equipment_tag"`
	Attachments        []string  `json:"attachments"`
	CompletionEvidence []string  `json:"completion_evidence"`
	Notes              []string  `json:"notes"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type AuditEntry struct {
	ID          string            `json:"id"`
	Actor       string            `json:"actor"`
	Action      string            `json:"action"`
	WorkOrderID string            `json:"work_order"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
}

type SyncOperation struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	WorkOrderID string `json:"work_order_id"`
	Status      Status `json:"status,omitempty"`
	Note        string `json:"note,omitempty"`
	Evidence    string `json:"evidence,omitempty"`
}

type SyncResult struct {
	ID    string `json:"id"`
	State string `json:"state"`
	Error string `json:"error,omitempty"`
}
