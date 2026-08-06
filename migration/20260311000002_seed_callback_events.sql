-- +goose Up
-- +goose StatementBegin
-- Payout Events
INSERT INTO `callback_events` (`uuid`, `event`, `label`, `event_group`, `is_active`, `created_at`, `updated_at`) VALUES
(UUID(), 'PAYOUT.DONE', 'Payout Done', 'Payout', 1, now(), now()),
(UUID(), 'PAYOUT.PENDING', 'Payout Pending', 'Payout', 1, now(), now()),
(UUID(), 'PAYOUT.DELAYED', 'Payout Delayed', 'Payout', 1, now(), now()),
(UUID(), 'PAYOUT.CANCELLED', 'Payout Cancelled', 'Payout', 1, now(), now()),
(UUID(), 'PAYOUT.SUCCESS', 'Payout Success', 'Payout', 1, now(), now()),
(UUID(), 'PAYOUT.FAILED', 'Payout Failed', 'Payout', 1, now(), now());

-- Payment Events
INSERT INTO `callback_events` (`uuid`, `event`, `label`, `event_group`, `is_active`, `created_at`, `updated_at`) VALUES
(UUID(), 'PAYMENT.VIRTUAL-ACCOUNT.PAID', 'Payment Virtual Account Paid', 'Payment', 1, now(), now()),
(UUID(), 'PAYMENT.CREDIT-CARD.PAID', 'Payment Credit Card Paid', 'Payment', 1, now(), now()),
(UUID(), 'PAYMENT.QRIS-MPM.PAID', 'Payment QRIS MPM Paid', 'Payment', 1, now(), now()),
(UUID(), 'PAYMENT.ACCESS-TOKEN-B2B', 'Access Token B2B', 'Payment', 1, now(), now()),
(UUID(), 'PAYMENT.ACCESS-TOKEN-B2B.TEST', 'Access Token B2B Test', 'Payment', 1, now(), now()),
(UUID(), 'PAYMENT.VIRTUAL-ACCOUNT.TEST', 'Virtual Account Test', 'Payment', 1, now(), now()),
(UUID(), 'PAYMENT.QRIS-MPM.TEST', 'QRIS MPM Test', 'Payment', 1, now(), now());

-- Wallet Events
INSERT INTO `callback_events` (`uuid`, `event`, `label`, `event_group`, `is_active`, `created_at`, `updated_at`) VALUES
(UUID(), 'WALLET.TOP-UP', 'Wallet Top Up', 'Wallet', 1, now(), now()),
(UUID(), 'WALLET.USER-ACTIVATION', 'Wallet User Activation', 'Wallet', 1, now(), now()),
(UUID(), 'WALLET.USER-ACTIVATION.KYC', 'Wallet User Activation KYC', 'Wallet', 1, now(), now()),
(UUID(), 'WALLET.ACCOUNT_LINKAGE.ACTIVATION', 'Wallet Account Linkage Activation', 'Wallet', 1, now(), now()),
(UUID(), 'WALLET.TRANSACTION', 'Wallet Transaction', 'Wallet', 1, now(), now()),
(UUID(), 'WALLET.SNAP.QRIS-MPM', 'Wallet SNAP QRIS MPM', 'Wallet', 1, now(), now()),
(UUID(), 'WALLET.SNAP.DIRECT-DEBIT', 'Wallet SNAP Direct Debit', 'Wallet', 1, now(), now()),
(UUID(), 'WALLET.SNAP.QRIS-MPM.TEST', 'Wallet SNAP QRIS MPM Test', 'Wallet', 1, now(), now()),
(UUID(), 'WALLET.SNAP.DIRECT-DEBIT.TEST', 'Wallet SNAP Direct Debit Test', 'Wallet', 1, now(), now());

-- Refund Events
INSERT INTO `callback_events` (`uuid`, `event`, `label`, `event_group`, `is_active`, `created_at`, `updated_at`) VALUES
(UUID(), 'REFUND.PENDING', 'Refund Pending', 'Refund', 1, now(), now()),
(UUID(), 'REFUND.WAITING_BANK_TRANSFER', 'Refund Waiting Bank Transfer', 'Refund', 1, now(), now()),
(UUID(), 'REFUND.SUCCESS', 'Refund Success', 'Refund', 1, now(), now()),
(UUID(), 'REFUND.FAILED', 'Refund Failed', 'Refund', 1, now(), now()),
(UUID(), 'REFUND.CANCELLED', 'Refund Cancelled', 'Refund', 1, now(), now());

-- Virtual Card Events
INSERT INTO `callback_events` (`uuid`, `event`, `label`, `event_group`, `is_active`, `created_at`, `updated_at`) VALUES
(UUID(), 'VIRTUAL-CARD.NOTIFICATION', 'Virtual Card Notification', 'Virtual Card', 1, now(), now());

-- Merchant Top Up Events
INSERT INTO `callback_events` (`uuid`, `event`, `label`, `event_group`, `is_active`, `created_at`, `updated_at`) VALUES
(UUID(), 'MERCHANT-TOP-UP.SUCCESS', 'Merchant Top Up Success', 'Merchant Top Up', 1, now(), now());

-- Sub Account Registration Events
INSERT INTO `callback_events` (`uuid`, `event`, `label`, `event_group`, `is_active`, `created_at`, `updated_at`) VALUES
(UUID(), 'SUB.ACTIVATION.APPROVED', 'Sub Account Activation Approved', 'Sub Account', 1, now(), now()),
(UUID(), 'SUB.ACTIVATION.PENDING', 'Sub Account Activation Pending', 'Sub Account', 1, now(), now()),
(UUID(), 'SUB.ACTIVATION.REJECTED', 'Sub Account Activation Rejected', 'Sub Account', 1, now(), now()),
(UUID(), 'SUB_ACCOUNT_REGISTRATION.TEST', 'Sub Account Registration Test', 'Sub Account', 1, now(), now());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM `callback_events` WHERE 1=1;
-- +goose StatementEnd
