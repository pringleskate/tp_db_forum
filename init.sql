CREATE
EXTENSION IF NOT EXISTS citext;

DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS forums CASCADE;
DROP TABLE IF EXISTS threads CASCADE;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS forum_users;
DROP TABLE IF EXISTS votes;

CREATE TABLE users
(
    ID       SERIAL NOT NULL PRIMARY KEY,
    nickname CITEXT NOT NULL UNIQUE COLLATE "POSIX",
    fullname TEXT   NOT NULL,
    email    CITEXT NOT NULL UNIQUE,
    about    TEXT
);


CREATE TABLE forums
(
    ID        SERIAL                             NOT NULL PRIMARY KEY,
    slug      CITEXT                             NOT NULL UNIQUE,
    threads   INTEGER DEFAULT 0                  NOT NULL,
    posts     INTEGER DEFAULT 0                  NOT NULL,
    title     TEXT                               NOT NULL,
    user_nick CITEXT REFERENCES users (nickname) NOT NULL
);


CREATE TABLE threads
(
    ID      SERIAL                                 NOT NULL PRIMARY KEY,
    author  CITEXT                                 NOT NULL REFERENCES users (nickname),
    created TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
    forum   CITEXT REFERENCES forums (slug)        NOT NULL,
    message TEXT                                   NOT NULL,
    slug    CITEXT UNIQUE,
    title   TEXT                                   NOT NULL,
    votes   INTEGER                  DEFAULT 0
);


CREATE TABLE forum_users
(
    forumID INTEGER REFERENCES forums (ID),
    userID  INTEGER REFERENCES users (ID)
);
ALTER TABLE IF EXISTS forum_users ADD CONSTRAINT uniq UNIQUE (forumID, userID);

CREATE TABLE votes
(
    user_nick CITEXT REFERENCES users (nickname) NOT NULL,
    voice     BOOLEAN                            NOT NULL,
    thread    INTEGER REFERENCES threads (ID)    NOT NULL
);
ALTER TABLE IF EXISTS votes ADD CONSTRAINT uniq_votes UNIQUE (user_nick, thread);


CREATE TABLE posts
(

    id      integer                                NOT NULL PRIMARY KEY,
    author  citext                                 NOT NULL REFERENCES users (nickname),
    created text                                   NOT NULL,

    forum   CITEXT REFERENCES forums (slug)        NOT NULL,

    edited  boolean DEFAULT false                  NOT NULL,
    message text                                   NOT NULL,
    parent  integer DEFAULT 0                      NOT NULL,
    thread  INTEGER REFERENCES public.threads (ID) NOT NULL,
    path    INTEGER[] DEFAULT '{0}':: INTEGER [] NOT NULL
);

CREATE SEQUENCE if not exists post_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE post_id_seq OWNED BY posts.id;
ALTER TABLE ONLY posts ALTER COLUMN id SET DEFAULT nextval('post_id_seq'::regclass);
SELECT pg_catalog.setval('post_id_seq', 1, false);

--user indexes
CREATE
INDEX idx_nick_nick ON users (nickname);
CREATE
INDEX idx_nick_email ON users (email);
CREATE
INDEX idx_nick_cover ON users (nickname, fullname, about, email);

--forum indexes
CREATE
INDEX idx_forum_slug ON forums using hash(slug);

--thread indexes
CREATE
INDEX idx_thread_id ON threads(id);
CREATE
INDEX idx_thread_slug ON threads(slug);
CREATE
INDEX idx_thread_coverage ON threads (forum, created, id, slug, author, title, message, votes);

--vote indexes
CREATE
INDEX idx_vote ON votes(thread, voice);

--forum user indexes
CREATE
INDEX idx_forum_user ON forum_users (forumID, userID);

--post indexes
CREATE
INDEX post_author_forum_index ON posts USING btree (author, forum);
CREATE
INDEX post_forum_index ON posts USING btree (forum);
CREATE
INDEX post_parent_index ON posts USING btree (parent);
CREATE
INDEX post_path_index ON posts USING gin (path);
CREATE
INDEX post_thread_index ON posts USING btree (thread);
