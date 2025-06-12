CREATE TABLE messages (
    id serial primary key,
    created timestamp not null default now(),
    content text not null,
    author varchar(32) not null,
);
