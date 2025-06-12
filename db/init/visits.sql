CREATE TYPE VisitType AS ENUM ('Pick', 'Create');

CREATE TABLE visits (
    username varchar(32) not null,
    type VisitType not null,
    created timestamp not null default now()
);

CREATE INDEX idx_visits_username_type ON visits (username, type);
