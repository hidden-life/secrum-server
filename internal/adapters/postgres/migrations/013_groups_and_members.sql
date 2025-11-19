create table groups
(
    id         uuid primary key,
    name       text not null,
    avatar_url text,
    created_by uuid not null references users(id) on delete restrict,
    created_at timestamp with time zone not null default now(),
    updated_at timestamp with time zone not null default now(),
    is_active  boolean not null default true
);

alter table groups owner to secrum_user;

do $$
begin
    if not exists (select 1 from pg_type where typname = 'group_member_role') then
create type group_member_role as enum ('owner', 'admin', 'member');
end if;
end$$;

create table group_members
(
    group_id  uuid not null references groups(id) on delete cascade,
    user_id   uuid not null references users(id) on delete cascade,
    role      group_member_role not null default 'member',
    joined_at timestamp with time zone not null default now(),
    is_active boolean not null default true,
    primary key (group_id, user_id)
);

alter table group_members owner to secrum_user;

create index idx_group_members_user_id on group_members(user_id);
create index idx_group_members_group_id on group_members(group_id);

alter table messages
    add column if not exists group_id uuid null references groups(id);

create index if not exists idx_messages_group_id on messages(group_id);