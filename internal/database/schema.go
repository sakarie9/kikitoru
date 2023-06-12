package database

var Schema = `
create table if not exists t_work
(
    id                text primary key,
    root_folder       text not null,
    dir 			  text not null,
    title             text not null,
    nsfw              boolean,
    release           date,
    dl_count          integer,
    price             integer,
    review_count      integer,
    rate_count        integer,
    rate_average_2dp  real,
    rate_count_detail jsonb,
    rank              jsonb,
    create_date 	  timestamp with time zone default current_timestamp,
    has_subtitle      bool default false
);

create table if not exists t_circle
(
    id text primary key,
    name text not null
);

create table if not exists t_va
(
    id uuid primary key,
    name text not null
);

create table if not exists t_user
(
    name text primary key,
    password text not null,
    "group" text not null
);

create table if not exists t_tag
(
    id int primary key,
    name text not null
);

create table if not exists t_review
(
    user_name text references t_user(name) on delete cascade,
    work_id text references t_work(id) on delete cascade,
    PRIMARY KEY(user_name, work_id),
    rating int,
    review_text text,
    created_at timestamp with time zone default current_timestamp,
    updated_at timestamp with time zone default current_timestamp,
    progress text
);

create table if not exists t_history
(
    id serial not null,
    user_name text not null references t_user(name) on delete cascade,
    work_id text not null references t_work(id) on delete cascade,
    PRIMARY KEY(user_name, work_id),
    file_index text not null,
    file_name text,
    play_time int,
    total_time int,
    created_at timestamp with time zone default current_timestamp,
    updated_at timestamp with time zone default current_timestamp
);

create table if not exists t_series
(
    id text primary key,
    name text not null
);

create table if not exists r_circle_work
(
    circle_id text references t_circle(id) on delete cascade,
    work_id text references t_work(id) on delete cascade,
	PRIMARY KEY(circle_id, work_id)
);

create table if not exists r_va_work
(
    va_id uuid references t_va(id) on delete cascade,
    work_id text references t_work(id) on delete cascade,
	PRIMARY KEY(va_id, work_id)
);

create table if not exists r_tag_work
(
    tag_id int references t_tag(id) on delete cascade ,
    work_id text references t_work(id) on delete cascade,
    PRIMARY KEY(tag_id, work_id)
);

create table if not exists r_series_work
(
    series_id text references t_series(id) on delete cascade,
    work_id text references t_work(id) on delete cascade,
    PRIMARY KEY(series_id, work_id)
);

drop table if exists t_file_nodes;
CREATE TABLE t_file_nodes 
(
    id text,
	index int,
	PRIMARY KEY(id, index),
	path text not null
);

DROP VIEW IF EXISTS "staticMetadata";
create view "staticMetadata" as
SELECT qv.*,
       jsonb_agg(jsonb_build_object('id', t_tag.id, 'name', t_tag.name)) AS tagObj
FROM (
         SELECT q.*,
                jsonb_agg(jsonb_build_object('id', t_va.id, 'name', t_va.name)) AS vaObj
         FROM (
                  SELECT substring(t_work.id,3)::integer as id,
                         t_work.title,
                         substring(t_circle.id,3)::integer as circle_id,
                         t_circle.name,
                         jsonb_build_object('id', substring(t_circle.id,3)::integer, 'name', t_circle.name) AS circleObj,
                         t_work.nsfw,
                         to_char(t_work.release, 'YYYY-MM-DD') as release,
                         t_work.dl_count,
                         t_work.price,
                         t_work.review_count,
                         t_work.rate_count,
                         t_work.rate_average_2dp,
                         t_work.rate_count_detail,
                         t_work.rank,
                         to_char(t_work.create_date, 'YYYY-MM-DD') as create_date,
                         t_work.has_subtitle
                  FROM t_work
                           JOIN r_circle_work on t_work.id = r_circle_work.work_id
                           JOIN t_circle ON t_circle.id = r_circle_work.circle_id
              ) AS q
                  JOIN r_va_work ON substring(r_va_work.work_id,3)::integer = q.id
                  JOIN t_va ON t_va.id = r_va_work.va_id
         GROUP BY q.id,q.title,q.circle_id,q.name,q.circleObj,q.nsfw,q.dl_count,q.price,q.review_count,q.rate_count,q.rate_average_2dp,q.rate_count_detail,q.rank,q.release,q.create_date,q.has_subtitle
     ) AS qv
         LEFT JOIN r_tag_work ON substring(r_tag_work.work_id,3)::integer = qv.id
         LEFT JOIN t_tag ON t_tag.id = r_tag_work.tag_id
GROUP BY qv.id,qv.title,qv.circle_id,qv.name,qv.circleObj,qv.vaObj,qv.nsfw,qv.dl_count,qv.price,qv.review_count,qv.rate_count,qv.rate_average_2dp,qv.rate_count_detail,qv.rank,qv.release,qv.create_date,qv.has_subtitle
;

DROP MATERIALIZED VIEW IF EXISTS search_view;
create materialized view search_view as
SELECT t_work.id,
       title,
       t_circle.name,
       string_agg(DISTINCT t_va.name,',') AS vas,
       string_agg(DISTINCT t_tag.name,',') AS tags
FROM t_work
INNER JOIN t_circle ON EXISTS (
    SELECT 1
    FROM r_circle_work
    WHERE work_id = t_work.id
    AND circle_id = t_circle.id
)
INNER JOIN t_va ON EXISTS (
    SELECT 1
    FROM r_va_work
    WHERE work_id = t_work.id
    AND va_id = t_va.id
)
LEFT JOIN t_tag ON EXISTS (
    SELECT 1
    FROM r_tag_work
    WHERE work_id = t_work.id
    AND tag_id = t_tag.id
)
GROUP BY t_work.id,title,t_circle.name;

-- 创建索引
create extension if not exists pg_trgm;

create or replace function record_to_text(anyelement) returns text as $$  
  select $1::text;                        
$$ language sql strict immutable;  

create index if not exists idx_sv_1 on search_view using gin (record_to_text(search_view) gin_trgm_ops);
-- create index idx_sv_2 on search_view using gist (record_to_text(search_view) gist_trgm_ops);
`
