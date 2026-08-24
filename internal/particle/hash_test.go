package particle

import "testing"

func TestPointHashDeterministic(t *testing.T) {
	first := PointHash("P-01")
	second := PointHash("P-01")
	if first != second || first == 0 {
		t.Fatalf("hash values: %d %d", first, second)
	}
	if PointHash("P-01") == PointHash("P-02") {
		t.Fatal("distinct points must hash differently")
	}
}

func TestPointBucket(t *testing.T) {
	bucket := PointBucket("P-01", 8)
	if bucket < 0 || bucket >= 8 {
		t.Fatalf("bucket = %d", bucket)
	}
	if PointBucket("P-01", 0) != 0 {
		t.Fatal("zero buckets must map to zero")
	}
}

func TestTableReadWrite(t *testing.T) {
	table := NewTable()
	table.Register("P-01", "OR-1", 100)
	table.Write("P-01", 42, 1.0, now())
	entry, ok := table.Read("P-01")
	if !ok || !entry.HasSample || entry.LastCount != 42 {
		t.Fatalf("entry = %+v ok=%v", entry, ok)
	}
	if table.Count() != 1 {
		t.Fatalf("count = %d", table.Count())
	}
	missing, ok := table.Read("P-99")
	if ok || missing.HasSample {
		t.Fatal("missing point must not resolve")
	}
}
