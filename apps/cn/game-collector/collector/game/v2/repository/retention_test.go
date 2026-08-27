package repository

import "testing"

func TestTaskResultRetentionUsesP01SafeWindow(t *testing.T) {
	if got := normalizeRetentionConfig(RetentionConfig{}).CollectTaskResultsDays; got != 90 {
		t.Fatalf("default task-result retention = %d, want 90", got)
	}
	if got := normalizeRetentionConfig(RetentionConfig{CollectTaskResultsDays: 7}).CollectTaskResultsDays; got != 30 {
		t.Fatalf("minimum task-result retention = %d, want 30", got)
	}
	if got := normalizeRetentionConfig(RetentionConfig{CollectTaskResultsDays: 60}).CollectTaskResultsDays; got != 60 {
		t.Fatalf("configured task-result retention = %d, want 60", got)
	}
}
