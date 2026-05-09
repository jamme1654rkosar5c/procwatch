package monitor

import (
	"testing"
)

func TestNotifyRuleStore_Add_And_Get(t *testing.T) {
	s := NewNotifyRuleStore()
	err := s.Add(NotifyRule{Process: "nginx", EventType: "down", Channels: []string{"slack"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := s.Get("nginx")
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].EventType != "down" {
		t.Errorf("expected event type 'down', got %q", rules[0].EventType)
	}
}

func TestNotifyRuleStore_Add_EmptyProcess(t *testing.T) {
	s := NewNotifyRuleStore()
	err := s.Add(NotifyRule{Process: "", EventType: "down"})
	if err == nil {
		t.Fatal("expected error for empty process")
	}
}

func TestNotifyRuleStore_Add_EmptyEventType(t *testing.T) {
	s := NewNotifyRuleStore()
	err := s.Add(NotifyRule{Process: "nginx", EventType: ""})
	if err == nil {
		t.Fatal("expected error for empty event type")
	}
}

func TestNotifyRuleStore_ChannelsFor_Match(t *testing.T) {
	s := NewNotifyRuleStore()
	_ = s.Add(NotifyRule{Process: "nginx", EventType: "down", Channels: []string{"slack", "email"}})
	_ = s.Add(NotifyRule{Process: "nginx", EventType: "cpu", Channels: []string{"pagerduty"}})

	ch := s.ChannelsFor("nginx", "down")
	if len(ch) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(ch))
	}
}

func TestNotifyRuleStore_ChannelsFor_Wildcard(t *testing.T) {
	s := NewNotifyRuleStore()
	_ = s.Add(NotifyRule{Process: "nginx", EventType: "*", Channels: []string{"slack"}})

	ch := s.ChannelsFor("nginx", "cpu")
	if len(ch) != 1 || ch[0] != "slack" {
		t.Errorf("expected [slack], got %v", ch)
	}
}

func TestNotifyRuleStore_ChannelsFor_Dedup(t *testing.T) {
	s := NewNotifyRuleStore()
	_ = s.Add(NotifyRule{Process: "nginx", EventType: "down", Channels: []string{"slack"}})
	_ = s.Add(NotifyRule{Process: "nginx", EventType: "down", Channels: []string{"slack"}})

	ch := s.ChannelsFor("nginx", "down")
	if len(ch) != 1 {
		t.Errorf("expected deduped channels, got %v", ch)
	}
}

func TestNotifyRuleStore_Delete(t *testing.T) {
	s := NewNotifyRuleStore()
	_ = s.Add(NotifyRule{Process: "nginx", EventType: "down", Channels: []string{"slack"}})
	s.Delete("nginx")
	if len(s.Get("nginx")) != 0 {
		t.Error("expected rules to be deleted")
	}
}

func TestNotifyRuleStore_All_ReturnsCopy(t *testing.T) {
	s := NewNotifyRuleStore()
	_ = s.Add(NotifyRule{Process: "nginx", EventType: "down", Channels: []string{"slack"}})
	all := s.All()
	all["nginx"] = nil
	if len(s.Get("nginx")) == 0 {
		t.Error("All() should return a copy, not a reference")
	}
}
