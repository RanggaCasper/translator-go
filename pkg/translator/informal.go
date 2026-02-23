package translator

import (
	"regexp"
	"strings"
)

// FormalizeToInformal converts formal Indonesian text to informal style
func FormalizeToInformal(text string) string {
	if text == "" {
		return text
	}

	// Dictionary of formal -> informal replacements
	replacements := map[string]string{
		// Pronouns
		`\bAnda semua\b`: "kalian",
		`\bAnda\b`:       "kamu",
		`\bIa\b`:         "dia",
		`\bDia\b`:        "Dia",
		`\bSaya\b`:       "Aku",
		`\bsaya\b`:       "aku",
		`\bKami\b`:       "Kita",
		`\bkami\b`:       "kita",

		// Verbs
		`\bmengatakan\b`:    "bilang",
		`\bmemberikan\b`:    "kasih",
		`\bmembuat\b`:       "bikin",
		`\bmemakan\b`:       "makan",
		`\bmeminum\b`:       "minum",
		`\bmengambil\b`:     "ambil",
		`\bmelihat\b`:       "lihat",
		`\bmendengar\b`:     "denger",
		`\bmendengarkan\b`:  "dengerin",
		`\bmembawa\b`:       "bawa",
		`\bmenggunakan\b`:   "pake",
		`\bmelakukan\b`:     "lakuin",
		`\bmendapatkan\b`:   "dapet",
		`\bmenemukan\b`:     "nemuin",
		`\bmencari\b`:       "cari",
		`\bmenunggu\b`:      "tunggu",
		`\bmembantu\b`:      "bantu",
		`\bmembeli\b`:       "beli",
		`\bmenjual\b`:       "jual",
		`\bmengerti\b`:      "ngerti",
		`\bmengetahui\b`:    "tau",
		`\bmemahami\b`:      "paham",
		`\bmenjadi\b`:       "jadi",
		`\bmemiliki\b`:      "punya",
		`\bmeninggalkan\b`:  "tinggalin",
		`\bmengikuti\b`:     "ikutin",
		`\bmenunjukkan\b`:   "tunjukin",
		`\bmenceritakan\b`:  "ceritain",
		`\bmenjelaskan\b`:   "jelasin",
		`\bmeminta\b`:       "minta",
		`\bmenawarkan\b`:    "nawarin",
		`\bmengirim\b`:      "kirim",
		`\bmenghubungi\b`:   "hubungin",
		`\bmemutuskan\b`:    "mutusin",
		`\bmemperbaiki\b`:   "benerin",
		`\bmemulai\b`:       "mulai",
		`\bmelanjutkan\b`:   "lanjut",
		`\bmenyelesaikan\b`: "selesain",
		`\bmenyiapkan\b`:    "siapin",
		`\bmenyuruh\b`:      "nyuruh",
		`\bmenyampaikan\b`:  "nyampein",
		`\bmenanyakan\b`:    "nanyain",

		// Common phrases
		`\btidak ada\b`:    "nggak ada",
		`\bTidak ada\b`:    "Nggak ada",
		`\bapakah\b`:       "",
		`\btetapi\b`:       "tapi",
		`\bsedang\b`:       "lagi",
		`\bakan\b`:         "mau",
		`\btelah\b`:        "udah",
		`\bsudah\b`:        "udah",
		`\bbelum\b`:        "belom",
		`\btidak\b`:        "nggak",
		`\bTidak\b`:        "Nggak",
		`\bkemudian\b`:     "terus",
		`\bseperti\b`:      "kayak",
		`\bbagaimana\b`:    "gimana",
		`\bmengapa\b`:      "kenapa",
		`\bdi mana\b`:      "dimana",
		`\bhanya\b`:        "cuma",
		`\bkarena\b`:       "soalnya",
		`\bnamun\b`:        "tapi",
		`\bagar\b`:         "biar",
		`\buntuk\b`:        "buat",
		`\bkepada\b`:       "ke",
		`\bsegera\b`:       "cepet",
		`\bselalu\b`:       "terus",
		`\bterlalu\b`:      "kelewat",
		`\bbenar\b`:        "bener",
		`\bbenarkah\b`:     "beneran?",
		`\bterima kasih\b`: "makasih",
		`\bTerima kasih\b`: "Makasih",
		`\bpermisi\b`:      "eh",
		`\btersebut\b`:     "itu",
		`\bberkata\b`:      "bilang",
		`\bberbicara\b`:    "ngomong",
		`\bberjalan\b`:     "jalan",
		`\bberlari\b`:      "lari",
		`\bberusaha\b`:     "usaha",
		`\bberpikir\b`:     "mikir",
		`\bbertemu\b`:      "ketemu",
		`\bberhenti\b`:     "berhenti",
		`\bberangkat\b`:    "pergi",

		// Time / connectors
		`\bselanjutnya\b`:  "abis ini",
		`\bsebelumnya\b`:   "tadi",
		`\bsebetulnya\b`:   "sebenernya",
		`\bsebenarnya\b`:   "sebenernya",
		`\bbarangkali\b`:   "mungkin",
		`\bseharusnya\b`:   "harusnya",
		`\bsebaiknya\b`:    "mending",
		`\bsilakan\b`:      "coba",
		`\bdipersilakan\b`: "silakan",
		`\bdimohon\b`:      "tolong",
		`\bharap\b`:        "tolong",
		`\bapabila\b`:      "kalau",
		`\bjika\b`:         "kalau",
		`\bdikarenakan\b`:  "soalnya",
		`\bdapat\b`:        "bisa",
		`\btidak dapat\b`:  "nggak bisa",

		// Emotional expressions
		`\baku tidak tahu\b`:      "aku nggak tau",
		`\bsaya tidak tahu\b`:     "aku nggak tau",
		`\baku tidak mengerti\b`:  "aku nggak ngerti",
		`\bsaya tidak mengerti\b`: "aku nggak ngerti",
		`\btidak mungkin\b`:       "nggak mungkin",
		`\bini tidak mungkin\b`:   "ini nggak mungkin",
		`\btidak apa-apa\b`:       "nggak apa-apa",
		`\btidak masalah\b`:       "nggak masalah",
		`\btidak peduli\b`:        "nggak peduli",
		`\btidak usah\b`:          "nggak usah",
		`\btidak perlu\b`:         "nggak perlu",
	}

	result := text
	for formal, informal := range replacements {
		re := regexp.MustCompile(formal)
		result = re.ReplaceAllString(result, informal)
	}

	return result
}

// SingleLine ensures output is single line (for VTT)
func SingleLine(s string) string {
	re := regexp.MustCompile(`\s*\n+\s*`)
	s = re.ReplaceAllString(s, " ")

	re = regexp.MustCompile(`[ \t]+`)
	s = re.ReplaceAllString(s, " ")

	return strings.TrimSpace(s)
}
