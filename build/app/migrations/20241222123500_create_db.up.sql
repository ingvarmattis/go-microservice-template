begin;

create schema   if not exists example;
create table    if not exists example.services
(
    service_name text primary key
);

alter table example.services owner to postgres;

end;
