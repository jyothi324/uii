--
-- PostgreSQL database dump
--

\restrict bcoc8bNTMmwfFRWX1evxh33hGVysKnGxJNbmCn1Cx0RKwa2pS9R86zzmpHDqZgr

-- Dumped from database version 17.4
-- Dumped by pg_dump version 17.6 (Debian 17.6-1.pgdg13+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: addresses; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.addresses (
    id bigint NOT NULL,
    state text NOT NULL,
    county text NOT NULL,
    street character varying(255) NOT NULL,
    unit_suite character varying(255) NOT NULL,
    city character varying(255) NOT NULL,
    zip text DEFAULT 'false'::text
);


ALTER TABLE public.addresses OWNER TO tulipdev;

--
-- Name: addresses_id_seq; Type: SEQUENCE; Schema: public; Owner: tulipdev
--

CREATE SEQUENCE public.addresses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.addresses_id_seq OWNER TO tulipdev;

--
-- Name: addresses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tulipdev
--

ALTER SEQUENCE public.addresses_id_seq OWNED BY public.addresses.id;

--
-- Name: profiles; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.profiles (
    id               SERIAL PRIMARY KEY,
    name             VARCHAR(255) NOT NULL,
    phone            VARCHAR(50)  NOT NULL,
    email            VARCHAR(255) NOT NULL,
    same_as_shipping BOOLEAN      NOT NULL DEFAULT false,

    shipping_address JSONB NOT NULL DEFAULT '{}'::jsonb,
    billing_address  JSONB NOT NULL DEFAULT '{}'::jsonb,
    qb_customer_id   TEXT,
    details          JSONB NOT NULL DEFAULT '{}'::jsonb,

    created_at       TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP    NOT NULL DEFAULT NOW()
);

ALTER TABLE public.profiles OWNER TO tulipdev;

CREATE INDEX idx_profiles_name  ON public.profiles (LOWER(name));
CREATE INDEX idx_profiles_email ON public.profiles (LOWER(email));
CREATE INDEX idx_profiles_phone ON public.profiles (LOWER(phone));
CREATE UNIQUE INDEX idx_profiles_qb_customer_id_unique
  ON public.profiles (qb_customer_id)
  WHERE qb_customer_id IS NOT NULL AND qb_customer_id <> '';



--
-- Name: agents; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.agents (
    id character varying(27) NOT NULL,
    name character varying(100) NOT NULL,
    email character varying(255) NOT NULL,
    password character varying(255) NOT NULL,
    password_hashed character varying(255) NOT NULL,
    details jsonb,
    role character varying(20),
    assignee_type character varying(20),
    superuser boolean DEFAULT false,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE public.agents OWNER TO tulipdev;

--
-- Name: associate_customers; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.associate_customers (
    id text NOT NULL,
    details jsonb,
    name text NOT NULL,
    role text NOT NULL
);


ALTER TABLE public.associate_customers OWNER TO tulipdev;

--
-- Name: county_taxes; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.county_taxes (
    id bigint NOT NULL,
    name text NOT NULL,
    tax_percentage numeric
);


ALTER TABLE public.county_taxes OWNER TO tulipdev;

--
-- Name: county_taxes_id_seq; Type: SEQUENCE; Schema: public; Owner: tulipdev
--

CREATE SEQUENCE public.county_taxes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.county_taxes_id_seq OWNER TO tulipdev;

--
-- Name: county_taxes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tulipdev
--

ALTER SEQUENCE public.county_taxes_id_seq OWNED BY public.county_taxes.id;


--
-- Name: customers; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.customers (
    id text NOT NULL,
    name text,
    email text,
    phone text
);


ALTER TABLE public.customers OWNER TO tulipdev;








--
-- Name: estimates; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.estimates (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    lead_id character varying(20) DEFAULT NULL::character varying,
    name text NOT NULL,
    material_total numeric NOT NULL,
    service_total numeric NOT NULL,
    details jsonb,
    status text NOT NULL
);


ALTER TABLE public.estimates OWNER TO tulipdev;

--
-- Name: files; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.files (
    id bigint NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    size text NOT NULL,
    uploaded_at timestamp without time zone NOT NULL,
    project_type text,
    stored_file_name text NOT NULL,
    category text,
    lead_id character varying(20),
    file_path text NOT NULL
);


ALTER TABLE public.files OWNER TO tulipdev;

--
-- Name: files_id_seq; Type: SEQUENCE; Schema: public; Owner: tulipdev
--

CREATE SEQUENCE public.files_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.files_id_seq OWNER TO tulipdev;

--
-- Name: files_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tulipdev
--

ALTER SEQUENCE public.files_id_seq OWNED BY public.files.id;










--
-- Name: lead_associates; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.lead_associates (
    lead_id character varying(20) NOT NULL,
    associate_customer_id text NOT NULL
);


ALTER TABLE public.lead_associates OWNER TO tulipdev;

--
-- Name: leads; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.leads (
    id character varying(20) NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    lead_type character varying(20) NOT NULL,
    job_name character varying(100) NOT NULL,
    delivery_type character varying(20),
    status text NOT NULL,
    details jsonb,
    referrer jsonb,
    detailed_lead_status character varying(50),
    payment_term character varying(20),
    pricing_type character varying(20),
    favorite text,
    profile_pic text,
    sales1_person_id character varying(27),
    sales2_person_id character varying(27),
    rejection_reason character varying(255)
);


ALTER TABLE public.leads OWNER TO tulipdev;

--
-- Name: manager_residential_entries; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.manager_residential_entries (
    id bigint NOT NULL,
    residential_entry_name text,
    length bigint DEFAULT 0 NOT NULL
);


ALTER TABLE public.manager_residential_entries OWNER TO tulipdev;

--
-- Name: manager_residential_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: tulipdev
--

CREATE SEQUENCE public.manager_residential_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.manager_residential_entries_id_seq OWNER TO tulipdev;

--
-- Name: manager_residential_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tulipdev
--

ALTER SEQUENCE public.manager_residential_entries_id_seq OWNED BY public.manager_residential_entries.id;


--
-- Name: manager_residential_types; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.manager_residential_types (
    id bigint NOT NULL,
    residential_entry_type text,
    length bigint DEFAULT 0 NOT NULL
);


ALTER TABLE public.manager_residential_types OWNER TO tulipdev;

--
-- Name: manager_residential_types_id_seq; Type: SEQUENCE; Schema: public; Owner: tulipdev
--

CREATE SEQUENCE public.manager_residential_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.manager_residential_types_id_seq OWNER TO tulipdev;

--
-- Name: manager_residential_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tulipdev
--

ALTER SEQUENCE public.manager_residential_types_id_seq OWNED BY public.manager_residential_types.id;


--
-- Name: material_rows; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.material_rows (
    id bigint NOT NULL,
    details jsonb,
    estimate_id text,
    residential_entry_id bigint
);


ALTER TABLE public.material_rows OWNER TO tulipdev;

--
-- Name: material_rows_id_seq; Type: SEQUENCE; Schema: public; Owner: tulipdev
--

CREATE SEQUENCE public.material_rows_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.material_rows_id_seq OWNER TO tulipdev;

--
-- Name: material_rows_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tulipdev
--

ALTER SEQUENCE public.material_rows_id_seq OWNED BY public.material_rows.id;


--
-- Name: material_templates; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.material_templates (
    id bigint NOT NULL,
    name text NOT NULL,
    manager_residential_entry_id bigint,
    manager_residential_type_id bigint,
    category text NOT NULL,
    sub_category jsonb
);


ALTER TABLE public.material_templates OWNER TO tulipdev;

--
-- Name: material_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: tulipdev
--

CREATE SEQUENCE public.material_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.material_templates_id_seq OWNER TO tulipdev;

--
-- Name: material_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tulipdev
--

ALTER SEQUENCE public.material_templates_id_seq OWNED BY public.material_templates.id;


--
-- Name: pickup_locations; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.pickup_locations (
    id bigint NOT NULL,
    name text NOT NULL,
    street text NOT NULL,
    unit_suite text,
    city text NOT NULL,
    state text NOT NULL,
    zip text NOT NULL,
    country text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE public.pickup_locations OWNER TO tulipdev;

--
-- Name: pickup_locations_id_seq; Type: SEQUENCE; Schema: public; Owner: tulipdev
--

CREATE SEQUENCE public.pickup_locations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.pickup_locations_id_seq OWNER TO tulipdev;

--
-- Name: pickup_locations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tulipdev
--

ALTER SEQUENCE public.pickup_locations_id_seq OWNED BY public.pickup_locations.id;


--
-- Name: residential_entries; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.residential_entries (
    id bigint NOT NULL,
    manager_residential_entry_id bigint,
    manager_residential_type_id bigint,
    details text,
    lead_id character varying(20),
    estimate_id text
);


ALTER TABLE public.residential_entries OWNER TO tulipdev;

--
-- Name: residential_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: tulipdev
--

CREATE SEQUENCE public.residential_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.residential_entries_id_seq OWNER TO tulipdev;

--
-- Name: residential_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tulipdev
--

ALTER SEQUENCE public.residential_entries_id_seq OWNED BY public.residential_entries.id;


--
-- Name: service_categories; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.service_categories (
    id bigint NOT NULL,
    name text
);


ALTER TABLE public.service_categories OWNER TO tulipdev;

--
-- Name: service_categories_id_seq; Type: SEQUENCE; Schema: public; Owner: tulipdev
--

CREATE SEQUENCE public.service_categories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.service_categories_id_seq OWNER TO tulipdev;

--
-- Name: service_categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tulipdev
--

ALTER SEQUENCE public.service_categories_id_seq OWNED BY public.service_categories.id;


--
-- Name: service_groups; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.service_groups (
    id bigint NOT NULL,
    name text NOT NULL
);


ALTER TABLE public.service_groups OWNER TO tulipdev;

--
-- Name: service_groups_id_seq; Type: SEQUENCE; Schema: public; Owner: tulipdev
--

CREATE SEQUENCE public.service_groups_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.service_groups_id_seq OWNER TO tulipdev;

--
-- Name: service_groups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tulipdev
--

ALTER SEQUENCE public.service_groups_id_seq OWNED BY public.service_groups.id;






--
-- Name: service_rows; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.service_rows (
    id bigint NOT NULL,
    details jsonb,
    estimate_id text,
    residential_entry_id bigint
);


ALTER TABLE public.service_rows OWNER TO tulipdev;

--
-- Name: service_rows_id_seq; Type: SEQUENCE; Schema: public; Owner: tulipdev
--

CREATE SEQUENCE public.service_rows_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.service_rows_id_seq OWNER TO tulipdev;

--
-- Name: service_rows_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tulipdev
--

ALTER SEQUENCE public.service_rows_id_seq OWNED BY public.service_rows.id;





--
-- Name: services; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.services (
    id bigint NOT NULL,
    details jsonb
);


ALTER TABLE public.services OWNER TO tulipdev;

--
-- Name: services_id_seq; Type: SEQUENCE; Schema: public; Owner: tulipdev
--

CREATE SEQUENCE public.services_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.services_id_seq OWNER TO tulipdev;

--
-- Name: services_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tulipdev
--

ALTER SEQUENCE public.services_id_seq OWNED BY public.services.id;



--
-- Name: stage_tasks; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.stage_tasks (
    id character varying(36) NOT NULL,
    stage_id character varying(36) NOT NULL,
    name character varying(255) NOT NULL,
    due_date timestamp with time zone NOT NULL,
    duration bigint NOT NULL,
    completed boolean DEFAULT false,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE public.stage_tasks OWNER TO tulipdev;

--
-- Name: stage_users; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.stage_users (
    id character varying(36) NOT NULL,
    stage_id character varying(36) NOT NULL,
    user_id character varying(36) NOT NULL,
    name character varying(255) NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE public.stage_users OWNER TO tulipdev;

--
-- Name: stages; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.stages (
    id character varying(36) NOT NULL,
    lead_id character varying(20) NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    state character varying(20),
    date timestamp with time zone,
    notes text,
    team character varying(255),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    residential_entry_id bigint,
    CONSTRAINT chk_stages_state CHECK (((state)::text = ANY (ARRAY[('Not Started'::character varying)::text, ('In Progress'::character varying)::text, ('Completed'::character varying)::text])))
);


ALTER TABLE public.stages OWNER TO tulipdev;

--
-- Name: tasks; Type: TABLE; Schema: public; Owner: tulipdev
--

CREATE TABLE public.tasks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    title text NOT NULL,
    description text,
    due_date text,
    completed_at timestamp without time zone,
    priority text NOT NULL,
    status text NOT NULL,
    task_type text NOT NULL
);


ALTER TABLE public.tasks OWNER TO tulipdev;

--
-- Name: tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: tulipdev
--

CREATE SEQUENCE public.tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.tasks_id_seq OWNER TO tulipdev;

--
-- Name: tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tulipdev
--

ALTER SEQUENCE public.tasks_id_seq OWNED BY public.tasks.id;


--
-- Name: addresses id; Type: DEFAULT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.addresses ALTER COLUMN id SET DEFAULT nextval('public.addresses_id_seq'::regclass);


--
-- Name: county_taxes id; Type: DEFAULT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.county_taxes ALTER COLUMN id SET DEFAULT nextval('public.county_taxes_id_seq'::regclass);





--
-- Name: files id; Type: DEFAULT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.files ALTER COLUMN id SET DEFAULT nextval('public.files_id_seq'::regclass);




--
-- Name: manager_residential_entries id; Type: DEFAULT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.manager_residential_entries ALTER COLUMN id SET DEFAULT nextval('public.manager_residential_entries_id_seq'::regclass);


--
-- Name: manager_residential_types id; Type: DEFAULT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.manager_residential_types ALTER COLUMN id SET DEFAULT nextval('public.manager_residential_types_id_seq'::regclass);


--
-- Name: material_rows id; Type: DEFAULT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.material_rows ALTER COLUMN id SET DEFAULT nextval('public.material_rows_id_seq'::regclass);


--
-- Name: material_templates id; Type: DEFAULT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.material_templates ALTER COLUMN id SET DEFAULT nextval('public.material_templates_id_seq'::regclass);


--
-- Name: pickup_locations id; Type: DEFAULT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.pickup_locations ALTER COLUMN id SET DEFAULT nextval('public.pickup_locations_id_seq'::regclass);


--
-- Name: residential_entries id; Type: DEFAULT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.residential_entries ALTER COLUMN id SET DEFAULT nextval('public.residential_entries_id_seq'::regclass);


--
-- Name: service_categories id; Type: DEFAULT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.service_categories ALTER COLUMN id SET DEFAULT nextval('public.service_categories_id_seq'::regclass);


--
-- Name: service_groups id; Type: DEFAULT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.service_groups ALTER COLUMN id SET DEFAULT nextval('public.service_groups_id_seq'::regclass);




--
-- Name: service_rows id; Type: DEFAULT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.service_rows ALTER COLUMN id SET DEFAULT nextval('public.service_rows_id_seq'::regclass);



--
-- Name: services id; Type: DEFAULT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.services ALTER COLUMN id SET DEFAULT nextval('public.services_id_seq'::regclass);


--
-- Name: tasks id; Type: DEFAULT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.tasks ALTER COLUMN id SET DEFAULT nextval('public.tasks_id_seq'::regclass);


--
-- Name: addresses addresses_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.addresses
    ADD CONSTRAINT addresses_pkey PRIMARY KEY (id);


--
-- Name: agents agents_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.agents
    ADD CONSTRAINT agents_pkey PRIMARY KEY (id);


--
-- Name: associate_customers associate_customers_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.associate_customers
    ADD CONSTRAINT associate_customers_pkey PRIMARY KEY (id);


--
-- Name: county_taxes county_taxes_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.county_taxes
    ADD CONSTRAINT county_taxes_pkey PRIMARY KEY (id);


--
-- Name: customers customers_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);




--
-- Name: estimates estimates_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.estimates
    ADD CONSTRAINT estimates_pkey PRIMARY KEY (id);


--
-- Name: files files_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_pkey PRIMARY KEY (id);


--
-- Name: lead_associates lead_associates_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.lead_associates
    ADD CONSTRAINT lead_associates_pkey PRIMARY KEY (lead_id, associate_customer_id);


--
-- Name: leads leads_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.leads
    ADD CONSTRAINT leads_pkey PRIMARY KEY (id);


--
-- Name: manager_residential_entries manager_residential_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.manager_residential_entries
    ADD CONSTRAINT manager_residential_entries_pkey PRIMARY KEY (id);


--
-- Name: manager_residential_types manager_residential_types_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.manager_residential_types
    ADD CONSTRAINT manager_residential_types_pkey PRIMARY KEY (id);


--
-- Name: material_rows material_rows_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.material_rows
    ADD CONSTRAINT material_rows_pkey PRIMARY KEY (id);


--
-- Name: material_templates material_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.material_templates
    ADD CONSTRAINT material_templates_pkey PRIMARY KEY (id);


--
-- Name: pickup_locations pickup_locations_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.pickup_locations
    ADD CONSTRAINT pickup_locations_pkey PRIMARY KEY (id);


--
-- Name: residential_entries residential_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.residential_entries
    ADD CONSTRAINT residential_entries_pkey PRIMARY KEY (id);


--
-- Name: service_categories service_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.service_categories
    ADD CONSTRAINT service_categories_pkey PRIMARY KEY (id);


--
-- Name: service_groups service_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.service_groups
    ADD CONSTRAINT service_groups_pkey PRIMARY KEY (id);



--
-- Name: service_rows service_rows_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.service_rows
    ADD CONSTRAINT service_rows_pkey PRIMARY KEY (id);




--
-- Name: services services_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.services
    ADD CONSTRAINT services_pkey PRIMARY KEY (id);




--
-- Name: stage_tasks stage_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.stage_tasks
    ADD CONSTRAINT stage_tasks_pkey PRIMARY KEY (id);


--
-- Name: stage_users stage_users_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.stage_users
    ADD CONSTRAINT stage_users_pkey PRIMARY KEY (id);


--
-- Name: stages stages_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.stages
    ADD CONSTRAINT stages_pkey PRIMARY KEY (id);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);


--
-- Name: agents uni_agents_email; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.agents
    ADD CONSTRAINT uni_agents_email UNIQUE (email);


--
-- Name: customers uni_customers_name; Type: CONSTRAINT; Schema: public; Owner: tulipdev
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT uni_customers_name UNIQUE (name);


--
-- Name: idx_agents_deleted_at; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_agents_deleted_at ON public.agents USING btree (deleted_at);


--
-- Name: idx_agents_email; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_agents_email ON public.agents USING btree (email);


--
-- Name: idx_county_taxes_name; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_county_taxes_name ON public.county_taxes USING btree (name);




--
-- Name: idx_estimates_deleted_at; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_estimates_deleted_at ON public.estimates USING btree (deleted_at);



--
-- Name: idx_leads_deleted_at; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_leads_deleted_at ON public.leads USING btree (deleted_at);


--
-- Name: idx_material_templates_name; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_material_templates_name ON public.material_templates USING btree (name);


--
-- Name: idx_pickup_locations_deleted_at; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_pickup_locations_deleted_at ON public.pickup_locations USING btree (deleted_at);


--
-- Name: idx_service_groups_name; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_service_groups_name ON public.service_groups USING btree (name);







--
-- Name: idx_stage_tasks_deleted_at; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_stage_tasks_deleted_at ON public.stage_tasks USING btree (deleted_at);


--
-- Name: idx_stage_tasks_stage_id; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_stage_tasks_stage_id ON public.stage_tasks USING btree (stage_id);


--
-- Name: idx_stage_users_deleted_at; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_stage_users_deleted_at ON public.stage_users USING btree (deleted_at);


--
-- Name: idx_stage_users_stage_id; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_stage_users_stage_id ON public.stage_users USING btree (stage_id);


--
-- Name: idx_stages_deleted_at; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_stages_deleted_at ON public.stages USING btree (deleted_at);


--
-- Name: idx_stages_lead_id; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_stages_lead_id ON public.stages USING btree (lead_id);


--
-- Name: idx_tasks_deleted_at; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_tasks_deleted_at ON public.tasks USING btree (deleted_at);


--
-- Name: idx_tasks_status; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_tasks_status ON public.tasks USING btree (status);


--
-- Name: idx_tasks_title; Type: INDEX; Schema: public; Owner: tulipdev
--

CREATE INDEX idx_tasks_title ON public.tasks USING btree (title);


--
-- PostgreSQL database dump complete
--

\unrestrict bcoc8bNTMmwfFRWX1evxh33hGVysKnGxJNbmCn1Cx0RKwa2pS9R86zzmpHDqZgr
