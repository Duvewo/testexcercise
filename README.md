**Известные проблемы**
1. Авторизация работает, но ничего не делает
2. Необходимо провести рефакторинг
3. Слайс магов не обновляется при выходе/смерти мага

Флаги:
```
var (
	FLAG_DBADDR  = flag.String("db-addr", os.Getenv("DB_ADDR"), "Database connectivity string")
	FLAG_SRVADDR = flag.String("srv-addr", os.Getenv("SRV_ADDR"), "Server address")
)
```

Схема __users__:
```
-- Table: public.users

-- DROP TABLE IF EXISTS public.users;

CREATE TABLE IF NOT EXISTS public.users
(
    id bigint NOT NULL DEFAULT nextval('users_id_seq'::regclass),
    username character varying COLLATE pg_catalog."default" NOT NULL,
    password character varying COLLATE pg_catalog."default" NOT NULL,
    health_points integer NOT NULL DEFAULT 100,
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT "unique user" UNIQUE (username)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.users
    OWNER to wizard;
```

