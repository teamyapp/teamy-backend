-- Seed tables
-- Users
INSERT INTO "user" (id, name, profile_url) VALUES (1, 'Luke Walker', 'http://static.teamyapp.com/user/1/profile.png');
INSERT INTO "user" (id, name, profile_url) VALUES (2, 'Mason Eyler', 'http://static.teamyapp.com/user/2/profile.png');
INSERT INTO "user" (id, name, profile_url) VALUES (3, 'Milton Cutler', 'http://static.teamyapp.com/user/3/profile.png');
INSERT INTO "user" (id, name, profile_url) VALUES (4, 'Milan Fender', 'http://static.teamyapp.com/user/4/profile.png');
INSERT INTO "user" (id, name, profile_url) VALUES (5, 'Martin Hitler ', 'http://static.teamyapp.com/user/5/profile.png');
INSERT INTO "user" (id, name, profile_url) VALUES (6, 'Myrnin Parker ', 'http://static.teamyapp.com/user/6/profile.png');
INSERT INTO "user" (id, name, profile_url) VALUES (7, 'Jade Alpin ', 'http://static.teamyapp.com/user/7/profile.png');
INSERT INTO "user" (id, name, profile_url) VALUES (8, 'Kyle Armster ', 'http://static.teamyapp.com/user/8/profile.png');

-- Tasks
INSERT INTO task (id, goal, owner_user_id) VALUES (1, 'Draft need attention, upcoming, and delivered API', 1);
INSERT INTO task (id, goal, due_at, owner_user_id) VALUES (2, 'Create/Delete/Edit task', '2021-10-31', 1);
INSERT INTO task (id, goal, due_at, owner_user_id) VALUES (3, 'To implement the repository storing the graph structures', '2021-10-31', 1);
INSERT INTO task (id, goal, owner_user_id) VALUES (4, 'Add unit tests to identity service core logic', 1);
INSERT INTO task (id, goal) VALUES (5, 'Move Identity service code into official Identity service repo');
INSERT INTO task (id, goal) VALUES (6, 'Productionise the core algorithm for graph based access control system');
INSERT INTO task (id, goal, owner_user_id) VALUES (7, 'Refactor identity service prototype with  PubSub interface', 1);
INSERT INTO task (id, goal, owner_user_id) VALUES (8, 'Write sample code for structural design pattern', 1);
INSERT INTO task (id, goal) VALUES (9, 'Create a Typescript client lib for identity service');
INSERT INTO task (id, goal) VALUES (10, 'Build a prototype of identity service with oauth working');

-- Task dependencies
INSERT INTO task_dependency (need_before, need_after) VALUES (5, 2);
INSERT INTO task_dependency (need_before, need_after) VALUES (4, 3);
INSERT INTO task_dependency (need_before, need_after) VALUES (1, 5);

-- Teams
INSERT INTO team (id, name, logo_url) VALUES (1, 'Teamy', 'http://static.teamyapp.com/team/1/logo.png');
INSERT INTO team (id, name, logo_url) VALUES (2, 'Google', 'http://static.teamyapp.com/team/2/logo.png');
INSERT INTO team (id, name, logo_url) VALUES (3, 'Apple', 'http://static.teamyapp.com/team/3/logo.png');
INSERT INTO team (id, name, logo_url) VALUES (4, 'Amazon', 'http://static.teamyapp.com/team/4/logo.png');
INSERT INTO team (id, name, logo_url) VALUES (5, 'Netflix', 'http://static.teamyapp.com/team/5/logo.png');
INSERT INTO team (id, name, logo_url) VALUES (6, 'Microsoft', 'http://static.teamyapp.com/team/6/logo.png');

-- Team members
INSERT INTO team_member (team_id, user_id, need_attention_task_id) VALUES (1, 1, 5);
INSERT INTO team_member (team_id, user_id) VALUES (1, 2);
INSERT INTO team_member (team_id, user_id, need_attention_task_id) VALUES (1, 3, 4);
INSERT INTO team_member (team_id, user_id, need_attention_task_id) VALUES (1, 4, 8);
INSERT INTO team_member (team_id, user_id) VALUES (1, 5);
INSERT INTO team_member (team_id, user_id) VALUES (2, 1);
INSERT INTO team_member (team_id, user_id) VALUES (3, 1);
INSERT INTO team_member (team_id, user_id) VALUES (3, 2);
INSERT INTO team_member (team_id, user_id) VALUES (3, 7);
INSERT INTO team_member (team_id, user_id) VALUES (3, 8);

-- Task statuses
INSERT INTO task_status (value, name) VALUES (1, 'Upcoming');
INSERT INTO task_status (value, name) VALUES (2, 'InProgress');
INSERT INTO task_status (value, name) VALUES (3, 'Delivered');

-- Team tasks
INSERT INTO team_task (team_id, task_id, task_status) VALUES (1, 1, 1);
INSERT INTO team_task (team_id, task_id, task_status) VALUES (1, 2, 1);
INSERT INTO team_task (team_id, task_id, task_status) VALUES (1, 3, 1);
INSERT INTO team_task (team_id, task_id, task_status) VALUES (1, 4, 1);
INSERT INTO team_task (team_id, task_id, task_status) VALUES (1, 5, 2);
INSERT INTO team_task (team_id, task_id, task_status) VALUES (1, 6, 2);
INSERT INTO team_task (team_id, task_id, task_status) VALUES (1, 7, 3);
INSERT INTO team_task (team_id, task_id, task_status) VALUES (1, 8, 3);
INSERT INTO team_task (team_id, task_id, task_status) VALUES (3, 9, 1);
INSERT INTO team_task (team_id, task_id, task_status) VALUES (3, 10, 1);

-- User state
INSERT INTO user_state (user_id, active_team_id) VALUES (1, 1);
INSERT INTO user_state (user_id, active_team_id) VALUES (2, 3);
INSERT INTO user_state (user_id) VALUES (6);
INSERT INTO user_state (user_id, active_team_id) VALUES (8, 3);
