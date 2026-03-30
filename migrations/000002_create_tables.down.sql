-- +migrate Down
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS refunds;
DROP TABLE IF EXISTS payment_transactions;
DROP TABLE IF EXISTS subscription_events;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS packages;
DROP TABLE IF EXISTS contact_unlocks;
DROP TABLE IF EXISTS credit_transactions;
DROP TABLE IF EXISTS credit_wallets;
DROP FUNCTION IF EXISTS update_updated_at();
