-- name: GetVulnerabilityByID :one
SELECT * FROM vulnerabilities
WHERE id = $1 LIMIT 1;

-- name: ListScanVulnerabilities :many
SELECT * FROM vulnerabilities
WHERE scan_id = $1
ORDER BY severity DESC, vulnerability_type ASC
LIMIT $2 OFFSET $3;

-- name: CreateVulnerability :one
INSERT INTO vulnerabilities (
  scan_id, vulnerability_type, file_path, line_start, line_end, 
  severity, description, remediation, code_snippet
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: CountVulnerabilitiesByScan :one
SELECT COUNT(*) FROM vulnerabilities
WHERE scan_id = $1;

-- name: CountVulnerabilitiesBySeverity :many
SELECT severity, COUNT(*) as count
FROM vulnerabilities
WHERE scan_id = $1
GROUP BY severity
ORDER BY 
  CASE 
    WHEN severity = 'Critical' THEN 1
    WHEN severity = 'High' THEN 2
    WHEN severity = 'Medium' THEN 3
    WHEN severity = 'Low' THEN 4
    ELSE 5
  END; 