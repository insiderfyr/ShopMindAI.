-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "citus";

-- Create conversations table (distributed by user_id)
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    title TEXT,
    model VARCHAR(50) NOT NULL DEFAULT 'gpt-3.5-turbo',
    system_prompt TEXT,
    metadata JSONB DEFAULT '{}',
    tokens_used INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Create messages table (distributed by conversation_id)
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system', 'function')),
    content TEXT NOT NULL,
    function_name TEXT,
    function_args JSONB,
    tokens INTEGER NOT NULL DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for performance at scale
-- Composite index for efficient pagination
CREATE INDEX idx_conversations_user_updated ON conversations(user_id, updated_at DESC, id DESC) 
WHERE deleted_at IS NULL;

-- Index for conversation lookup
CREATE INDEX idx_conversations_user_id ON conversations(user_id) 
WHERE deleted_at IS NULL;

-- Index for soft deletes
CREATE INDEX idx_conversations_deleted ON conversations(deleted_at) 
WHERE deleted_at IS NOT NULL;

-- Messages indexes
CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at DESC);

-- Full-text search indexes
CREATE INDEX idx_messages_content_trgm ON messages USING gin (content gin_trgm_ops);
CREATE INDEX idx_conversations_title_trgm ON conversations USING gin (title gin_trgm_ops);

-- Index for token tracking
CREATE INDEX idx_conversations_tokens ON conversations(user_id, tokens_used) 
WHERE deleted_at IS NULL;

-- Distribute tables using Citus
SELECT create_distributed_table('conversations', 'user_id');
SELECT create_distributed_table('messages', 'conversation_id');

-- Create co-location between related tables
SELECT create_reference_table('models');
CREATE TABLE models (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    max_tokens INTEGER NOT NULL,
    cost_per_token DECIMAL(10, 8) NOT NULL
);

-- Insert default models
INSERT INTO models (name, max_tokens, cost_per_token) VALUES
    ('gpt-3.5-turbo', 4096, 0.000002),
    ('gpt-4', 8192, 0.00003),
    ('gpt-4-turbo', 128000, 0.00001);

-- Create function for updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add trigger for updated_at
CREATE TRIGGER update_conversations_updated_at 
    BEFORE UPDATE ON conversations
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create partitioned tables for time-series data (messages)
-- This is for future when messages table grows beyond billions
CREATE TABLE messages_partitioned (
    LIKE messages INCLUDING ALL
) PARTITION BY RANGE (created_at);

-- Create monthly partitions (example for 2024)
CREATE TABLE messages_2024_01 PARTITION OF messages_partitioned
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
    
CREATE TABLE messages_2024_02 PARTITION OF messages_partitioned
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Add partition creation function for automation
CREATE OR REPLACE FUNCTION create_monthly_partition(table_name text, start_date date)
RETURNS void AS $$
DECLARE
    partition_name text;
    end_date date;
BEGIN
    partition_name := table_name || '_' || to_char(start_date, 'YYYY_MM');
    end_date := start_date + interval '1 month';
    
    EXECUTE format('CREATE TABLE IF NOT EXISTS %I PARTITION OF %I FOR VALUES FROM (%L) TO (%L)',
        partition_name, table_name, start_date, end_date);
END;
$$ LANGUAGE plpgsql;

-- Create analytics tables for fast aggregations
CREATE TABLE user_stats (
    user_id UUID PRIMARY KEY,
    total_conversations INTEGER NOT NULL DEFAULT 0,
    total_messages INTEGER NOT NULL DEFAULT 0,
    total_tokens_used BIGINT NOT NULL DEFAULT 0,
    last_active_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

SELECT create_distributed_table('user_stats', 'user_id');

-- Create materialized view for popular conversations
CREATE MATERIALIZED VIEW popular_conversations AS
SELECT 
    c.id,
    c.user_id,
    c.title,
    COUNT(m.id) as message_count,
    MAX(m.created_at) as last_message_at,
    c.created_at
FROM conversations c
LEFT JOIN messages m ON m.conversation_id = c.id
WHERE c.deleted_at IS NULL
GROUP BY c.id
HAVING COUNT(m.id) > 10
ORDER BY COUNT(m.id) DESC;

CREATE INDEX idx_popular_conversations_user ON popular_conversations(user_id);

-- Add table for session management
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    token_hash VARCHAR(64) NOT NULL,
    device_info JSONB,
    ip_address INET,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_token ON user_sessions(token_hash);
CREATE INDEX idx_user_sessions_user ON user_sessions(user_id, expires_at);
CREATE INDEX idx_user_sessions_expires ON user_sessions(expires_at) WHERE expires_at > NOW();

SELECT create_distributed_table('user_sessions', 'user_id');

-- Add permissions for read replicas
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly_user;

-- Add comment documentation
COMMENT ON TABLE conversations IS 'Stores chat conversations distributed by user_id for horizontal scaling';
COMMENT ON TABLE messages IS 'Stores messages distributed by conversation_id, co-located with conversations';
COMMENT ON INDEX idx_conversations_user_updated IS 'Primary index for paginated conversation lists';
COMMENT ON INDEX idx_messages_content_trgm IS 'Trigram index for full-text search on message content';