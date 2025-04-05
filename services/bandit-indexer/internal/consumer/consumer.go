package consumer

import (
	"context"
	"encoding/json"

	model "github.com/EbumbaE/bandit/services/bandit-indexer/internal"
	"github.com/pkg/errors"
)

type Admin interface {
	GetRule(ctx context.Context, ruleID string) (model.Rule, error)
}

type Storage interface {
	SaveRuleVariants(ctx context.Context, service, context string, variants []model.Variant) error
	SaveRuleVersion(ctx context.Context, service, context string, version uint64) error
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
