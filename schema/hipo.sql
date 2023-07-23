begin;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

create table jwallets
(
    address     text        not null,
    info        jsonb       not null,
    create_time timestamptz not null,
    notify_time timestamptz,

    primary key (address)
);
