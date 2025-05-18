package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/EbumbaE/bandit/pkg/logger"
	model "github.com/EbumbaE/bandit/services/rule-analytic/internal"
)

type Storage interface {
	ApplyAnalyticEvent(ctx context.Context, events []model.BanditEvent) error
	GetAnalyticEvents(ctx context.Context) ([]model.BanditEvent, error)
	DeleteAnalyticEvents(ctx context.Context, events []model.BanditEvent) error
	InsertHistoryBatch(ctx context.Context, batch []model.HistoryEvent) error
}

type Notifier interface {
	SendEvent(ctx context.Context, event model.BanditEvent) error
}

type Consumer struct {
	storage  Storage
	notifier Notifier

	historyChan    chan model.HistoryEvent
	analyticChan   chan model.BanditEvent
	shutdownChan   chan struct{}
	wg             sync.WaitGroup
	flushInterval  time.Duration
	senderInterval time.Duration
	maxBatchSize   int
}

func NewConsumer(storage Storage, notifier Notifier) *Consumer {
	c := &Consumer{
		storage:  storage,
		notifier: notifier,

		historyChan:    make(chan model.HistoryEvent, 10000),
		analyticChan:   make(chan model.BanditEvent, 10000),
		shutdownChan:   make(chan struct{}),
		flushInterval:  500 * time.Millisecond,
		senderInterval: 100 * time.Millisecond,
		maxBatchSize:   100,
	}

	c.wg.Add(3)
	go c.historyBatcher()
	go c.analyticBatcher()
	go c.eventSender()

	return c
}

type Event struct {
	Payload string           `json:"payload"`
	Action  model.ActionType `json:"action"`
	Amount  float64          `json:"amount"`
}

func (c *Consumer) Handle(ctx context.Context, msg []byte) error {
	unmarhaled := &Event{}
	if err := json.Unmarshal(msg, unmarhaled); err != nil {
		return errors.Wrapf(err, "unmarshal message: %s", string(msg))
	}

	if unmarhaled.Payload == "" {
		return nil
	}

	toHistory := model.HistoryEvent{
		Action: unmarhaled.Action,
		Amount: unmarhaled.Amount,
	}
	if err := json.Unmarshal([]byte(unmarhaled.Payload), &(toHistory.Payload)); err != nil {
		return errors.Wrapf(err, "unmarshal message: %s", string(msg))
	}

	reward := calculateReward(toHistory)

	toSend := model.BanditEvent{
		RuleID:      toHistory.Payload.RuleID,
		VariantID:   toHistory.Payload.VariantID,
		Reward:      reward,
		Count:       1,
		RuleVersion: toHistory.Payload.RuleVersion,
	}

	select {
	case c.analyticChan <- toSend:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case c.historyChan <- toHistory:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func calculateReward(history model.HistoryEvent) float64 {
	switch history.Action {
	case model.ClickActionType:
		return history.Amount * 0.3
	case model.ViewActionType:
		return history.Amount * 0.1
	case model.CartActionType:
		return history.Amount * 0.003
	case model.PurchaseActionType:
		return history.Amount * 0.3
	default:
		return 0
	}
}

func (c *Consumer) historyBatcher() {
	defer c.wg.Done()

	var batch []model.HistoryEvent
	ticker := time.NewTicker(c.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case event := <-c.historyChan:
			batch = append(batch, event)
			if len(batch) >= c.maxBatchSize {
				c.flushHistory(batch)
				batch = nil
			}
			ticker.Reset(c.flushInterval)
		case <-ticker.C:
			if len(batch) > 0 {
				c.flushHistory(batch)
				batch = nil
			}
			ticker.Reset(c.flushInterval)
		case <-c.shutdownChan:
			if len(batch) > 0 {
				c.flushHistory(batch)
			}
			ticker.Stop()
			return
		}
	}
}

func (c *Consumer) flushHistory(batch []model.HistoryEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.storage.InsertHistoryBatch(ctx, batch); err != nil {
		logger.Error("flush history batch", zap.Error(err))
	}
}

func (c *Consumer) analyticBatcher() {
	defer c.wg.Done()

	var batch []model.BanditEvent
	ticker := time.NewTicker(c.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case event := <-c.analyticChan:
			batch = append(batch, event)
			if len(batch) >= c.maxBatchSize {
				c.flushAnalytics(batch)
				batch = nil
			}
			ticker.Reset(c.flushInterval)
		case <-ticker.C:
			if len(batch) > 0 {
				c.flushAnalytics(batch)
				batch = nil
			}
			ticker.Reset(c.flushInterval)
		case <-c.shutdownChan:
			if len(batch) > 0 {
				c.flushAnalytics(batch)
			}
			ticker.Stop()
			return
		}
	}
}

func (c *Consumer) flushAnalytics(batch []model.BanditEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	aggregated := make(map[string]model.BanditEvent)
	for _, event := range batch {
		key := fmt.Sprintf("%s:%s:%d", event.RuleID, event.VariantID, event.RuleVersion)
		if existing, exists := aggregated[key]; exists {
			existing.Count += event.Count
			existing.Reward += event.Reward
			aggregated[key] = existing
		} else {
			aggregated[key] = event
		}
	}

	var aggregatedBatch []model.BanditEvent
	for _, event := range aggregated {
		aggregatedBatch = append(aggregatedBatch, event)
	}

	if err := c.storage.ApplyAnalyticEvent(ctx, aggregatedBatch); err != nil {
		logger.Error("apply analitic batch", zap.Error(err))
	}
}

func (c *Consumer) eventSender() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.senderInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.sendEvents()
			ticker.Reset(c.senderInterval)
		case <-c.shutdownChan:
			c.sendEvents()
			ticker.Stop()
			return
		}
	}
}

func (c *Consumer) sendEvents() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := c.storage.GetAnalyticEvents(ctx)
	if err != nil {
		logger.Error("get analitic batch", zap.Error(err))
		return
	}
	if len(events) == 0 {
		return
	}

	var toDelete []model.BanditEvent
	for _, event := range events {
		err := c.notifier.SendEvent(ctx, event)
		if err != nil {
			logger.Error("send analitic event", zap.Error(err))
			continue
		}
		toDelete = append(toDelete, event)
	}

	if len(toDelete) > 0 {
		_ = c.storage.DeleteAnalyticEvents(ctx, toDelete)
	}
}

func (c *Consumer) Close() error {
	close(c.shutdownChan)
	c.wg.Wait()
	close(c.historyChan)
	close(c.analyticChan)

	return nil
}
