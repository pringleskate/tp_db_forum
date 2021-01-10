--SELECT pg_catalog.set_config('search_path', '', false);
CREATE EXTENSION IF NOT EXISTS citext;

DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE users
(
    ID       SERIAL NOT NULL PRIMARY KEY,

  --  nickname text NOT NULL unique ,
      nickname CITEXT NOT NULL UNIQUE COLLATE "POSIX",

    fullname TEXT   NOT NULL,

     email    CITEXT   NOT NULL UNIQUE,
    --email    text   NOT NULL UNIQUE,

    about    TEXT
);
--indexes
CREATE INDEX idx_nick_nick ON users (nickname);
CREATE INDEX idx_nick_email ON users (email);
CREATE INDEX idx_nick_cover ON users (nickname, fullname, about, email);

DROP TABLE IF EXISTS forums CASCADE;
CREATE TABLE forums
(
    ID        SERIAL                             NOT NULL PRIMARY KEY,

       slug      CITEXT                             NOT NULL UNIQUE,
    --slug      text                             NOT NULL UNIQUE,

    threads   INTEGER DEFAULT 0                  NOT NULL,
    posts     INTEGER DEFAULT 0                  NOT NULL,
    title     TEXT                               NOT NULL,

    user_nick CITEXT REFERENCES users (nickname) NOT NULL
    --user_nick text REFERENCES public.users (nickname) NOT NULL
);
--indexes
CREATE INDEX idx_forum_slug ON forums using hash(slug);

DROP TABLE IF EXISTS threads CASCADE;
CREATE TABLE threads
(
    ID      SERIAL                          NOT NULL PRIMARY KEY,

    author  CITEXT                          NOT NULL REFERENCES users (nickname),
    --author  TEXT                          NOT NULL REFERENCES public.users (nickname),

    --   created TEXT                            NOT NULL,
    created TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,

    forum   CITEXT REFERENCES forums (slug) NOT NULL,
    --forum   text REFERENCES public.forums (slug) NOT NULL,

    message TEXT                            NOT NULL,

    slug    CITEXT UNIQUE,
    --slug    text UNIQUE,


    title   TEXT                            NOT NULL,
    votes   INTEGER DEFAULT 0
);
--indexes
CREATE INDEX idx_thread_id ON threads(id);
CREATE INDEX idx_thread_slug ON threads(slug);
CREATE INDEX idx_thread_coverage ON threads (forum, created, id, slug, author, title, message, votes);

DROP TABLE IF EXISTS posts;
/*CREATE TABLE public.posts
(
    ID      SERIAL                          NOT NULL primary key,
    author  CITEXT                          NOT NULL REFERENCES users (nickname),
    created TIMESTAMP WITH TIME ZONE,
    edited  BOOLEAN DEFAULT false           NOT NULL,
    forum   CITEXT REFERENCES forums (slug) NOT NULL,
    message TEXT                            NOT NULL,
    parent  INTEGER DEFAULT 0               NOT NULL,
    thread  INTEGER REFERENCES threads (ID) NOT NULL,
    path    INTEGER[] DEFAULT '{0}':: INTEGER [] NOT NULL
);
--indexes
CREATE INDEX ON posts(thread, id, created, author, edited, message, parent, forum);
CREATE INDEX idx_post_thread_id_p_i ON posts (thread, (path[1]), id);
*/
DROP TABLE IF EXISTS forum_users;
CREATE TABLE forum_users
(
    forumID INTEGER REFERENCES forums (ID),
    userID  INTEGER REFERENCES users (ID)
);
ALTER TABLE IF EXISTS forum_users ADD CONSTRAINT uniq UNIQUE (forumID, userID);
CREATE INDEX idx_forum_user ON forum_users (forumID, userID);

DROP TABLE IF EXISTS votes;
CREATE TABLE votes
(
    user_nick CITEXT REFERENCES users (nickname) NOT NULL,
    --user_nick text REFERENCES users (nickname) NOT NULL,

    voice BOOLEAN NOT NULL,
    thread  INTEGER REFERENCES threads (ID) NOT NULL
    --  CONSTRAINT unique_vote UNIQUE (user_nick, thread)
);
ALTER TABLE IF EXISTS votes ADD CONSTRAINT uniq_votes UNIQUE (user_nick, thread);
CREATE INDEX idx_vote ON votes(thread, voice);


-----------------------------------------------------
CREATE TABLE posts (
    /* ID      SERIAL                          NOT NULL primary key,
     author  CITEXT                          NOT NULL REFERENCES users (nickname),
     created TIMESTAMP WITH TIME ZONE,
     edited  BOOLEAN DEFAULT false           NOT NULL,
     forum   CITEXT REFERENCES forums (slug) NOT NULL,
     message TEXT                            NOT NULL,
     parent  INTEGER DEFAULT 0               NOT NULL,
     thread  INTEGER REFERENCES threads (ID) NOT NULL,
     path    INTEGER[] DEFAULT '{0}':: INTEGER [] NOT NULL
         */
                              id integer NOT NULL PRIMARY KEY,
      author citext NOT NULL REFERENCES users (nickname),
            --                  author text NOT NULL REFERENCES public.users (nickname),

                              created text NOT NULL,

     forum  CITEXT REFERENCES forums (slug) NOT NULL,
                           --   forum  text REFERENCES public.forums (slug) NOT NULL,

                              edited boolean DEFAULT false NOT NULL,
                              message text NOT NULL,
                              parent integer DEFAULT 0 NOT NULL,
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

CREATE INDEX post_author_forum_index ON posts USING btree (author, forum);
CREATE INDEX post_forum_index ON posts USING btree (forum);
CREATE INDEX post_parent_index ON posts USING btree (parent);
CREATE INDEX post_path_index ON posts USING gin (path);
CREATE INDEX post_thread_index ON posts USING btree (thread);


/*SELECT pg_catalog.set_config('search_path', '', false);
CREATE EXTENSION IF NOT EXISTS citext;

DROP TABLE IF EXISTS public.users CASCADE;
CREATE TABLE public.users
(
    ID       SERIAL NOT NULL PRIMARY KEY,

    nickname text NOT NULL unique ,
  --  nickname CITEXT NOT NULL UNIQUE COLLATE "POSIX",

    fullname TEXT   NOT NULL,

   -- email    CITEXT   NOT NULL UNIQUE,
    email    text   NOT NULL UNIQUE,

    about    TEXT
);
--indexes
CREATE INDEX idx_nick_nick ON public.users (nickname);
CREATE INDEX idx_nick_email ON public.users (email);
CREATE INDEX idx_nick_cover ON public.users (nickname, fullname, about, email);

DROP TABLE IF EXISTS public.forums CASCADE;
CREATE TABLE public.forums
(
    ID        SERIAL                             NOT NULL PRIMARY KEY,

 --   slug      CITEXT                             NOT NULL UNIQUE,
    slug      text                             NOT NULL UNIQUE,

    threads   INTEGER DEFAULT 0                  NOT NULL,
    posts     INTEGER DEFAULT 0                  NOT NULL,
    title     TEXT                               NOT NULL,

   -- user_nick CITEXT REFERENCES users (nickname) NOT NULL
    user_nick text REFERENCES public.users (nickname) NOT NULL
);
--indexes
CREATE INDEX idx_forum_slug ON public.forums using hash(slug);

DROP TABLE IF EXISTS public.threads CASCADE;
CREATE TABLE public.threads
(
    ID      SERIAL                          NOT NULL PRIMARY KEY,

   -- author  CITEXT                          NOT NULL REFERENCES users (nickname),
    author  TEXT                          NOT NULL REFERENCES public.users (nickname),

 --   created TEXT                            NOT NULL,
   created TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,

   -- forum   CITEXT REFERENCES forums (slug) NOT NULL,
    forum   text REFERENCES public.forums (slug) NOT NULL,

    message TEXT                            NOT NULL,

   -- slug    CITEXT UNIQUE,
    slug    text UNIQUE,


    title   TEXT                            NOT NULL,
    votes   INTEGER DEFAULT 0
);
--indexes
CREATE INDEX idx_thread_id ON public.threads(id);
CREATE INDEX idx_thread_slug ON public.threads(slug);
CREATE INDEX idx_thread_coverage ON public.threads (forum, created, id, slug, author, title, message, votes);

DROP TABLE IF EXISTS public.posts;
/*CREATE TABLE public.posts
(
    ID      SERIAL                          NOT NULL primary key,
    author  CITEXT                          NOT NULL REFERENCES users (nickname),
    created TIMESTAMP WITH TIME ZONE,
    edited  BOOLEAN DEFAULT false           NOT NULL,
    forum   CITEXT REFERENCES forums (slug) NOT NULL,
    message TEXT                            NOT NULL,
    parent  INTEGER DEFAULT 0               NOT NULL,
    thread  INTEGER REFERENCES threads (ID) NOT NULL,
    path    INTEGER[] DEFAULT '{0}':: INTEGER [] NOT NULL
);
--indexes
CREATE INDEX ON posts(thread, id, created, author, edited, message, parent, forum);
CREATE INDEX idx_post_thread_id_p_i ON posts (thread, (path[1]), id);
*/
DROP TABLE IF EXISTS public.forum_users;
CREATE TABLE public.forum_users
(
    forumID INTEGER REFERENCES public.forums (ID),
    userID  INTEGER REFERENCES public.users (ID)
);
ALTER TABLE IF EXISTS public.forum_users ADD CONSTRAINT uniq UNIQUE (forumID, userID);
CREATE INDEX idx_forum_user ON public.forum_users (forumID, userID);

DROP TABLE IF EXISTS public.votes;
CREATE TABLE public.votes
(
   -- user_nick CITEXT REFERENCES users (nickname) NOT NULL,
    user_nick text REFERENCES public.users (nickname) NOT NULL,

    voice BOOLEAN NOT NULL,
    thread  INTEGER REFERENCES public.threads (ID) NOT NULL
  --  CONSTRAINT unique_vote UNIQUE (user_nick, thread)
);
ALTER TABLE IF EXISTS public.votes ADD CONSTRAINT uniq_votes UNIQUE (user_nick, thread);
CREATE INDEX idx_vote ON public.votes(thread, voice);


-----------------------------------------------------
CREATE TABLE public.posts (
                             /* ID      SERIAL                          NOT NULL primary key,
                              author  CITEXT                          NOT NULL REFERENCES users (nickname),
                              created TIMESTAMP WITH TIME ZONE,
                              edited  BOOLEAN DEFAULT false           NOT NULL,
                              forum   CITEXT REFERENCES forums (slug) NOT NULL,
                              message TEXT                            NOT NULL,
                              parent  INTEGER DEFAULT 0               NOT NULL,
                              thread  INTEGER REFERENCES threads (ID) NOT NULL,
                              path    INTEGER[] DEFAULT '{0}':: INTEGER [] NOT NULL
                                  */
                             id integer NOT NULL PRIMARY KEY,
                           --  author citext NOT NULL REFERENCES public.users (nickname),
                             author text NOT NULL REFERENCES public.users (nickname),

                             created text NOT NULL,

                            -- forum  CITEXT REFERENCES public.forums (slug) NOT NULL,
                             forum  text REFERENCES public.forums (slug) NOT NULL,

                             edited boolean DEFAULT false NOT NULL,
                             message text NOT NULL,
                             parent integer DEFAULT 0 NOT NULL,
                             thread  INTEGER REFERENCES public.threads (ID) NOT NULL,
                             path    INTEGER[] DEFAULT '{0}':: INTEGER [] NOT NULL
);

CREATE SEQUENCE if not exists public.post_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.post_id_seq OWNED BY public.posts.id;
ALTER TABLE ONLY public.posts ALTER COLUMN id SET DEFAULT nextval('public.post_id_seq'::regclass);
SELECT pg_catalog.setval('public.post_id_seq', 1, false);

CREATE INDEX post_author_forum_index ON public.posts USING btree (author, forum);
CREATE INDEX post_forum_index ON public.posts USING btree (forum);
CREATE INDEX post_parent_index ON public.posts USING btree (parent);
CREATE INDEX post_path_index ON public.posts USING gin (path);
CREATE INDEX post_thread_index ON public.posts USING btree (thread);

*/