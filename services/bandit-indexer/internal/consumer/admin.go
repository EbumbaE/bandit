package consumer

import (
	"context"
	"encoding/json"

	core "github.com/EbumbaE/bandit/services/bandit-core/v6"
	"github.com/pkg/errors"

	model "github.com/EbumbaE/bandit/services/bandit-indexer/internal"
	"github.com/EbumbaE/bandit/services/bandit-indexer/internal/storage"
)

type Notifier interface {
	Send(ctx context.Context, ruleID string) error
}

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
	DeleteBandit(ctx context.Context, ruleID string) error

	AddArm(ctx context.Context, ruleID string, v model.Arm) (model.Arm, error)
	SetArmState(ctx context.Context, variantID string, state model.StateType) error
	DeleteArm(ctx context.Context, variantID string) error
}

type AdminConsumer struct {
	storage  Storage
	admin    Admin
	notifier Notifier
}

func NewAdminConsumer(admin Admin, storage Storage, notifier Notifier) *AdminConsumer {
	return &AdminConsumer{
		admin:    admin,
		storage:  storage,
		notifier: notifier,
	}
}

type AdminEvent struct {
	Type      string `json:"type"`
	Action    string `json:"action"`
	RuleID    string `json:"rule_id"`
	VariantID string `json:"variant_id"`
}

func (c *AdminConsumer) Handle(ctx context.Context, msg []byte) error {
	event := &AdminEvent{}
	if err := json.Unmarshal(msg, event); err != nil {
		return errors.Wrapf(err, "unmarshal message: %s", string(msg))
	}

	var err error
	switch event.Type {
	case "rule":
		err = c.ruleAction(ctx, event.Action, event.RuleID)
	case "variant":
		err = c.variantAction(ctx, event.Action, event.RuleID, event.VariantID)
	}
	if err != nil {
		return errors.Wrapf(err, "exec action[%v]", event)
	}

	return c.notifier.Send(ctx, event.RuleID)
}

func (c *AdminConsumer) ruleAction(ctx context.Context, action string, ruleID string) error {
	switch action {
	case "create":
		bandit, err := c.admin.GetBandit(ctx, ruleID)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return errors.Wrap(err, "admin.GetBandit")
		}

		banditCore := core.NewDefaultGaussianBandit()
		bandit.Config, err = banditCore.Serialize()
		if err != nil {
			return errors.Wrap(err, "banditCore.Serialize")
		}

		if _, err = c.storage.CreateBandit(ctx, bandit); err != nil {
			return errors.Wrap(err, "storage.CreateBandit")
		}

	case "delete":
		if err := c.storage.DeleteBandit(ctx, ruleID); err != nil {
			return errors.Wrap(err, "storage.CreateBandit")
		}

	case "inactive":
		if err := c.storage.SetBanditState(ctx, ruleID, model.StateTypeDisable); err != nil {
			return errors.Wrap(err, "storage.SetBanditState StateTypeDisable")
		}

	case "active":
		if err := c.storage.SetBanditState(ctx, ruleID, model.StateTypeEnable); err != nil {
			return errors.Wrap(err, "storage.SetBanditState StateTypeEnable")
		}
	}

	return nil
}

func (c *AdminConsumer) variantAction(ctx context.Context, action string, ruleID, variantID string) error {
	switch action {
	case "create":
		arm, err := c.admin.GetArm(ctx, ruleID, variantID)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return errors.Wrap(err, "admin.GetBandit")
		}

		armCore := core.NewDefaultGaussianArm()
		arm.Config, err = armCore.Serialize()
		if err != nil {
			return errors.Wrap(err, "armCore.Serialize")
		}

		if _, err = c.storage.AddArm(ctx, ruleID, arm); err != nil {
			return errors.Wrap(err, "storage.CreateBandit")
		}

	case "delete":
		if err := c.storage.DeleteArm(ctx, variantID); err != nil {
			return errors.Wrap(err, "storage.DeleteArm")
		}

	case "inactive":
		if err := c.storage.SetArmState(ctx, variantID, model.StateTypeDisable); err != nil {
			return errors.Wrap(err, "storage.SetArmState StateTypeDisable")
		}

	case "active":
		if err := c.storage.SetArmState(ctx, variantID, model.StateTypeEnable); err != nil {
			return errors.Wrap(err, "storage.SetArmState StateTypeEnable")
		}
	}

	return nil
}
