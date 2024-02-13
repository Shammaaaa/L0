create table public.order
(
    order_uid varchar(19) primary key not null,
    data      jsonb default '{}'::jsonb not null
);

create unique index if not exists uq_order_uid
    on public.order (order_uid);
