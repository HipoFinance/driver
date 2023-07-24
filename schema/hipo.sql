begin;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

create table jwallets
(
    address     text        not null,
    round_since integer     not null,
    info        jsonb       not null,
    create_time timestamptz not null,
    notify_time timestamptz,

    primary key (address, round_since)
);

create table memos
(
    key     text    not null,
    memo    jsonb   not null,

    primary key (key)
);
