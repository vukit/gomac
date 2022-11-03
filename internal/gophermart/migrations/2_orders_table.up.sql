create type order_status as enum ('NEW', 'REGISTERED', 'PROCESSING', 'INVALID', 'PROCESSED');

create table orders (
    "order_id"      serial primary key,
    "client_id"     int not null references clients on delete cascade,
    "order_number"  character varying not null,
    "accrual"       double precision default 0,
    "status"        order_status,
    "uploaded_at"   timestamp with time zone,
    unique("order_number")
);

create index "orders_order_number_idx" ON orders ("order_number");