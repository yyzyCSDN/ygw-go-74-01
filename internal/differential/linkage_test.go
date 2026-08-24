package differential

import (
	"testing"
)

// TestApplyFanSwitchCommitsTarget reproduces the reported defect at the linkage
// level: ApplyFanSwitch must commit the new fan as the active target and record
// it in history, not merely stage a value that Target() never reads.
func TestApplyFanSwitchCommitsTarget(t *testing.T) {
	linkage := NewLinkage()
	linkage.SetTarget("OR-1", "EX-1")
	rev0 := linkage.Revision("OR-1")

	linkage.ApplyFanSwitch("OR-1", "EX-2")

	target, ok := linkage.Target("OR-1")
	if !ok || target != "EX-2" {
		t.Fatalf("target after apply = %q, want EX-2", target)
	}
	if got := linkage.Revision("OR-1"); got <= rev0 {
		t.Fatalf("revision = %d, want > %d", got, rev0)
	}
	if history := linkage.History("OR-1"); len(history) == 0 || history[len(history)-1] != "EX-2" {
		t.Fatalf("history = %v, want EX-2 recorded", history)
	}
}

// TestApplyFanSwitchIgnoresEmpty guards the no-op guard for empty fan ids.
func TestApplyFanSwitchIgnoresEmpty(t *testing.T) {
	linkage := NewLinkage()
	linkage.SetTarget("OR-1", "EX-1")
	rev0 := linkage.Revision("OR-1")

	linkage.ApplyFanSwitch("OR-1", "")

	target, ok := linkage.Target("OR-1")
	if !ok || target != "EX-1" {
		t.Fatalf("target after empty apply = %q, want EX-1", target)
	}
	if got := linkage.Revision("OR-1"); got != rev0 {
		t.Fatalf("revision = %d, want %d (empty apply must not bump)", got, rev0)
	}
}

// TestSetTargetRecordsInitialTarget confirms the startup path still seeds the
// initial target and history is empty until a runtime switch occurs.
func TestSetTargetRecordsInitialTarget(t *testing.T) {
	linkage := NewLinkage()
	linkage.SetTarget("OR-1", "EX-1")

	target, ok := linkage.Target("OR-1")
	if !ok || target != "EX-1" {
		t.Fatalf("target = %q, want EX-1", target)
	}
	if got := linkage.Revision("OR-1"); got != 1 {
		t.Fatalf("revision = %d, want 1", got)
	}
	if history := linkage.History("OR-1"); len(history) != 0 {
		t.Fatalf("history = %v, want empty before any switch", history)
	}
}

// TestSnapshotReflectsSwitch confirms the exported snapshot (used for
// persistence and the console) follows a committed fan switch.
func TestSnapshotReflectsSwitch(t *testing.T) {
	linkage := NewLinkage()
	linkage.SetTarget("OR-1", "EX-1")
	linkage.ApplyFanSwitch("OR-1", "EX-2")

	snap := linkage.Snapshot()
	if snap["OR-1"] != "EX-2" {
		t.Fatalf("snapshot target = %q, want EX-2", snap["OR-1"])
	}
}
