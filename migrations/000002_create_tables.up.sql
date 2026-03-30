-- +migrate Up

-- Shared updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- credit_wallets
CREATE TABLE credit_wallets (
    user_id      UUID PRIMARY KEY,
    balance      INT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    total_earned INT NOT NULL DEFAULT 0,
    total_spent  INT NOT NULL DEFAULT 0,
    version      BIGINT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TRIGGER trg_credit_wallets_updated_at
    BEFORE UPDATE ON credit_wallets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- credit_transactions
CREATE TABLE credit_transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES credit_wallets(user_id),
    amount          INT NOT NULL,
    type            credit_tx_type NOT NULL,
    ref_type        VARCHAR(50),
    ref_id          UUID,
    balance_after   INT NOT NULL,
    description     VARCHAR(500) NOT NULL,
    idempotency_key VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_credit_tx_user_created ON credit_transactions(user_id, created_at DESC);
CREATE INDEX idx_credit_tx_ref ON credit_transactions(ref_type, ref_id);
CREATE INDEX idx_credit_tx_type ON credit_transactions(type, created_at DESC);
CREATE UNIQUE INDEX idx_credit_tx_idempotency ON credit_transactions(idempotency_key)
    WHERE idempotency_key IS NOT NULL;

-- contact_unlocks
CREATE TABLE contact_unlocks (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buyer_id     UUID NOT NULL,
    product_id   UUID NOT NULL,
    seller_id    UUID NOT NULL,
    credit_tx_id UUID REFERENCES credit_transactions(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(buyer_id, product_id)
);

-- packages
CREATE TABLE packages (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code              VARCHAR(50) UNIQUE NOT NULL,
    name              VARCHAR(100) NOT NULL,
    tier              SMALLINT NOT NULL DEFAULT 0,
    price_monthly     BIGINT NOT NULL DEFAULT 0,
    price_yearly      BIGINT NOT NULL DEFAULT 0,
    credits_per_month INT NOT NULL DEFAULT 0,
    search_boost      BOOLEAN NOT NULL DEFAULT FALSE,
    search_priority   BOOLEAN NOT NULL DEFAULT FALSE,
    badge_name        VARCHAR(50),
    priority_support  BOOLEAN NOT NULL DEFAULT FALSE,
    featured_slots    INT NOT NULL DEFAULT 0,
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order        INT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TRIGGER trg_packages_updated_at
    BEFORE UPDATE ON packages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

INSERT INTO packages (code, name, tier, price_monthly, price_yearly, credits_per_month, search_boost, search_priority, badge_name, priority_support, featured_slots, sort_order)
VALUES
    ('curator', 'Curator', 0, 0, 0, 5, false, false, null, false, 0, 1),
    ('style', 'Style', 1, 99000, 989880, 50, true, false, 'Style Curator', false, 3, 2),
    ('elite', 'Elite', 2, 249000, 2489880, 200, true, true, 'Elite', true, 10, 3);

-- subscriptions
CREATE TABLE subscriptions (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID NOT NULL,
    package_id         UUID NOT NULL REFERENCES packages(id),
    billing_cycle      billing_cycle NOT NULL,
    price_paid         BIGINT NOT NULL,
    discount_code      VARCHAR(50),
    credits_allocated  INT NOT NULL DEFAULT 0,
    credits_used       INT NOT NULL DEFAULT 0,
    status             subscription_status NOT NULL DEFAULT 'pending',
    auto_renew         BOOLEAN NOT NULL DEFAULT TRUE,
    start_date         TIMESTAMPTZ,
    end_date           TIMESTAMPTZ,
    payment_tx_id      UUID,
    cancelled_at       TIMESTAMPTZ,
    cancel_reason      VARCHAR(500),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_subscriptions_user_status ON subscriptions(user_id, status);
CREATE INDEX idx_subscriptions_expiring ON subscriptions(status, end_date)
    WHERE status = 'active';
CREATE INDEX idx_subscriptions_autorenew ON subscriptions(status, auto_renew, end_date)
    WHERE status = 'active' AND auto_renew = TRUE;
CREATE TRIGGER trg_subscriptions_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- subscription_events
CREATE TABLE subscription_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id),
    from_status     subscription_status,
    to_status       subscription_status NOT NULL,
    reason          VARCHAR(500),
    actor_type      VARCHAR(20) NOT NULL,
    actor_id        UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- payment_transactions
CREATE TABLE payment_transactions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL,
    order_ref        VARCHAR(100) UNIQUE NOT NULL,
    gateway_tx_id    VARCHAR(255),
    amount           BIGINT NOT NULL,
    currency         VARCHAR(3) NOT NULL DEFAULT 'VND',
    method           payment_method NOT NULL,
    purpose          VARCHAR(50) NOT NULL,
    purpose_ref_id   UUID,
    status           payment_status NOT NULL DEFAULT 'pending',
    gateway_response JSONB,
    payment_url      VARCHAR(2000),
    ip_address       INET,
    expires_at       TIMESTAMPTZ,
    paid_at          TIMESTAMPTZ,
    failure_reason   VARCHAR(500),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_payment_tx_user ON payment_transactions(user_id, created_at DESC);
CREATE INDEX idx_payment_tx_order_ref ON payment_transactions(order_ref);
CREATE TRIGGER trg_payment_tx_updated_at
    BEFORE UPDATE ON payment_transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- refunds
CREATE TABLE refunds (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_tx_id     UUID REFERENCES payment_transactions(id),
    credit_tx_id      UUID REFERENCES credit_transactions(id),
    user_id           UUID NOT NULL,
    refund_type       VARCHAR(20) NOT NULL,
    amount            BIGINT NOT NULL,
    reason            VARCHAR(500),
    status            refund_status NOT NULL DEFAULT 'pending',
    requested_by      UUID NOT NULL,
    approved_by       UUID,
    gateway_refund_id VARCHAR(255),
    gateway_response  JSONB,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TRIGGER trg_refunds_updated_at
    BEFORE UPDATE ON refunds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- invoices
CREATE TABLE invoices (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    user_id        UUID NOT NULL,
    payment_tx_id  UUID REFERENCES payment_transactions(id),
    subscription_id UUID REFERENCES subscriptions(id),
    subtotal       BIGINT NOT NULL,
    tax_amount     BIGINT NOT NULL DEFAULT 0,
    total          BIGINT NOT NULL,
    description    VARCHAR(500),
    status         invoice_status NOT NULL DEFAULT 'draft',
    billing_name   VARCHAR(200),
    billing_email  VARCHAR(200),
    tax_code       VARCHAR(50),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TRIGGER trg_invoices_updated_at
    BEFORE UPDATE ON invoices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- outbox_events
CREATE TABLE outbox_events (
    id             BIGSERIAL PRIMARY KEY,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id   VARCHAR(100) NOT NULL,
    event_type     VARCHAR(100) NOT NULL,
    payload        JSONB NOT NULL,
    status         outbox_status NOT NULL DEFAULT 'pending',
    retry_count    INT NOT NULL DEFAULT 0,
    max_retries    INT NOT NULL DEFAULT 5,
    published_at   TIMESTAMPTZ,
    last_error     VARCHAR(1000),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_outbox_pending ON outbox_events(status, id)
    WHERE status = 'pending';
