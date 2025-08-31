CREATE TABLE payments (
                          id UUID PRIMARY KEY,
                          idempotency_key VARCHAR(255) UNIQUE NOT NULL,
                          user_id UUID NOT NULL,
                          amount BIGINT NOT NULL,
                          status VARCHAR(20) NOT NULL,
                          service_id VARCHAR(100) NOT NULL,
                          client_number VARCHAR(100) NOT NULL,
                          created_at TIMESTAMP DEFAULT NOW(),
                          updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE balance (
                          user_id UUID PRIMARY KEY,
                          available_balance BIGINT NOT NULL,
                          reserved_balance BIGINT DEFAULT 0,
                          updated_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO balance (user_id, available_balance, reserved_balance, updated_at)
VALUES ('550e8400-e29b-41d4-a716-446655440000', 10000, 0, NOW());
VALUES ('550e8400-e29b-41d4-a716-446655440001', 10000, 0, NOW());
VALUES ('550e8400-e29b-41d4-a716-446655440002', 10000, 0, NOW());
VALUES ('550e8400-e29b-41d4-a716-446655440003', 10000, 0, NOW());