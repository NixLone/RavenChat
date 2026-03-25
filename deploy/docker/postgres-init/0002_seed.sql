INSERT INTO users (id, username, phone, email, role, password_hash) VALUES
('11111111-1111-1111-1111-111111111111','alice','+15550000001','alice@corp.local','system_admin','$2a$10$7EqJtq98hPqEX7fNZaFWoOhiT6Qf3q5M6P6q6Yh6gMPkTFogyXvC'),
('22222222-2222-2222-2222-222222222222','bob','+15550000002','bob@corp.local','user','$2a$10$7EqJtq98hPqEX7fNZaFWoOhiT6Qf3q5M6P6q6Yh6gMPkTFogyXvC'),
('33333333-3333-3333-3333-333333333333','carol','+15550000003','carol@corp.local','security_auditor','$2a$10$7EqJtq98hPqEX7fNZaFWoOhiT6Qf3q5M6P6q6Yh6gMPkTFogyXvC')
ON CONFLICT DO NOTHING;

INSERT INTO user_profiles(user_id, display_name, job_title)
VALUES
('11111111-1111-1111-1111-111111111111','Alice Admin','Operations'),
('22222222-2222-2222-2222-222222222222','Bob Builder','Engineering'),
('33333333-3333-3333-3333-333333333333','Carol Auditor','Security')
ON CONFLICT DO NOTHING;

INSERT INTO chats (id, kind, name) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','direct','Alice/Bob'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb','group','Eng Team Group')
ON CONFLICT DO NOTHING;

INSERT INTO chat_members(chat_id,user_id,role) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','11111111-1111-1111-1111-111111111111','member'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','22222222-2222-2222-2222-222222222222','member'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb','11111111-1111-1111-1111-111111111111','member'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb','22222222-2222-2222-2222-222222222222','member'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb','33333333-3333-3333-3333-333333333333','member')
ON CONFLICT DO NOTHING;

INSERT INTO messages(id,chat_id,sender_id,body_ciphertext,status) VALUES
('abababab-abab-abab-abab-abababababab','aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','11111111-1111-1111-1111-111111111111','Welcome to RavenChat MVP.','sent'),
('cdcdcdcd-cdcd-cdcd-cdcd-cdcdcdcdcdcd','bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb','22222222-2222-2222-2222-222222222222','Group chat initialized.','sent')
ON CONFLICT DO NOTHING;

INSERT INTO groups(id,chat_id,name,kind) VALUES
('44444444-4444-4444-4444-444444444444','bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb','Eng Team','board')
ON CONFLICT DO NOTHING;

INSERT INTO group_members(group_id,user_id,role) VALUES
('44444444-4444-4444-4444-444444444444','11111111-1111-1111-1111-111111111111','group_admin'),
('44444444-4444-4444-4444-444444444444','22222222-2222-2222-2222-222222222222','group_member'),
('44444444-4444-4444-4444-444444444444','33333333-3333-3333-3333-333333333333','group_member')
ON CONFLICT DO NOTHING;

INSERT INTO boards(id, group_id, name) VALUES ('55555555-5555-5555-5555-555555555555','44444444-4444-4444-4444-444444444444','Engineering Board')
ON CONFLICT DO NOTHING;

INSERT INTO board_members(board_id,user_id,role) VALUES
('55555555-5555-5555-5555-555555555555','11111111-1111-1111-1111-111111111111','board_admin'),
('55555555-5555-5555-5555-555555555555','22222222-2222-2222-2222-222222222222','board_member')
ON CONFLICT DO NOTHING;

INSERT INTO board_columns(id,board_id,name,sort_order) VALUES
('66666666-6666-6666-6666-666666666661','55555555-5555-5555-5555-555555555555','To Do',1),
('66666666-6666-6666-6666-666666666662','55555555-5555-5555-5555-555555555555','In Progress',2),
('66666666-6666-6666-6666-666666666663','55555555-5555-5555-5555-555555555555','Done',3)
ON CONFLICT DO NOTHING;

INSERT INTO tasks(id,board_id,column_id,title,description,assignee_id,creator_id,due_date,status) VALUES
('77777777-7777-7777-7777-777777777771','55555555-5555-5555-5555-555555555555','66666666-6666-6666-6666-666666666661','Bootstrap MVP','Create first vertical slice','22222222-2222-2222-2222-222222222222','11111111-1111-1111-1111-111111111111', now() + interval '5 day','todo'),
('77777777-7777-7777-7777-777777777772','55555555-5555-5555-5555-555555555555','66666666-6666-6666-6666-666666666662','Implement chat ws','Add realtime updates','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222', now() + interval '2 day','in_progress')
ON CONFLICT DO NOTHING;
