create table websites (
    url text,
    title text,
    content text,
    update_time timestamp
);

create unique index websites__url on websites(url);

create table user_websites (
    user_uuid varchar(64),
    url text,
    access_time timestamp,
    group_name text
);

create unique index user_websites__user_url on user_websites(user_uuid, url);