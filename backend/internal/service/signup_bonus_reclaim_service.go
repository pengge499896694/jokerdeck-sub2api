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

// SignupBonusReclaimService 周期性回收"从未充值、注册满 TTL"的账号赠额。
// 当 signup_bonus_ttl_hours <= 0 时不做任何事（等价于关闭）。
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
	if s == nil || s.reclaimer == nil || s.ttl == nil || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.runOnce()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
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

	hours := s.ttl.GetSignupBonusTTLHours(ctx)
	if hours <= 0 {
		return // 未启用到期回收
	}

	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	result, err := s.reclaimer.ExecuteSignupBonusReclaim(ctx, &cutoff)
	if err != nil {
		log.Printf("[SignupBonusReclaim] reclaim failed: %v", err)
		return
	}
	if result != nil && result.ReclaimedCount > 0 {
		log.Printf("[SignupBonusReclaim] reclaimed %d accounts, total=%.8f (ttl=%dh)", result.ReclaimedCount, result.TotalAmount, hours)
	}
}
