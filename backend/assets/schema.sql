create table websites (
    user_uuid varchar(64),
    url text,
    title text,
    content text,
    accessTime integer,
    updateTime integer,
    groupName text
);

create index websites__user_uuid on websites(user_uuid);
