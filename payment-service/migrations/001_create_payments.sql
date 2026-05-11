CREATE TABLE payments (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    transaction_id VARCHAR(36),
    customer_email VARCHAR(100) NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL
);