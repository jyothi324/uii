-- Read-only Postgres role for the AI chatbot (text-to-SQL) feature.
--
-- This is the real safety backstop for lead-services' POST /api/ai/ask
-- endpoint: the app-level guardrail in services/sql_guardrail.go rejects
-- non-SELECT statements before execution, but that check is defense in
-- depth, not a substitute for running the generated query under a role
-- that is physically incapable of mutating data.
--
-- This script is NOT run automatically (lead-services uses GORM
-- auto-migrate for its own schema, not a migration-file system — see
-- CLAUDE.md). It's a one-time ops step per Postgres instance, run manually
-- via psql by whoever administers that environment's database. Because all
-- tenant databases (tenant_app, tenant_tampa, tenant_miami, ...) live on
-- the same physical Postgres instance, this only needs to be run once per
-- instance (dev, staging, prod each have their own).
--
-- Usage:
--   psql "host=<host> port=<port> user=<superuser> dbname=postgres" -f readonly_role.sql
--
-- After running, set AI_DB_USER=tulip_ai_readonly and AI_DB_PASSWORD=<the
-- password you choose below> in lead-services' .env file for that
-- environment.

-- 1. Create the role. Replace the password before running against any
--    real environment — do not commit a real password anywhere.
CREATE ROLE tulip_ai_readonly WITH LOGIN PASSWORD 'CHANGE_ME_BEFORE_RUNNING';

-- 2. For EACH tenant database on this instance, run the block below
--    (psql \c or a separate connection per database — GRANT/ALTER DEFAULT
--    PRIVILEGES are per-database, they don't cascade across databases).
--    Uncommented and ready to run as-is for tenant_app, the baseline
--    tenant every fresh clone gets from `make setup`. If this instance
--    has additional tenant databases (tenant_tampa, tenant_<name>, ...),
--    copy this whole block once per tenant and swap tenant_app for that
--    tenant's name in the four places it appears below.

\c tenant_app

GRANT CONNECT ON DATABASE tenant_app TO tulip_ai_readonly;
GRANT USAGE ON SCHEMA public TO tulip_ai_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO tulip_ai_readonly;

-- Ensures tables created later by GORM auto-migrate are automatically
-- readable too, without re-running this script on every schema change.
-- Must be run by the role that owns the tables (tulipdev, per schema.sql)
-- so it applies to objects that role creates going forward.
ALTER DEFAULT PRIVILEGES FOR ROLE tulipdev IN SCHEMA public
  GRANT SELECT ON TABLES TO tulip_ai_readonly;

-- Explicitly confirm this role has no write capability (belt-and-suspenders
-- — it shouldn't have been granted any INSERT/UPDATE/DELETE above, but this
-- makes the intent unambiguous if anyone edits this script later):
REVOKE INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER
  ON ALL TABLES IN SCHEMA public FROM tulip_ai_readonly;

-- agents is allow-listed for Bloom (services/ai_schema.go
-- AIQueryableTables) so it can resolve staff names/roles for questions
-- like "who is assigned to this lead", but the table also holds
-- password / password_hashed (literal credentials) and details (jsonb
-- that embeds the TOTP secret — see models.AgentDetails). The blanket
-- GRANT SELECT ON ALL TABLES above would otherwise let a generated query
-- read those straight out of this role, which is the actual safety
-- boundary here — services/sql_guardrail.go's validateAgentsColumnAccess
-- is defense in depth on top of this, not a substitute for it. Revoke
-- the table-wide grant and re-grant only the columns
-- AISchemaContext actually describes.
REVOKE SELECT ON agents FROM tulip_ai_readonly;
GRANT SELECT (id, name, email, role, assignee_type, superuser, created_at, updated_at)
  ON agents TO tulip_ai_readonly;

-- 3. Verify (run as tulip_ai_readonly against a tenant DB):
--   INSERT INTO leads (id) VALUES ('should-fail');   -- must error: permission denied
--   SELECT COUNT(*) FROM leads;                       -- must succeed
--   SELECT password FROM agents;                      -- must error: permission denied for column
--   SELECT id, name, email FROM agents;                -- must succeed
