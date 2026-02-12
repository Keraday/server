CREATE TABLE subscriptions (
    id UUID PRIMARY KEY ,
    service_name VARCHAR(50) NOT NULL,
    price INTEGER NOT NULL CHECK (price>=0),
    user_id UUID NOT NULL,
    start_date  DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_subscriptions_user_service_dates
ON subscriptions (user_id, service_name, start_date, end_date);

CREATE INDEX idx_subscriptions_user_id
ON subscriptions (user_id);

CREATE INDEX idx_subscriptions_start_date
ON subscriptions (start_date);


CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_subscriptions_updated_at
BEFORE UPDATE ON subscriptions
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();



