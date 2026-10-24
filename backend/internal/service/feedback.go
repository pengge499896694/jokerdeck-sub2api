package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	ErrFeedbackNotFound  = errors.New("feedback ticket not found")
	ErrFeedbackForbidden = errors.New("feedback ticket access denied")
	ErrFeedbackInvalid   = errors.New("invalid feedback ticket")
)

type FeedbackContext struct {
	Model     string            `json:"model,omitempty"`
	ErrorCode string            `json:"error_code,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

func (c FeedbackContext) JSON() []byte {
	b, _ := json.Marshal(c)
	return b
}

type FeedbackTicket struct {
	ID             int64             `json:"id"`
	UserID         int64             `json:"user_id"`
	UserEmail      string            `json:"user_email,omitempty"`
	Subject        string            `json:"subject"`
	Category       string            `json:"category"`
	Status         string            `json:"status"`
	Priority       string            `json:"priority"`
	Model          string            `json:"model,omitempty"`
	ErrorCode      string            `json:"error_code,omitempty"`
	RequestID      string            `json:"request_id,omitempty"`
	Context        json.RawMessage   `json:"context"`
	LastUserAt     time.Time         `json:"last_user_at"`
	LastAdminAt    *time.Time        `json:"last_admin_at,omitempty"`
	LastRemindedAt *time.Time        `json:"last_reminded_at,omitempty"`
	ClosedAt       *time.Time        `json:"closed_at,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	Messages       []FeedbackMessage `json:"messages,omitempty"`
}

type FeedbackMessage struct {
	ID           int64           `json:"id"`
	TicketID     int64           `json:"ticket_id"`
	AuthorType   string          `json:"author_type"`
	AuthorUserID *int64          `json:"author_user_id,omitempty"`
	AuthorEmail  string          `json:"author_email,omitempty"`
	Body         string          `json:"body"`
	Context      json.RawMessage `json:"context"`
	CreatedAt    time.Time       `json:"created_at"`
}

type FeedbackRepository interface {
	CreateTicket(context.Context, *FeedbackTicket, string) error
	ListTickets(context.Context, int64, bool, string, int, int) ([]FeedbackTicket, int64, error)
	GetTicket(context.Context, int64, int64, bool) (*FeedbackTicket, error)
	AddMessage(context.Context, int64, *FeedbackMessage, bool) error
	SetStatus(context.Context, int64, string) error
	DeleteTicket(context.Context, int64) error
	ListUnanswered(context.Context, time.Time) ([]FeedbackTicket, error)
	MarkReminded(context.Context, int64, time.Time) error
}

type FeedbackService struct {
	repo  FeedbackRepository
	email *EmailService
}

type FeedbackReminderService struct {
	feedback   *FeedbackService
	adminEmail string
	stop       chan struct{}
	done       chan struct{}
}

func NewFeedbackReminderService(feedback *FeedbackService, adminEmail string) *FeedbackReminderService {
	r := &FeedbackReminderService{feedback: feedback, adminEmail: strings.TrimSpace(adminEmail), stop: make(chan struct{}), done: make(chan struct{})}
	go r.run()
	return r
}

func (r *FeedbackReminderService) run() {
	defer close(r.done)
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			_, _ = r.feedback.SendTimeoutReminders(ctx, r.adminEmail, 24*time.Hour)
			cancel()
		case <-r.stop:
			return
		}
	}
}

func (r *FeedbackReminderService) Stop() {
	if r == nil {
		return
	}
	close(r.stop)
	<-r.done
}

func NewFeedbackService(repo FeedbackRepository, email *EmailService) *FeedbackService {
	return &FeedbackService{repo: repo, email: email}
}

func (s *FeedbackService) CreateTicket(ctx context.Context, userID int64, subject, category, body string, fc FeedbackContext) (*FeedbackTicket, error) {
	subject, body = strings.TrimSpace(subject), strings.TrimSpace(body)
	category = strings.TrimSpace(category)
	if userID <= 0 || subject == "" || len(subject) > 200 || body == "" || len(body) > 20000 {
		return nil, ErrFeedbackInvalid
	}
	if category == "" {
		category = "other"
	}
	if len(category) > 50 {
		return nil, ErrFeedbackInvalid
	}
	fc.Model = truncateFeedbackField(fc.Model, 200)
	fc.ErrorCode = truncateFeedbackField(fc.ErrorCode, 120)
	fc.RequestID = truncateFeedbackField(fc.RequestID, 200)
	if len(fc.Metadata) > 20 {
		return nil, ErrFeedbackInvalid
	}
	for k, v := range fc.Metadata {
		if len(k) > 64 || len(v) > 500 {
			return nil, ErrFeedbackInvalid
		}
	}
	t := &FeedbackTicket{UserID: userID, Subject: subject, Category: category, Status: "open", Priority: "normal", Model: fc.Model, ErrorCode: fc.ErrorCode, RequestID: fc.RequestID, Context: fc.JSON()}
	if err := s.repo.CreateTicket(ctx, t, body); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *FeedbackService) ListTickets(ctx context.Context, userID int64, admin bool, status string, page, pageSize int) ([]FeedbackTicket, int64, error) {
	return s.repo.ListTickets(ctx, userID, admin, strings.TrimSpace(status), page, pageSize)
}

func (s *FeedbackService) GetTicket(ctx context.Context, id, userID int64, admin bool) (*FeedbackTicket, error) {
	t, err := s.repo.GetTicket(ctx, id, userID, admin)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrFeedbackNotFound
	}
	return t, nil
}

func (s *FeedbackService) Reply(ctx context.Context, id, userID int64, admin bool, body string, fc FeedbackContext) error {
	body = strings.TrimSpace(body)
	if id <= 0 || body == "" || len(body) > 20000 {
		return ErrFeedbackInvalid
	}
	return s.repo.AddMessage(ctx, id, &FeedbackMessage{AuthorType: map[bool]string{true: "admin", false: "user"}[admin], AuthorUserID: &userID, Body: body, Context: fc.JSON()}, admin)
}

func (s *FeedbackService) SetStatus(ctx context.Context, id int64, status string) error {
	status = strings.TrimSpace(status)
	switch status {
	case "open", "pending_user", "pending_admin", "resolved", "closed":
	default:
		return ErrFeedbackInvalid
	}
	return s.repo.SetStatus(ctx, id, status)
}

func (s *FeedbackService) DeleteTicket(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrFeedbackInvalid
	}
	return s.repo.DeleteTicket(ctx, id)
}

func (s *FeedbackService) SendTimeoutReminders(ctx context.Context, adminEmail string, timeout time.Duration) (int, error) {
	if s.email == nil || strings.TrimSpace(adminEmail) == "" {
		return 0, nil
	}
	tickets, err := s.repo.ListUnanswered(ctx, time.Now().Add(-timeout))
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, t := range tickets {
		subject := "Feedback ticket waiting for reply: #" + stringID(t.ID)
		body := "<p>A feedback ticket has not received an admin reply within the configured timeout.</p>" +
			"<p><b>#" + stringID(t.ID) + "</b> " + htmlEscape(t.Subject) + "</p>"
		if err := s.email.SendEmail(ctx, adminEmail, subject, body); err == nil {
			if markErr := s.repo.MarkReminded(ctx, t.ID, time.Now()); markErr != nil {
				return sent, markErr
			}
			sent++
		}
	}
	return sent, nil
}

func stringID(id int64) string { return strconv.FormatInt(id, 10) }
func truncateFeedbackField(v string, max int) string {
	v = strings.TrimSpace(v)
	if len(v) > max {
		return v[:max]
	}
	return v
}
