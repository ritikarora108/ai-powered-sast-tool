-- +goose Up
-- SQL in this section is executed when the migration is applied

-- Vulnerabilities table
CREATE TABLE vulnerabilities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scan_id UUID NOT NULL REFERENCES scans(id),
    vulnerability_type VARCHAR(100) NOT NULL, -- OWASP category
    file_path TEXT NOT NULL,
    line_start INTEGER NOT NULL,
    line_end INTEGER NOT NULL,
    severity VARCHAR(20) NOT NULL, -- 'Critical', 'High', 'Medium', 'Low'
    description TEXT NOT NULL,
    remediation TEXT,
    code_snippet TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_vulnerabilities_scan_id ON vulnerabilities(scan_id);
CREATE INDEX idx_vulnerabilities_type ON vulnerabilities(vulnerability_type);
CREATE INDEX idx_vulnerabilities_severity ON vulnerabilities(severity);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back
DROP INDEX IF EXISTS idx_vulnerabilities_severity;
DROP INDEX IF EXISTS idx_vulnerabilities_type;
DROP INDEX IF EXISTS idx_vulnerabilities_scan_id;
DROP TABLE IF EXISTS vulnerabilities; 