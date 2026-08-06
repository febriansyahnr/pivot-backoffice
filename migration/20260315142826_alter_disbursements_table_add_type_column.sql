-- +goose Up
-- +goose StatementBegin
ALTER TABLE disbursements ADD COLUMN `type` VARCHAR(50) GENERATED ALWAYS AS (
	CASE
		WHEN metadata->>'$.xbDetail' IS NOT NULL THEN 'INTERNATIONAL_PAYOUT'
		WHEN metadata->>'$.cardFundedDetail' IS NOT NULL THEN 'CARD_FUNDED_PAYOUT'
		ELSE 'LOCAL_PAYOUT'
	END
) VIRTUAL AFTER purpose_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE disbursements DROP COLUMN `type`;
-- +goose StatementEnd
