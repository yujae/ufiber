-------- User 관리
CREATE TABLE public.user
(
    id character varying(500) COLLATE pg_catalog."default" NOT NULL,
    pw character varying(128) COLLATE pg_catalog."default" NOT NULL,
    nick character varying(20) COLLATE pg_catalog."default" NOT NULL,
    superuser boolean NOT NULL DEFAULT false,
    active boolean NOT NULL DEFAULT false,
    joined timestamp with time zone NOT NULL,
    activated timestamp with time zone,
    login timestamp with time zone,
    activekey character varying(32),
    CONSTRAINT pk_user PRIMARY KEY (id)
)
    WITH (
        OIDS = FALSE
        )
    TABLESPACE pg_default;

ALTER TABLE public.user OWNER to postgres;

CREATE INDEX ix_user_nick
    ON public.user USING btree
    (nick COLLATE pg_catalog."default" varchar_pattern_ops)
    TABLESPACE pg_default;

CREATE INDEX ix_user_activekey
    ON public.user USING btree
        (activekey COLLATE pg_catalog."default" varchar_pattern_ops)
    TABLESPACE pg_default;



-- drop table public.user cascade;

-- 임시 유저 생성 (비번 : 1)
-- insert into public.user (id, pw, nick, joined) values ('1', '$2a$10$TVfUPM8ZJmPK3EbyfBsFwe2vGi7LaI1mQCBfqatsP2jnp7Z/5HQCq', 'nick',now());

select * from public.user;


select name, setting
from pg_settings
where name = 'data_directory';