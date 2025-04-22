package consumer

import (
	"context"
	"encoding/json"

	model "github.com/EbumbaE/bandit/services/bandit-indexer/internal"
	"github.com/pkg/errors"
)

type Admin interface {
	GetBandit(ctx context.Context, ruleID string) (model.Bandit, error)
	CheckBandit(ctx context.Context, ruleID string) (bool, error)
	GetBanditState(ctx context.Context, ruleID string) (model.StateType, error)
	CheckArm(ctx context.Context, ruleID, variantID string) (bool, error)
	GetArm(ctx context.Context, ruleID, variantID string) (model.Arm, error)
	GetArmState(ctx context.Context, ruleID, variantID string) (model.StateType, error)
}

type Storage interface {
	CreateBandit(ctx context.Context, bandit model.Bandit) (model.Bandit, error)
	SetBanditState(ctx context.Context, ruleID string, state model.StateType) error
	AddArm(ctx context.Context, ruleID string, v model.Arm) (model.Arm, error)
	SetArmState(ctx context.Context, variantID string, state model.StateType) error
}

type Consumer struct {
	storage Storage
	admin   Admin
}

func NewConsumer(admin Admin, storage Storage) *Consumer {
	return &Consumer{
		admin:   admin,
		storage: storage,
	}
}

type Event struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	Id     string `json:"id"`
}

func (c *Consumer) Handle(ctx context.Context, msg []byte) error {
	event := &Event{}
	if err := json.Unmarshal(msg, event); err != nil {
		return errors.Wrapf(err, "unmarshal message: %s", string(msg))
	}

	switch event.Type {
	case "rule":
		return ruleAction(ctx, event.Action, event.Id)
	case "variant":
		return variantAction(ctx, event.Action, event.Id)
	}

	return nil
}

func ruleAction(ctx context.Context, action string, Id string) error {
	switch action {
	case "create":
	case "delete":
	case "inactive":
	case "active":
	}

	return nil
}

func variantAction(ctx context.Context, action string, Id string) error {
	switch action {
	case "create":
	case "delete":
	case "inactive":
	case "active":
	}

	return nil
}
