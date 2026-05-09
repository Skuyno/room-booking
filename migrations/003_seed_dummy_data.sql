insert into user_roles (id, code)
values
    (1, 'admin'),
    (2, 'user')
on conflict (id) do nothing;

insert into booking_statuses (id, code)
values
    (1, 'active'),
    (2, 'cancelled')
on conflict (id) do nothing;

insert into users (id, email, role_id)
values
    ('11111111-1111-1111-1111-111111111111', 'admin@example.com', 1),
    ('22222222-2222-2222-2222-222222222222', 'user@example.com', 2)
on conflict (id) do nothing;