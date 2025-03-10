-- gauge collection
create table if not exists gauge(
name text not null unique PRIMARY KEY,
value double precision not null,
id serial not null unique,
created_at timestamp with time zone not null default now()
);

-- counter collection
create table if not exists counter(
name text not null unique PRIMARY KEY,
value int not null,
id serial not null unique,
created_at timestamp with time zone not null default now()
);
