-- call-service/migrations/000001_create_calls_table.up.sql
CREATE TABLE calls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'открыта',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL
);