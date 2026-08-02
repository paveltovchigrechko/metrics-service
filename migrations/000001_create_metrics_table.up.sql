CREATE TABLE IF NOT EXISTS metrics (
    id VARCHAR(255) NOT NULL,
    mtype VARCHAR(50) NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION,
    hash VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Composite Primary Key: Guarantees unique metrics per name/type combination
    PRIMARY KEY (id, mtype),

    -- Ensure mtype is restricted to expected values
    CONSTRAINT check_mtype_valid 
        CHECK (mtype IN ('counter', 'gauge')),

    -- Ensure counters have a delta value and gauges have a value
    CONSTRAINT check_metric_value_presence 
        CHECK (
            (mtype = 'counter' AND delta IS NOT NULL AND value IS NULL) OR
            (mtype = 'gauge' AND value IS NOT NULL AND delta IS NULL)
        )
);

-- Index to optimize lookups by ID alone (for example, SELECT * FROM metrics WHERE id = $1)
CREATE INDEX IF NOT EXISTS idx_metrics_id ON metrics (id);