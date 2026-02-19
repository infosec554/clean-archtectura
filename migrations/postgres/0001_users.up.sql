
CREATE EXTENSION IF NOT EXISTS pgcrypto;



create table if not exists users(
    id uuid primary key default gen_random_uuid(),
    first_name varchar not null ,
    last_name varchar not null ,
    email varchar,
    password varchar,
    created_at timestamp default now(),
    updated_at timestamp default now()
);
