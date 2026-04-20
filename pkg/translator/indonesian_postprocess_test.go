package translator

import (
	"strings"
	"testing"
)

func TestEnhanceIndonesianSubtitle_PrefixAndPunctuation(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "hyphen prefix uses next word letters",
			in:   "Th-Makasih",
			want: "Ma-Makasih",
		},
		{
			name: "single letter prefix follows next word",
			in:   "T-Nggak",
			want: "N-Nggak",
		},
		{
			name: "prefix follows artinya",
			in:   "I-Artinya",
			want: "A-Artinya",
		},
		{
			name: "prefix follows ayo",
			in:   "L-Ayo",
			want: "A-Ayo",
		},
		{
			name: "prefix with trailing sentence",
			in:   "I-Artinya, kesan seseorang terhadap sesuatu!",
			want: "A-Artinya, kesan seseorang terhadap sesuatu!",
		},
		{
			name: "prefix token after punctuation",
			in:   "Hmm, L-Ayo sekarang!",
			want: "Hmm, A-Ayo sekarang!",
		},
		{
			name: "split ellipsis and quote spacing",
			in:   "Nggak ada. Hanya saja. ..",
			want: "Nggak ada. Hanya saja...",
		},
		{
			name: "joint punctuation",
			in:   "Apa? !",
			want: "Apa?!",
		},
		{
			name: "ellipsis then question mark at line end",
			in:   "Kalau begitu. .. Bolehkah aku minta satu permintaan lagi dari\n?",
			want: "Kalau begitu... Bolehkah aku minta satu permintaan lagi dari?",
		},
		{
			name: "closing quote spacing",
			in:   "\"Ikatan Datang dalam Segala Bentuk. \"",
			want: "\"Ikatan Datang dalam Segala Bentuk.\"",
		},
		{
			name: "stutter capitalizes next word",
			in:   "Oh... Haruskah kamu pergi?",
			want: "Ha... Haruskah kamu pergi?",
		},
		{
			name: "stutter capitalizes merged next word",
			in:   "Ini...semua cuma ilusi!",
			want: "Sem... Semua cuma ilusi!",
		},
		{
			name: "does not force terminal period",
			in:   "Yang sepi",
			want: "Yang sepi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnhanceIndonesianSubtitle(tt.in)
			if got != tt.want {
				t.Fatalf("EnhanceIndonesianSubtitle(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestPostProcessSubtitleContent_IndonesianCachedVTT(t *testing.T) {
	in := "WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nI-Artinya, kesan seseorang terhadap sesuatu!\n\n00:00:01.000 --> 00:00:02.000\nL-Ayo mulai kelasnya!\n"
	got := PostProcessSubtitleContent(in, "id")

	if !strings.Contains(got, "A-Artinya, kesan seseorang terhadap sesuatu!") {
		t.Fatalf("expected I-Artinya to be normalized, got: %q", got)
	}

	if !strings.Contains(got, "A-Ayo mulai kelasnya!") {
		t.Fatalf("expected L-Ayo to be normalized, got: %q", got)
	}
}

func TestPostProcessSubtitleContent_PunctuationLines(t *testing.T) {
	input := "WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nKejar dia! Dia punya pecahan Permata Suci.\n!\n\n00:00:01.000 --> 00:00:02.000\nDi tempat terakhir itu, aku mencium bau tinta.\n.\n\n00:00:02.000 --> 00:00:03.000\nDia mencoba untuk punya Permata Suci.\n...\n\n00:00:03.000 --> 00:00:04.000\nNggak ada wajah dalam refleksi.\n! Inuyasha!\n"
	got := PostProcessSubtitleContent(input, "id")

	if !strings.Contains(got, "Kejar dia! Dia punya pecahan Permata Suci.!") {
		t.Fatalf("expected exclamation punctuation to append to previous line, got: %q", got)
	}

	if !strings.Contains(got, "Di tempat terakhir itu, aku mencium bau tinta.") {
		t.Fatalf("expected trailing dot line to be removed, got: %q", got)
	}

	if strings.Contains(got, "Dia mencoba untuk punya Permata Suci....") {
		t.Fatalf("expected ellipsis-only line to be removed, got: %q", got)
	}

	if strings.Contains(got, "! Inuyasha!") {
		t.Fatalf("expected leading punctuation on next line to be stripped, got: %q", got)
	}

	if !strings.Contains(got, "Inuyasha!") {
		t.Fatalf("expected final line to remain as Inuyasha!, got: %q", got)
	}
}

func TestPostProcessSubtitleContent_DropsLongCueBlocks(t *testing.T) {
	input := "WEBVTT\n\n00:00:00.000 --> 00:00:02.480 line:20%\nDeserted Island Survival\nDays\nJuly 19: Set out\nJuly 20-August 3: Special test\nAugust 4-10: Cruise (free time)\nAugust 11: Return, activity ends\n\n00:00:02.500 --> 00:00:03.000\nHalo!\n"
	got := PostProcessSubtitleContent(input, "id")

	if strings.Contains(got, "Deserted Island Survival") {
		t.Fatalf("expected long cue block to be removed, got: %q", got)
	}

	if strings.Contains(got, "July 19: Set out") {
		t.Fatalf("expected long cue block dates to be removed, got: %q", got)
	}

	if !strings.Contains(got, "Halo!") {
		t.Fatalf("expected remaining short cue to stay, got: %q", got)
	}
}

func TestPostProcessSubtitleContent_DropsOverWordCueBlocks(t *testing.T) {
	input := "WEBVTT\n\n00:00:36.390 --> 00:00:40.390 line:20%\nPoin kelas yang diperoleh oleh tiga kelompok teratas akan ditransfer dari tahun-tahun tiga kelompok terbawah. Poin kelas akan dibagi rata antar kelas dalam grup, berapa pun jumlah anggotanya.\n\n00:00:40.500 --> 00:00:41.500\nIni tetap ada.\n"
	got := PostProcessSubtitleContent(input, "id")

	if strings.Contains(got, "Poin kelas yang diperoleh") {
		t.Fatalf("expected over-word cue block to be removed, got: %q", got)
	}

	if !strings.Contains(got, "Ini tetap ada.") {
		t.Fatalf("expected short cue to remain, got: %q", got)
	}
}

func TestPostProcessSubtitleContent_DropsSymbolOnlyLines(t *testing.T) {
	input := "WEBVTT\n\n00:00:00.000 --> 00:00:02.480 line:20%\nKelihatannya bukan kasus terburuk yang mungkin terjadi pada.\n,\n.\n!\n/\n+\n-\n\n00:00:02.500 --> 00:00:03.000\nHalo!\n"
	got := PostProcessSubtitleContent(input, "id")

	if strings.Contains(got, ",\n") || strings.Contains(got, ".\n") || strings.Contains(got, "/\n") || strings.Contains(got, "+\n") || strings.Contains(got, "-\n") {
		t.Fatalf("expected symbol-only lines to be removed, got: %q", got)
	}

	if !strings.Contains(got, "Kelihatannya bukan kasus terburuk yang mungkin terjadi pada.") {
		t.Fatalf("expected normal subtitle text to remain, got: %q", got)
	}

	if !strings.Contains(got, "Halo!") {
		t.Fatalf("expected short cue to remain, got: %q", got)
	}
}

func TestPostProcessSubtitleContent_PreservesFormattingTags(t *testing.T) {
	input := "WEBVTT\n\n00:00:00.000 --> 00:00:01.000\n<I>Kekayaan, ketenaran, kekuasaan...</I>\n\n00:00:01.000 --> 00:00:02.000\n<b>Semua itu pernah dimiliki satu orang.</b>\n"
	got := PostProcessSubtitleContent(input, "id")

	if !strings.Contains(got, "<i>Kekayaan, ketenaran, kekuasaan...</i>") {
		t.Fatalf("expected italic tags to be preserved and normalized, got: %q", got)
	}

	if !strings.Contains(got, "<b>Semua itu pernah dimiliki satu orang.</b>") {
		t.Fatalf("expected bold tags to be preserved, got: %q", got)
	}
}

func TestPostProcessSubtitleContent_RepairsBrokenOpeningFormattingTag(t *testing.T) {
	input := "WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nI>Kekayaan, ketenaran, kekuasaan...</I>\n"
	got := PostProcessSubtitleContent(input, "id")

	if !strings.Contains(got, "<i>Kekayaan, ketenaran, kekuasaan...</i>") {
		t.Fatalf("expected broken opening tag to be repaired, got: %q", got)
	}
}

func TestLooksUntranslatedEnglish(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{
			name: "english sentence",
			in:   "If you need some help, y'know...",
			want: true,
		},
		{
			name: "mixed with english majority",
			in:   "I'd be willing to grup Bersamamu, kamu tahu?",
			want: true,
		},
		{
			name: "indonesian sentence",
			in:   "Aku akan dengan senang hati meminjamkanmu bantuan.",
			want: false,
		},
		{
			name: "single english token only",
			in:   "Classroom of the Elite",
			want: false,
		},
		{
			name: "english sentence with function words",
			in:   "May the high bishop bless the birth Of these new couples.",
			want: true,
		},
		{
			name: "short english sentence still detected",
			in:   "Do it a bit extravagantly.",
			want: true,
		},
		{
			name: "mixed english and indonesian line",
			in:   "The supreme gods that preside over Langit tak berujung",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := looksUntranslatedEnglish(tt.in)
			if got != tt.want {
				t.Fatalf("looksUntranslatedEnglish(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
