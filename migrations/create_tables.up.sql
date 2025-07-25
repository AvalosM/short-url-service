create table if not exists short_urls (
    id varchar(6) primary key,
    long_url text not null,

    created_at timestamp default now() not null,
    updated_at timestamp default now() not null
);

create table if not exists short_url_metrics (
    short_url_id varchar(6) not null references short_urls(id) on delete cascade,
    visit_count bigint  not null,
    unique_visit_count bigint not null,
    timestamp timestamp default now() not null
);

create index if not exists idx_short_urls_long_url on short_urls(long_url);
create index if not exists idx_short_url_statistics_short_url_id_timestamp on short_url_metrics using btree (short_url_id, timestamp);