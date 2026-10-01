package finance

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestParseFinanceFiltersMonthRange(t *testing.T) {
	tests := []struct {
		query string
		start string
		end   string
		error bool
	}{
		{query: "start_month=2026-08&end_month=2026-10", start: "2026-08-01", end: "2026-11-01"},
		{query: "start_month=2026-12&end_month=2027-01", start: "2026-12-01", end: "2027-02-01"},
		{query: "month=2026-09", start: "2026-09-01", end: "2026-10-01"},
		{query: "start_month=2026-07", start: "2026-07-01"},
		{query: "end_month=2026-07", end: "2026-08-01"},
		{query: "start_month=2026-10&end_month=2026-09", error: true},
		{query: "start_month=2026-13", error: true},
		{query: "month=2026-09&end_month=2026-10", error: true},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("GET", "/finance/usage?"+tt.query, nil)
			f, err := parseFinanceFilters(c)
			if tt.error {
				if err == nil {
					t.Fatal("expected invalid range")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, check := range []struct {
				got  *time.Time
				want string
			}{{f.StartTime, tt.start}, {f.EndTime, tt.end}} {
				if check.want == "" {
					if check.got != nil {
						t.Fatalf("unexpected bound: %v", check.got)
					}
				} else if check.got == nil || check.got.Format("2006-01-02") != check.want {
					t.Fatalf("bound = %v, want %s", check.got, check.want)
				}
			}
		})
	}
}

func TestSafeCSVText(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "normal text", want: "normal text"},
		{input: "=SUM(A1:A2)", want: "'=SUM(A1:A2)"},
		{input: "  @command", want: "'  @command"},
		{input: "-1+2", want: "'-1+2"},
		{input: "note\r\nnext", want: "note  next"},
	}
	for _, tt := range tests {
		if got := safeCSVText(tt.input); got != tt.want {
			t.Errorf("safeCSVText(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
