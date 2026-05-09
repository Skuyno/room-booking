create table user_roles (
    id smallint primary key,
    code text not null unique
);

create table booking_statuses (
    id smallint primary key,
    code text not null unique
);

create table users (
    id uuid primary key,
    email text not null unique,
    role_id smallint not null references user_roles(id),
    password_hash text,
    created_at timestamptz not null default now()
);

create table rooms (
    id uuid primary key,
    name text not null,
    description text,
    capacity integer check (capacity is null or capacity > 0),
    created_at timestamptz not null default now()
);

create table schedules (
    id uuid primary key,
    room_id uuid not null references rooms(id) on delete cascade,
    days_of_week smallint[] not null,
    start_time time not null,
    end_time time not null,
    created_at timestamptz not null default now(),

    constraint schedules_room_unique unique (room_id),
    constraint schedules_days_non_empty check (cardinality(days_of_week) > 0),
    constraint schedules_days_valid check (days_of_week <@ array[1,2,3,4,5,6,7]::smallint[]),
    constraint schedules_time_valid check (start_time < end_time),
    constraint schedules_start_step check (extract(minute from start_time) in (0, 30)),
    constraint schedules_end_step check (extract(minute from end_time) in (0, 30))
);

create table bookings (
    id uuid primary key,
    room_id uuid not null references rooms(id) on delete restrict,
    user_id uuid not null references users(id) on delete restrict,
    status_id smallint not null references booking_statuses(id),
    start_at timestamptz not null,
    end_at timestamptz not null,
    conference_link text,
    created_at timestamptz not null default now(),
    cancelled_at timestamptz,

    constraint bookings_time_valid check (start_at < end_at),
    constraint bookings_duration_30m check (end_at = start_at + interval '30 minutes')
);
