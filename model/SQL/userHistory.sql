-------- User 관리
CREATE TABLE public.user_history
(
    id serial primary key ,
    user_id character varying(500),
    msg character varying(500),
    accessed timestamp with time zone NOT NULL
)
    WITH (
        OIDS = FALSE
        )
    TABLESPACE pg_default;

ALTER TABLE public.user_history OWNER to postgres;

-- drop table public.user_history cascade;


select * from user_history;