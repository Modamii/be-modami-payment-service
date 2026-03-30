-- +migrate Up
CREATE TYPE credit_tx_type AS ENUM (
    'purchase', 'unlock', 'refund', 'reward', 'subscription_alloc', 'expire', 'admin_adjust'
);

CREATE TYPE payment_status AS ENUM (
    'pending', 'processing', 'success', 'failed', 'expired', 'cancelled'
);

CREATE TYPE payment_method AS ENUM (
    'vnpay', 'momo', 'zalopay', 'bank_transfer', 'credit_card'
);

CREATE TYPE subscription_status AS ENUM (
    'pending', 'active', 'expired', 'cancelled', 'failed'
);

CREATE TYPE billing_cycle AS ENUM ('monthly', 'yearly');

CREATE TYPE refund_status AS ENUM (
    'pending', 'approved', 'processing', 'completed', 'rejected'
);

CREATE TYPE invoice_status AS ENUM ('draft', 'issued', 'paid', 'voided');

CREATE TYPE outbox_status AS ENUM ('pending', 'published', 'failed');
