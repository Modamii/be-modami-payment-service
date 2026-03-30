-- +migrate Down
DROP TYPE IF EXISTS outbox_status;
DROP TYPE IF EXISTS invoice_status;
DROP TYPE IF EXISTS refund_status;
DROP TYPE IF EXISTS billing_cycle;
DROP TYPE IF EXISTS subscription_status;
DROP TYPE IF EXISTS payment_method;
DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS credit_tx_type;
