begin;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

create table stakes
(
    address       text        not null,
    round_since   bigint      not null,
    hash          text        not null,
    state         text        not null,
    retried       integer     not null,
    info          jsonb       not null,
    create_time   timestamptz not null,
    retry_time    timestamptz,
    sent_time     timestamptz,
    verified_time timestamptz,

    primary key (hash)
);

create table unstakes
(
    address       text        not null,
    tokens        numeric(40) not null,
    hash          text        not null,
    state         text        not null,
    retried       integer     not null,
    info          jsonb       not null,
    create_time   timestamptz not null,
    retry_time    timestamptz,
    sent_time     timestamptz,
    verified_time timestamptz,
    
    primary key (hash)
);

create table memos
(
    key     text    not null,
    memo    jsonb   not null,

    primary key (key)
);
