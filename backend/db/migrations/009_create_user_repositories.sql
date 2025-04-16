-- +goose Up
-- SQL in this section is executed when the migration is applied

-- Create the user_repositories join table
CREATE TABLE user_repositories (
    user_id UUID REFERENCES users(id),
    repository_id UUID REFERENCES repositories(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, repository_id)
);

-- Create index to speed up repository lookups by user
CREATE INDEX idx_user_repositories_user_id ON user_repositories(user_id);
CREATE INDEX idx_user_repositories_repository_id ON user_repositories(repository_id);

-- Populate the table with existing repositories by connecting them to their creators
INSERT INTO user_repositories (user_id, repository_id)
SELECT created_by, id FROM repositories
WHERE created_by IS NOT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back
DROP INDEX IF EXISTS idx_user_repositories_repository_id;
DROP INDEX IF EXISTS idx_user_repositories_user_id;
DROP TABLE IF EXISTS user_repositories; 