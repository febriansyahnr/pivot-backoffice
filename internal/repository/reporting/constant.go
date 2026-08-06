package reportingRepository

const balanceHistorySelectQuery = `SELECT
	transaction_id AS uuid, IF(reference_id != '', reference_id, '-') AS merchant_reference_id,
	transaction_type AS type, channel, settlement_status AS status, source_created_by AS created_by, status_updated_at AS updated_at,
	currency, IF(amount < 0, ABS(amount), 0) AS debit, IF(amount > 0, amount, 0) AS credit, fee,
	settlement_at, settlement_model,
	CASE
		WHEN balance_type = 'DISBURSEMENT' THEN 'Payout Balance'
		WHEN balance_type = 'PAYMENT' THEN 'Payment Balance'
		WHEN balance_type = 'WALLET' THEN 'Wallet Balance'
		WHEN balance_type = 'VIRTUAL_TERMINAL' THEN 'Virtual Terminal Balance'
		ELSE '-'
	END AS balance_type,
	IFNULL(additional_info->>'$.bankReferenceNo', '') AS bank_reference,
	IFNULL(additional_info->>'$.beneficiaryBankName', '') AS beneficiary_bank_name,
	IFNULL(additional_info->>'$.beneficiaryAccountNo', '') AS beneficiary_account_no,
	IFNULL(additional_info->>'$.beneficiaryName', '') AS beneficiary_account_name,
	source_id AS reference_id, reason_type, reason_description, remarks, created_at
FROM
	report_balance_histories`
