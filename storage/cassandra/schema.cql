-- Keyspace: single DC lab, replication 1 (prod: RF=3 across DCs).
CREATE KEYSPACE IF NOT EXISTS lab
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

-- Table modeled BY QUERY: timeline by user, newest first.
-- Partition key (user_id) spreads users; clustering (created_at DESC)
-- orders inside a partition. Query MUST include user_id.
CREATE TABLE IF NOT EXISTS lab.timeline (
    user_id    INT,
    created_at TIMESTAMP,
    post_id    UUID,
    body       TEXT,
    PRIMARY KEY (user_id, created_at, post_id)
) WITH CLUSTERING ORDER BY (created_at DESC, post_id ASC);
