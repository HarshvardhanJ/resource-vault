-- 000002_resource_vault_catalog.sql
-- Extend the catalog for NITC Resource Vault.

-- Academic units replace the narrower branch-only concept while preserving the existing
-- tables during migration. Existing application code may continue to use branches until
-- the catalog refactor is complete.
CREATE TABLE IF NOT EXISTS academic_units (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'department',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS course_academic_units (
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    academic_unit_id UUID NOT NULL REFERENCES academic_units(id) ON DELETE CASCADE,
    PRIMARY KEY (course_id, academic_unit_id)
);

CREATE TABLE IF NOT EXISTS metadata_suggestions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    proposed_changes JSONB NOT NULL,
    reason TEXT,
    submitted_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'PENDING',
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_academic_units_active_code ON academic_units(active, code);
CREATE INDEX IF NOT EXISTS idx_course_academic_units_unit_course ON course_academic_units(academic_unit_id, course_id);
CREATE INDEX IF NOT EXISTS idx_metadata_suggestions_status_created ON metadata_suggestions(status, created_at DESC);

-- Seed the official academic units supplied for the initial NITC Resource Vault catalog.
INSERT INTO academic_units (code, name, type) VALUES
('BARCH', 'B. Arch', 'academic_programme'),
('BT', 'Biotechnology', 'department'),
('CHE', 'Chemical Engineering', 'department'),
('CE', 'Civil Engineering', 'department'),
('CSE', 'Computer Science and Engineering', 'department'),
('EEE', 'Electrical and Electronics Engineering', 'department'),
('ECE', 'Electronics and Communication Engineering', 'department'),
('ENE', 'Energy Engineering', 'department'),
('EP', 'Engineering Physics', 'department'),
('HSS', 'Humanities and Social Sciences', 'department'),
('MSE', 'Materials Science and Engineering', 'department'),
('ME', 'Mechanical Engineering', 'department'),
('PE', 'Production Engineering', 'department'),
('ITEP', '4-year Integrated Teacher Education Programme (ITEP) B.Sc–B.Ed', 'academic_programme')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    type = EXCLUDED.type,
    active = TRUE;

-- Seed the current branch records with the same canonical display names where possible.
UPDATE branches SET name = CASE code
    WHEN 'BARCH' THEN 'B. Arch'
    WHEN 'BT' THEN 'Biotechnology'
    WHEN 'CH' THEN 'Chemical Engineering'
    WHEN 'CE' THEN 'Civil Engineering'
    WHEN 'CSE' THEN 'Computer Science and Engineering'
    WHEN 'EEE' THEN 'Electrical and Electronics Engineering'
    WHEN 'ECE' THEN 'Electronics and Communication Engineering'
    WHEN 'ENE' THEN 'Energy Engineering'
    WHEN 'EP' THEN 'Engineering Physics'
    WHEN 'HSS' THEN 'Humanities and Social Sciences'
    WHEN 'MSE' THEN 'Materials Science and Engineering'
    WHEN 'ME' THEN 'Mechanical Engineering'
    WHEN 'PE' THEN 'Production Engineering'
    WHEN 'ITEP' THEN '4-year Integrated Teacher Education Programme (ITEP) B.Sc-B.Ed'
    ELSE name END;

-- Backfill course-academic-unit relationships from the existing course_branches table.
INSERT INTO course_academic_units (course_id, academic_unit_id)
SELECT cb.course_id, au.id
FROM course_branches cb
JOIN branches b ON b.id = cb.branch_id
JOIN academic_units au ON au.code = CASE b.code
    WHEN 'CH' THEN 'CHE'
    ELSE b.code END
ON CONFLICT DO NOTHING;

-- Resource storage state must distinguish archival backend from application state.
ALTER TABLE resources
    ADD COLUMN IF NOT EXISTS storage_identifier TEXT;

ALTER TABLE resources
    ADD COLUMN IF NOT EXISTS archive_status TEXT NOT NULL DEFAULT 'NOT_APPLICABLE';

ALTER TABLE resources
    ADD COLUMN IF NOT EXISTS scan_status TEXT NOT NULL DEFAULT 'NOT_RUN';

ALTER TABLE resources
    ADD COLUMN IF NOT EXISTS reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE resources
    ADD COLUMN IF NOT EXISTS rejection_reason TEXT;

ALTER TABLE resources
    ADD COLUMN IF NOT EXISTS removed_at TIMESTAMPTZ;

ALTER TABLE audit_events
    ADD COLUMN IF NOT EXISTS request_id TEXT;
