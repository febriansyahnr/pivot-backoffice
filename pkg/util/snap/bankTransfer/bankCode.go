package bankTransfer

import (
	"strings"
	"sync"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

var (
	bankDB *BankDB
	once   sync.Once
)

type Bank struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	ChannelCode string `json:"channelCode"`
}

type BankDB struct {
	Banks []Bank `json:"banks"`
}

func NewBankDB() *BankDB {
	once.Do(func() {
		banks := []Bank{
			{Code: "002", Name: "BANK RAKYAT INDONESIA", ChannelCode: "BRI"},
			{Code: "003", Name: "Bank Indonesia Eximbank (formerly Bank Ekspor Indonesia)", ChannelCode: "EXIMBANK"},
			{Code: "008", Name: "BANK MANDIRI", ChannelCode: "MANDIRI"},
			{Code: "009", Name: "BANK NEGARA INDONESIA", ChannelCode: "BNI"},
			{Code: "011", Name: "BANK DANAMON INDONESIA", ChannelCode: "DANAMON"},
			{Code: "011", Name: "Bank Danamon UUS", ChannelCode: "DANAMON_UUS"},
			{Code: "013", Name: "BANK PERMATA", ChannelCode: "PERMATA"},
			{Code: "013", Name: "Bank Permata UUS", ChannelCode: "PERMATA_UUS"},
			{Code: "014", Name: "BANK BCA", ChannelCode: "BCA"},
			{Code: "016", Name: "BANK MAYBANK INDONESIA", ChannelCode: "MAYBANK"},
			{Code: "019", Name: "BANK PANIN INDONESIA", ChannelCode: "PANIN"},
			{Code: "020", Name: "Bank Arta Niaga Kencana", ChannelCode: "ARTA_NIAGA_KENCANA"},
			{Code: "022", Name: "BANK CIMB NIAGA", ChannelCode: "CIMB"},
			{Code: "022", Name: "Bank CIMB Niaga Syariah", ChannelCode: "CIMB_SYR"},
			{Code: "022", Name: "Bank CIMB Niaga UUS", ChannelCode: "CIMB_UUS"},
			{Code: "023", Name: "BANK UOB INDONESIA", ChannelCode: "UOB"},
			{Code: "023", Name: "TMRW by UOB Indonesia", ChannelCode: "TMRW"},
			{Code: "026", Name: "Bank Lippo", ChannelCode: "LIPPO"},
			{Code: "028", Name: "BANK OCBC NISP", ChannelCode: "OCBC"},
			{Code: "028", Name: "Bank OCBC NISP UUS", ChannelCode: "OCBC_UUS"},
			{Code: "028", Name: "BPR Danagung Abadi", ChannelCode: "DANAGUNG_ABADI"},
			{Code: "028", Name: "BPR Danagung Bakti", ChannelCode: "DANAGUNG_BAKTI"},
			{Code: "028", Name: "BPR Danagung Ramulti", ChannelCode: "DANAGUNG_RAMULTI"},
			{Code: "030", Name: "American Express Bank LTD", ChannelCode: "AMEX"},
			{Code: "031", Name: "BANK CITIBANK", ChannelCode: "CITIBANK"},
			{Code: "032", Name: "JP Morgan Chase Bank", ChannelCode: "JPMORGAN"},
			{Code: "033", Name: "Bank of America Merill-Lynch", ChannelCode: "BAML"},
			{Code: "034", Name: "Bank ING Indonesia", ChannelCode: "ING"},
			{Code: "036", Name: "Bank China Construction", ChannelCode: "BCC"},
			{Code: "036", Name: "Bank Multicor", ChannelCode: "MULTICOR"},
			{Code: "037", Name: "BANK ARTHA GRAHA INTERNATIONAL", ChannelCode: "ARTHA"},
			{Code: "039", Name: "Bank Credit Agricole Indosuez", ChannelCode: "CREDIT_AGRICOLE"},
			{Code: "040", Name: "The Bangkok Bank Company", ChannelCode: "BANGKOK_BANK"},
			{Code: "042", Name: "BANK MUFG", ChannelCode: "MUFG"},
			{Code: "045", Name: "Bank Sumitomo Mitsui Indonesia", ChannelCode: "SUMITOMO"},
			{Code: "046", Name: "BANK DBS INDONESIA", ChannelCode: "DBS"},
			{Code: "046", Name: "Digibank", ChannelCode: "DIGIBANK"},
			{Code: "047", Name: "Bank Resona Perdania", ChannelCode: "RESONA"},
			{Code: "048", Name: "Bank Mizuho Indonesia", ChannelCode: "MIZUHO"},
			{Code: "050", Name: "BANK STANDARD CHARTERED", ChannelCode: "STANDARD_CHARTERED"},
			{Code: "052", Name: "Bank ABN AMRO", ChannelCode: "ABN_AMRO"},
			{Code: "053", Name: "Bank Keppel Tatlee Buana", ChannelCode: "KEPPEL"},
			{Code: "054", Name: "BANK CAPITAL INDONESIA", ChannelCode: "CAPITAL"},
			{Code: "059", Name: "Korea Exchange Bank Danamon (KEB Indonesia)", ChannelCode: "KEB_INDONESIA"},
			{Code: "061", Name: "BANK ANZ INDONESIA", ChannelCode: "ANZ"},
			{Code: "067", Name: "Deutsche Bank", ChannelCode: "DEUTSCHE"},
			{Code: "068", Name: "Bank Woori Indonesia", ChannelCode: "WOORI"},
			{Code: "069", Name: "BANK OF CHINA", ChannelCode: "BOC"},
			{Code: "076", Name: "BANK BUMI ARTA", ChannelCode: "BUMI_ARTA"},
			{Code: "087", Name: "BANK HSBC", ChannelCode: "HSBC"},
			{Code: "087", Name: "Bank Hongkong and Shanghai Bank Corporation (HSBC) Indonesia (formerly Bank Ekonomi Raharja) UUS", ChannelCode: "HSBC_UUS"},
			{Code: "088", Name: "BANK ANTAR DAERAH", ChannelCode: "ANTARDAERAH"},
			{Code: "089", Name: "BANK RABO", ChannelCode: "RABOBANK"},
			{Code: "089", Name: "Bank Haga", ChannelCode: "HAGA"},
			{Code: "093", Name: "Bank IFI", ChannelCode: "IFI"},
			{Code: "095", Name: "JTRUST BANK", ChannelCode: "JTRUST"},
			{Code: "097", Name: "BANK MAYAPADA", ChannelCode: "MAYAPADA"},
			{Code: "110", Name: "BANK BJB", ChannelCode: "BJB"},
			{Code: "111", Name: "BANK DKI", ChannelCode: "DKI"},
			{Code: "111", Name: "Bank DKI UUS", ChannelCode: "DKI_UUS"},
			{Code: "112", Name: "BPD DIY", ChannelCode: "DAERAH_ISTIMEWA"},
			{Code: "112", Name: "BPD Daerah Istimewa Yogyakarta (DIY) UUS", ChannelCode: "DAERAH_ISTIMEWA_UUS"},
			{Code: "113", Name: "BANK JATENG", ChannelCode: "JAWA_TENGAH"},
			{Code: "113", Name: "BPD Jawa Tengah UUS", ChannelCode: "JAWA_TENGAH_UUS"},
			{Code: "114", Name: "BANK JATIM", ChannelCode: "JAWA_TIMUR"},
			{Code: "114", Name: "BPD Jawa Timur UUS", ChannelCode: "JAWA_TIMUR_UUS"},
			{Code: "115", Name: "BPD JAMBI", ChannelCode: "JAMBI"},
			{Code: "115", Name: "BPD Jambi UUS", ChannelCode: "JAMBI_UUS"},
			{Code: "116", Name: "BPD ACEH", ChannelCode: "ACEH"},
			{Code: "116", Name: "BPD Aceh Syariah", ChannelCode: "ACEH_SYR"},
			{Code: "116", Name: "BPD Aceh UUS", ChannelCode: "ACEH_UUS"},
			{Code: "117", Name: "BPD Sumatera Utara (SUMUT)", ChannelCode: "SUMUT"},
			{Code: "117", Name: "BPD Sumatera Utara (SUMUT) UUS", ChannelCode: "SUMUT_UUS"},
			{Code: "118", Name: "BPD Sumatera Barat (SUMBAR) (Bank Nagari)", ChannelCode: "SUMATERA_BARAT"},
			{Code: "118", Name: "BPD Sumatera Barat (SUMBAR) UUS", ChannelCode: "SUMATERA_BARAT_UUS"},
			{Code: "119", Name: "BPD Riau Kepri Syariah", ChannelCode: "RIAU_DAN_KEPRI_SYR"},
			{Code: "120", Name: "BANK SUMSELBABEL", ChannelCode: "SUMSEL_DAN_BABEL"},
			{Code: "120", Name: "BPD Sumsel Dan Babel UUS", ChannelCode: "SUMSEL_DAN_BABEL_UUS"},
			{Code: "121", Name: "BANK LAMPUNG", ChannelCode: "LAMPUNG"},
			{Code: "122", Name: "BPD KALIMANTAN SELATAN", ChannelCode: "KALIMANTAN_SELATAN"},
			{Code: "122", Name: "BPD Kalimantan Selatan UUS", ChannelCode: "KALIMANTAN_SELATAN_UUS"},
			{Code: "123", Name: "BANK KALIMANTAN BARAT", ChannelCode: "KALIMANTAN_BARAT"},
			{Code: "123", Name: "BPD Kalimantan Barat UUS", ChannelCode: "KALIMANTAN_BARAT_UUS"},
			{Code: "124", Name: "BPD Kaltim Kaltara", ChannelCode: "KALIMANTAN_TIMUR"},
			{Code: "124", Name: "BPD Kaltim Kaltara UUS", ChannelCode: "KALIMANTAN_TIMUR_UUS"},
			{Code: "125", Name: "BANK KALIMANTAN TENGAH", ChannelCode: "KALIMANTAN_TENGAH"},
			{Code: "126", Name: "BPD SULSELBAR", ChannelCode: "SULSELBAR"},
			{Code: "126", Name: "BPD Sulawesi Selatan dan Barat (SULSELBAR) UUS", ChannelCode: "SULSELBAR_UUS"},
			{Code: "127", Name: "BPD Sulut Gorontalo", ChannelCode: "SULUTGO"},
			{Code: "128", Name: "Bank NTB Syariah", ChannelCode: "NTB_SYR"},
			{Code: "129", Name: "BPD BALI", ChannelCode: "BALI"},
			{Code: "130", Name: "BANK NUSA TENGGARA TIMUR", ChannelCode: "NUSA_TENGGARA_TIMUR"},
			{Code: "131", Name: "BANK MALUKU DAN MALUKU UTARA", ChannelCode: "MALUKU"},
			{Code: "132", Name: "BANK PAPUA", ChannelCode: "PAPUA"},
			{Code: "133", Name: "BANK BENGKULU", ChannelCode: "BENGKULU"},
			{Code: "134", Name: "BANK SULAWESI TENGAH", ChannelCode: "SULAWESI"},
			{Code: "135", Name: "BPD SULAWESI TENGGARA", ChannelCode: "SULAWESI_TENGGARA"},
			{Code: "137", Name: "BANK BANTEN", ChannelCode: "BANTEN"},
			{Code: "145", Name: "BANK BNP", ChannelCode: "BNP_PARIBAS"},
			{Code: "145", Name: "Bank Nusantara Parahyangan", ChannelCode: "NUSANTARA_PARAHYANGAN"},
			{Code: "146", Name: "BANK OF INDIA INDONESIA", ChannelCode: "INDIA"},
			{Code: "147", Name: "BANK MUAMALAT", ChannelCode: "MUAMALAT"},
			{Code: "151", Name: "BANK MESTIKA DHARMA", ChannelCode: "MESTIKA_DHARMA"},
			{Code: "152", Name: "BANK SHINHAN", ChannelCode: "SHINHAN"},
			{Code: "153", Name: "BANK SINARMAS", ChannelCode: "SINARMAS"},
			{Code: "153", Name: "Bank Sinarmas UUS", ChannelCode: "SINARMAS_UUS"},
			{Code: "157", Name: "BANK MASPION", ChannelCode: "MASPION"},
			{Code: "159", Name: "Bank Hagakita", ChannelCode: "HAGAKITA"},
			{Code: "161", Name: "BANK GANESHA", ChannelCode: "GANESHA"},
			{Code: "162", Name: "Bank Windu Kentjana", ChannelCode: "WINDU_KENTJANA"},
			{Code: "164", Name: "BANK ICBC INDONESIA", ChannelCode: "ICBC"},
			{Code: "166", Name: "Bank Harmoni International", ChannelCode: "HARMONI"},
			{Code: "167", Name: "BANK QNB INDONESIA", ChannelCode: "QNB_INDONESIA"},
			{Code: "200", Name: "BANK TABUNGAN NEGARA", ChannelCode: "BTN"},
			{Code: "212", Name: "BANK WOORI SAUDARA", ChannelCode: "WOORI_SAUDARA"},
			{Code: "212", Name: "Bank Himpunan Saudara 1906", ChannelCode: "HIMPUNAN_SAUDARA"},
			{Code: "213", Name: "PT Bank SMBC Indonesia Tbk", ChannelCode: "SMBC_INDONESIA"},
			{Code: "213", Name: "PT BANK BTPN TBK", ChannelCode: "TABUNGAN_PENSIUNAN_NASIONAL"},
			{Code: "213", Name: "Jenius", ChannelCode: "JENIUS"},
			{Code: "333", Name: "KOP INTIDANA", ChannelCode: "KOP_INTIDANA"},
			{Code: "405", Name: "Bank Syariah Nasional", ChannelCode: "BSN_UUS"},
			{Code: "405", Name: "BANK VICTORIA SYARIAH", ChannelCode: "VICTORIA_SYR"},
			{Code: "405", Name: "Bank Swaguna", ChannelCode: "SWAGUNA"},
			{Code: "405", Name: "Bank Tabungan Negara (BTN) UUS", ChannelCode: "BTN_UUS"},
			{Code: "422", Name: "Bank Syariah Indonesia (formerly BRI Syariah)", ChannelCode: "BRI_SYR"},
			{Code: "425", Name: "BANK BJB SYARIAH", ChannelCode: "BJB_SYR"},
			{Code: "426", Name: "BANK MEGA", ChannelCode: "MEGA"},
			{Code: "427", Name: "Bank Syariah Indonesia (formerly BNI SYARIAH)", ChannelCode: "BNI_SYR"},
			{Code: "441", Name: "Bank KB Indonesia", ChannelCode: "KBID"},
			{Code: "451", Name: "BANK SYARIAH INDONESIA", ChannelCode: "BSI"},
			{Code: "459", Name: "Krom Bank Indonesia", ChannelCode: "KROM"},
			{Code: "466", Name: "Bank Andara (formerly Bank Sri Partha)", ChannelCode: "ANDARA"},
			{Code: "472", Name: "Bank Saqu Indonesia", ChannelCode: "SAQU"},
			{Code: "484", Name: "BANK KEB HANA", ChannelCode: "HANA"},
			{Code: "485", Name: "BANK MNC INTERNATIONAL", ChannelCode: "MNC_INTERNASIONAL"},
			{Code: "490", Name: "BANK NEO COMMERCE", ChannelCode: "BNC"},
			{Code: "491", Name: "Bank Mitraniaga", ChannelCode: "MITRANIAGA"},
			{Code: "494", Name: "BANK RAYA", ChannelCode: "AGRONIAGA"},
			{Code: "498", Name: "BANK SBI INDONESIA", ChannelCode: "SBI_INDONESIA"},
			{Code: "501", Name: "BANK BCA DIGITAL", ChannelCode: "BCA_DIGITAL"},
			{Code: "501", Name: "Bank Royal Indonesia", ChannelCode: "ROYAL"},
			{Code: "503", Name: "BANK NATIONAL NOBU", ChannelCode: "NATIONALNOBU"},
			{Code: "506", Name: "BANK MEGA SYARIAH", ChannelCode: "MEGA_SYR"},
			{Code: "513", Name: "BANK INA PERDANA", ChannelCode: "INA_PERDANA"},
			{Code: "517", Name: "PANIN DUBAI SYARIAH"}, // ?
			{Code: "520", Name: "BANK PRIMA MASTER", ChannelCode: "PRIMA_MASTER"},
			{Code: "521", Name: "KB BUKOPIN SYARIAH", ChannelCode: "BUKOPIN_SYR"},
			{Code: "521", Name: "Bank Persyarikatan Indonesia", ChannelCode: "PERSYARIKATAN"},
			{Code: "523", Name: "BANK SAHABAT SAMPOERNA", ChannelCode: "SAHABAT_SAMPOERNA"},
			{Code: "525", Name: "BANK BARCLAYS (formerly Bank Barclays)", ChannelCode: "BARCLAYS"},
			{Code: "526", Name: "BANK OKE", ChannelCode: "OKE"},
			{Code: "526", Name: "Bank Dinar Indonesia", ChannelCode: "DINAR_INDONESIA"},
			{Code: "531", Name: "BANK AMAR INDONESIA", ChannelCode: "AMAR"},
			{Code: "531", Name: "Anglomas Internasional Bank", ChannelCode: "ANGLOMAS"},
			{Code: "535", Name: "SEABANK", ChannelCode: "SEABANK"},
			{Code: "536", Name: "BANK BCA SYARIAH", ChannelCode: "BCA_SYR"},
			{Code: "542", Name: "BANK JAGO", ChannelCode: "JAGO"},
			{Code: "547", Name: "BTPN SYARIAH", ChannelCode: "BTPN_SYARIAH"},
			{Code: "548", Name: "Bank Multi Arta Sentosa", ChannelCode: "MULTI_ARTA_SENTOSA"},
			{Code: "548", Name: "Bank Multiarta Sentosa (MAS)", ChannelCode: "MAS"},
			{Code: "553", Name: "Bank Hibank Indonesia", ChannelCode: "HIBANK"},
			{Code: "555", Name: "BANK INDEX", ChannelCode: "INDEX_SELINDO"},
			{Code: "558", Name: "Bank Pundi (formerly Bank Eksekutif)", ChannelCode: "PUNDI"},
			{Code: "559", Name: "BANK Centratama Nasional Bank (CNB)", ChannelCode: "CNB"},
			{Code: "562", Name: "Super Bank Indonesia", ChannelCode: "SUPERBANK"},
			{Code: "564", Name: "Bank Mandiri Taspen Pos (Bank Sinar Harapan Bali)", ChannelCode: "MANDIRI_TASPEN"},
			{Code: "566", Name: "BANK VICTORIA", ChannelCode: "VICTORIA_INTERNASIONAL"},
			{Code: "567", Name: "ALLO BANK", ChannelCode: "HARDA_INTERNASIONAL"},
			{Code: "600", Name: "BPR Supra Artapersada", ChannelCode: "SUPRA"},
			{Code: "608", Name: "MANDIRI - BPR", ChannelCode: "MANDIRI_BPR"},
			{Code: "688", Name: "BPR KS (Karyajatnika Sedaya)", ChannelCode: "KS"},
			{Code: "770", Name: "ALTO PAY"}, // ?
			{Code: "775", Name: "SGOD PAY"}, // ?
			{Code: "777", Name: "FINNET"},   // ?
			{Code: "789", Name: "IMKAS"},    // ?
			{Code: "800", Name: "ALTOCASH"}, // ?
			{Code: "808", Name: "XL TUNAI"}, // ?
			{Code: "867", Name: "BANK Eka Bumi Artha", ChannelCode: "EKA"},
			{Code: "886", Name: "ALTOPAY"},            // ?
			{Code: "888", Name: "ISO"},                // ?
			{Code: "889", Name: "BPR KRYJATNIKA SDY"}, // ?
			{Code: "898", Name: "TELKOM"},             // ?
			{Code: "899", Name: "DOKU"},               // ?
			{Code: "920", Name: "ATMI"},               // ?
			{Code: "921", Name: "ATMI-EN"},            // ?
			{Code: "945", Name: "IBK BANK", ChannelCode: "IBK"},
			{Code: "945", Name: "Bank Agris (Bank Finconesia)", ChannelCode: "AGRIS"},
			{Code: "946", Name: "Bank Merincorp", ChannelCode: "MERINCORP"},
			{Code: "947", Name: "BANK ALADIN SYARIAH", ChannelCode: "ALADIN"},
			{Code: "948", Name: "Bank OCBC Indonesia", ChannelCode: "OCBC_INDONESIA"},
			{Code: "949", Name: "Bank CTBC (China Trust) Indonesia", ChannelCode: "CTBC"},
			{Code: "950", Name: "BANK COMMONWEALTH", ChannelCode: "COMMONWEALTH"},
			{Code: "972", Name: "PANIN SYARIAH", ChannelCode: "PANIN_SYR"},
			{Code: "987", Name: "ATMBPLUS"},           // ?
			{Code: "998", Name: "DIGITAL SOLUSI PTM"}, // ?
			// WALLET CODE
			{Code: "GOPAY", Name: "GoPay", ChannelCode: "GOPAY"},
			{Code: "OVO", Name: "OVO", ChannelCode: "OVO"},
			{Code: "SHOPEEPAY", Name: "ShopeePay", ChannelCode: "SHOPEEPAY"},
			{Code: "DANA", Name: "DANA", ChannelCode: "DANA"},
			{Code: "LINKAJA", Name: "LinkAja", ChannelCode: "LINKAJA"},
		}

		bankDB = &BankDB{Banks: banks}
	})

	return bankDB
}

// List is a function to get all banks
func (db *BankDB) List() []Bank {
	return db.Banks
}

// FindByCode is a function to find a bank by its code
func (db *BankDB) FindByCode(code string) *Bank {
	for _, bank := range db.Banks {
		if bank.Code == code {
			return &bank
		}
	}
	return nil
}

// FindByName is a function to find a bank by its name
func (db *BankDB) FindByName(name string) *Bank {
	for _, bank := range db.Banks {
		if strings.Contains(strings.ToLower(bank.Name), strings.ToLower(name)) {
			return &bank
		}
	}
	return nil
}

// FindByChannelCode is a function to find a bank by its code
func (db *BankDB) FindByChannelCode(channelCode string) *Bank {
	for _, bank := range db.Banks {
		if bank.ChannelCode == channelCode {
			return &bank
		}
	}
	return nil
}

// func IsWalletCode(code string) bool {
// 	return strings.EqualFold(code, "GOPAY") ||
// 		strings.EqualFold(code, "OVO") ||
// 		strings.EqualFold(code, "SHOPEEPAY") ||
// 		strings.EqualFold(code, "DANA") ||
// 		strings.EqualFold(code, "LINKAJA") ||
// 		strings.EqualFold(code, "911")
// }

func GetStatusFromLatestTransactionCode(code string) string {
	var status string
	switch code {
	// Mark as success
	case "00":
		status = constant.SnapCoreBankTransferStatusSuccess
	// Mark as pending
	case "01", "02", "03":
		status = constant.SnapCoreBankTransferStatusPending
	// Mark as failed
	case "04", "05", "06", "07":
		status = constant.SnapCoreBankTransferStatusFailed
	default:
		status = constant.SnapCoreBankTransferStatusPending
	}
	return status
}
