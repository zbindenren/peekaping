CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY,
    status_page_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    style VARCHAR(30) NOT NULL DEFAULT 'warning',
    active BOOLEAN NOT NULL DEFAULT true,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (status_page_id) REFERENCES status_pages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_incidents_status_page_active ON incidents(status_page_id, active, created_at DESC);
