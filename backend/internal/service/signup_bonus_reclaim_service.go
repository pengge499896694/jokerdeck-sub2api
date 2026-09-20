package service

import (
	"context"
	"log"
	"sync"
	"time"
)

// signupBonusReclaimer 是回收服务对 AdminService 的窄依赖，便于测试替换。
type signupBonusReclaimer interface {
	ExecuteSignupBonusReclaim(ctx context.Context, olderThan *time.Time) (*SignupBonusReclaimResult, error)
}

// signupBonusTTLProvider 提供当前配置的注册赠额有效期（小时，<=0 表示关闭）。
type signupBonusTTLProvider interface {
	GetSignupBonusTTLHours(ctx context.Context) int
}

// SignupBonusReclaimService 周期性回收连续 30 天未使用且无充值记录的账号余额。
type SignupBonusReclaimService struct {
	reclaimer signupBonusReclaimer
	ttl       signupBonusTTLProvider
	interval  time.Duration
	stopCh    chan struct{}
	stopOnce  sync.Once
	wg        sync.WaitGroup
}

func NewSignupBonusReclaimService(reclaimer signupBonusReclaimer, ttl signupBonusTTLProvider, interval time.Duration) *SignupBonusReclaimService {
	return &SignupBonusReclaimService{
		reclaimer: reclaimer,
		ttl:       ttl,
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

func (s *SignupBonusReclaimService) Start() {
	// Manual-only: never reclaim balances immediately after deployment.
}

func (s *SignupBonusReclaimService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *SignupBonusReclaimService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.reclaimer.ExecuteSignupBonusReclaim(ctx, nil)
	if err != nil {
		log.Printf("[SignupBonusReclaim] reclaim failed: %v", err)
		return
	}
	if result != nil && result.ReclaimedCount > 0 {
		log.Printf("[InactiveBalanceReclaim] reclaimed %d accounts, total=%.8f (inactive_days=30)", result.ReclaimedCount, result.TotalAmount)
	}
}
