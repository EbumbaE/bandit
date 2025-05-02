package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/exp/rand"
	"golang.org/x/sync/errgroup"

	"github.com/EbumbaE/bandit/services/rule-test/internal/metrics"
	"github.com/EbumbaE/bandit/services/rule-test/internal/notifier"
)

const (
	service = "rule-test"

	ruleContextFormat     = "load_test_%d"
	loadRuleContextFormat = "load_test_%d"
)

func init() {
	rand.Seed(uint64(time.Now().Unix()))
}

type DillerClient interface {
	GetRuleData(ctx context.Context, service, context string) (string, string, error)
}

type AdminClient interface {
	CreateGaussianBanditIfExist(ctx context.Context) error

	CreateRule(ctx context.Context, service, context string) (string, error)

	AddVariant(ctx context.Context, ruleID string) (string, error)
	DisableVariant(ctx context.Context, ruleID, variantID string) error
}

type Notifier interface {
	SendAnalytic(ctx context.Context, action notifier.ActionType, amount float64, payload string) error
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

func (p *Provider) DoLoadTest(ctx context.Context, parallelCount, targetRPS int, duration time.Duration) error {
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

			if err := p.doCycle(gCtx, localCtx, targetRPS, duration); err != nil {
				return errors.Wrap(err, "doCycle")
			}

			return nil
		})
	}

	return errGr.Wait()
}

func (p *Provider) DoEfficiencyTest(ctx context.Context, targetRPS int, duration time.Duration) error {
	if err := p.admin.CreateGaussianBanditIfExist(ctx); err != nil {
		return errors.Wrap(err, "create bandit")
	}

	localCtx := fmt.Sprintf(ruleContextFormat, rand.Intn(1_000_000_000))

	ruleID, err := p.admin.CreateRule(ctx, service, localCtx)
	if err != nil {
		return errors.Wrap(err, "create rule")
	}

	time.Sleep(1000 * time.Microsecond)

	if err := p.doCycle(ctx, localCtx, targetRPS, duration); err != nil {
		return errors.Wrap(err, "1 doCycle")
	}

	variantID, err := p.admin.AddVariant(ctx, ruleID)
	if err != nil {
		return errors.Wrap(err, "create rule")
	}

	time.Sleep(1000 * time.Microsecond)

	if err := p.doCycle(ctx, localCtx, targetRPS, duration); err != nil {
		return errors.Wrap(err, "2 doCycle")
	}

	if err := p.admin.DisableVariant(ctx, ruleID, variantID); err != nil {
		return errors.Wrap(err, "create rule")
	}

	time.Sleep(1000 * time.Microsecond)

	if err := p.doCycle(ctx, localCtx, targetRPS, duration); err != nil {
		return errors.Wrap(err, "3 doCycle")
	}

	return nil
}

func (p *Provider) doCycle(ctx context.Context, localCtx string, targetRPS int, duration time.Duration) error {
	totalRequests := targetRPS * int(duration.Seconds())
	interval := time.Second / time.Duration(targetRPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range totalRequests {
		select {
		case <-ticker.C:
			startTime := time.Now()
			data, payload, err := p.diller.GetRuleData(ctx, service, localCtx)
			if err != nil {
				return errors.Wrap(err, "get rule data")
			}

			metrics.ResponceTime.WithLabelValues("GetRuleData", "ok").Observe(float64(time.Since(startTime).Milliseconds()))

			if err := p.doAnalytic(ctx, payload); err != nil {
				return errors.Wrap(err, "doAnalytic")
			}

			metrics.DataCounter.WithLabelValues(localCtx, string(data)).Inc()
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (p *Provider) doAnalytic(ctx context.Context, payload string) error {
	if err := p.notifier.SendAnalytic(ctx, notifier.ViewActionType, float64(len(payload)), payload); err != nil {
		return errors.Wrap(err, "send analytic")
	}

	return nil
}
