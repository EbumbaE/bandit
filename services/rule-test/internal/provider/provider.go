package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"
	"golang.org/x/sync/errgroup"

	"github.com/EbumbaE/bandit/pkg/logger"
	"github.com/EbumbaE/bandit/services/rule-test/internal/metrics"
	"github.com/EbumbaE/bandit/services/rule-test/internal/notifier"
)

const (
	service     = "rule-test"
	ruleContext = "efficiency_test"

	loadRuleContextFormat = "load_test_%d"
)

func init() {
	rand.Seed(uint64(time.Now().Unix()))
}

type DillerClient interface {
	GetRuleData(ctx context.Context, service, context string) ([]byte, []byte, error)
}

type AdminClient interface {
	CreateGaussianBanditIfExist(ctx context.Context) error

	CreateRule(ctx context.Context, service, context string) (string, error)

	AddVariant(ctx context.Context, ruleID string) (string, error)
	DisableVariant(ctx context.Context, ruleID, variantID string) error
}

type Notifier interface {
	SendAnalytic(ctx context.Context, action notifier.ActionType, amount float64, payload []byte) error
}

type Provider struct {
	diller   DillerClient
	admin    AdminClient
	notifier Notifier
}

func NewProvider(diller DillerClient, admin AdminClient, notifier Notifier) *Provider {
	return &Provider{
		diller:   diller,
		admin:    admin,
		notifier: notifier,
	}
}

func (p *Provider) DoLoadTest(ctx context.Context, parallelCount, cycleAmount int) error {
	if err := p.admin.CreateGaussianBanditIfExist(ctx); err != nil {
		return errors.Wrap(err, "create bandit")
	}

	errGr, gCtx := errgroup.WithContext(ctx)

	for range parallelCount {
		errGr.Go(func() error {
			localCtx := fmt.Sprintf(loadRuleContextFormat, rand.Intn(1_000_000_000))

			_, err := p.admin.CreateRule(gCtx, service, localCtx)
			if err != nil {
				return errors.Wrap(err, "create rule")
			}

			if err := p.doCycle(gCtx, ruleContext, cycleAmount); err != nil {
				return errors.Wrap(err, "doCycle")
			}

			return nil
		})
	}

	return errGr.Wait()
}

func (p *Provider) DoEfficiencyTest(ctx context.Context, cycleAmount int) error {
	if err := p.admin.CreateGaussianBanditIfExist(ctx); err != nil {
		return errors.Wrap(err, "create bandit")
	}

	ruleID, err := p.admin.CreateRule(ctx, service, ruleContext)
	if err != nil {
		return errors.Wrap(err, "create rule")
	}

	if err := p.doCycle(ctx, ruleContext, cycleAmount); err != nil {
		return errors.Wrap(err, "1 doCycle")
	}

	variantID, err := p.admin.AddVariant(ctx, ruleID)
	if err != nil {
		return errors.Wrap(err, "create rule")
	}

	if err := p.doCycle(ctx, ruleContext, cycleAmount); err != nil {
		return errors.Wrap(err, "2 doCycle")
	}

	if err := p.admin.DisableVariant(ctx, ruleID, variantID); err != nil {
		return errors.Wrap(err, "create rule")
	}

	if err := p.doCycle(ctx, ruleContext, cycleAmount); err != nil {
		return errors.Wrap(err, "3 doCycle")
	}

	return nil
}

func (p *Provider) doCycle(ctx context.Context, localCtx string, n int) error {
	for range n {
		startTime := time.Now()

		data, payload, err := p.diller.GetRuleData(ctx, service, localCtx)
		if err != nil {
			return errors.Wrap(err, "create rule")
		}

		metrics.SummaryResponceTime.Observe(float64(time.Since(startTime).Milliseconds()))

		if err := p.notifier.SendAnalytic(ctx, notifier.ViewActionType, 1, payload); err != nil {
			return errors.Wrap(err, "send analytic")
		}

		metrics.DataCounter.WithLabelValues(string(data)).Inc()
		logger.Info("[DoEfficiencyTest] data", zap.String("data", string(data)))
	}

	return nil
}
