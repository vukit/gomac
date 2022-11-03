create table clients (
    "client_id" serial primary key,
    "login" varchar(64) not null,
    "password" char(64) not null,
    unique ("login")
);

create index "clients_login_idx" ON clients ("login");
