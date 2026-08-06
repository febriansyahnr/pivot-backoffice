package constant

import "time"

const (
	EventMerchantCreated = "MERCHANT.CREATED"
	EventMerchantUpdated = "MERCHANT.UPDATED"
)

const (
	MerchantStatusActive      = "ACTIVE"
	MerchantStatusInactive    = "INACTIVE"    // deprecated due kyc feature
	MerchantStatusClosed      = "CLOSED"      // Account has been closed due to main merchant request
	MerchantStatusBlocked     = "BLOCKED"     // Account Blocked by Harsya due to any terms and policy violation
	MerchantStatusDormant     = "DORMANT"     // Account has not been have transaction in any balance (account) for a period of time
	MerchantStatusCreated     = "CREATED"     // Account created and still in inactive mode
	MerchantStatusDeactivated = "DEACTIVATED" // Main merchant intentionally deactivated the submerchant
)

const (
	MerchantStatusByIDCacheKey = "backend-portal:merchant:%s:status"
)

const (
	MerchantStatusField       = "status"
	MerchantReasonStatusField = "reason_status"
)

const (
	MerchantKYCTypeKYC    = "KYC"
	MerchantKYCTypeNonKYC = "NON_KYC"

	MerchantPICNotInvited   = "NOT_INVITED"
	MerchantPICInvited      = UserStatusInvited
	MerchantPICActive       = UserStatusActive
	MerchantTypeMerchant    = "MERCHANT"
	MerchantTypeSubMerchant = "SUB_MERCHANT"
)

const (
	KYCStatusWaitingForDocument = "WAITING_FOR_DOCUMENT"
	KYCStatusInReview           = "IN_REVIEW"
	KYCStatusRejected           = "REJECTED"
	KYCStatusApproved           = "APPROVED"
	KYCStatusNeedResubmission   = "NEED_RESUBMISSION"
	KYCStatusNotRequired        = "NOT_REQUIRED"
)

const (
	// reserved for QR Registration
	MerchantDocumentTypeOwnerNationalIdentityCard  = "NationalIdentityCard"       // KTP Pemilik
	MerchantDocumentTypeBusinessLicense            = "BusinessLicense"            // NIB, SIUP, atau lisensi perdagangan secara umum
	MerchantDocumentTypeTaxIdentification          = "TaxIdentification"          // NPWP Perusahaan
	MerchantDocumentTypeBusinessRegistration       = "BusinessRegistration"       // TDP (Tanda Daftar Perusahaan) / versi baru NIB
	MerchantDocumentTypeCertificateIncorporation   = "CertificateIncorporation"   // Akta Pendirian
	MerchantDocumentTypeCertificateNo40            = "CertificateNo40"            // Akta UU No. 40 Tahun 2007
	MerchantDocumentTypeCertificateLastAmendment   = "CertificateLastAmendment"   // Akta Perubahan Terakhir
	MerchantDocumentTypeCertificateDeedAmendment   = "CertificateDeedAmendment"   // Akta Perubahan
	MerchantDocumentTypeCertificateAmendmentAct    = "CertificateAmendmentAct"    // Akta Perubahan sesuai ketentuan hukum
	MerchantDocumentTypeCertificateEstablishment   = "CertificateEstablishment"   // Salinan Akta Pendirian
	MerchantDocumentTypeCertificateTaxRegistration = "CertificateTaxRegistration" // SKT (Surat Keterangan Terdaftar) Pajak
	MerchantDocumentTypeBusinessEnvironmentPhoto   = "BusinessEnvironmentPhoto"   // Foto Lingkungan Usaha

	MerchantDocumentTypeBusinessProfile                  = "BusinessProfile"                  // Profil Perusahaan & Dokumen Deskripsi & flow business perusahaan
	MerchantDocumentTypeRegulatoryApprovalLicense        = "RegulatoryApprovalLicense"        // Perizinan OJK, Kominfo, BPOM, dll
	MerchantDocumentTypeAdminNationalIdentityCard        = "AdminNationalIdentityCard"        // KTP Admin
	MerchantDocumentTypeShareholderIdentityCard          = "ShareholderIdentityCard"          // KTP Pemegang Saham
	MerchantDocumentTypeOwnerTaxIdentification           = "OwnerTaxIdentification"           // NPWP Pemilik
	MerchantDocumentTypeShareholderStructureDeed         = "ShareholderStructureDeed"         // Akta Susunan Pemegang Saham
	MerchantDocumentTypeCertificateIncorporationApproval = "CertificateIncorporationApproval" // Surat Persetujuan Kementerian Hukum dan Hak Asasi Manusia Republik Indonesia (Kemenkumham)
	MerchantDocumentTypeBoardOfManagement                = "BoardOfManagement"                // Susunan Pengurus

	// Information That this document is not required
	MerchantDocumentTypeDirectorNationalIdentityCard = "DirectorNationalIdentityCard" // KTP Direksi
	MerchantDocumentTypePlaceOfEstablishment         = "PlaceOfEstablishment"         // Tempat Usaha
	MerchantDocumentTypeDateOfEstablishment          = "DateOfEstablishment"          // Tanggal Pendirian
)

const (
	MerchantBODPositionDirector              = "Director"
	MerchantBODPositionPresidentDirector     = "President Director"
	MerchantBODPositionCommissioner          = "Commissioner"
	MerchantBODPositionPresidentCommissioner = "President Commissioner"
	MerchantBODPositionShareholder           = "Shareholder"
)

const (
	MerchantDocumentStatusNotSubmitted = "NOT_SUBMITTED"
	MerchantDocumentStatusSubmitted    = "SUBMITTED"
)

const (
	MerchantRiskLevelLow     = "LOW"
	MerchantRiskLevelLowMid  = "LOW_MID"
	MerchantRiskLevelMid     = "MID"
	MerchantRiskLevelMidHigh = "MID_HIGH"
	MerchantRiskLevelHigh    = "HIGH"
)

var ValidMerchantRiskLevels = []string{
	MerchantRiskLevelLow,
	MerchantRiskLevelLowMid,
	MerchantRiskLevelMid,
	MerchantRiskLevelMidHigh,
	MerchantRiskLevelHigh,
}

func IsValidRiskLevel(riskLevel string) bool {
	if riskLevel == "" {
		return true
	}

	for _, validLevel := range ValidMerchantRiskLevels {
		if riskLevel == validLevel {
			return true
		}
	}
	return false
}

const (
	MerchantBulkCreateSubMerchantSessionIDCacheKey = "backend-portal:sub-merchant:%s:bulk-create:%s" // $1 is merchant_id and $2 is session_id
	MerchantBulkCreateSubMerchantSessionIDCacheTTL = 24 * time.Hour
	MerchantReservedShortNameKey                   = "backend-portal:reserved-short-name"
)

const (
	MerchantReservedShortNameDefaultFileName = "active_reserved_shortname.xlsx"
	MerchantReservedShortNameBackupFileName  = "backup_reserved_shortname_%s.xlsx" // $date
)
