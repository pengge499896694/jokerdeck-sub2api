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
// SignupBonusReclaimService 周期性回收注册超过24小时且无充值记录的账号余额。
type SignupBonusReclaimService struct {
	reclaimer signupBonusReclaimer
	interval  time.Duration
	stopCh    chan struct{}
	stopOnce  sync.Once
	wg        sync.WaitGroup
}

func NewSignupBonusReclaimService(reclaimer signupBonusReclaimer, interval time.Duration) *SignupBonusReclaimService {
	return &SignupBonusReclaimService{
		reclaimer: reclaimer,
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

func (s *SignupBonusReclaimService) Start() {
	if s == nil || s.interval <= 0 { return }
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.runOnce()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C: s.runOnce()
			case <-s.stopCh: return
			}
		}
	}()
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

	cutoff := time.Now().UTC().Add(-24 * time.Hour)
	result, err := s.reclaimer.ExecuteSignupBonusReclaim(ctx, &cutoff)
	if err != nil {
		log.Printf("[SignupBonusReclaim] reclaim failed: %v", err)
		return
	}
	if result != nil && result.ReclaimedCount > 0 {
		log.Printf("[SignupBonusReclaim] reclaimed %d accounts, total=%.8f (ttl_hours=24)", result.ReclaimedCount, result.TotalAmount)
	}
}
