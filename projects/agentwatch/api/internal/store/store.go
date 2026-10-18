package store

import (
	"context"
	"errors"

	"github.com/AndersonPaulinoDev/agentwatch/api/internal/domain"
)

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")

type Store interface {
	CreateUser(context.Context, domain.User) error
	UserByEmail(context.Context, string) (domain.User, error)
	UserByID(context.Context, string) (domain.User, error)

	CreateProject(context.Context, domain.Project) error
	ListProjects(context.Context) ([]domain.Project, error)

	UpsertAgent(context.Context, domain.Agent) error
	ListAgents(context.Context) ([]domain.Agent, error)
	AgentByID(context.Context, string) (domain.Agent, error)

	CreateTask(context.Context, domain.Task) error
	UpdateTask(context.Context, domain.Task) error
	TaskByID(context.Context, string) (domain.Task, error)
	ListTasks(context.Context) ([]domain.Task, error)

	AppendEvent(context.Context, domain.Event) error
	ListEvents(context.Context, int) ([]domain.Event, error)

	CreateApproval(context.Context, domain.Approval) error
	ApprovalByID(context.Context, string) (domain.Approval, error)
	UpdateApproval(context.Context, domain.Approval) error
	ListApprovals(context.Context) ([]domain.Approval, error)
}
