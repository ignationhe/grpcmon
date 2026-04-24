package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/user/grpcmon/internal/probe"
)

func makeResult(target string, serving bool, latency time.Duration, err error) probe.Result {
	return probe.Result{
		Target:    target,
		Serving:   serving,
		Latency:   latency,
		Timestamp: time.Now(),
		Err:       err,
	}
}

func TestDashboard_Update_Serving(t *testing.T) {
	d := New(80)
	d.Update(makeResult("localhost:50051", true, 5*time.Millisecond, nil))
	if d.RowCount() != 1 {
		t.Fatalf("expected 1 row, got %d", d.RowCount())
	}
	row := d.rows["localhost:50051"]
	if row.Status != "SERVING" {
		t.Errorf("expected SERVING, got %s", row.Status)
	}
}

func TestDashboard_Update_NotServing(t *testing.T) {
	d := New(80)
	d.Update(makeResult("localhost:50052", false, 10*time.Millisecond, nil))
	row := d.rows["localhost:50052"]
	if row.Status != "NOT_SERVING" {
		t.Errorf("expected NOT_SERVING, got %s", row.Status)
	}
}

func TestDashboard_Update_Error(t *testing.T) {
	d := New(80)
	d.Update(makeResult("localhost:50053", false, 0, errors.New("connection refused")))
	row := d.rows["localhost:50053"]
	if row.Status != "ERROR" {
		t.Errorf("expected ERROR, got %s", row.Status)
	}
}

func TestDashboard_Render_ContainsHeaders(t *testing.T) {
	d := New(80)
	d.Update(makeResult("svc:9090", true, 3*time.Millisecond, nil))
	out := d.Render()
	for _, hdr := range []string{"TARGET", "STATUS", "LATENCY", "UPDATED"} {
		if !strings.Contains(out, hdr) {
			t.Errorf("render missing header %q", hdr)
		}
	}
	if !strings.Contains(out, "svc:9090") {
		t.Error("render missing target")
	}
}

func TestDashboard_DefaultWidth(t *testing.T) {
	d := New(0)
	if d.width != 80 {
		t.Errorf("expected default width 80, got %d", d.width)
	}
}
