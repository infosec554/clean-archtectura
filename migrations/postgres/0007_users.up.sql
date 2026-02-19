
CREATE EXTENSION IF NOT EXISTS pgcrypto;



create table if not exists users(
    id uuid primary key default gen_random_uuid(),
    pinfl varchar not null check ( length(pinfl) = 14) unique ,
    passport varchar not null ,
    first_name varchar not null ,
    last_name varchar not null ,
    middle_name varchar,
    email varchar,
    phone varchar,
    password varchar,
    created_at timestamp default now(),
    updated_at timestamp default now()
);
