create unique index bookings_one_active_per_room_start_idx
    on bookings(room_id, start_at)
    where status_id = 1;

create index bookings_user_status_created_idx
    on bookings(user_id, status_id, created_at desc);

create index bookings_user_start_idx
    on bookings(user_id, start_at);

create index bookings_room_start_idx
    on bookings(room_id, start_at);
