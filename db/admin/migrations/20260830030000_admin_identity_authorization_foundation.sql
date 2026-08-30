-- V3-P0.5.2-A Admin identity and authorization foundation.
-- +goose Up

-- The released Admin contract allowed exactly one account. Refuse to guess
-- which legacy row should become Owner if that invariant was already broken.
-- +goose StatementBegin
DO $migration$
BEGIN
    IF (SELECT count(*) FROM public.gfa_admin_account) > 1 THEN
        RAISE EXCEPTION 'cannot migrate Admin identity: legacy account table contains more than one row';
    END IF;
END
$migration$;
-- +goose StatementEnd

ALTER TABLE public.gfa_admin_account
    ADD COLUMN username text,
    ADD COLUMN display_name text,
    ADD COLUMN role text,
    ADD COLUMN status text,
    ADD COLUMN last_login_at timestamp(0) without time zone;

UPDATE public.gfa_admin_account
SET username = 'owner',
    display_name = 'Owner',
    role = 'owner',
    status = 'active';

ALTER TABLE public.gfa_admin_account
    ALTER COLUMN username SET NOT NULL,
    ALTER COLUMN display_name SET NOT NULL,
    ALTER COLUMN role SET NOT NULL,
    ALTER COLUMN status SET NOT NULL,
    ALTER COLUMN session_version SET DEFAULT 1,
    ADD CONSTRAINT gfa_admin_account_username_check CHECK (
        username = lower(btrim(username))
        AND username ~ '^[a-z0-9][a-z0-9._-]{2,63}$'
    ),
    ADD CONSTRAINT gfa_admin_account_display_name_check CHECK (
        btrim(display_name) <> '' AND display_name = btrim(display_name) AND length(display_name) <= 128
    ),
    ADD CONSTRAINT gfa_admin_account_role_check CHECK (role IN ('owner', 'developer', 'operator')),
    ADD CONSTRAINT gfa_admin_account_status_check CHECK (status IN ('active', 'disabled')),
    ADD CONSTRAINT gfa_admin_account_session_version_check CHECK (session_version > 0);

CREATE UNIQUE INDEX uq_gfa_admin_account_username_ci
    ON public.gfa_admin_account (lower(username));
CREATE INDEX idx_gfa_admin_account_role_status
    ON public.gfa_admin_account (role, status, id);

-- The old singleton insert used id=1 explicitly, which did not necessarily
-- advance the owned sequence. Make the next multi-account insert safe.
SELECT setval(
    'public.gfa_admin_account_id_seq',
    GREATEST(COALESCE(max(id), 1), 1),
    count(*) > 0
)
FROM public.gfa_admin_account;

ALTER TABLE public.gfa_admin_audit_log
    ADD COLUMN operator_account_id bigint,
    ADD COLUMN operator_name text,
    ADD COLUMN operator_role text;

-- Every historical row came from the one legacy account when that account
-- exists. With no account, retain the old operator text as a system snapshot.
UPDATE public.gfa_admin_audit_log audit
SET operator_account_id = account.id,
    operator_name = account.display_name,
    operator_role = account.role
FROM public.gfa_admin_account account;

UPDATE public.gfa_admin_audit_log
SET operator_name = COALESCE(NULLIF(btrim(operator), ''), 'system'),
    operator_role = 'system'
WHERE operator_name IS NULL;

ALTER TABLE public.gfa_admin_audit_log
    ALTER COLUMN operator_name SET NOT NULL,
    ALTER COLUMN operator_role SET NOT NULL,
    ADD CONSTRAINT fk_gfa_admin_audit_operator_account FOREIGN KEY (operator_account_id)
        REFERENCES public.gfa_admin_account(id) ON DELETE RESTRICT,
    ADD CONSTRAINT gfa_admin_audit_operator_name_check CHECK (
        btrim(operator_name) <> '' AND length(operator_name) <= 128
    ),
    ADD CONSTRAINT gfa_admin_audit_operator_role_check CHECK (
        operator_role IN ('owner', 'developer', 'operator', 'system')
    );

CREATE INDEX idx_gfa_admin_audit_operator_account_created
    ON public.gfa_admin_audit_log (operator_account_id, created_at DESC)
    WHERE operator_account_id IS NOT NULL;

-- This expand-contract migration intentionally has no Down section. Removing
-- multi-account identity would discard role, status, and durable audit data.
