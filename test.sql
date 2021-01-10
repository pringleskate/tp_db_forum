insert into users(nickname, fullname, email, about) VALUES ('pkaterinaa', 'Kate', 'lala', 'about');

insert into forums(slug, title, user_nick) VALUES ('forum', 'test', 'pkaterinaa');

insert into  threads(author, forum, message, slug, title) VALUES
                    ('pkaterinaa', 'forum', 'message', 'thread', 'title');

insert into forum_users(forumid, userid) VALUES (1, 1);

--"INSERT INTO posts (author, created, forum, message, parent, thread, path) VALUES ($1,$2,$3,$4,$5,$6, array[(select currval('posts_id_seq')::integer)]) RETURNING ID"
--"INSERT INTO posts (author, created, forum, message, parent, thread, path) VALUES ($1,$2,$3,$4,$5,$6, (SELECT path FROM posts WHERE id = $5) || (select currval('posts_id_seq')::integer)) RETURNING ID"

insert into posts (author, created, forum, message, parent, thread, path) VALUES ('pkaterinaa', 'created', 'forum', 'post1', 0 , 1, array[(select currval('post_id_seq')::integer)]);
insert into posts (author, created, forum, message, parent, thread, path) VALUES ('pkaterinaa', 'created', 'forum', 'post1', 0 , 1, (SELECT path FROM posts WHERE id = 1) || (select currval('post_id_seq')::integer));