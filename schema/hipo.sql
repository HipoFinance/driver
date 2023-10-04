begin;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

create table stakes
(
    address       text        not null,
    round_since   bigint      not null,
    hash          text        not null,
    state         text        not null,
    retry_count   integer     not null,
    info          jsonb       not null,
    created_at    timestamptz not null,
    retried_at    timestamptz,
    sent_at       timestamptz,
    verified_at   timestamptz,

    primary key (hash)
);

create table unstakes
(
    address       text        not null,
    tokens        numeric(40) not null,
    hash          text        not null,
    state         text        not null,
    retry_count   integer     not null,
    info          jsonb       not null,
    created_at    timestamptz not null,
    retried_at    timestamptz,
    sent_at       timestamptz,
    verified_at   timestamptz,
    
    primary key (hash)
);

create table memos
(
    key     text    not null,
    memo    jsonb   not null,

    primary key (key)
);
