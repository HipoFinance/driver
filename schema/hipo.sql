begin;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

create table stakes
(
    address      text        not null,
    round_since  integer     not null,
    hash         text        not null,
    state        text        not null,
    retried      integer     not null,
    info         jsonb       not null,
    create_time  timestamptz not null,
    retry_time   timestamptz,
    success_time timestamptz,

-- @TODO: hash can be used as primary key by single
    primary key (address, round_since, hash)
);

create table unstakes
(
    address      text        not null,
    tokens       bigint      not null,
    hash         text        not null,
    state        text        not null,
    retried      integer     not null,
    info         jsonb       not null,
    create_time  timestamptz not null,
    retry_time   timestamptz,
    success_time timestamptz,
    
-- @TODO: hash can be used as primary key by single
    primary key (address, tokens, hash)
);

create table memos
(
    key     text    not null,
    memo    jsonb   not null,

    primary key (key)
);
