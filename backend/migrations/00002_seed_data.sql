-- +goose Up
-- +goose StatementBegin
INSERT INTO
    users (id, name, email, password)
VALUES (
        '11111111-1111-1111-1111-111111111111',
        'Test User',
        'test@example.com',
        crypt (
            'password123',
            gen_salt ('bf', 12)
        )
    ),
    (
        '22222222-2222-2222-2222-222222222222',
        'Alex Teammate',
        'alex@example.com',
        crypt (
            'password123',
            gen_salt ('bf', 12)
        )
    ) ON CONFLICT (email) DO NOTHING;

INSERT INTO
    projects (
        id,
        name,
        description,
        owner_id
    )
VALUES (
        '33333333-3333-3333-3333-333333333333',
        'TaskFlow Demo Project',
        'Seed project for reviewer quick testing',
        '11111111-1111-1111-1111-111111111111'
    ) ON CONFLICT (id) DO NOTHING;

INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, creator_id, due_date)
VALUES
  ('44444444-4444-4444-4444-444444444444', 'Set up wireframes', 'Initial planning and sketches', 'todo', 'medium', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', NOW()::DATE + INTERVAL '5 day'),
  ('55555555-5555-5555-5555-555555555555', 'Implement auth endpoints', 'Register and login flow with JWT', 'in_progress', 'high', '33333333-3333-3333-3333-333333333333', '22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', NOW()::DATE + INTERVAL '2 day'),
  ('66666666-6666-6666-6666-666666666666', 'Prepare deployment checklist', 'Document GCP rollout steps', 'done', 'low', '33333333-3333-3333-3333-333333333333', NULL, '11111111-1111-1111-1111-111111111111', NULL)
ON CONFLICT (id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tasks
WHERE
    id IN (
        '44444444-4444-4444-4444-444444444444',
        '55555555-5555-5555-5555-555555555555',
        '66666666-6666-6666-6666-666666666666'
    );

DELETE FROM projects
WHERE
    id = '33333333-3333-3333-3333-333333333333';

DELETE FROM users
WHERE
    id IN (
        '11111111-1111-1111-1111-111111111111',
        '22222222-2222-2222-2222-222222222222'
    );
-- +goose StatementEnd