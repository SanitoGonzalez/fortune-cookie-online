CREATE TYPE VisitType AS ENUM ('Pick', 'Create');

CREATE TABLE visits (
    username varchar(32) not null,
    visit_type VisitType not null,
    count int not null default 0,
    last_visited timestamp not null default now(),

    primary key(username, visit_type)
);