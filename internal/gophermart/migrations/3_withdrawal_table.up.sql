create table withdrawals (
    "withdrawal_id" serial primary key,
    "client_id"     int not null references clients on delete cascade,
    "order_number"  character varying not null,
    "sum"           double precision default 0,
    "processed_at"  timestamp with time zone
);

create index "withdrawals_number_idx" ON withdrawals ("order_number");