package services

// AIQueryableTables is the allow-list of tables the AI chatbot's generated
// SQL is permitted to reference. Kept in sync by hand with AISchemaContext
// below and enforced by sql_guardrail.go — the model is never given (and
// the guardrail never permits) access to anything outside this set.
var AIQueryableTables = []string{
	"leads",
	"jobs",
	"estimates",
	"customers",
	"agents",
	"stages",
	"tasks",
	"lead_associates",
	"associate_customers",
	// Stage detail + reference lookup tables — no jsonb-buried data, and
	// either a clear stage_id/task_id join to a table above or a
	// standalone lookup with no PII.

	"stage_tasks",

	"stage_users",
	"county_taxes",

	"pickup_locations",

	// Residential entry / property-type taxonomy, tied to leads/estimates
	// via lead_id/estimate_id.
	"manager_residential_entries",
	"manager_residential_types",
	"residential_entries",
	// Services & materials catalog. Most per-line detail lives in a
	// `details` jsonb column rather than plain scalar columns — see the
	// caveat in AISchemaContext below.
	"services",
	"service_categories",
	"service_groups",

	"service_rows",

	"material_rows",
	"material_templates",

	// Physical warehouse inventory — owned by the separate `inventory`
	// service, not lead-services, but it reads/writes these same
	// tenant_<name> Postgres databases (same RDS instance, same DB user,
	// confirmed against both dev and prod task definitions), so the AI's
	// read-only role already has SELECT on them via the blanket
	// "GRANT SELECT ON ALL TABLES IN SCHEMA public" in readonly_role.sql —
	// no DB-level grant changes were needed to add these, only this
	// allow-list entry and the schema description below. Distinct from
	// material_rows/material_templates above, which are lead-services'
	// own per-estimate line items, not actual stock.
	"materials",
	"materialchunks",
	"suppliers",
	"purchases",
	"purchase_receiveds",
}

// If a table or column below no longer matches the live schema, update
// this constant and AIQueryableTables together — they're cross-referenced
// by hand, nothing enforces they stay in sync automatically.
const AISchemaContext = `You are generating a single read-only PostgreSQL SELECT query to answer a
business question about a home-services CRM tenant database. Only the
following tables exist for your purposes — do not reference any other
table, and do not use information_schema, pg_catalog, or any system table.

TABLE leads
  id varchar PK, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz,
  lead_type varchar, job_name varchar, delivery_type varchar, status text,
  detailed_lead_status varchar, payment_term varchar, pricing_type varchar,
  sales1_person_id varchar (-> agents.id), sales2_person_id varchar (-> agents.id),
  rejection_reason varchar, details jsonb, referrer jsonb
  -- status is one of NEW, IN_PROGRESS, COMPLETED, REJECTED. A lead becomes a
  -- "job" specifically when status reaches COMPLETED (i.e. it's sold) — at
  -- that point a row is created in TABLE jobs below. For "how many
  -- jobs"/"leads converted to jobs"/"leads sold" questions, use TABLE jobs
  -- (or "leads WHERE status = 'COMPLETED'") — do not invent a status value
  -- like 'CONVERTED' or 'JOB', it does not exist.
  -- ALWAYS add "deleted_at IS NULL" unless the question explicitly asks about
  -- deleted/rejected/historical records.
  --
  -- "customer"/"client"/"profile" questions (e.g. "which customer has the
  -- most leads"): this product has no true customer entity linked to
  -- leads — TABLE customers is NOT connected to leads (see its note above).
  -- Instead, the app's own "Lead Profiles" feature defines a profile as one
  -- distinct value of details->>'Phone' — that IS the customer identity
  -- here. Answer these questions the same way: GROUP BY details->>'Phone'
  -- on this table (filtering deleted_at IS NULL), never by joining to
  -- customers. job_name is the job's name, not a person's name — do not
  -- present it as the customer's name, but you may mention it as context
  -- (e.g. "the phone number ending in ... has the most leads, most
  -- recently for '<job_name>'"). If details->>'Phone' is null/empty for
  -- the relevant rows, say the phone number is missing rather than
  -- omitting that from the answer.

TABLE jobs
  id text PK, job_id text, lead_id text (-> leads.id), estimate_id text (-> estimates.id),
  stages jsonb, details jsonb, created_at timestamptz, updated_at timestamptz
  -- One row per lead once it's sold (leads.status = 'COMPLETED') — this row
  -- IS "the job". No soft-delete column. details is a jsonb blob containing
  -- jobName and Total (the job's dollar value) — use "details->>'Total'" for
  -- job value questions. stages is a cached snapshot of stage/task progress;
  -- for live per-stage state prefer joining TABLE stages/stage_tasks by
  -- lead_id instead of parsing this blob.

TABLE estimates
  id text PK, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz,
  lead_id varchar (-> leads.id), name text, material_total numeric, service_total numeric,
  status text, details jsonb
  -- ALWAYS add "deleted_at IS NULL" unless asked about deleted estimates.

TABLE customers
  id text PK, name text, email text, phone text
  -- No soft-delete column on this table — do not add deleted_at for customers.
  -- IMPORTANT: there is NO column on TABLE leads that references this table
  -- (no customer_id anywhere), and no join table between them either. Do
  -- NOT join leads to this table, and do not use it to answer questions
  -- about which "customer"/"client"/"profile" has the most/fewest leads —
  -- see the note on TABLE leads below for how those questions actually
  -- get answered in this product.

TABLE agents
  id varchar PK, name varchar, email varchar, role varchar, assignee_type varchar,
  superuser boolean, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz
  -- Staff/sales users, not customers. ALWAYS add "deleted_at IS NULL" unless
  -- asked about deactivated/removed agents.

TABLE stages
  id varchar PK, lead_id varchar (-> leads.id), name varchar, description text,
  state varchar (one of 'Not Started', 'In Progress', 'Completed'), date timestamptz,
  notes text, team varchar, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz
  -- Job pipeline stages for a lead. ALWAYS add "deleted_at IS NULL".

TABLE tasks
  id bigint PK, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz,
  title text, description text, due_date text, completed_at timestamp, priority text,
  status text, task_type text
  -- ALWAYS add "deleted_at IS NULL" unless asked about deleted tasks.

TABLE lead_associates
  lead_id varchar (-> leads.id), associate_customer_id text (-> associate_customers.id)
  -- Join table linking a lead to associated customer/company records.

TABLE associate_customers
  id text PK, name text, role text, details jsonb


TABLE stage_tasks
  id varchar PK, stage_id varchar (-> stages.id), name varchar, due_date timestamptz,
  duration bigint, completed boolean, created_at timestamptz, updated_at timestamptz,
  deleted_at timestamptz
  -- Checklist tasks within a stage (distinct from the top-level TABLE tasks).
  -- ALWAYS add "deleted_at IS NULL".


TABLE stage_users
  id varchar PK, stage_id varchar (-> stages.id), user_id varchar (-> agents.id), name varchar,
  created_at timestamptz, updated_at timestamptz, deleted_at timestamptz
  -- Agents assigned to a stage. ALWAYS add "deleted_at IS NULL".

TABLE county_taxes
  id bigint PK, name text, tax_percentage numeric
  -- Standalone tax-rate lookup by county name. No soft-delete column.


TABLE pickup_locations
  id bigint PK, name text, street text, unit_suite text, city text, state text, zip text,
  country text, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz
  -- Branch/pickup locations. ALWAYS add "deleted_at IS NULL".



TABLE manager_residential_entries
  id bigint PK, residential_entry_name text
  -- Catalog of residential entry names (e.g. "Kitchen", "Bathroom") — the
  -- type-of-work taxonomy, not a specific job's data. See residential_entries
  -- for the actual per-lead instances.

TABLE manager_residential_types
  id bigint PK, residential_entry_type text
  -- Catalog of residential entry types, paired with manager_residential_entries.

TABLE residential_entries
  id bigint PK, manager_residential_entry_id bigint (-> manager_residential_entries.id),
  manager_residential_type_id bigint (-> manager_residential_types.id), details text,
  lead_id varchar (-> leads.id), estimate_id text (-> estimates.id)
  -- A specific residential-entry instance attached to a lead/estimate.

TABLE services
  id bigint PK, details jsonb
  -- Almost everything about a service (name, price, etc.) lives inside the
  -- details jsonb column — use "details->>'fieldName'" to read it. If unsure
  -- what keys exist, prefer service_categories/service_groups/service_types
  -- (plain-column lookups) over guessing jsonb keys here.

TABLE service_categories
  id bigint PK, name text

TABLE service_groups
  id bigint PK, name text



TABLE service_rows
  id bigint PK, details jsonb, estimate_id text (-> estimates.id),
  residential_entry_id bigint (-> residential_entries.id)
  -- Per-estimate service line items; pricing/quantity fields are inside the
  -- details jsonb column — use "details->>'fieldName'" to read it.



TABLE material_rows
  id bigint PK, details jsonb, estimate_id text (-> estimates.id),
  residential_entry_id bigint (-> residential_entries.id)
  -- Per-estimate material line items; pricing/quantity fields are inside the
  -- details jsonb column — use "details->>'fieldName'" to read it.

TABLE material_templates
  id bigint PK, name text, manager_residential_entry_id bigint (-> manager_residential_entries.id),
  manager_residential_type_id bigint (-> manager_residential_types.id), category text,
  sub_category jsonb

-- Physical warehouse inventory (owned by the separate "inventory" service —
-- see the comment on these five entries in AIQueryableTables). Do not
-- confuse these with material_rows/material_templates above: those are
-- what was quoted on one estimate, these are actual stock on hand.

TABLE materials
  id bigint PK, serialnum text, name text, category text, threshold integer,
  available integer, retail_cost numeric, pro_cost numeric, unit text, type text,
  sub_category text, origin text, finish text, base_color text, level text,
  lot_no text, dimensions text, square_feet numeric, created_at timestamptz,
  updated_at timestamptz
  -- The catalog of stocked materials (e.g. one row per granite/quartz
  -- product). "available" is a live count, kept in sync on every
  -- materialchunks status change (see below) — for "how much of X do we
  -- have in stock"/"what's available" questions, just use this column
  -- directly rather than counting materialchunks yourself, unless the
  -- question needs a breakdown (by location, lot, etc.). "threshold" is
  -- the reorder point: "what's running low"/"what needs reordering" means
  -- available <= threshold. No soft-delete column.

TABLE materialchunks
  id bigint PK, material_id bigint (-> materials.id),
  purchase_received_id bigint (-> purchase_receiveds.id), serialnum text,
  slabno text, dimensions text, lot text, present_location text,
  retail_cost numeric, pro_cost numeric, status text, base_color text,
  level text, job_id text, bin text, supplier_details jsonb, notes text,
  inventory_type text, invoice_no text, created_at timestamptz, updated_at timestamptz
  -- One row per physical, individually-trackable piece (e.g. one slab) of a
  -- material. status is one of 'available' (in stock, unassigned),
  -- 'reserved' (held for a job but not yet assigned), 'assigned'
  -- (currently allocated to a job), or 'returned'. job_id is only
  -- populated while status is 'reserved'/'assigned' and is a plain string
  -- that matches TABLE jobs.job_id by convention — it is NOT a
  -- database-enforced foreign key, so treat a join to jobs as a best-effort
  -- text match, not a guaranteed one. supplier_details is a jsonb snapshot
  -- of the supplying vendor at receipt time — prefer joining
  -- purchase_received_id -> purchase_receiveds -> purchases -> suppliers
  -- for authoritative, current supplier info instead. No soft-delete column.

TABLE suppliers
  id bigint PK, name text, phone text, email text, category text, website text,
  details jsonb, created_at timestamptz, updated_at timestamptz
  -- Vendors materials are purchased from. Address, payment terms, and
  -- referrer info live inside the details jsonb column — e.g.
  -- "details->'address'->>'city'". No soft-delete column.

TABLE purchases
  id bigint PK, name text, material_id bigint (-> materials.id), quantity text,
  status text, rejection_reason text, current_price text,
  supplier_id bigint (-> suppliers.id), dimensions text, square_feet numeric,
  purchase_items jsonb, created_at timestamptz, updated_at timestamptz,
  bill_id text, inventory_type text, note text, county_tax_id bigint (-> county_taxes.id),
  county_name text, tax_percentage numeric, delivery_charges numeric,
  additional_charges jsonb, return_reason text
  -- A purchase order placed with a supplier for a material. quantity and
  -- current_price are stored as text, not numeric — cast with e.g.
  -- "NULLIF(quantity, '')::numeric" before doing arithmetic on them.
  -- Further per-line detail lives in the purchase_items/additional_charges
  -- jsonb columns. No soft-delete column.

TABLE purchase_receiveds
  id bigint PK, purchase_id bigint (-> purchases.id), invoice_id text,
  price numeric, received_quantity numeric, image text,
  created_at timestamptz, updated_at timestamptz
  -- One row per actual delivery against a purchase order — a single
  -- purchase can be received in more than one shipment. No soft-delete column.

General rules:
- Only ever produce a single SELECT (or a SELECT built with a WITH clause) statement.
- Never produce INSERT, UPDATE, DELETE, DROP, ALTER, TRUNCATE, GRANT, CREATE, COPY, or any
  other statement that is not a read.
- Never end the query with more than one statement (no semicolon-separated statements).
- Prefer aggregate functions (COUNT, SUM, AVG) when the question asks "how many"/"how much".
- Use "created_at" for "created" questions and relative date ranges like "yesterday" or
  "last month" should be computed with standard PostgreSQL date functions
  (e.g. created_at >= CURRENT_DATE - INTERVAL '1 day' AND created_at < CURRENT_DATE for "yesterday").
- If the question cannot be answered from the tables above, explain that in the "explanation"
  field and produce a trivially-empty query such as "SELECT 1 WHERE false".
`
