begin;

create schema   if not exists auth;
create table    if not exists auth.users
(
    id            serial        primary key,
    email_id      bigint        not null,
    email         text          not null,
    password_hash bigint        not null,
    last_login_at timestamp     default now(),
    created_at    timestamp     default now(),
    updated_at    timestamp     default now()
);

alter table auth.users owner to postgres;

create unique index if not exists idx_users_email_id on auth.users (email_id);

end;
