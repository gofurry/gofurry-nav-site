-- Canonical gfa baseline, version 20260823000000.
-- Derived from the audited production schema and normalized to omit owners, ACLs, and data.
-- +goose Up

CREATE SEQUENCE public."gfa_admin_account_id_seq"
    AS bigint
    START WITH 1
    INCREMENT BY 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE SEQUENCE public."gfa_admin_audit_log_id_seq"
    AS bigint
    START WITH 1
    INCREMENT BY 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE TABLE public."gfa_admin_account" (
    "id" bigint DEFAULT nextval('gfa_admin_account_id_seq'::regclass) NOT NULL,
    "password_hash" text NOT NULL,
    "session_version" bigint NOT NULL,
    "created_at" timestamp(0) without time zone NOT NULL,
    "updated_at" timestamp(0) without time zone NOT NULL,
    "password_updated_at" timestamp(0) without time zone
);

CREATE TABLE public."gfa_admin_audit_log" (
    "id" bigint DEFAULT nextval('gfa_admin_audit_log_id_seq'::regclass) NOT NULL,
    "action" character varying(64) NOT NULL,
    "resource" character varying(128) NOT NULL,
    "target_id" character varying(128),
    "operator" character varying(64) NOT NULL,
    "session_version" bigint DEFAULT 0 NOT NULL,
    "request_id" character varying(128),
    "ip_address" character varying(64),
    "user_agent" text,
    "before_data" text,
    "after_data" text,
    "created_at" timestamp(0) without time zone NOT NULL
);

ALTER SEQUENCE public."gfa_admin_account_id_seq" OWNED BY public."gfa_admin_account"."id";
ALTER SEQUENCE public."gfa_admin_audit_log_id_seq" OWNED BY public."gfa_admin_audit_log"."id";

ALTER TABLE ONLY public."gfa_admin_account" ADD CONSTRAINT "gfa_admin_account_created_at_not_null" NOT NULL created_at;
ALTER TABLE ONLY public."gfa_admin_account" ADD CONSTRAINT "gfa_admin_account_id_not_null" NOT NULL id;
ALTER TABLE ONLY public."gfa_admin_account" ADD CONSTRAINT "gfa_admin_account_password_hash_not_null" NOT NULL password_hash;
ALTER TABLE ONLY public."gfa_admin_account" ADD CONSTRAINT "gfa_admin_account_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfa_admin_account" ADD CONSTRAINT "gfa_admin_account_session_version_not_null" NOT NULL session_version;
ALTER TABLE ONLY public."gfa_admin_account" ADD CONSTRAINT "gfa_admin_account_updated_at_not_null" NOT NULL updated_at;
ALTER TABLE ONLY public."gfa_admin_audit_log" ADD CONSTRAINT "gfa_admin_audit_log_action_not_null" NOT NULL action;
ALTER TABLE ONLY public."gfa_admin_audit_log" ADD CONSTRAINT "gfa_admin_audit_log_created_at_not_null" NOT NULL created_at;
ALTER TABLE ONLY public."gfa_admin_audit_log" ADD CONSTRAINT "gfa_admin_audit_log_id_not_null" NOT NULL id;
ALTER TABLE ONLY public."gfa_admin_audit_log" ADD CONSTRAINT "gfa_admin_audit_log_operator_not_null" NOT NULL operator;
ALTER TABLE ONLY public."gfa_admin_audit_log" ADD CONSTRAINT "gfa_admin_audit_log_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfa_admin_audit_log" ADD CONSTRAINT "gfa_admin_audit_log_resource_not_null" NOT NULL resource;
ALTER TABLE ONLY public."gfa_admin_audit_log" ADD CONSTRAINT "gfa_admin_audit_log_session_version_not_null" NOT NULL session_version;




-- Baseline intentionally has no Down section: dropping an application schema is not a safe rollback.
