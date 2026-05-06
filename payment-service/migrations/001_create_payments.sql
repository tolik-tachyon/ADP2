CREATE TABLE paymentspsql -U postgres -d order_db (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    transaction_id VARCHAR(36),
    amount BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL
);