package monitor

import (
	"testing"
	"time"
)

func newTestTrend() *TrendAnalyzer {
	return NewTrendAnalyzer(5*time.Minute, 3)
}

func TestTrendAnalyzer_StableWhenTooFewSamples(t *testing.T) {
	ta := newTestTrend()
	ta.Record("nginx", 10.0)
	ta.Record("nginx", 20.0)

	res := ta.Analyze("nginx")
	if res.Direction != TrendStable {
		t.Fatalf("expected stable with 2 samples, got %s", res.Direction)
	}
	if res.Samples != 2 {
		t.Fatalf("expected 2 samples, got %d", res.Samples)
	}
}

func TestTrendAnalyzer_Rising(t *testing.T) {
	ta := newTestTrend()
	ta.Record("nginx", 10.0)
	ta.Record("nginx", 20.0)
	ta.Record("nginx", 30.0)

	res := ta.Analyze("nginx")
	if res.Direction != TrendRising {
		t.Fatalf("expected rising, got %s", res.Direction)
	}
	if res.Delta != 20.0 {
		t.Fatalf("expected delta 20, got %f", res.Delta)
	}
}

func TestTrendAnalyzer_Falling(t *testing.T) {
	ta := newTestTrend()
	ta.Record("nginx", 90.0)
	ta.Record("nginx", 50.0)
	ta.Record("nginx", 20.0)

	res := ta.Analyze("nginx")
	if res.Direction != TrendFalling {
		t.Fatalf("expected falling, got %s", res.Direction)
	}
	if res.Delta != -70.0 {
		t.Fatalf("expected delta -70, got %f", res.Delta)
	}
}

func TestTrendAnalyzer_Stable_EqualValues(t *testing.T) {
	ta := newTestTrend()
	ta.Record("nginx", 50.0)
	ta.Record("nginx", 50.0)
	ta.Record("nginx", 50.0)

	res := ta.Analyze("nginx")
	if res.Direction != TrendStable {
		t.Fatalf("expected stable, got %s", res.Direction)
	}
}

func TestTrendAnalyzer_UnknownProcess_Stable(t *testing.T) {
	ta := newTestTrend()
	res := ta.Analyze("unknown")
	if res.Direction != TrendStable {
		t.Fatalf("expected stable for unknown process, got %s", res.Direction)
	}
	if res.Samples != 0 {
		t.Fatalf("expected 0 samples, got %d", res.Samples)
	}
}

func TestTrendAnalyzer_WindowEviction(t *testing.T) {
	ta := NewTrendAnalyzer(50*time.Millisecond, 2)
	ta.Record("nginx", 100.0)
	time.Sleep(60 * time.Millisecond)
	ta.Record("nginx", 10.0)

	// only 1 sample within window after eviction
	res := ta.Analyze("nginx")
	if res.Samples != 1 {
		t.Fatalf("expected 1 sample after eviction, got %d", res.Samples)
	}
	if res.Direction != TrendStable {
		t.Fatalf("expected stable with 1 sample, got %s", res.Direction)
	}
}

func TestTrendAnalyzer_All_ReturnsCopy(t *testing.T) {
	ta := newTestTrend()
	ta.Record("nginx", 1)
	ta.Record("nginx", 2)
	ta.Record("nginx", 3)
	ta.Record("redis", 5)
	ta.Record("redis", 3)
	ta.Record("redis", 1)

	all := ta.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all["nginx"].Direction != TrendRising {
		t.Errorf("nginx: expected rising, got %s", all["nginx"].Direction)
	}
	if all["redis"].Direction != TrendFalling {
		t.Errorf("redis: expected falling, got %s", all["redis"].Direction)
	}
}
