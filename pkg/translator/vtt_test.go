package translator

import (
	"reflect"
	"testing"
)

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

func TestCollectVTTCueBatches_GroupsByTimestamp(t *testing.T) {
	lines := []string{
		"WEBVTT",
		"",
		"00:17.976 --> 00:19.853",
		"A SPRING BREEZE BLOWS",
		"THROUGH THE YOZAKURAS",
		"",
		"00:20.437 --> 00:23.732",
		"Sui! You know that Asano kid",
		"who married into the Yozakura family?",
	}

	blocked := markLongCueBlocks(lines)
	cues := collectVTTCueBatches(lines, blocked)

	if len(cues) != 2 {
		t.Fatalf("expected 2 cue batches, got %d", len(cues))
	}

	if cues[0].originalText != "A SPRING BREEZE BLOWS\nTHROUGH THE YOZAKURAS" {
		t.Fatalf("unexpected first cue text: %q", cues[0].originalText)
	}

	if cues[1].originalText != "Sui! You know that Asano kid\nwho married into the Yozakura family?" {
		t.Fatalf("unexpected second cue text: %q", cues[1].originalText)
	}

	if lines[2] != "00:00:17.976 --> 00:00:19.853" {
		t.Fatalf("expected normalized first timestamp, got %q", lines[2])
	}

	if lines[6] != "00:00:20.437 --> 00:00:23.732" {
		t.Fatalf("expected normalized second timestamp, got %q", lines[6])
	}
}

func TestApplyTranslatedCue_DistributesTranslatedLinesToCueSlots(t *testing.T) {
	lines := []string{
		"WEBVTT",
		"",
		"00:00:17.976 --> 00:00:19.853",
		"A SPRING BREEZE BLOWS",
		"THROUGH THE YOZAKURAS",
		"",
	}

	cue := vttCueBatch{
		textLineIndices: []int{3, 4},
		originalText:    "A SPRING BREEZE BLOWS\nTHROUGH THE YOZAKURAS",
	}

	applyTranslatedCue(lines, cue, "ANGIN MUSIM SEMI BERHEMBUS\nMELALUI KELUARGA YOZAKURA", "id")

	want := []string{"ANGIN MUSIM SEMI BERHEMBUS", "MELALUI KELUARGA YOZAKURA"}
	got := []string{lines[3], lines[4]}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected cue replacement: got %#v want %#v", got, want)
	}
}
