-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.users
(
    "ID" serial NOT NULL,
    login text COLLATE pg_catalog."default" NOT NULL,
    password bytea NOT NULL,
    CONSTRAINT users_pkey PRIMARY KEY ("ID")
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.users
    OWNER to postgres;

CREATE INDEX IF NOT EXISTS "IDX_USER_LOGIN"
    ON public.users USING btree
    (login COLLATE pg_catalog."default" ASC NULLS LAST)
    WITH (deduplicate_items=True)
    TABLESPACE pg_default;

CREATE TABLE IF NOT EXISTS public.content
(
    id serial NOT NULL,
    user_id integer NOT NULL,
    type integer NOT NULL,
    data bytea,
    "desc" bytea,
    meta text COLLATE pg_catalog."default",
    CONSTRAINT content_pkey PRIMARY KEY (id),
    CONSTRAINT fk_users FOREIGN KEY (user_id)
        REFERENCES public.users ("ID") MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.content
    OWNER to postgres;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE public.content;
DROP TABLE public.users;
-- +goose StatementEnd
