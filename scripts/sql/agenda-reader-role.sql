-- Creates the read-only "agenda_reader" Postgres role used by agenda-service
-- (docs/conversations/14-agenda-service-stage2-reader-and-role.md).
--
-- Why this file exists: roles are cluster-wide objects, so `pg_dump` (which
-- dumps ONE database) never includes them -- but the dump DOES include the
-- GRANTs that reference this role. Restoring a dump onto a fresh Postgres
-- therefore fails at the first `GRANT ... TO agenda_reader` unless this runs
-- first. Run it BEFORE loading dump.sql; the dump then re-applies the grants.
--
-- The password is passed in as the psql variable `agenda_password` (from
-- AGENDA_DB_PASSWORD in .env) and quoted safely by :'...'. Feed it over
-- stdin rather than `psql -v` on the command line, so it never shows up in
-- the host's process list. scripts/restore-test.sh does exactly this.
--
-- Grants the dump restores (listed here for reference / a non-restore setup):
--   GRANT USAGE ON SCHEMA public TO agenda_reader;
--   GRANT SELECT ON TABLE public.items TO agenda_reader;

CREATE ROLE agenda_reader LOGIN PASSWORD :'agenda_password';
