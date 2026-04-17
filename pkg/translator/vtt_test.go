package translator

import "testing"

func TestMarkLongCueBlocks_SkipsBlocksWithMoreThanThreeTextLines(t *testing.T) {
	lines := []string{
		"WEBVTT",
		"",
		"00:00:00.000 --> 00:00:02.480 line:20%",
		"Deserted Island Survival",
		"Days",
		"July 19: Set out",
		"July 20-August 3: Special test",
		"August 4-10: Cruise (free time)",
		"August 11: Return, activity ends",
		"",
		"00:00:02.500 --> 00:00:03.000",
		"Halo!",
	}

	blocked := markLongCueBlocks(lines)

	for _, idx := range []int{2, 3, 4, 5, 6, 7, 8} {
		if !blocked[idx] {
			t.Fatalf("expected line %d to be blocked", idx)
		}
	}

	if blocked[10] || blocked[11] {
		t.Fatalf("expected short cue to remain unblocked")
	}
}

func TestMarkLongCueBlocks_SkipsSingleLineCueWhenWordsExceedLimit(t *testing.T) {
	lines := []string{
		"WEBVTT",
		"",
		"00:00:36.390 --> 00:00:40.390 line:20%",
		"Poin kelas yang diperoleh oleh tiga kelompok teratas akan ditransfer dari tahun-tahun tiga kelompok terbawah. Poin kelas akan dibagi rata antar kelas dalam grup, berapa pun jumlah anggotanya.",
		"",
		"00:00:41.000 --> 00:00:42.000",
		"Lanjut.",
	}

	blocked := markLongCueBlocks(lines)

	if !blocked[2] || !blocked[3] {
		t.Fatalf("expected over-word cue lines to be blocked")
	}

	if blocked[5] || blocked[6] {
		t.Fatalf("expected short cue to remain unblocked")
	}
}
