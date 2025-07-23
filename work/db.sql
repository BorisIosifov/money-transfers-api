DROP TABLE IF EXISTS public.messages;
DROP TABLE IF EXISTS public.invoices;
DROP TABLE IF EXISTS public.transactions;
DROP TABLE IF EXISTS public.requests;
DROP TABLE IF EXISTS public.rates;
DROP TABLE IF EXISTS public.users_accounts;
DROP TABLE IF EXISTS public.accounts;
DROP TABLE IF EXISTS public.banks;
DROP TABLE IF EXISTS public.sessions;
DROP TABLE IF EXISTS public.users;

-- Пользователи (users)
-- user_id
-- email
-- password
-- phone
-- type (password, google, facebook, etc.)
-- external_user_id
-- telegram_id
-- name
-- ctime

CREATE TABLE public.users (
    user_id SERIAL PRIMARY KEY,
    email character varying DEFAULT ''::character varying NOT NULL,
    password character varying DEFAULT ''::character varying NOT NULL,
    phone character varying DEFAULT ''::character varying NOT NULL,
    type character varying DEFAULT ''::character varying NOT NULL,
    external_user_id integer DEFAULT 0 NOT NULL,
    telegram_chat_id integer DEFAULT 0 NOT NULL,
    name character varying DEFAULT ''::character varying NOT NULL,
    role character varying DEFAULT ''::character varying NOT NULL,
    ctime timestamp without time zone DEFAULT now() NOT NULL
);


-- Сессии (sessions)
-- session_id
-- user_id
-- data (json)
-- ctime

CREATE TABLE public.sessions (
    session_id character varying DEFAULT ''::character varying NOT NULL PRIMARY KEY,
    user_id integer DEFAULT NULL REFERENCES public.users(user_id),
    data json DEFAULT '{}'::json NOT NULL,
    ctime timestamp without time zone DEFAULT now() NOT NULL
);


-- Банки (banks)
-- bank_id
-- country
-- name

CREATE TABLE public.banks (
    bank_id SERIAL PRIMARY KEY,
    country character varying DEFAULT ''::character varying NOT NULL,
    bank_external_id character varying DEFAULT ''::character varying NOT NULL,
    name character varying DEFAULT ''::character varying NOT NULL,
    ctime timestamp without time zone DEFAULT now() NOT NULL
);


-- Счета (accounts)
-- account_id
-- country
-- phone
-- name
-- bank_id
-- ctime

CREATE TABLE public.accounts (
    account_id SERIAL PRIMARY KEY,
    country character varying DEFAULT ''::character varying NOT NULL,
    phone character varying DEFAULT ''::character varying NOT NULL,
    name character varying DEFAULT ''::character varying NOT NULL,
    bank_id integer DEFAULT NULL REFERENCES public.banks(bank_id),
    ctime timestamp without time zone DEFAULT now() NOT NULL
);


-- Пользователи-Счета (users_accounts)
-- user_id
-- account_id
-- ctime

CREATE TABLE public.users_accounts (
    user_id integer DEFAULT NULL REFERENCES public.users(user_id),
    account_id integer DEFAULT NULL REFERENCES public.accounts(account_id),
    ctime timestamp without time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.users_accounts
    ADD CONSTRAINT users_accounts_pkey PRIMARY KEY (user_id, account_id);


-- Курсы (rates)
-- rate_id
-- currency_from
-- currency_to
-- rate
-- ctime

CREATE TABLE public.rates (
    rate_id SERIAL PRIMARY KEY,
    currency_from character varying DEFAULT ''::character varying NOT NULL,
    currency_to character varying DEFAULT ''::character varying NOT NULL,
    rate real DEFAULT 0 NOT NULL,
    ctime timestamp without time zone DEFAULT now() NOT NULL
);


-- Заявки (requests)
-- request_id
-- user_id
-- currency
-- rate_id
-- amount
-- country_from
-- country_to
-- account_to
-- status
-- ctime

CREATE TABLE public.requests (
    request_id SERIAL PRIMARY KEY,
    user_id integer DEFAULT NULL REFERENCES public.users(user_id),
    currency character varying DEFAULT ''::character varying NOT NULL,
    rate_id integer DEFAULT NULL REFERENCES public.rates(rate_id),
    amount real DEFAULT 0 NOT NULL,
    country_to character varying DEFAULT ''::character varying NOT NULL,
    account_to integer DEFAULT NULL REFERENCES public.accounts(account_id),
    status integer DEFAULT 0 NOT NULL,
    ctime timestamp without time zone DEFAULT now() NOT NULL
);


-- Транзакции (transactions)
-- transaction_id
-- account_to
-- request_from
-- request_to
-- status
-- ctime

CREATE TABLE public.transactions (
    transaction_id SERIAL PRIMARY KEY,
    account_to integer DEFAULT NULL REFERENCES public.accounts(account_id),
    request_from integer DEFAULT NULL REFERENCES public.requests(request_id),
    request_to integer DEFAULT NULL REFERENCES public.requests(request_id),
    status integer DEFAULT NULL,
    ctime timestamp without time zone DEFAULT now() NOT NULL
);


-- Инвойсы (invoices)
-- invoice_id
-- transaction_id
-- ctime

CREATE TABLE public.invoices (
    invoice_id SERIAL PRIMARY KEY,
    transaction_id integer DEFAULT NULL REFERENCES public.transactions(transaction_id),
    ctime timestamp without time zone DEFAULT now() NOT NULL
);


-- Сообщения (messages)
-- message_id
-- user_id
-- type (from user, to user)
-- transaction_id
-- text
-- ctime

CREATE TABLE public.messages (
    message_id SERIAL PRIMARY KEY,
    user_from integer DEFAULT NULL REFERENCES public.users(user_id),
    user_to integer DEFAULT NULL REFERENCES public.users(user_id),
    type integer DEFAULT NULL,
    transaction_id integer DEFAULT NULL REFERENCES public.transactions(transaction_id),
    text text DEFAULT '' NOT NULL,
    ctime timestamp without time zone DEFAULT now() NOT NULL
);
