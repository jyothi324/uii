
--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4
-- Dumped by pg_dump version 17.5 (Ubuntu 17.5-1.pgdg24.04+1)

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
-- Name: addresses; Type: TABLE; Schema: public; Owner: postgres
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


ALTER TABLE public.addresses OWNER TO postgres;

--
-- Name: addresses_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.addresses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.addresses_id_seq OWNER TO postgres;

--
-- Name: addresses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.addresses_id_seq OWNED BY public.addresses.id;


--
-- Name: agents; Type: TABLE; Schema: public; Owner: postgres
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


ALTER TABLE public.agents OWNER TO postgres;

--
-- Name: associate_customers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.associate_customers (
    id text NOT NULL,
    details jsonb,
    name text NOT NULL,
    role text NOT NULL
);


ALTER TABLE public.associate_customers OWNER TO postgres;

--
-- Name: county_taxes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.county_taxes (
    id bigint NOT NULL,
    name text NOT NULL,
    tax_percentage numeric
);


ALTER TABLE public.county_taxes OWNER TO postgres;

--
-- Name: county_taxes_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.county_taxes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.county_taxes_id_seq OWNER TO postgres;

--
-- Name: county_taxes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.county_taxes_id_seq OWNED BY public.county_taxes.id;


--
-- Name: customers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.customers (
    id text NOT NULL,
    name text,
    email text,
    phone text
);


ALTER TABLE public.customers OWNER TO postgres;








--
-- Name: estimates; Type: TABLE; Schema: public; Owner: postgres
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


ALTER TABLE public.estimates OWNER TO postgres;

--
-- Name: files; Type: TABLE; Schema: public; Owner: postgres
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


ALTER TABLE public.files OWNER TO postgres;

--
-- Name: files_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.files_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.files_id_seq OWNER TO postgres;

--
-- Name: files_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.files_id_seq OWNED BY public.files.id;






--
-- Name: lead_associates; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.lead_associates (
    lead_id character varying(20) NOT NULL,
    associate_customer_id text NOT NULL
);


ALTER TABLE public.lead_associates OWNER TO postgres;

--
-- Name: leads; Type: TABLE; Schema: public; Owner: postgres
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


ALTER TABLE public.leads OWNER TO postgres;

--
-- Name: manager_residential_entries; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.manager_residential_entries (
    id bigint NOT NULL,
    residential_entry_name text,
    length bigint DEFAULT 0 NOT NULL
);


ALTER TABLE public.manager_residential_entries OWNER TO postgres;

--
-- Name: manager_residential_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.manager_residential_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.manager_residential_entries_id_seq OWNER TO postgres;

--
-- Name: manager_residential_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.manager_residential_entries_id_seq OWNED BY public.manager_residential_entries.id;


--
-- Name: manager_residential_types; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.manager_residential_types (
    id bigint NOT NULL,
    residential_entry_type text,
    length bigint DEFAULT 0 NOT NULL
);


ALTER TABLE public.manager_residential_types OWNER TO postgres;

--
-- Name: manager_residential_types_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.manager_residential_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.manager_residential_types_id_seq OWNER TO postgres;

--
-- Name: profiles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE IF NOT EXISTS profiles (
    id               SERIAL PRIMARY KEY,
    name             VARCHAR(255) NOT NULL,
    phone            VARCHAR(50)  NOT NULL,
    email            VARCHAR(255) NOT NULL,
    same_as_shipping BOOLEAN      NOT NULL DEFAULT false,
    shipping_address JSONB       NOT NULL,
    billing_address  JSONB       NOT NULL,
    qb_customer_id   TEXT,
    details          JSONB        NOT NULL,
    created_at       TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_profiles_name ON profiles (LOWER(name));
CREATE INDEX IF NOT EXISTS idx_profiles_email ON profiles (LOWER(email));
CREATE INDEX IF NOT EXISTS idx_profiles_phone ON profiles (LOWER(phone));
CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_qb_customer_id_unique
  ON profiles (qb_customer_id)
  WHERE qb_customer_id IS NOT NULL AND qb_customer_id <> '';

--
-- Name: manager_residential_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.manager_residential_types_id_seq OWNED BY public.manager_residential_types.id;


--
-- Name: material_rows; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.material_rows (
    id bigint NOT NULL,
    details jsonb,
    estimate_id text,
    residential_entry_id bigint
);


ALTER TABLE public.material_rows OWNER TO postgres;

--
-- Name: material_rows_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.material_rows_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.material_rows_id_seq OWNER TO postgres;

--
-- Name: material_rows_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.material_rows_id_seq OWNED BY public.material_rows.id;


--
-- Name: material_templates; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.material_templates (
    id bigint NOT NULL,
    name text NOT NULL,
    manager_residential_entry_id bigint,
    manager_residential_type_id bigint,
    category text NOT NULL,
    sub_category jsonb
);


ALTER TABLE public.material_templates OWNER TO postgres;

--
-- Name: material_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.material_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.material_templates_id_seq OWNER TO postgres;

--
-- Name: material_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.material_templates_id_seq OWNED BY public.material_templates.id;


--
-- Name: pickup_locations; Type: TABLE; Schema: public; Owner: postgres
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


ALTER TABLE public.pickup_locations OWNER TO postgres;

--
-- Name: pickup_locations_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.pickup_locations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.pickup_locations_id_seq OWNER TO postgres;

--
-- Name: pickup_locations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.pickup_locations_id_seq OWNED BY public.pickup_locations.id;


--
-- Name: residential_entries; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.residential_entries (
    id bigint NOT NULL,
    manager_residential_entry_id bigint,
    manager_residential_type_id bigint,
    details text,
    lead_id character varying(20),
    estimate_id text
);


ALTER TABLE public.residential_entries OWNER TO postgres;

--
-- Name: residential_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.residential_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.residential_entries_id_seq OWNER TO postgres;

--
-- Name: residential_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.residential_entries_id_seq OWNED BY public.residential_entries.id;


--
-- Name: service_categories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.service_categories (
    id bigint NOT NULL,
    name text
);


ALTER TABLE public.service_categories OWNER TO postgres;

--
-- Name: service_categories_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.service_categories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.service_categories_id_seq OWNER TO postgres;

--
-- Name: service_categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.service_categories_id_seq OWNED BY public.service_categories.id;


--
-- Name: service_groups; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.service_groups (
    id bigint NOT NULL,
    name text NOT NULL
);


ALTER TABLE public.service_groups OWNER TO postgres;

--
-- Name: service_groups_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.service_groups_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.service_groups_id_seq OWNER TO postgres;

--
-- Name: service_groups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.service_groups_id_seq OWNED BY public.service_groups.id;






--
-- Name: service_rows; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.service_rows (
    id bigint NOT NULL,
    details jsonb,
    estimate_id text,
    residential_entry_id bigint
);


ALTER TABLE public.service_rows OWNER TO postgres;

--
-- Name: service_rows_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.service_rows_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.service_rows_id_seq OWNER TO postgres;

--
-- Name: service_rows_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.service_rows_id_seq OWNED BY public.service_rows.id;




--
-- Name: services; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.services (
    id bigint NOT NULL,
    details jsonb
);


ALTER TABLE public.services OWNER TO postgres;

--
-- Name: services_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.services_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.services_id_seq OWNER TO postgres;

--
-- Name: services_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.services_id_seq OWNED BY public.services.id;



--
-- Name: stage_tasks; Type: TABLE; Schema: public; Owner: postgres
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


ALTER TABLE public.stage_tasks OWNER TO postgres;

--
-- Name: stage_users; Type: TABLE; Schema: public; Owner: postgres
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


ALTER TABLE public.stage_users OWNER TO postgres;

--
-- Name: stages; Type: TABLE; Schema: public; Owner: postgres
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


ALTER TABLE public.stages OWNER TO postgres;

--
-- Name: tasks; Type: TABLE; Schema: public; Owner: postgres
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


ALTER TABLE public.tasks OWNER TO postgres;

--
-- Name: tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.tasks_id_seq OWNER TO postgres;

--
-- Name: tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.tasks_id_seq OWNED BY public.tasks.id;


--
-- Name: addresses id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.addresses ALTER COLUMN id SET DEFAULT nextval('public.addresses_id_seq'::regclass);


--
-- Name: county_taxes id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.county_taxes ALTER COLUMN id SET DEFAULT nextval('public.county_taxes_id_seq'::regclass);





--
-- Name: files id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.files ALTER COLUMN id SET DEFAULT nextval('public.files_id_seq'::regclass);




--
-- Name: manager_residential_entries id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.manager_residential_entries ALTER COLUMN id SET DEFAULT nextval('public.manager_residential_entries_id_seq'::regclass);


--
-- Name: manager_residential_types id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.manager_residential_types ALTER COLUMN id SET DEFAULT nextval('public.manager_residential_types_id_seq'::regclass);


--
-- Name: material_rows id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.material_rows ALTER COLUMN id SET DEFAULT nextval('public.material_rows_id_seq'::regclass);


--
-- Name: material_templates id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.material_templates ALTER COLUMN id SET DEFAULT nextval('public.material_templates_id_seq'::regclass);


--
-- Name: pickup_locations id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pickup_locations ALTER COLUMN id SET DEFAULT nextval('public.pickup_locations_id_seq'::regclass);


--
-- Name: residential_entries id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.residential_entries ALTER COLUMN id SET DEFAULT nextval('public.residential_entries_id_seq'::regclass);


--
-- Name: service_categories id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.service_categories ALTER COLUMN id SET DEFAULT nextval('public.service_categories_id_seq'::regclass);


--
-- Name: service_groups id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.service_groups ALTER COLUMN id SET DEFAULT nextval('public.service_groups_id_seq'::regclass);





--
-- Name: service_rows id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.service_rows ALTER COLUMN id SET DEFAULT nextval('public.service_rows_id_seq'::regclass);




--
-- Name: services id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.services ALTER COLUMN id SET DEFAULT nextval('public.services_id_seq'::regclass);


--
-- Name: tasks id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks ALTER COLUMN id SET DEFAULT nextval('public.tasks_id_seq'::regclass);


--
-- Data for Name: addresses; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.addresses (id, state, county, street, unit_suite, city, zip) FROM stdin;
1	Florida	USA	456 Market Street		Miami	62701
2	Kentucky	USA	veill		asdf	533341
3	Florida		456 Market Street		Miami	62701
4	Louisiana		veill		asdfa	533341
5	Florida		456 Market Street		Miami	62701
6	Delaware		veill		asdfa	533341
56	Florida	USA	123 Warehouse Blvd		Tampa	62704
59	Kansas	USA	14658 S HAGAN ST	4208	OLATHE	66062
7	Florida	USA	456 Market Street		Miami	62701
8	Hawaii	USA	veill		asdf	533341
9	Florida	USA	123 Warehouse Blvd		Tampa	62704
10	Florida	USA	123 Warehouse Blvd		Tampa	62704
11	Florida	USA	Chakali Veedhi		Rajam	532127
12	Florida	USA	Chakali Veedhi		Rajam	532127
13	Florida		456 Market Street		Miami	62701
14	Louisiana		veill		asdfasdf	533341
15	Florida		123 Warehouse Blvd		Tampa	62704
16	Alaska		veill		asdfa	533341
21	Florida		456 Market Street		Miami	62701
22	Florida		456 Market Street		Miami	62701
19	Illinois	USA	sun street	sdhbjsdbhjdshbj	miami	12354
20	Illinois	USA	sun street	sdhbjsdbhjdshbj	miami	12354
23	Florida		123 Warehouse Blvd		Tampa	62704
24	Florida		123 Warehouse Blvd		Tampa	62704
25	Florida		123 Warehouse Blvd		Tampa	62704
26	Manitoba		gjhhbjh		asdfa	533341
27	Baja California Sur		456 Market Street		Miami	62701
28	California		veill		asdfa	533341
29	Florida	USA	789 East Avenue		Orlando	62702
30	Florida	USA	789 East Avenue		Orlando	62702
31	Florida		456 Market Street		Miami	62701
32	Kentucky		veill		asdfa	533341
60	Arkansas	USA	Plot no 123&124, Sampada Homes, Flat G2	Pragathi Nagar	HYDERABAD	500090
61	Florida	USA	456 Market Street		Miami	62701
33	Florida	USA	123 Warehouse Blvd		Tampa	62704
34	Florida	USA	123 Warehouse Blvd		Tampa	62704
37	Florida	USA	123 Warehouse Blvd		Tampa	62704
38	Florida	USA	123 Warehouse Blvd		Tampa	62704
62	Chihuahua	Mexico	Plot no 123&124, Sampada Homes, Flat G2	Pragathi Nagar	HYDERABAD	500090
63	Florida	USA	789 East Avenue		Orlando	62702
39	Florida	USA	456 Market Street		Miami	62701
40	Illinois	USA	Chakali Veedhi		Rajam	532127
64	Florida	USA	789 East Avenue		Orlando	62702
65	Kansas	USA	14658 S Hagan ST		Olathe	66062
43	Florida	USA	456 Market Street		Miami	62701
44	Michigan	USA	36671 Grand River Avenue	apt 204	Farmington	48335
45	Florida	USA	456 Market Street		Miami	62701
46	Florida	USA	456 Market Street		Miami	62701
35	Florida	USA	456 Market Street		Miami	62701
36	Kansas	USA	veill		asdfa	533341
41	Hawaii	USA	789 East Avenue		Orlando	62702
42	Florida	USA	789 East Avenue		Orlando	62702
47	Florida	USA	123 Warehouse Blvd		Tampa	62704
48	Michigan	USA	36671 Grand River Avenue		Canton	48335
49	Florida	USA	456 Market Street		Miami	62701
50	Kansas	USA	veill		asdfa	533341
51	Florida	USA	456 Market Street		Miami	62701
52	Saskatchewan	Canada	veill		asdfa	533341
53	Florida	USA	123 Warehouse Blvd		Tampa	62704
54	Florida	USA	123 Warehouse Blvd		Tampa	62704
66	Kansas	USA	14658 S Hagan ST		Olathe	66062
67	Prince Edward Island	Canada	Plot no 123&124, Sampada Homes, Flat G2	Pragathi Nagar	HYDERABAD	500090
57	Florida	USA	789 East Avenue		Orlando	62702
58	Florida	USA	789 East Avenue		Orlando	62702
87	Florida	USA	456 Market Street		Miami	62701
88	Kansas	USA	veill		asdfa	533341
55	Florida	USA	123 Warehouse Blvd		Tampa	62704
68	Prince Edward Island	Canada	Plot no 123&124, Sampada Homes, Flat G2	Pragathi Nagar	HYDERABAD	500090
69	Florida	USA	456 Market Street		Miami	62701
70	Manitoba	Canada	14658 S Hagan ST		olathe	66062
71	New Brunswick	Canada	3650	1	Canton	48188
72	New Brunswick	Canada	3650	1	Canton	48188
73	Florida	USA	789 East Avenue		Orlando	62702
74	Florida	USA	789 East Avenue		Orlando	62702
75	Florida	USA	123 Warehouse Blvd		Tampa	62704
76	Florida	USA	123 Warehouse Blvd		Tampa	62704
77	Florida	USA	456 Market Street		Miami	62701
78	Campeche	Mexico	14658 S Hagan ST		olathe	66062
79	Florida	USA	456 Market Street		Miami	62701
80	Florida	USA	789 East Avenue		Orlando	62702
81	Florida	USA	789 East Avenue		Orlando	62702
82	Florida	USA	789 East Avenue		Orlando	62702
83	Alaska	USA	123 Warehouse Blvd		Tampa	62704
84	Manitoba	Canada	123 Warehouse Blvd		Tampa	62704
85	Florida	USA	456 Market Street		Miami	62701
86	Kansas	USA	veill		asdfa	533341
89	Florida	USA	456 Market Street		Miami	62701
90	Kansas	USA	veill		asdfa	533341
91	Florida	USA	456 Market Street		Miami	62701
92	Kansas	USA	14658 S Hagan St		Olathe	66062-9002
93	Florida	USA	456 Market Street		Miami	62701
94	Kansas	USA	13911 Marty ST	4208	Fort Myers	66062
95	Kansas	USA	14658 S Hagan ST		Olathe	66062
96	Kansas	USA	14658 S Hagan ST		Olathe	66062
97	Kansas	USA	14658 S Hagan ST		Olathe	66062
98	Kansas	USA	14658 S Hagan ST		Olathe	66062
99	Colorado	USA	13911 Marty ST	4208	Olathe	7657
100	Colorado	USA	13911 Marty ST	4208	Olathe	7657
101	Florida	USA	456 Market Street		Miami	62701
102	Alaska	USA	veill		asdfa	533341
103	Florida	USA	123 Warehouse Blvd		Tampa	62704
104	Manitoba	Canada	Prakasam		asdfa	533341
105	Florida	USA	456 Market Street		Miami	62701
106	Manitoba	Canada	Plot no 123&124, Sampada Homes, Flat G2	Pragathi Nagar	HYDERABAD	500090
107	Kansas	USA	14658 S Hagan ST		Olathe	66062
108	Kansas	USA	14658 S Hagan ST		Olathe	66062
109	Kansas	USA	14658 S HAGAN ST	4208	Olathe	66062
110	Kansas	USA	14658 S HAGAN ST	4208	Olathe	66062
111	Florida	USA	123 Warehouse Blvd		Tampa	62704
112	Florida	USA	123 Warehouse Blvd		Tampa	62704
113	Florida	USA	123 Warehouse Blvd		Tampa	62704
114	Florida	USA	123 Warehouse Blvd		Tampa	62704
115	Florida	USA	123 Warehouse Blvd		Tampa	62704
116	Florida	USA	123 Warehouse Blvd		Tampa	62704
117	Florida	USA	789 East Avenue		Orlando	62702
118	Florida	USA	789 East Avenue		Orlando	62702
119	Florida	USA	123 Warehouse Blvd		Tampa	62704
120	Florida	USA	123 Warehouse Blvd		Tampa	62704
121	Florida	USA	123 Warehouse Blvd		Tampa	62704
122	Florida	USA	123 Warehouse Blvd		Tampa	62704
123	Florida	USA	123 Warehouse Blvd		Tampa	62704
124	Florida	USA	123 Warehouse Blvd		Tampa	62704
125	Florida	USA	123 Warehouse Blvd		Tampa	62704
126	Florida	USA	123 Warehouse Blvd		Tampa	62704
127	Florida	USA	123 Warehouse Blvd		Tampa	62704
128	Florida	USA	123 Warehouse Blvd		Tampa	62704
129	Florida	USA	123 Warehouse Blvd		Tampa	62704
130	Florida	USA	123 Warehouse Blvd		Tampa	62704
131	Florida	USA	456 Market Street		Miami	62701
132	Nova Scotia	Canada	12qw		olathe	12345
133	Florida	USA	456 Market Street		Miami	62701
134	Kansas	USA	14658 S Hagan St		Olathe	66062-9002
135	Florida	USA	456 Market Street		Miami	62701
136	Florida	USA	12301 Metro Parkway		Fort Myers	33966
137	Kansas	USA	14658 S Hagan ST		Olathe	66062
138	Kansas	USA	14658 S Hagan ST		Olathe	66062
139	Florida	USA	456 Market Street		Miami	62701
140	Manitoba	Canada	Plot no 123&124, Sampada Homes, Flat G2	Pragathi Nagar	HYDERABAD	500090
141	Florida	USA	123 Warehouse Blvd		Tampa	62704
142	Florida	USA	123 Warehouse Blvd		Tampa	62704
143	Florida	USA	123 Warehouse Blvd		Tampa	62704
144	Florida	USA	123 Warehouse Blvd		Tampa	62704
145	Florida	USA	123 Warehouse Blvd		Tampa	62704
146	Florida	USA	123 Warehouse Blvd		Tampa	62704
\.


--
-- Data for Name: agents; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.agents (id, name, email, password, password_hashed, details, role, assignee_type, superuser, created_at, updated_at, deleted_at) FROM stdin;
659808	durga	durga@gmail.com	gowtham	$2a$14$1R5mUGBRqZX2cA.GLwed8Oxo4v6jArpVCogwmA81nGAqKz6Odhuqq	{"phone": "1234567890", "branch": "it", "totpsecret": "", "totpenabled": false}	MANAGER		f	2025-08-11 14:47:53.300708+00	2025-08-11 14:47:53.300708+00	\N
910454	Pradeep	pradeepyandamuri532@gmail.com	pradeep@123	$2a$14$Muv8sBrhgg4G6GAGTDhltuhqcECtNTPwyLM3dhKSlw9xpRkoZglKW	{"phone": "8074604985", "branch": "IT", "totpsecret": "", "totpenabled": false}	MANAGER		f	2025-08-11 15:41:31.362483+00	2025-08-11 15:41:31.362483+00	\N
221472	raja	raja@gmail.com	Kiran1970#	$2a$14$8gpYAlAGESOXYQQARQbVpueBAd9FbatjMUnSusdIXyrfOgCRdyAoi	{"phone": "1234567890", "branch": "test", "totpsecret": "QEIMXSJGE6TGSY7DO744L6OCX3B4RZES", "totpenabled": true}	MANAGER		f	2025-08-12 08:50:21.008659+00	2025-08-12 08:50:29.328921+00	\N
351994	Dheeraj Pasham	rinkelreddy@yahoo.com	password	$2a$14$RV/bMIK5BNYtxgyFRwiQ8.oKnZZfUUYyyAk48K5tdsJwsBLwK9r8u	{"phone": "4088399420", "branch": "hyderabad", "totpsecret": "", "totpenabled": false}	MANAGER		f	2025-08-12 23:08:25.962133+00	2025-08-12 23:08:25.962133+00	\N
581600	manager	manager@gmail.com	getout612	$2a$14$HFQXJ8nIm.W1I9eJcIehtejwWB/P3uzld7JBLYMf.bVxmnTcnlfxe	{"phone": "9121925025", "branch": "Florida", "totpsecret": "", "totpenabled": false}	MANAGER		f	2025-08-13 03:23:11.603934+00	2025-08-13 03:23:11.603934+00	\N
038230	Dheerajreddy Pasham	dheerajpasham7@gmail.com	pAssw0rd@007	$2a$14$rRVXGMMzX4NZRqph.PEK3e2t.OGK.HKn965a9ke7zs5E2/OS9GWhy	{"phone": "4085904123", "branch": "hyderabad", "totpsecret": "", "totpenabled": false}	MANAGER		f	2025-08-14 15:54:20.415938+00	2025-08-14 15:54:20.415938+00	\N
999307	Shankar-Rao Mata	shankar@gmail.com	Kiranb1970#	$2a$14$.gFQ1PVKFsnB58q0Q3wMd.azHv/TPMDY86pbzvU4Lf.mspcLv/pMK	{"phone": "5058148672", "branch": "Test", "totpsecret": "", "totpenabled": false}	MANAGER		f	2025-08-19 04:00:28.097965+00	2025-08-19 04:00:28.097965+00	\N
472784	Shankar-Rao Mata	shankar1@gmail.com	Kiran1970#	$2a$14$ZRD7dYLWznjgv2raEURdfeCF4jb2msix.2Bv61CfDCzAAIE6f5/1m	{"phone": "5058148672", "branch": "Jaffa", "totpsecret": "", "totpenabled": false}	MANAGER		f	2025-08-19 04:12:46.073287+00	2025-08-19 04:12:46.073287+00	\N
047793	temp	temp@gmail.com	Kiran1970#	$2a$14$3Wu8vMYXkRIPrHAmjstg2.5.DAXUBsRkPby09CXsaZeNkvKM3og2C	{"phone": "5058148672", "branch": "PAG", "totpsecret": "", "totpenabled": false}	MANAGER		f	2025-08-19 04:45:07.150731+00	2025-08-19 04:45:07.150731+00	\N
164028	temp1	temp1@gmail.com	Kiran1970#	$2a$14$Pf4Z.xA5QivRPWkUIXUZGu6bwh7caDRmsSgGik4S15o3MQbrJ5GOe	{"phone": "5058148672", "branch": "RAG", "totpsecret": "GTADGHFLOMEDYO5EM522VKAB65U2HZMC", "totpenabled": true}	MANAGER		f	2025-08-19 05:10:27.210247+00	2025-08-19 05:10:36.327455+00	\N
877682	Test1	test@gmail.com	Kiran1970#	$2a$14$vMaevK0OwdUVPJapd2mNsu5JIfxcmufbmMpEfZmrQF7ktA4pD7EYa	{"phone": "1234456754", "branch": "Test", "totpsecret": "", "totpenabled": false}	MANAGER		f	2025-08-19 05:19:45.919195+00	2025-08-19 05:19:45.919195+00	\N
063103	Shankar-Rao Mata	shankar123@gmail.com	Kiran1970#	$2a$14$Z76QKmnJkKvqOPFK92dkr.nTV5gRyYFnw5aYVNwDSwTgDTiFXUvwq	{"phone": "5058148671", "branch": "PAG", "totpsecret": "", "totpenabled": false}	MANAGER		f	2025-08-21 07:18:49.185635+00	2025-08-21 07:18:49.185635+00	\N
\.


--
-- Data for Name: associate_customers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.associate_customers (id, details, name, role) FROM stdin;
918287	{"Email": "sainsg333@gmail.com", "Phone": "9640164778", "company": {"id": "247325", "name": "NSG"}, "LastName": "gowtham", "FirstName": "sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
631632	{"Email": "sainsg333@gmail.com", "Phone": "9640164778", "company": {"id": "247325", "name": "NSG"}, "LastName": "gowtham", "FirstName": "sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
468342	{"Email": "sainsg333@gmail.com", "Phone": "9640164778", "company": {"id": "247325", "name": "NSG"}, "LastName": "gowtham", "FirstName": "sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
898055	{"Email": "sainsg333@gmail.com", "Phone": "9640164778", "company": {"id": "247325", "name": "NSG"}, "LastName": "gowtham", "FirstName": "sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
787375	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
326700	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
240421	{"Email": "pradeep@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
013482	{"Email": "pradeep@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
085947	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
496853	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
164745	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
824131	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
960451	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
275808	{"Email": "sainsg333@gmail.com", "Phone": "9640164778", "company": {"id": "212184", "name": "Tulip"}, "LastName": "manohar", "FirstName": "aditya", "company_id": "212184", "CompanyName": "Tulip"}		ARCHITECT
150019	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
419649	{"Email": "sainsg333@gmail.com", "Phone": "9640164778", "company": {"id": "212184", "name": "Tulip"}, "LastName": "manohar", "FirstName": "aditya", "company_id": "212184", "CompanyName": "Tulip"}		ARCHITECT
125112	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		DESIGNER
859013	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		DESIGNER
825054	{"Email": "shankar@gmail.com", "Phone": "5058148671", "company": {"id": "212184", "name": "Tulip"}, "LastName": "mata", "FirstName": "shankar", "company_id": "212184", "CompanyName": "Tulip"}		DESIGNER
581012	{"Email": "test1@gmail.com", "Phone": "9998887771", "company": {"id": "963695", "name": "Adhitya groups"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "963695", "CompanyName": "Adhitya groups"}		ARCHITECT
918814	{"Email": "sainsg333@gmail.com", "Phone": "9640164778", "company": {"id": "212184", "name": "Tulip"}, "LastName": "manohar", "FirstName": "aditya", "company_id": "212184", "CompanyName": "Tulip"}		ARCHITECT
804615	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		DESIGNER
557090	{"Email": "shankar@gmail.com", "Phone": "5058148672", "company": {"id": "", "name": ""}, "LastName": "jaffa", "FirstName": "test", "company_id": "", "CompanyName": ""}		ARCHITECT
041036	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
105420	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "", "name": ""}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "", "CompanyName": ""}		ARCHITECT
286414	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
067404	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "", "name": ""}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "", "CompanyName": ""}		ARCHITECT
120226	{"Email": "", "Phone": "1232222334", "company": {"id": "", "name": ""}, "LastName": "ande", "FirstName": "hite", "company_id": "", "CompanyName": ""}		DESIGNER
604561	{"Email": "rinkelreddy@yahoo.com", "Phone": "8136506083", "company": {"id": "247325", "name": "NSG"}, "LastName": "Pasham", "FirstName": "Rishi", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
491637	{"Email": "shankar21@gmail.com", "Phone": "1234567890", "company": {"id": "963695", "name": "Adhitya groups"}, "LastName": "", "FirstName": "", "company_id": "963695", "CompanyName": "Adhitya groups"}		ARCHITECT
704761	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
624754	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
646089	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		CONTRACTOR
514220	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
056713	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		DESIGNER
120480	{"Email": "shankar@gmail.com", "Phone": "5058148672", "company": {"id": "247325", "name": "NSG"}, "LastName": "Mata", "FirstName": "Shankar-Rao", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
947079	{"Email": "", "Phone": "1231231212", "company": {"id": "963695", "name": "Adhitya groups"}, "LastName": "", "FirstName": "", "company_id": "963695", "CompanyName": "Adhitya groups"}		ARCHITECT
411972	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
809855	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
819745	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
301041	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
776072	{"Email": "", "Phone": "1231231234", "company": {"id": "258245", "name": "Shankar-Rao Mata"}, "LastName": "", "FirstName": "", "company_id": "258245", "CompanyName": "Shankar-Rao Mata"}		ARCHITECT
344611	{"Email": "", "Phone": "4085904123", "company": {"id": "424103", "name": "durga"}, "LastName": "Ande", "FirstName": "Mounika", "company_id": "424103", "CompanyName": "durga"}		ARCHITECT
778225	{"Email": "mounika.ande@gmail.com", "Phone": "4088399420", "company": {"id": "212184", "name": "Tulip"}, "LastName": "ande", "FirstName": "Rishi", "company_id": "212184", "CompanyName": "Tulip"}		ARCHITECT
791659	{"Email": "", "Phone": "123434343", "company": {"id": "355381", "name": "Rishi"}, "LastName": "", "FirstName": "", "company_id": "355381", "CompanyName": "Rishi"}		ARCHITECT
952419	{"Email": "", "Phone": "1231231212", "company": {"id": "963695", "name": "Adhitya groups"}, "LastName": "", "FirstName": "", "company_id": "963695", "CompanyName": "Adhitya groups"}		ARCHITECT
589276	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
373889	{"Email": "Gowtham@gmail.com", "Phone": "9998887771", "company": {"id": "247325", "name": "NSG"}, "LastName": "Gowtham", "FirstName": "Sai", "company_id": "247325", "CompanyName": "NSG"}		ARCHITECT
\.


--
-- Data for Name: county_taxes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.county_taxes (id, name, tax_percentage) FROM stdin;
1	Jackson	7.2
2	Jefferson	8.5
3	Madison	6.800000190734863
7	maincounty	7.599999904632568
10	johnson	7.400000095367432
12	newcounty1	9.899999618530273
\.


--
-- Data for Name: customers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.customers (id, name, email, phone) FROM stdin;
247325	NSG	\N	\N
212184	Tulip	\N	\N
963695	Adhitya groups	\N	\N
773920	Balaji	\N	\N
424103	durga	durga@gmail.com	1234567891
258245	Shankar-Rao Mata	\N	\N
355381	Rishi	\N	\N
\.




--
-- Data for Name: estimates; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.estimates (id, created_at, updated_at, deleted_at, lead_id, name, material_total, service_total, details, status) FROM stdin;
005640	2025-08-11 15:13:48.169658+00	2025-08-11 15:13:48.169658+00	\N	153714	Comprehensive Estimate - 1 Residential Entries	200	50	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 218, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
828551	2025-08-11 15:16:42.655872+00	2025-08-11 15:16:42.655872+00	\N	338697	Comprehensive Estimate - 2 Residential Entries	310	100	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 410, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
138115	2025-08-11 15:24:05.292174+00	2025-08-11 15:24:05.292174+00	\N	278004	Comprehensive Estimate - 1 Residential Entries	100	40	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 140, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
299571	2025-08-11 16:43:24.6614+00	2025-08-11 16:43:24.6614+00	\N	379122	Comprehensive Estimate - 1 Residential Entries	100	40	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 136.08, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
449661	2025-08-11 17:08:33.01404+00	2025-08-11 17:08:33.01404+00	\N	990977	Comprehensive Estimate - 2 Residential Entries	301	110	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 411, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
440111	2025-08-11 17:33:37.782785+00	2025-08-11 17:33:37.782785+00	\N	087879	Comprehensive Estimate - 1 Residential Entries	100	40	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 140, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
104900	2025-08-11 17:49:36.88935+00	2025-08-11 17:56:25.471643+00	\N	788865	Comprehensive Estimate - 1 Residential Entries	2650	190	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 2632.68, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
348917	2025-08-11 18:23:49.194208+00	2025-08-11 18:23:49.194208+00	\N	764542	Comprehensive Estimate - 1 Residential Entries	1100	80	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 1044.3, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
761845	2025-08-11 17:44:08.307286+00	2025-08-11 18:44:27.553399+00	\N	893916	Comprehensive Estimate - 1 Residential Entries	500	91	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 515.35, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
602959	2025-08-11 19:10:07.768493+00	2025-08-11 19:10:07.768493+00	\N	028735	Comprehensive Estimate - 1 Residential Entries	0	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 60, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
198231	2025-08-11 19:12:53.956665+00	2025-08-11 19:12:53.956665+00	\N	945419	Comprehensive Estimate - 1 Residential Entries	0	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 60, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
466116	2025-08-11 19:16:42.922826+00	2025-08-11 19:16:42.922826+00	\N	756697	Comprehensive Estimate - 1 Residential Entries	0	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 60, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
947144	2025-08-11 19:21:22.824984+00	2025-08-11 19:21:22.824984+00	\N	512161	Comprehensive Estimate - 1 Residential Entries	1100	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 1160, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
791963	2025-08-11 19:23:34.607089+00	2025-08-11 19:23:34.607089+00	\N	198032	Comprehensive Estimate - 1 Residential Entries	0	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 60, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
468618	2025-08-11 20:07:15.305619+00	2025-08-11 20:07:15.305619+00	\N	966384	Comprehensive Estimate - 2 Residential Entries	0	120	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 120, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
189986	2025-08-11 20:10:16.850237+00	2025-08-11 20:10:16.850237+00	\N	553579	Comprehensive Estimate - 1 Residential Entries	0	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 47.7, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
260462	2025-08-11 20:13:35.24901+00	2025-08-11 20:13:35.24901+00	\N	971784	Comprehensive Estimate - 1 Residential Entries	0	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 58.32, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
790224	2025-08-11 20:29:59.640863+00	2025-08-11 20:29:59.640863+00	\N	115148	Comprehensive Estimate - 1 Residential Entries	0	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 58.08, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
659728	2025-08-11 20:41:00.102132+00	2025-08-11 20:41:00.102132+00	\N	920579	Comprehensive Estimate - 1 Residential Entries	0	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 53.1, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
811282	2025-08-11 20:48:46.566247+00	2025-08-11 20:50:05.240349+00	\N	078083	Comprehensive Estimate - 1 Residential Entries	0	40	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 39.4, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
112352	2025-08-11 20:51:21.028023+00	2025-08-11 20:51:21.028023+00	\N	571272	Comprehensive Estimate - 1 Residential Entries	0	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 53.1, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
846594	2025-08-12 04:50:38.495187+00	2025-08-12 05:13:33.713649+00	\N	573182	Comprehensive Estimate - 1 Residential Entries	1510	120	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 1511.01, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
604666	2025-08-12 09:37:36.925266+00	2025-08-12 09:41:41.940655+00	\N	297805	Comprehensive Estimate - 0 Residential Entries	0	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 64.32, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
359478	2025-08-12 05:23:04.192754+00	2025-08-12 05:28:58.144999+00	\N	684877	Comprehensive Estimate - 1 Residential Entries	1770	150	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 1699.2, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
640014	2025-08-12 09:46:08.254119+00	2025-08-12 09:46:08.254119+00	\N	820700	Comprehensive Estimate - 2 Residential Entries	100	120	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 191.84, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
365344	2025-08-12 06:34:34.872804+00	2025-08-12 06:40:53.872088+00	\N	022652	Comprehensive Estimate - 1 Residential Entries	3500	180	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 3227.36, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
153822	2025-08-12 08:38:43.016073+00	2025-08-12 09:05:25.173951+00	\N	884011	Comprehensive Estimate - 1 Residential Entries	2390	200	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 2292.15, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
436116	2025-08-12 09:50:05.232693+00	2025-08-12 09:50:05.232693+00	\N	440604	Comprehensive Estimate - 2 Residential Entries	400	140	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 524.88, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
849736	2025-08-12 07:07:27.355911+00	2025-08-12 07:42:12.179651+00	\N	766166	Comprehensive Estimate - 1 Residential Entries	2290	210	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 2212.5, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
941746	2025-08-12 08:36:13.582745+00	2025-08-12 08:36:13.582745+00	\N	304586	Comprehensive Estimate - 1 Residential Entries	500	80	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 513.3, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
307848	2025-08-12 09:03:54.363506+00	2025-08-12 09:10:18.823848+00	\N	301498	Comprehensive Estimate - 2 Residential Entries	170	120	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 252.88, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
720330	2025-08-12 09:05:32.210402+00	2025-08-12 09:12:10.08504+00	\N	973844	Comprehensive Estimate - 2 Residential Entries	1200	170	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 1189.16, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
416334	2025-08-12 09:16:05.883936+00	2025-08-12 09:16:05.883936+00	\N	554098	Comprehensive Estimate - 1 Residential Entries	2000	72	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 1806.78, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
420632	2025-08-12 23:13:44.265714+00	2025-08-12 23:13:44.265714+00	\N	467243	Comprehensive Estimate - 1 Residential Entries	100	0	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 93.5, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
561785	2025-08-12 10:10:22.543271+00	2025-08-12 10:16:00.852782+00	\N	701868	Comprehensive Estimate - 2 Residential Entries	4900	270	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 4534.09, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
549052	2025-08-12 10:41:53.746101+00	2025-08-12 13:14:24.946045+00	\N	487281	Comprehensive Estimate - 1 Residential Entries	4900	280	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 4496.24, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
800426	2025-08-13 00:07:05.54667+00	2025-08-13 00:07:05.54667+00	\N	317521	Comprehensive Estimate - 1 Residential Entries	0	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 53.1, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
575388	2025-08-12 23:56:05.167791+00	2025-08-12 23:57:06.419659+00	\N	253703	Comprehensive Estimate - 1 Residential Entries	1240	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 1128.4, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
141628	2025-08-13 00:25:01.716981+00	2025-08-13 00:25:01.716981+00	\N	022137	Comprehensive Estimate - 2 Residential Entries	500	120	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 579.7, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
367470	2025-08-13 00:42:21.677692+00	2025-08-13 00:42:21.677692+00	\N	323585	Comprehensive Estimate - 1 Residential Entries	2210	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 2102.02, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
491094	2025-08-13 01:03:45.443809+00	2025-08-13 01:03:45.443809+00	\N	148872	Comprehensive Estimate - 2 Residential Entries	2320	120	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 2483.92, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
096722	2025-08-13 01:04:36.019838+00	2025-08-13 01:04:36.019838+00	\N	858123	Comprehensive Estimate - 1 Residential Entries	10	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 61.18, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
296995	2025-08-13 01:10:48.037827+00	2025-08-13 01:10:48.037827+00	\N	688174	Comprehensive Estimate - 2 Residential Entries	150	120	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 252.45, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
221078	2025-08-13 01:18:35.253557+00	2025-08-13 01:18:35.253557+00	\N	341276	Comprehensive Estimate - 1 Residential Entries	1000	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 1030.32, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
419955	2025-08-14 16:24:53.88393+00	2025-08-14 16:24:53.88393+00	\N	653535	Comprehensive Estimate - 2 Residential Entries	2210	120	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 2138.94, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
874354	2025-08-14 15:44:13.774766+00	2025-08-14 15:49:43.87731+00	\N	043649	Comprehensive Estimate - 1 Residential Entries	2200	191	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 2355.14, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
422170	2025-08-14 15:57:58.440348+00	2025-08-14 15:57:58.440348+00	\N	806460	Comprehensive Estimate - 1 Residential Entries	200	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 256.1, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
215051	2025-08-13 18:53:37.053639+00	2025-08-13 19:02:48.39241+00	\N	535395	Comprehensive Estimate - 2 Residential Entries	4400	380	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 4388.04, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
764222	2025-08-14 02:23:57.217316+00	2025-08-14 02:23:57.217316+00	\N	249777	Comprehensive Estimate - 1 Residential Entries	200	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 230.1, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
276098	2025-08-14 05:46:01.52165+00	2025-08-14 05:46:01.52165+00	\N	616587	Comprehensive Estimate - 1 Residential Entries	900	70	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 845.84, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
878573	2025-08-14 06:05:09.791622+00	2025-08-14 06:05:09.791622+00	\N	747298	Comprehensive Estimate - 1 Residential Entries	500	140	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 587.52, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
181659	2025-08-13 03:43:36.323261+00	2025-08-14 12:26:32.685806+00	\N	870087	Comprehensive Estimate - 1 Residential Entries	3600	280	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 3391.12, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
328348	2025-08-14 13:20:33.562802+00	2025-08-14 13:20:33.562802+00	\N	244629	Comprehensive Estimate - 2 Residential Entries	0	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 59.1, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
836278	2025-08-14 14:35:50.892264+00	2025-08-14 14:36:40.192911+00	\N	110545	Comprehensive Estimate - 2 Residential Entries	200	131	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 326.04, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
034042	2025-08-14 16:02:22.743345+00	2025-08-14 16:06:12.45215+00	\N	496909	Comprehensive Estimate - 0 Residential Entries	100	0	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 92.2, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
677087	2025-08-14 16:18:10.840147+00	2025-08-14 16:18:10.840147+00	\N	484015	Comprehensive Estimate - 2 Residential Entries	1468	120	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 1457.78, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
542229	2025-08-14 16:30:40.56553+00	2025-08-14 16:34:12.486166+00	\N	255076	Comprehensive Estimate - 0 Residential Entries	30	0	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 27.54, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
009177	2025-08-14 18:06:12.637643+00	2025-08-14 18:06:12.637643+00	\N	220509	Comprehensive Estimate - 1 Residential Entries	10	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 61.95, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
000848	2025-08-14 18:25:48.325624+00	2025-08-14 18:25:48.325624+00	\N	076315	Comprehensive Estimate - 1 Residential Entries	190	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 221.25, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
152654	2025-08-14 21:22:23.139345+00	2025-08-14 21:22:23.139345+00	\N	598530	Comprehensive Estimate - 2 Residential Entries	1020	41	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 938.99, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
295267	2025-08-14 21:29:32.535239+00	2025-08-14 21:31:35.286985+00	\N	605988	Comprehensive Estimate - 0 Residential Entries	20	0	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 18.7, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
066615	2025-08-19 08:09:46.461796+00	2025-08-19 08:09:46.461796+00	\N	818324	Comprehensive Estimate - 1 Residential Entries	300	50	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 321.3, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
045813	2025-08-19 08:13:35.8654+00	2025-08-21 06:51:42.678204+00	\N	673065	Comprehensive Estimate - 1 Residential Entries	1000	60	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 1060, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
585455	2025-08-21 09:48:56.84765+00	2025-08-21 14:46:51.683084+00	\N	660420	Comprehensive Estimate - 1 Residential Entries	1500	110	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 1490.86, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
204729	2025-08-22 02:43:17.413617+00	2025-08-22 02:43:17.413617+00	\N	888866	Comprehensive Estimate - 1 Residential Entries	200	90	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 290, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
211806	2025-08-22 03:03:24.656133+00	2025-08-22 03:03:24.656133+00	\N	556657	Comprehensive Estimate - 1 Residential Entries	140	90	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 211.14, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
106485	2025-08-22 03:13:49.836381+00	2025-08-22 03:13:49.836381+00	\N	652375	Comprehensive Estimate - 2 Residential Entries	1600	160	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 1645.6, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
004083	2025-08-22 06:06:16.378366+00	2025-08-22 06:06:16.378366+00	\N	820904	Comprehensive Estimate - 0 Residential Entries	0	0	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 0, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
258800	2025-08-22 06:06:20.17161+00	2025-08-22 06:06:20.17161+00	\N	635119	Comprehensive Estimate - 0 Residential Entries	0	0	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 0, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
100565	2025-08-22 09:13:35.111985+00	2025-08-22 09:13:35.111985+00	\N	580823	Comprehensive Estimate - 1 Residential Entries	0	80	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 70.8, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
378895	2025-08-22 09:57:57.677006+00	2025-08-22 09:57:57.677006+00	\N	142588	Comprehensive Estimate - 1 Residential Entries	100	80	{"Tax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "TaxID": 0, "Total": 175.68, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "LaborCost": 0, "DiscountID": 0}	draft
\.


--
-- Data for Name: files; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.files (id, name, type, size, uploaded_at, project_type, stored_file_name, category, lead_id, file_path) FROM stdin;
1	Screenshot from 2025-07-29 23-55-45.png	png	183	2025-08-11 20:43:04.511034	Residential Remodel - Kitchen	2f064e2a-2479-48ae-af6d-ad6fc8762a9e.png	CAD	153714	filestorage/2f064e2a-2479-48ae-af6d-ad6fc8762a9e.png
2	Screenshot from 2025-07-31 11-03-38.png	png	4939	2025-08-11 20:43:04.589639	Residential Remodel - Kitchen	1604661f-46c1-49be-8cf6-b5db28b021e3.png	CAD	153714	filestorage/1604661f-46c1-49be-8cf6-b5db28b021e3.png
3	Screenshot from 2025-07-30 20-23-43.png	png	10498	2025-08-11 20:43:04.629179	Residential Remodel - Kitchen	67afcd6f-7c75-46eb-aa28-406b34e23181.png	CAD	153714	filestorage/67afcd6f-7c75-46eb-aa28-406b34e23181.png
4	Screenshot from 2025-08-11 20-21-59.png	png	43581	2025-08-11 20:43:04.671073	Residential Remodel - Kitchen	a1a7921a-ecb2-4fcd-82da-428d9dea0bdf.png	CAD	153714	filestorage/a1a7921a-ecb2-4fcd-82da-428d9dea0bdf.png
6	Screenshot from 2025-07-31 22-21-29.png	png	27632	2025-08-11 22:27:11.546778	Residential New-Construction - Master Bedroom	0713ac23-6968-4a2b-ae47-37ed483228b0.png	All	334790	filestorage/0713ac23-6968-4a2b-ae47-37ed483228b0.png
10	Screenshot from 2025-07-31 11-44-59.png	png	94461	2025-08-11 23:15:54.768654	Residential New-Construction - Master Bedroom	ac6e8ae5-19c4-4f66-9fbd-26ba9e0b2f87.png	All	788865	filestorage/ac6e8ae5-19c4-4f66-9fbd-26ba9e0b2f87.png
11	Screenshot from 2025-07-31 22-21-29.png	png	27632	2025-08-11 23:15:54.803759	Residential New-Construction - Master Bedroom	1cb75758-e85b-47b1-8c26-b8d83ed5d438.png	All	788865	filestorage/1cb75758-e85b-47b1-8c26-b8d83ed5d438.png
12	Screenshot from 2025-08-07 14-45-18.png	image/png	77655	2025-08-11 23:21:16.661392	Residential New-Construction - Master Bathroom	0f44015e-ce57-4125-83b1-1e27b55e976b.png	CAD	788865	filestorage/0f44015e-ce57-4125-83b1-1e27b55e976b.png
14	Screenshot from 2025-07-31 22-21-29.png	image/png	27632	2025-08-11 23:58:55.018741	Residential New-Construction - Master Bathroom	ae044752-4fc5-4798-a0b5-bde2bba79eeb.png	Fabrication	764542	filestorage/ae044752-4fc5-4798-a0b5-bde2bba79eeb.png
15	Screenshot from 2025-07-31 11-44-59.png	image/png	94461	2025-08-12 01:34:40.472192	Residential New-Construction - Master Bathroom	7680a856-3100-42c1-8f61-54e4cebb75da.png	Installation	512161	filestorage/7680a856-3100-42c1-8f61-54e4cebb75da.png
17	Screenshot from 2025-08-07 14-45-18.png	png	77655	2025-08-12 10:18:26.434382	Residential New-Construction - Master Bedroom	56bcda15-ae60-450b-a7fa-f62b9a864193.png	All	573182	filestorage/56bcda15-ae60-450b-a7fa-f62b9a864193.png
18	Screenshot from 2025-08-07 14-45-18.png	image/png	77655	2025-08-12 11:44:54.769722	Apartment - Kitchen	8bec240d-514e-496d-aedd-a59ed9243e2e.png	Bills	684877	filestorage/8bec240d-514e-496d-aedd-a59ed9243e2e.png
20	Screenshot from 2025-08-07 14-43-33.png	png	77655	2025-08-12 12:02:16.824824	Residential New-Construction - Master Bathroom	7435001b-4351-4806-8e2e-fbad0b30a7bc.png	All	022652	filestorage/7435001b-4351-4806-8e2e-fbad0b30a7bc.png
22	Screenshot from 2025-08-07 14-43-33.png	png	77655	2025-08-12 12:28:00.698262	Residential New-Construction - Master Bedroom	1f6a9d7b-4b6c-45e4-9900-7bb52bca10bc.png	All	304586	filestorage/1f6a9d7b-4b6c-45e4-9900-7bb52bca10bc.png
23	Screenshot from 2025-08-07 14-43-33.png	image/png	77655	2025-08-12 12:41:15.58157	Residential New-Construction - Bathroom	f1157f37-4970-4fe7-a34e-0624bd37794b.png	Bills	766166	filestorage/f1157f37-4970-4fe7-a34e-0624bd37794b.png
25	Screenshot from 2025-08-07 13-15-49.png	png	41230	2025-08-12 14:04:11.125832	Residential New-Construction - Master Bedroom	9a418f3f-3dc2-4d95-bb05-65dd08a9cf1a.png	All	884011	filestorage/9a418f3f-3dc2-4d95-bb05-65dd08a9cf1a.png
27	assignment.pdf	pdf	4021302	2025-08-12 14:29:37.738531	Residential Remodel - Bathroom	46c11246-6b2e-490c-a319-1cdf4a487be5.pdf	Template	301498	filestorage/46c11246-6b2e-490c-a319-1cdf4a487be5.pdf
28	Screenshot from 2025-08-07 13-15-49.png	png	41230	2025-08-12 14:31:16.327933	Residential New-Construction - Master Bathroom	af9d615f-8e34-4b44-ba26-fcccf434a8be.png	All	973844	filestorage/af9d615f-8e34-4b44-ba26-fcccf434a8be.png
29	Screenshot from 2025-07-29 23-55-45.png	png	183	2025-08-12 15:14:51.077244	Residential New-Construction - Bathroom	614797c2-5c49-49ed-b659-7b208f9b1b55.png	CAD	820700	filestorage/614797c2-5c49-49ed-b659-7b208f9b1b55.png
30	Screenshot from 2025-07-31 11-03-38.png	png	4939	2025-08-12 15:14:51.177242	Residential New-Construction - Bathroom	83ccf148-7d08-45be-8b0c-478b52ac733a.png	CAD	820700	filestorage/83ccf148-7d08-45be-8b0c-478b52ac733a.png
31	Screenshot from 2025-07-30 20-23-43.png	png	10498	2025-08-12 15:14:51.22924	Residential New-Construction - Bathroom	a64597c6-cb89-4a56-acf3-0954e16af305.png	CAD	820700	filestorage/a64597c6-cb89-4a56-acf3-0954e16af305.png
32	Screenshot from 2025-08-11 20-21-59.png	png	43581	2025-08-12 15:14:51.291467	Residential New-Construction - Bathroom	10d9d789-e8d1-4157-951b-0f0ec3ece8a1.png	CAD	820700	filestorage/10d9d789-e8d1-4157-951b-0f0ec3ece8a1.png
33	Screenshot from 2025-08-07 13-15-49.png	png	41230	2025-08-12 15:34:09.384489	Commercial - Bathroom	b39533b7-c24f-4c3a-afd5-ad2f044f8499.png	All	701868	filestorage/b39533b7-c24f-4c3a-afd5-ad2f044f8499.png
35	Screenshot from 2025-08-07 14-43-33.png	image/png	77655	2025-08-12 10:48:36.600134	Residential New-Construction - Other	ce0be2ba-2db9-4322-a2dc-6cf2cb87e9c0.png	Bills	487281	filestorage/ce0be2ba-2db9-4322-a2dc-6cf2cb87e9c0.png
36	Screenshot from 2025-08-07 13-15-49.png	image/png	41230	2025-08-12 11:13:53.135194	Residential New-Construction - Master Bathroom	7251affa-9ede-45da-9070-0dc6340ae77a.png	Template	487281	filestorage/7251affa-9ede-45da-9070-0dc6340ae77a.png
37	Screenshot from 2025-08-07 14-43-33.png	png	77655	2025-08-12 12:55:29.992599	Residential New-Construction - Master Bedroom	b428e5c5-dc50-4ac3-8957-9d33c858c31e.png	All	858123	filestorage/b428e5c5-dc50-4ac3-8957-9d33c858c31e.png
38	Screenshot 2025-08-05 at 7.03.23 PM.png	image/png	41350	2025-08-12 23:14:27.842786	Residential Remodel - Kitchen	c493b60d-fc92-4655-8cb3-321d00e4ba44.png	Templates	467243	filestorage/c493b60d-fc92-4655-8cb3-321d00e4ba44.png
39	Screenshot 2025-08-05 at 6.49.23 PM.png	image/png	135390	2025-08-12 23:19:05.883729	Residential Remodel - Kitchen	642c002f-b0a9-4dd7-b214-98fc8f1686b9.png	Other	467243	filestorage/642c002f-b0a9-4dd7-b214-98fc8f1686b9.png
40	Screenshot 2025-08-05 at 6.19.51 PM.png	image/png	12599	2025-08-12 23:19:17.318821	Residential Remodel - Kitchen	576b47bb-bd72-44ec-90ab-3a8bc527f77b.png	Fabrication	467243	filestorage/576b47bb-bd72-44ec-90ab-3a8bc527f77b.png
41	Screenshot 2025-08-05 at 7.01.52 PM.png	png	75064	2025-08-12 23:50:15.12999	Residential New-Construction - Bathroom	ef68c44d-b5b7-4c9d-97ea-e72902a6935a.png	Service	253703	filestorage/ef68c44d-b5b7-4c9d-97ea-e72902a6935a.png
42	Screenshot 2025-08-05 at 7.03.23 PM.png	image/png	41350	2025-08-13 00:27:21.545604	Residential New-Construction - Master Bathroom	784b76a5-7534-47af-ab1a-4dae50f63c0e.png	Cutting	022137	filestorage/784b76a5-7534-47af-ab1a-4dae50f63c0e.png
43	Screenshot 2025-08-05 at 6.24.14 PM.png	image/png	63217	2025-08-13 00:27:32.450187	Residential New-Construction - Master Bathroom	126f7005-aaae-4d12-b3ed-fb90df42cb38.png	Cutting	022137	filestorage/126f7005-aaae-4d12-b3ed-fb90df42cb38.png
44	Screenshot 2025-08-05 at 7.01.52 PM.png	image/png	75064	2025-08-13 00:43:29.936098	Commercial - Master Bathroom	61958902-1829-416e-9f4f-3730d7d10f5f.png	Bills	323585	filestorage/61958902-1829-416e-9f4f-3730d7d10f5f.png
45	Screenshot 2025-08-05 at 6.19.51 PM.png	image/png	12599	2025-08-13 01:05:00.59684	Residential New-Construction - Master Bedroom	2812d2b7-b637-43ba-8926-ee0fbe6d36df.png	Template	148872	filestorage/2812d2b7-b637-43ba-8926-ee0fbe6d36df.png
46	Hyderabad_2_Chennai.pdf	pdf	457852	2025-08-13 01:09:54.227934	Residential Remodel - Bathroom	8abf2801-ec94-434c-b347-e57d6d9fb663.pdf	Template	688174	filestorage/8abf2801-ec94-434c-b347-e57d6d9fb663.pdf
47	RAG_AI.pdf	pdf	1285700	2025-08-13 01:09:54.231221	Residential New-Construction - Master Bathroom	4f1c3f20-92db-4a7a-b704-67b350f14409.pdf	CAD	688174	filestorage/4f1c3f20-92db-4a7a-b704-67b350f14409.pdf
48	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-13 07:46:11.419507	Residential New-Construction - Master Bedroom	c61ddbb1-ecdd-4366-a374-92e49137156b.png	Template	688174	filestorage/c61ddbb1-ecdd-4366-a374-92e49137156b.png
49	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-13 07:46:11.596544	Residential New-Construction - Master Bedroom	30f3d4a6-fa43-412d-b37b-9dc0c73e8d97.png	Template	688174	filestorage/30f3d4a6-fa43-412d-b37b-9dc0c73e8d97.png
50	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-13 07:46:11.69518	Residential New-Construction - Master Bedroom	1968f46f-9b57-4d06-9993-3932c6735917.png	Template	688174	filestorage/1968f46f-9b57-4d06-9993-3932c6735917.png
51	Screenshot from 2025-07-31 22-21-29.png	png	27632	2025-08-13 03:41:57.726036	Commercial - Master Bedroom	f8f2035e-bc01-42cf-8763-cff037d5bc93.png	All	870087	filestorage/f8f2035e-bc01-42cf-8763-cff037d5bc93.png
52	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-13 11:52:02.649559	Commercial - Master Bathroom	bb19820e-c86e-4fbe-a167-0b86f41a14ec.png	Template	870087	filestorage/bb19820e-c86e-4fbe-a167-0b86f41a14ec.png
53	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-13 11:52:02.854449	Commercial - Master Bathroom	b9dd1dce-4b90-409f-b215-bcd46dc24b0a.png	Template	870087	filestorage/b9dd1dce-4b90-409f-b215-bcd46dc24b0a.png
54	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-13 11:52:03.010485	Commercial - Master Bathroom	2658329d-dd98-4ba1-b521-fffc5322bbaf.png	Template	870087	filestorage/2658329d-dd98-4ba1-b521-fffc5322bbaf.png
55	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-13 12:05:39.993229	Commercial - Master Bathroom	d4c83cac-ab77-4dd4-bf5e-68488b0c8d9b.png	Installation	870087	filestorage/d4c83cac-ab77-4dd4-bf5e-68488b0c8d9b.png
58	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-13 12:07:03.643161	Commercial - Master Bathroom	0e76669a-d4ec-4a08-afd6-f380560c0c34.png	Fabrication	870087	filestorage/0e76669a-d4ec-4a08-afd6-f380560c0c34.png
59	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-13 12:07:03.878832	Commercial - Master Bathroom	ddbfb032-cb5e-43d1-9438-a42cd8ba47e3.png	Fabrication	870087	filestorage/ddbfb032-cb5e-43d1-9438-a42cd8ba47e3.png
60	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-13 12:07:04.049308	Commercial - Master Bathroom	0460cc17-b83c-44f0-bb7b-a026eaaa4a99.png	Fabrication	870087	filestorage/0460cc17-b83c-44f0-bb7b-a026eaaa4a99.png
62	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-13 12:09:42.048824	Residential Remodel - Bathroom	dfd1ab83-acd1-42ee-a88a-584dc072da3b.png	Template	688174	filestorage/dfd1ab83-acd1-42ee-a88a-584dc072da3b.png
63	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-13 12:09:48.014194	Residential Remodel - Bathroom	47b6f8be-8954-4be7-af4e-9a3d0a294ab4.png	CAD	688174	filestorage/47b6f8be-8954-4be7-af4e-9a3d0a294ab4.png
64	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-13 12:09:48.207864	Residential Remodel - Bathroom	80135489-dece-4bfc-9a3c-32551ba03076.png	CAD	688174	filestorage/80135489-dece-4bfc-9a3c-32551ba03076.png
65	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-13 12:09:48.338727	Residential Remodel - Bathroom	f383d0a3-9f19-4f79-a1ab-61b684a04e24.png	CAD	688174	filestorage/f383d0a3-9f19-4f79-a1ab-61b684a04e24.png
66	Screenshot from 2025-08-11 20-21-59.png	image/png	43581	2025-08-13 12:09:48.461164	Residential Remodel - Bathroom	6302d51a-f580-46ae-9241-bbd0a1308ec2.png	CAD	688174	filestorage/6302d51a-f580-46ae-9241-bbd0a1308ec2.png
67	icon.png	image/png	68123	2025-08-13 12:10:09.364119	Residential Remodel - Bathroom	e9747450-28c7-441d-a1ee-19821701396e.png	CAD	688174	filestorage/e9747450-28c7-441d-a1ee-19821701396e.png
68	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-13 12:32:01.549611	Residential New-Construction - Master Bedroom	dde247be-861e-4466-b596-79269d6f563b.png	CAD	688174	filestorage/dde247be-861e-4466-b596-79269d6f563b.png
69	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-13 12:32:01.778383	Residential New-Construction - Master Bedroom	e6b4b729-eba7-4fc7-a2a3-e7d3163d8f90.png	CAD	688174	filestorage/e6b4b729-eba7-4fc7-a2a3-e7d3163d8f90.png
70	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-13 12:32:01.895093	Residential New-Construction - Master Bedroom	c0549444-ad9e-4b00-86c9-898d00453d37.png	CAD	688174	filestorage/c0549444-ad9e-4b00-86c9-898d00453d37.png
56	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-13 12:05:40.195136	Commercial - Master Bathroom	c739a73e-8391-47dc-b0a6-469480704d76.png	Installation	870087	filestorage/c739a73e-8391-47dc-b0a6-469480704d76.png
71	Screenshot from 2025-08-07 13-15-49.png	image/png	41230	2025-08-13 07:21:55.9452	Commercial - Master Bathroom	c11451ca-e3c4-4576-9777-6431b6401d2e.png	All	870087	filestorage/c11451ca-e3c4-4576-9777-6431b6401d2e.png
72	Screenshot from 2025-08-07 13-15-49.png	image/png	41230	2025-08-13 07:22:12.920257	Commercial - Master Bathroom	21cee804-969f-430d-8f84-48e3210f57b9.png	Template	870087	filestorage/21cee804-969f-430d-8f84-48e3210f57b9.png
73	Screenshot from 2025-08-12 18-32-35.png	image/png	35900	2025-08-13 07:22:38.399923	Commercial - Master Bathroom	4dc54f4f-84da-43ab-a62c-5e264b4c1d65.png	Service	870087	filestorage/4dc54f4f-84da-43ab-a62c-5e264b4c1d65.png
74	Screenshot from 2025-08-12 18-32-35.png	image/png	35900	2025-08-13 07:23:09.050639	Commercial - Master Bathroom	980cb3a4-a93c-4e8e-a54d-6b754931b64f.png	Service	870087	filestorage/980cb3a4-a93c-4e8e-a54d-6b754931b64f.png
77	Screenshot from 2025-07-31 22-21-29.png	image/png	27632	2025-08-13 07:24:17.450181	Commercial - Master Bathroom	ec261d9b-b170-4344-9aad-05478f468bf2.png	Service	870087	filestorage/ec261d9b-b170-4344-9aad-05478f468bf2.png
78	Screenshot from 2025-07-31 11-44-59.png	image/png	94461	2025-08-13 07:24:36.749903	Commercial - Master Bathroom	83fb5a57-5f2f-4c31-a2c5-ec9b608f55ee.png	All	870087	filestorage/83fb5a57-5f2f-4c31-a2c5-ec9b608f55ee.png
83	Screenshot from 2025-08-07 13-15-49.png	image/png	41230	2025-08-13 15:17:17.981632	Commercial - Master Bathroom	26c95086-2a57-43ec-b30a-a396e3a0c43d.png	Fabrication	870087	filestorage/26c95086-2a57-43ec-b30a-a396e3a0c43d.png
84	Screenshot from 2025-07-31 22-21-29.png	image/png	27632	2025-08-13 15:17:26.618694	Commercial - Master Bathroom	6a812baa-76b6-4348-b6a9-51d200550b1a.png	Installation	870087	filestorage/6a812baa-76b6-4348-b6a9-51d200550b1a.png
85	Screenshot from 2025-07-31 22-21-29.png	image/png	27632	2025-08-13 15:17:37.45019	Commercial - Master Bathroom	7569d44d-51d6-4693-8724-b162b5f7012f.png	Service Call	870087	filestorage/7569d44d-51d6-4693-8724-b162b5f7012f.png
87	Screenshot from 2025-08-07 13-15-49.png	png	41230	2025-08-14 00:19:05.622252	Commercial - Master Bedroom	92ac850c-4124-4832-8d12-1113f9393d10.png	All	535395	filestorage/92ac850c-4124-4832-8d12-1113f9393d10.png
88	Screenshot from 2025-08-12 18-32-35.png	image/png	35900	2025-08-14 00:25:27.153328	Commercial - Master Bathroom	95203e49-4a52-4974-b756-1112c07146b9.png	Installation	535395	filestorage/95203e49-4a52-4974-b756-1112c07146b9.png
89	Screenshot from 2025-08-07 13-15-49.png	image/png	41230	2025-08-14 00:33:48.772685	Apartment - Master Bedroom	94de154b-6b97-4bed-a0a7-1d6463738666.png	Other	535395	filestorage/94de154b-6b97-4bed-a0a7-1d6463738666.png
90	Screenshot from 2025-08-07 14-43-33.png	image/png	77655	2025-08-14 00:34:30.387706	Apartment - Master Bedroom	5ad0d113-853d-41cd-9f52-502a0d5907bc.png	Cutting	535395	filestorage/5ad0d113-853d-41cd-9f52-502a0d5907bc.png
91	Screenshot from 2025-07-31 22-21-29.png	image/png	27632	2025-08-14 00:34:40.306061	Apartment - Master Bedroom	fbf0d699-349a-4b19-bc78-248c6659b440.png	Cutting	535395	filestorage/fbf0d699-349a-4b19-bc78-248c6659b440.png
92	Screenshot from 2025-08-12 18-32-35.png	image/png	35900	2025-08-14 00:34:49.672988	Apartment - Master Bedroom	e7374945-ec0a-4045-bd45-841be8ad9f28.png	Fabrication	535395	filestorage/e7374945-ec0a-4045-bd45-841be8ad9f28.png
93	Screenshot from 2025-07-31 22-21-29.png	png	27632	2025-08-14 11:11:35.350773	Residential New-Construction - New Corridor	ee57ced9-9ff1-4405-b954-98f498381246.png	Installation	616587	filestorage/ee57ced9-9ff1-4405-b954-98f498381246.png
94	Screenshot from 2025-07-31 22-21-29.png	png	27632	2025-08-14 11:30:56.094088	Commercial - Master Bedroom	1c6ea98b-cbaa-4f70-8ac0-0851051aeaa2.png	Cutting	747298	filestorage/1c6ea98b-cbaa-4f70-8ac0-0851051aeaa2.png
95	Docker.dmg	application/x-apple-diskimage	565611371	2025-08-14 20:50:55.5778	Residential New-Construction - Master Bedroom	5119764f-a7a2-4219-967a-1c5957fb4e00.dmg	CAD	688174	filestorage/5119764f-a7a2-4219-967a-1c5957fb4e00.dmg
97	harshith_reddy_06.pdf	pdf	218558	2025-08-14 15:43:01.10038	Residential New-Construction - Master Bedroom	177b2ed3-3830-4a1a-b09e-fc7efe0cb8c7.pdf	CAD	043649	filestorage/177b2ed3-3830-4a1a-b09e-fc7efe0cb8c7.pdf
98	FormSubject.pdf	application/pdf	167694	2025-08-14 15:47:26.557455	Residential New-Construction - Bathroom	1d011f49-4aac-4843-b948-9a06bdebfd98.pdf	Other	043649	filestorage/1d011f49-4aac-4843-b948-9a06bdebfd98.pdf
101	1754039078440.docx	application/vnd.openxmlformats-officedocument.wordprocessingml.document	21774	2025-08-14 15:48:27.592213	Residential New-Construction - Bathroom	8a75bdde-23ab-4ae5-8a81-8e466063a4aa.docx	Service Call	043649	filestorage/8a75bdde-23ab-4ae5-8a81-8e466063a4aa.docx
102	Screenshot 2025-08-05 at 7.03.23 PM.png	png	41350	2025-08-14 15:56:55.176006	Residential New-Construction - Kitchen	ea98c18f-ca8d-412f-8280-84c98dc1d2e0.png	CAD	806460	filestorage/ea98c18f-ca8d-412f-8280-84c98dc1d2e0.png
103	Screenshot 2025-08-05 at 7.01.52 PM.png	image/png	75064	2025-08-14 15:58:33.840499	Residential New-Construction - Kitchen	0ffae244-f54f-4d68-baad-841c7464119b.png	Bills	806460	filestorage/0ffae244-f54f-4d68-baad-841c7464119b.png
76	Screenshot from 2025-07-31 22-21-29.png	image/png	27632	2025-08-13 07:23:52.710075	Apartment - Master Bathroom	8c6e8ee5-a9b9-47f1-a447-80bc4edcb965.png	CAD	870087	filestorage/8c6e8ee5-a9b9-47f1-a447-80bc4edcb965.png
80	Screenshot from 2025-07-31 22-21-29.png	image/png	27632	2025-08-13 15:16:49.137992	Commercial - Master Bathroom	b57dddd7-9df3-425c-b6fb-5eddcd872755.png	Template	870087	filestorage/b57dddd7-9df3-425c-b6fb-5eddcd872755.png
81	Screenshot from 2025-08-07 14-43-33.png	image/png	77655	2025-08-13 15:16:57.184623	Commercial - Master Bathroom	63913f0d-7fc8-4af0-a853-26ef3bdda2a0.png	Other	870087	filestorage/63913f0d-7fc8-4af0-a853-26ef3bdda2a0.png
75	Screenshot from 2025-08-07 14-43-33.png	image/png	77655	2025-08-13 07:23:34.56558	Residential New-Construction - Master Bathroom	f7fce048-eb79-4d56-9b00-c128538a37d4.png	CAD	870087	filestorage/f7fce048-eb79-4d56-9b00-c128538a37d4.png
104	Screenshot 2025-08-05 at 7.03.23 PM.png	image/png	41350	2025-08-14 16:06:45.430404	Residential New-Construction - Bathroom	8118a018-6cfe-4fd1-ba4a-ffcd3e00aa1c.png	Templates	496909	filestorage/8118a018-6cfe-4fd1-ba4a-ffcd3e00aa1c.png
105	Screenshot 2025-08-05 at 6.49.23 PM.png	image/png	135390	2025-08-14 16:07:36.105867	Residential New-Construction - Bathroom	c300771d-7afb-4ad3-8ce3-dcbc808f8cd4.png	Bills	496909	filestorage/c300771d-7afb-4ad3-8ce3-dcbc808f8cd4.png
106	Screenshot 2025-08-05 at 7.03.23 PM.png	image/png	41350	2025-08-14 16:08:48.286696	Residential New-Construction - Bathroom	a510de29-7c08-45a4-b918-3d00f1e1d559.png	Template	496909	filestorage/a510de29-7c08-45a4-b918-3d00f1e1d559.png
107	Screenshot 2025-08-05 at 7.01.52 PM.png	image/png	75064	2025-08-14 16:08:54.361815	Residential New-Construction - Bathroom	ef6c0c3e-0bda-4ec5-94fe-53580aa50d36.png	CAD	496909	filestorage/ef6c0c3e-0bda-4ec5-94fe-53580aa50d36.png
108	Screenshot 2025-08-05 at 6.49.23 PM.png	image/png	135390	2025-08-14 16:09:00.117807	Residential New-Construction - Bathroom	b3256426-9d42-4747-997c-9821be695076.png	Cutting	496909	filestorage/b3256426-9d42-4747-997c-9821be695076.png
109	Screenshot 2025-08-05 at 6.24.14 PM.png	image/png	63217	2025-08-14 16:09:06.295759	Residential New-Construction - Bathroom	f23e7bb0-7aa0-49f3-99f1-0589ae211798.png	Fabrication	496909	filestorage/f23e7bb0-7aa0-49f3-99f1-0589ae211798.png
110	Screenshot 2025-08-05 at 7.03.23 PM.png	image/png	41350	2025-08-14 16:09:14.519694	Residential New-Construction - Bathroom	f5502135-56ce-412b-b5d2-1f1e878de84b.png	Installation	496909	filestorage/f5502135-56ce-412b-b5d2-1f1e878de84b.png
111	Screenshot 2025-08-05 at 6.49.23 PM.png	image/png	135390	2025-08-14 16:09:14.521617	Residential New-Construction - Bathroom	de7257c5-bfe6-45bb-adea-8974d381fd80.png	Installation	496909	filestorage/de7257c5-bfe6-45bb-adea-8974d381fd80.png
112	Screenshot 2025-08-05 at 6.19.51 PM.png	image/png	12599	2025-08-14 16:09:14.523182	Residential New-Construction - Bathroom	62ff8740-c4b8-4ad9-929b-298b12889296.png	Installation	496909	filestorage/62ff8740-c4b8-4ad9-929b-298b12889296.png
113	Screenshot 2025-08-05 at 7.03.23 PM.png	image/png	41350	2025-08-14 16:09:27.687732	Residential New-Construction - Bathroom	e37258fe-aee2-4440-b9f6-13f32f1a8d8c.png	Service Call	496909	filestorage/e37258fe-aee2-4440-b9f6-13f32f1a8d8c.png
114	Screenshot 2025-08-05 at 7.01.52 PM.png	image/png	75064	2025-08-14 16:09:27.689552	Residential New-Construction - Bathroom	0d19dd27-8b7a-435b-9f11-7c3a0da6a510.png	Service Call	496909	filestorage/0d19dd27-8b7a-435b-9f11-7c3a0da6a510.png
115	Screenshot 2025-08-05 at 6.49.23 PM.png	image/png	135390	2025-08-14 16:09:27.691216	Residential New-Construction - Bathroom	a42a99ed-89cb-4479-ac48-d0f69da8978a.png	Service Call	496909	filestorage/a42a99ed-89cb-4479-ac48-d0f69da8978a.png
116	Screenshot 2025-08-05 at 6.24.14 PM.png	image/png	63217	2025-08-14 16:09:27.692563	Residential New-Construction - Bathroom	e9c9c1ba-7db2-4b03-9909-525dcdc9e5ee.png	Service Call	496909	filestorage/e9c9c1ba-7db2-4b03-9909-525dcdc9e5ee.png
117	Screenshot 2025-08-05 at 6.19.51 PM.png	image/png	12599	2025-08-14 16:09:27.693826	Residential New-Construction - Bathroom	c0a0b974-ed08-41ae-94f5-d227ff84a356.png	Service Call	496909	filestorage/c0a0b974-ed08-41ae-94f5-d227ff84a356.png
118	Screenshot 2025-08-05 at 7.03.23 PM.png	png	41350	2025-08-14 16:16:34.950099	Residential New-Construction - Bathroom	61a07afa-8a7b-4240-8352-ef7e0f250a3d.png	Template	484015	filestorage/61a07afa-8a7b-4240-8352-ef7e0f250a3d.png
119	Screenshot 2025-08-05 at 7.01.52 PM.png	png	75064	2025-08-14 16:16:34.951167	Apartment - Master Bathroom	4e12c3e8-5594-47ae-85ae-455d4f8d761d.png	Template	484015	filestorage/4e12c3e8-5594-47ae-85ae-455d4f8d761d.png
120	Screenshot 2025-08-05 at 6.49.23 PM.png	png	135390	2025-08-14 16:16:34.951778	Residential New-Construction - Bathroom	bfc64039-89be-4209-ae5f-dd13a222881d.png	CAD	484015	filestorage/bfc64039-89be-4209-ae5f-dd13a222881d.png
121	Screenshot 2025-08-05 at 6.19.51 PM.png	png	12599	2025-08-14 16:16:34.952196	Apartment - Master Bathroom	bf71ce3b-085e-430d-8e68-ad6299cb83d0.png	CAD	484015	filestorage/bf71ce3b-085e-430d-8e68-ad6299cb83d0.png
123	Screenshot 2025-08-05 at 7.01.52 PM.png	png	75064	2025-08-14 16:23:37.696586	Apartment - New Corridor	39288b8e-7817-4069-a5bc-fe6ffafaf4fd.png	Template	653535	filestorage/39288b8e-7817-4069-a5bc-fe6ffafaf4fd.png
124	Screenshot 2025-08-05 at 6.49.23 PM.png	png	135390	2025-08-14 16:23:37.697206	Residential New-Construction - Master Bedroom	28ec84de-3ed3-4b81-bcd8-a51b1687171f.png	CAD	653535	filestorage/28ec84de-3ed3-4b81-bcd8-a51b1687171f.png
125	Screenshot 2025-08-05 at 6.24.14 PM.png	png	63217	2025-08-14 16:23:37.697694	Apartment - New Corridor	385d9cef-3d05-4f24-9221-9718dfbce84b.png	CAD	653535	filestorage/385d9cef-3d05-4f24-9221-9718dfbce84b.png
126	Screenshot 2025-08-05 at 6.19.51 PM.png	png	12599	2025-08-14 16:23:37.698107	Apartment - New Corridor	95191787-d8b1-48f2-b4e9-36e9cb4bec2f.png	CAD	653535	filestorage/95191787-d8b1-48f2-b4e9-36e9cb4bec2f.png
127	Screenshot 2025-08-05 at 7.01.52 PM.png	image/png	75064	2025-08-14 16:29:29.978388	Residential New-Construction - Master Bathroom	5ab9bbaf-0d18-493c-a25d-f067489ba0df.png	Service Call	653535	filestorage/5ab9bbaf-0d18-493c-a25d-f067489ba0df.png
128	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 16:44:31.29129	Residential New-Construction - Master Bathroom	353be6db-750d-4af3-aafa-6c7dc48baac6.png	Other	043649	filestorage/353be6db-750d-4af3-aafa-6c7dc48baac6.png
129	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 16:44:31.29397	Residential New-Construction - Master Bathroom	6c7d37e0-db54-4009-a297-4d9ff4ba6d77.png	Other	043649	filestorage/6c7d37e0-db54-4009-a297-4d9ff4ba6d77.png
130	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 16:44:31.295338	Residential New-Construction - Master Bathroom	6ae91268-6be5-4db9-a0a9-e253827895dc.png	Other	043649	filestorage/6ae91268-6be5-4db9-a0a9-e253827895dc.png
131	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 16:44:40.511264	Residential New-Construction - Master Bathroom	ce8d2cd9-b2f5-4cbf-8bfd-519e53c94668.png	Cutting	043649	filestorage/ce8d2cd9-b2f5-4cbf-8bfd-519e53c94668.png
132	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 16:44:40.513395	Residential New-Construction - Master Bathroom	612e3625-f443-4a73-b7c1-d6f4cec116a6.png	Cutting	043649	filestorage/612e3625-f443-4a73-b7c1-d6f4cec116a6.png
133	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 16:44:40.51494	Residential New-Construction - Master Bathroom	96442a5e-6947-41cf-a239-178d320d6994.png	Cutting	043649	filestorage/96442a5e-6947-41cf-a239-178d320d6994.png
134	Screenshot from 2025-08-11 20-21-59.png	image/png	43581	2025-08-14 16:44:40.51615	Residential New-Construction - Master Bathroom	8590b75a-5b1f-4c8d-8add-82014e699da3.png	Cutting	043649	filestorage/8590b75a-5b1f-4c8d-8add-82014e699da3.png
135	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 16:49:22.506491	Residential New-Construction - Master Bathroom	ac4f3321-b715-4eae-a637-7efe914e8ae8.png	CAD	516289	filestorage/ac4f3321-b715-4eae-a637-7efe914e8ae8.png
136	icon.png	image/png	68123	2025-08-14 16:49:40.65627	Residential New-Construction - Master Bathroom	0662cabf-6216-40e5-bc45-3dcabee84933.png	CAD	516289	filestorage/0662cabf-6216-40e5-bc45-3dcabee84933.png
137	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 16:53:16.292165	Residential New-Construction - Master Bathroom	91b2a42f-094d-4a08-8657-d5de7cc3014a.png	Templates	516289	filestorage/91b2a42f-094d-4a08-8657-d5de7cc3014a.png
138	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 16:53:16.295098	Residential New-Construction - Master Bathroom	32866674-fca6-4707-8a41-55a9644d2861.png	Templates	516289	filestorage/32866674-fca6-4707-8a41-55a9644d2861.png
139	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 16:53:16.296679	Residential New-Construction - Master Bathroom	84089c17-4099-409a-8061-d8780ac548cb.png	Templates	516289	filestorage/84089c17-4099-409a-8061-d8780ac548cb.png
141	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 16:53:54.874516	Apartment - Other	ce88105a-770d-4366-bd6b-6a7ec5fd6e55.png	CAD	653535	filestorage/ce88105a-770d-4366-bd6b-6a7ec5fd6e55.png
142	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 16:59:16.741184	Residential New-Construction - Master Bathroom	b38a6fbf-9e19-4af1-bf64-de47b980f5bc.png	CAD	334790	filestorage/b38a6fbf-9e19-4af1-bf64-de47b980f5bc.png
143	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 16:59:16.743848	Residential New-Construction - Master Bathroom	45ec87c2-969d-4748-a438-f9f5ffead8c7.png	CAD	334790	filestorage/45ec87c2-969d-4748-a438-f9f5ffead8c7.png
144	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 16:59:16.745619	Residential New-Construction - Master Bathroom	9c5c70df-6a35-41b1-b62c-62cf52894ad7.png	CAD	334790	filestorage/9c5c70df-6a35-41b1-b62c-62cf52894ad7.png
145	Screenshot from 2025-07-29 23-55-45.png	png	183	2025-08-14 17:07:19.917455	Residential Remodel - Kitchen	8fb3bb15-70f2-4ace-86a6-f7d0594a6fb4.png	CAD	690103	filestorage/8fb3bb15-70f2-4ace-86a6-f7d0594a6fb4.png
146	Screenshot from 2025-07-31 11-03-38.png	png	4939	2025-08-14 17:07:19.918328	Residential Remodel - Kitchen	a0c308d0-10d6-40a7-b18e-c3222c0c33c4.png	CAD	690103	filestorage/a0c308d0-10d6-40a7-b18e-c3222c0c33c4.png
147	Screenshot from 2025-07-30 20-23-43.png	png	10498	2025-08-14 17:07:19.918736	Residential Remodel - Kitchen	86819483-2a62-4461-a3e1-1ccdbbf626fb.png	CAD	690103	filestorage/86819483-2a62-4461-a3e1-1ccdbbf626fb.png
148	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 17:08:00.530832	Residential Remodel - Kitchen	5a9e15ad-3a0b-4e7d-8f3f-4cdf2ecd8343.png	Templates	690103	filestorage/5a9e15ad-3a0b-4e7d-8f3f-4cdf2ecd8343.png
149	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 17:08:00.533454	Residential Remodel - Kitchen	f9deadf3-88bb-48c8-a6b2-00ba70d73f64.png	Templates	690103	filestorage/f9deadf3-88bb-48c8-a6b2-00ba70d73f64.png
150	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 17:08:00.535042	Residential Remodel - Kitchen	54842ef2-4148-40e6-ab75-054f58696b40.png	Templates	690103	filestorage/54842ef2-4148-40e6-ab75-054f58696b40.png
151	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 17:08:09.616422	Residential Remodel - Kitchen	d6749908-dd5f-4d4d-8f24-42279d9ecfcd.png	Bills	690103	filestorage/d6749908-dd5f-4d4d-8f24-42279d9ecfcd.png
152	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 17:08:09.61845	Residential Remodel - Kitchen	2122eb52-a914-4a10-8ede-b38c38e5724f.png	Bills	690103	filestorage/2122eb52-a914-4a10-8ede-b38c38e5724f.png
153	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 17:08:09.62	Residential Remodel - Kitchen	62bdb676-def1-4824-9f0b-d3883937074c.png	Bills	690103	filestorage/62bdb676-def1-4824-9f0b-d3883937074c.png
154	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 22:40:57.902123	Residential Remodel - Kitchen	b67f09d2-15a2-46fc-94ef-961211e57d7b.png	Bills	690103	filestorage/b67f09d2-15a2-46fc-94ef-961211e57d7b.png
155	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 22:40:58.436498	Residential Remodel - Kitchen	dd635298-fe22-49ee-8ace-8c7acb443dd9.png	Bills	690103	filestorage/dd635298-fe22-49ee-8ace-8c7acb443dd9.png
156	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 22:40:58.696445	Residential Remodel - Kitchen	3a799fe9-562c-4c63-b55e-7c08e92e3eff.png	Bills	690103	filestorage/3a799fe9-562c-4c63-b55e-7c08e92e3eff.png
157	Screenshot from 2025-08-01 05-38-10.png	image/png	1122979	2025-08-14 22:42:03.981862	Residential Remodel - Kitchen	f5bea032-bb85-4f5b-a8ba-40a327d7c361.png	Bills	690103	filestorage/f5bea032-bb85-4f5b-a8ba-40a327d7c361.png
158	Screenshot from 2025-08-11 20-21-59.png	image/png	43581	2025-08-14 22:42:58.412757	Residential Remodel - Kitchen	ab0d13ab-b717-4c2c-b5ec-fa127896857c.png	Bills	690103	filestorage/ab0d13ab-b717-4c2c-b5ec-fa127896857c.png
159	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 22:44:04.264006	Residential Remodel - Kitchen	6dbe878d-8576-451a-b398-74e0dbc4747d.png	Bills	690103	filestorage/6dbe878d-8576-451a-b398-74e0dbc4747d.png
160	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 22:44:04.555057	Residential Remodel - Kitchen	8013b8ee-fa01-4e8a-81bd-4b4d4ffaf7c9.png	Bills	690103	filestorage/8013b8ee-fa01-4e8a-81bd-4b4d4ffaf7c9.png
161	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 22:44:04.661336	Residential Remodel - Kitchen	ac1e7d1d-cc31-40c8-a71f-70a3382b087f.png	Bills	690103	filestorage/ac1e7d1d-cc31-40c8-a71f-70a3382b087f.png
162	Screenshot from 2025-08-11 20-21-59.png	image/png	43581	2025-08-14 22:44:04.770544	Residential Remodel - Kitchen	3474b951-c7e8-402f-ab36-da049ecec0d5.png	Bills	690103	filestorage/3474b951-c7e8-402f-ab36-da049ecec0d5.png
163	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 22:45:55.332328	Residential Remodel - Kitchen	8bfdb1ab-fb3d-4461-b08b-fe058b736d08.png	Template	690103	filestorage/8bfdb1ab-fb3d-4461-b08b-fe058b736d08.png
164	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 22:45:55.55112	Residential Remodel - Kitchen	10139395-b7ec-44e0-a2fe-7517426a3b92.png	Template	690103	filestorage/10139395-b7ec-44e0-a2fe-7517426a3b92.png
165	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 22:45:55.674138	Residential Remodel - Kitchen	6c0b35a7-1d29-4b45-914e-2f0f4e36752c.png	Template	690103	filestorage/6c0b35a7-1d29-4b45-914e-2f0f4e36752c.png
166	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 22:48:43.612484	Residential Remodel - Kitchen	ebf86a70-8596-4fe9-ad3d-0fae1b63562c.png	Cutting	690103	filestorage/ebf86a70-8596-4fe9-ad3d-0fae1b63562c.png
167	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 22:50:30.366274	Residential Remodel - Kitchen	1be4d056-eb12-484f-a814-e5542d88a6b7.png	CAD	690103	filestorage/1be4d056-eb12-484f-a814-e5542d88a6b7.png
168	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 22:54:46.611348	Residential New-Construction - Master Bathroom	e0d47b50-7ef9-41ff-bbf8-bbaef585516e.png	Installation	043649	filestorage/e0d47b50-7ef9-41ff-bbf8-bbaef585516e.png
169	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 22:54:46.872486	Residential New-Construction - Master Bathroom	672ff193-a014-4402-a717-b9043f3abe26.png	Installation	043649	filestorage/672ff193-a014-4402-a717-b9043f3abe26.png
170	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 22:54:47.008898	Residential New-Construction - Master Bathroom	69f202b8-dde8-461f-94db-9575ebea53cb.png	Installation	043649	filestorage/69f202b8-dde8-461f-94db-9575ebea53cb.png
171	Screenshot from 2025-08-11 20-21-59.png	image/png	43581	2025-08-14 22:54:47.155166	Residential New-Construction - Master Bathroom	b3047150-fdd4-44a2-8991-4f3a54fe7470.png	Installation	043649	filestorage/b3047150-fdd4-44a2-8991-4f3a54fe7470.png
172	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 22:57:15.268748	Residential Remodel - Kitchen	7a198b41-46d7-41e4-9c96-b240ffb38271.png	Template	690103	filestorage/7a198b41-46d7-41e4-9c96-b240ffb38271.png
173	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 22:57:15.434597	Residential Remodel - Kitchen	df3b1926-104d-4a4a-a02d-3ee7b8045dfb.png	Template	690103	filestorage/df3b1926-104d-4a4a-a02d-3ee7b8045dfb.png
174	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 22:57:15.557192	Residential Remodel - Kitchen	d0424131-278e-4907-8172-9c5de1709895.png	Template	690103	filestorage/d0424131-278e-4907-8172-9c5de1709895.png
175	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 22:58:29.328471	Residential Remodel - Kitchen	b0c9d325-0418-43a2-827e-1e67ca51764b.png	Cutting	690103	filestorage/b0c9d325-0418-43a2-827e-1e67ca51764b.png
176	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 22:58:29.782512	Residential Remodel - Kitchen	ca44174c-ed56-4cc7-8e10-13a36e5c29db.png	Cutting	690103	filestorage/ca44174c-ed56-4cc7-8e10-13a36e5c29db.png
177	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 22:58:29.938647	Residential Remodel - Kitchen	654798a9-6f68-434f-8c69-794a6db1f958.png	Cutting	690103	filestorage/654798a9-6f68-434f-8c69-794a6db1f958.png
178	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 23:00:25.914518	Residential Remodel - Kitchen	4fdb7462-648b-487e-bd3b-0d07778d3824.png	Installation	690103	filestorage/4fdb7462-648b-487e-bd3b-0d07778d3824.png
179	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 23:00:26.059041	Residential Remodel - Kitchen	a4462e9f-c88a-4f08-ad35-acfd1d9388f8.png	Installation	690103	filestorage/a4462e9f-c88a-4f08-ad35-acfd1d9388f8.png
180	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 23:00:26.160735	Residential Remodel - Kitchen	78d8a554-056b-4935-b3d2-1515148dd5f9.png	Installation	690103	filestorage/78d8a554-056b-4935-b3d2-1515148dd5f9.png
181	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 23:01:21.437037	Residential New-Construction - Master Bathroom	2a2052bb-48a3-452f-90cb-7ec9c04d0519.png	CAD	043649	filestorage/2a2052bb-48a3-452f-90cb-7ec9c04d0519.png
182	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 23:01:21.658293	Residential New-Construction - Master Bathroom	94d020bc-ef9d-4e1b-8a1b-f75640e8c152.png	CAD	043649	filestorage/94d020bc-ef9d-4e1b-8a1b-f75640e8c152.png
256	image.jpg	image/jpeg	90567	2025-08-22 10:18:26.366943	Commercial - Master Bathroom	ddd887c7-4f25-4c4d-aa48-10593d2198e5.jpg	CAD	235859	filestorage/ddd887c7-4f25-4c4d-aa48-10593d2198e5.jpg
183	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 23:01:22.128506	Residential New-Construction - Master Bathroom	774f0b12-3620-4931-b8a5-95e15f1c8dfa.png	CAD	043649	filestorage/774f0b12-3620-4931-b8a5-95e15f1c8dfa.png
184	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 23:02:34.147298	Residential New-Construction - Master Bedroom	63520a88-7c06-41e9-a78b-f1c97a1bdf9f.png	Bills	043649	filestorage/63520a88-7c06-41e9-a78b-f1c97a1bdf9f.png
185	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 23:02:34.318544	Residential New-Construction - Master Bedroom	3366c209-eca2-458e-a309-edb8114b51b1.png	Bills	043649	filestorage/3366c209-eca2-458e-a309-edb8114b51b1.png
186	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 23:02:34.472923	Residential New-Construction - Master Bedroom	71756c86-54fb-4dc9-b1c3-da74619891b9.png	Bills	043649	filestorage/71756c86-54fb-4dc9-b1c3-da74619891b9.png
187	Screenshot from 2025-08-11 20-21-59.png	image/png	43581	2025-08-14 23:02:34.616116	Residential New-Construction - Master Bedroom	16191f42-eb10-43f3-bd92-ada147980264.png	Bills	043649	filestorage/16191f42-eb10-43f3-bd92-ada147980264.png
188	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 17:35:21.171268	Residential New-Construction - Master Bedroom	8129bebe-bf2e-41d5-97e2-b2ccfe0366b6.png	CAD	688174	filestorage/8129bebe-bf2e-41d5-97e2-b2ccfe0366b6.png
189	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 17:35:21.173937	Residential New-Construction - Master Bedroom	997ca43e-baec-46bb-ad41-2091be17e1e2.png	CAD	688174	filestorage/997ca43e-baec-46bb-ad41-2091be17e1e2.png
190	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 17:35:21.175306	Residential New-Construction - Master Bedroom	bbb32294-c844-4bef-93b7-f8924e927f03.png	CAD	688174	filestorage/bbb32294-c844-4bef-93b7-f8924e927f03.png
191	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 23:06:23.629866	Residential New-Construction - Bathroom	638032b7-d912-405f-9d02-eedf6c4cd46a.png	CAD	043649	filestorage/638032b7-d912-405f-9d02-eedf6c4cd46a.png
192	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 23:06:23.859769	Residential New-Construction - Bathroom	fb1968bc-4bff-4deb-a506-6e9cb2faf924.png	CAD	043649	filestorage/fb1968bc-4bff-4deb-a506-6e9cb2faf924.png
193	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 23:06:23.986074	Residential New-Construction - Bathroom	4abc19c7-12dd-4eca-9512-19d1631072c5.png	CAD	043649	filestorage/4abc19c7-12dd-4eca-9512-19d1631072c5.png
194	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 23:11:39.719494	Residential New-Construction - Master Bathroom	28ae7d29-cbae-46e7-880b-7058731e75d0.png	Template	043649	filestorage/28ae7d29-cbae-46e7-880b-7058731e75d0.png
195	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 23:11:39.95895	Residential New-Construction - Master Bathroom	dedc4c98-9f17-4b7b-bd79-9efaedc4966f.png	Template	043649	filestorage/dedc4c98-9f17-4b7b-bd79-9efaedc4966f.png
196	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 23:11:40.086697	Residential New-Construction - Master Bathroom	c2927934-4d66-45d2-afa8-ecab58c34f32.png	Template	043649	filestorage/c2927934-4d66-45d2-afa8-ecab58c34f32.png
197	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 23:14:50.507107	Residential New-Construction - Bathroom	ed735c61-e2e0-47cd-8d63-eb52538f004f.png	Cutting	043649	filestorage/ed735c61-e2e0-47cd-8d63-eb52538f004f.png
198	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 23:14:50.669367	Residential New-Construction - Bathroom	b5d67eb2-ae63-44d0-a2f5-47d59735fb81.png	Cutting	043649	filestorage/b5d67eb2-ae63-44d0-a2f5-47d59735fb81.png
199	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 23:14:50.796439	Residential New-Construction - Bathroom	b667e828-388e-4659-b265-58c67976e671.png	Cutting	043649	filestorage/b667e828-388e-4659-b265-58c67976e671.png
200	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 23:17:06.096077	Residential Remodel - Kitchen	65f8286a-76dd-4159-875e-1e6fdc9b11f6.png	Template	690103	filestorage/65f8286a-76dd-4159-875e-1e6fdc9b11f6.png
201	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 23:17:06.234745	Residential Remodel - Kitchen	b06ed8a7-9b9f-43b8-b85b-ea69af02c01a.png	Template	690103	filestorage/b06ed8a7-9b9f-43b8-b85b-ea69af02c01a.png
202	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 23:17:06.365678	Residential Remodel - Kitchen	4a263b21-1ba9-49ac-8108-ecff44d8f14c.png	Template	690103	filestorage/4a263b21-1ba9-49ac-8108-ecff44d8f14c.png
203	Screenshot from 2025-08-11 20-21-59.png	image/png	43581	2025-08-14 23:17:06.505967	Residential Remodel - Kitchen	98afd55b-ec13-42a4-8ea8-a51ec808dd75.png	Template	690103	filestorage/98afd55b-ec13-42a4-8ea8-a51ec808dd75.png
204	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 23:17:35.403296	Residential Remodel - Kitchen	8e1bf29e-65e2-4af3-adb4-a7b940047e34.png	Template	690103	filestorage/8e1bf29e-65e2-4af3-adb4-a7b940047e34.png
205	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 23:17:35.595448	Residential Remodel - Kitchen	e96c7662-147c-4ad0-82bd-919354aea6c2.png	Template	690103	filestorage/e96c7662-147c-4ad0-82bd-919354aea6c2.png
206	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 23:17:35.710911	Residential Remodel - Kitchen	9a4092c0-d28e-4d60-821f-24fa4c056db0.png	Template	690103	filestorage/9a4092c0-d28e-4d60-821f-24fa4c056db0.png
207	Screenshot from 2025-08-11 20-21-59.png	image/png	43581	2025-08-14 23:17:35.825613	Residential Remodel - Kitchen	de715bf8-0bda-4ee8-9fab-a9241b167e4f.png	Template	690103	filestorage/de715bf8-0bda-4ee8-9fab-a9241b167e4f.png
208	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 23:17:59.393516	Residential Remodel - Kitchen	b7337277-e818-49f2-99f1-571078c05a2b.png	Template	690103	filestorage/b7337277-e818-49f2-99f1-571078c05a2b.png
209	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 23:18:09.667929	Residential Remodel - Kitchen	400b5b8d-8925-49ff-b623-5c28399d6bf0.png	Cutting	690103	filestorage/400b5b8d-8925-49ff-b623-5c28399d6bf0.png
210	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 23:18:09.84341	Residential Remodel - Kitchen	e0395773-b481-434f-b474-6ee4f95f8e23.png	Cutting	690103	filestorage/e0395773-b481-434f-b474-6ee4f95f8e23.png
211	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 23:18:09.95574	Residential Remodel - Kitchen	d0bde577-e97b-47fc-862d-fce863f73660.png	Cutting	690103	filestorage/d0bde577-e97b-47fc-862d-fce863f73660.png
212	Screenshot from 2025-08-11 20-21-59.png	image/png	43581	2025-08-14 23:18:26.182436	Residential Remodel - Kitchen	f639cef3-7a22-4f82-b666-a0fd746c5fb0.png	Cutting	690103	filestorage/f639cef3-7a22-4f82-b666-a0fd746c5fb0.png
213	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 23:18:30.771437	Residential Remodel - Kitchen	66bab47a-ac00-47bc-8471-6ed11b5a3356.png	Cutting	690103	filestorage/66bab47a-ac00-47bc-8471-6ed11b5a3356.png
214	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 23:18:50.471639	Residential Remodel - Kitchen	b488331e-f9ee-45d4-8da9-aad4b950fc72.png	Cutting	690103	filestorage/b488331e-f9ee-45d4-8da9-aad4b950fc72.png
215	WhatsApp Image 2025-07-05 at 02.42.11_89f1f265.jpg	image/jpeg	40161	2025-08-14 17:53:03.670733	Residential New-Construction - Master Bathroom	395cf759-be0e-4625-8b28-7255ab9d0d59.jpg	Service Call	043649	filestorage/395cf759-be0e-4625-8b28-7255ab9d0d59.jpg
216	WhatsApp Image 2025-07-05 at 02.42.11_89f1f265.jpg	image/jpeg	40161	2025-08-14 17:55:17.731624	Residential New-Construction - Master Bathroom	1f36b587-3095-4077-acf8-08a5a80dbc11.jpg	Fabrication	043649	filestorage/1f36b587-3095-4077-acf8-08a5a80dbc11.jpg
217	1754039078440.docx	application/vnd.openxmlformats-officedocument.wordprocessingml.document	21774	2025-08-14 18:03:56.860768	Residential Remodel - Kitchen	39dda46c-b027-4a07-81ef-6f00ca4247c4.docx	Template	690103	filestorage/39dda46c-b027-4a07-81ef-6f00ca4247c4.docx
218	Screenshot from 2025-07-29 23-55-45.png	image/png	183	2025-08-14 18:06:05.751277	Residential Remodel - Kitchen	41f80e7a-7f70-44c8-be51-9c541ab1a7ca.png	Cutting	690103	filestorage/41f80e7a-7f70-44c8-be51-9c541ab1a7ca.png
219	Screenshot from 2025-07-31 11-03-38.png	image/png	4939	2025-08-14 18:06:05.753742	Residential Remodel - Kitchen	687f68bc-c507-4aad-b036-72bf8122b0b9.png	Cutting	690103	filestorage/687f68bc-c507-4aad-b036-72bf8122b0b9.png
220	Screenshot from 2025-07-30 20-23-43.png	image/png	10498	2025-08-14 18:06:05.755113	Residential Remodel - Kitchen	767dc399-a341-4803-9548-f7acaaf1ea1e.png	Cutting	690103	filestorage/767dc399-a341-4803-9548-f7acaaf1ea1e.png
221	Screenshot 2025-08-05 at 7.03.23 PM.png	png	41350	2025-08-14 18:24:51.682487	Residential Remodel - Bathroom	e1a5e2ec-efca-4fdd-8279-570bd2f323ff.png	Template	076315	filestorage/e1a5e2ec-efca-4fdd-8279-570bd2f323ff.png
222	Screenshot 2025-08-05 at 7.01.52 PM.png	png	75064	2025-08-14 18:24:51.683707	Residential Remodel - Bathroom	535692d9-5f1e-46fe-accb-35f1c9594983.png	CAD	076315	filestorage/535692d9-5f1e-46fe-accb-35f1c9594983.png
223	Screenshot 2025-08-05 at 6.49.23 PM.png	png	135390	2025-08-14 18:24:51.684255	Residential Remodel - Bathroom	7c3cdb0d-283e-4b1c-9348-49bed5dd95c2.png	Cutting	076315	filestorage/7c3cdb0d-283e-4b1c-9348-49bed5dd95c2.png
224	Screenshot 2025-08-05 at 6.19.51 PM.png	image/png	12599	2025-08-14 18:26:53.12285	Residential Remodel - Bathroom	b0998ae7-fa57-4571-999d-daec51692d8b.png	Template	076315	filestorage/b0998ae7-fa57-4571-999d-daec51692d8b.png
225	Screenshot 2024-09-06 at 6.33.00 PM.png	image/png	339716	2025-08-14 18:27:09.662226	Residential Remodel - Bathroom	2dc02de7-00f2-4466-b96c-f06b4fa2b2be.png	CAD	076315	filestorage/2dc02de7-00f2-4466-b96c-f06b4fa2b2be.png
226	Screenshot 2024-09-06 at 6.33.00 PM.png	image/png	339716	2025-08-14 18:27:29.539117	Residential Remodel - Bathroom	19dc7937-0c50-4e4e-8880-902f775b7c69.png	Installation	076315	filestorage/19dc7937-0c50-4e4e-8880-902f775b7c69.png
227	Screenshot 2025-08-05 at 7.03.23 PM.png	png	41350	2025-08-14 21:14:53.935094	Residential New-Construction - Bathroom	fb07f415-e7a9-4893-b974-0858683a0436.png	Service	598530	filestorage/fb07f415-e7a9-4893-b974-0858683a0436.png
79	Screenshot from 2025-07-31 11-44-59.png	image/png	94461	2025-08-13 15:16:21.652979	Commercial - Master Bathroom	fb254811-6465-4ec4-8402-ef99e0f4e285.png	Template	870087	filestorage/fb254811-6465-4ec4-8402-ef99e0f4e285.png
61	Screenshot from 2025-08-11 20-21-59.png	image/png	43581	2025-08-13 12:07:04.194642	Residential Remodel - Bathroom	c15eab25-6259-47c8-bdb4-5bc62c2d2926.png	Template	870087	filestorage/c15eab25-6259-47c8-bdb4-5bc62c2d2926.png
237	assignment_track.pdf	application/pdf	111541	2025-08-20 17:44:08.688856	Commercial - Master Bathroom	bbd713ee-f260-49a3-b462-91cc59778973.pdf	CAD	870087	filestorage/bbd713ee-f260-49a3-b462-91cc59778973.pdf
238	Swarnalatha_Resume.pdf	application/pdf	447751	2025-08-20 17:44:56.037455	Commercial - Master Bathroom	58e25ac8-919d-43b1-b00e-4dfbbc1bb2ad.pdf	CAD	870087	filestorage/58e25ac8-919d-43b1-b00e-4dfbbc1bb2ad.pdf
240	Screenshot from 2025-08-07 13-15-49.png	png	41230	2025-08-21 09:46:46.332595		64b9f7f7-02d9-42b4-a945-f2856ffa01b7.png	Profile	660420	filestorage/64b9f7f7-02d9-42b4-a945-f2856ffa01b7.png
241	Screenshot from 2025-07-31 22-21-29.png	png	27632	2025-08-21 09:46:46.333764	Other	d3bbb842-86cd-4a21-828e-2d0a326865d5.png	Other	660420	filestorage/d3bbb842-86cd-4a21-828e-2d0a326865d5.png
246	Screenshot 2025-08-05 at 6.24.14 PM.png	png	63217	2025-08-22 03:11:25.690212	Commercial - Master Bathroom	da8fb0c3-25b5-4926-b32c-6fa780d520f0.png	Template	652375	filestorage/da8fb0c3-25b5-4926-b32c-6fa780d520f0.png
247	Screenshot 2025-08-05 at 7.03.23 PM.png	image/png	41350	2025-08-22 03:15:54.98403	Commercial - Master Bathroom	3994d558-cb25-4d3c-8ad9-53a4264c0764.png	Other	652375	filestorage/3994d558-cb25-4d3c-8ad9-53a4264c0764.png
248	Screenshot from 2025-07-31 22-21-29.png	image/png	27632	2025-08-22 05:06:42.081879	Residential New-Construction - Master Bedroom	2392fc63-8d56-4b8e-89bf-cfce743c7bc8.png	Other	673065	filestorage/2392fc63-8d56-4b8e-89bf-cfce743c7bc8.png
249	Screenshot from 2025-08-07 14-45-18.png	image/png	77655	2025-08-22 05:08:48.317651	Residential New-Construction - Master Bedroom	58c0d31d-dd10-4255-ac77-4201ef8d5fb3.png	Other	673065	filestorage/58c0d31d-dd10-4255-ac77-4201ef8d5fb3.png
242	Screenshot 2025-08-05 at 7.01.52 PM.png	png	75064	2025-08-22 02:42:23.90716	Residential New-Construction - Bathroom	baa02ecc-532d-43a1-81b2-9d6cb08e98de.png	Other	888866	filestorage/baa02ecc-532d-43a1-81b2-9d6cb08e98de.png
243	Screenshot 2025-08-05 at 6.24.14 PM.png	image/png	63217	2025-08-22 02:47:45.883272	Residential New-Construction - Bathroom	d040618b-3bcd-4848-8673-8d8b9e59fa76.png	Other	888866	filestorage/d040618b-3bcd-4848-8673-8d8b9e59fa76.png
244	Screenshot 2025-08-05 at 6.49.23 PM.png	image/png	135390	2025-08-22 02:48:18.723863	Residential New-Construction - Bathroom	73a88ff3-d1c0-49f2-a69b-4a2dbae8c96a.png	Other	888866	filestorage/73a88ff3-d1c0-49f2-a69b-4a2dbae8c96a.png
245	Screenshot 2025-08-05 at 6.24.14 PM.png	image/png	63217	2025-08-22 02:51:53.59194	Residential New-Construction - Bathroom	dedddaac-07fe-40db-b02e-31c023649735.png	Other	888866	filestorage/dedddaac-07fe-40db-b02e-31c023649735.png
250	Screenshot from 2025-08-21 19-17-11.png	image/png	10127	2025-08-22 05:11:48.017872	Residential New-Construction - Master Bedroom	cc32b897-4899-4358-997f-cf879db4b9ca.png	Other	673065	filestorage/cc32b897-4899-4358-997f-cf879db4b9ca.png
251	previewjpg	application/octet-stream	4009464	2025-08-22 05:25:47.517772	Commercial - Master Bathroom	c04c5199-a479-4fae-8ea3-0e6420cc8153	Other	652375	filestorage/c04c5199-a479-4fae-8ea3-0e6420cc8153
252	Screenshot from 2025-07-31 22-21-29.png	png	27632	2025-08-22 14:16:05.577615	Other	7da10406-3aeb-4af2-b5cf-ef5d8f235564.png	Other	580823	filestorage/7da10406-3aeb-4af2-b5cf-ef5d8f235564.png
253	Screenshot from 2025-07-31 22-21-29.png	png	27632	2025-08-22 09:43:35.309963	Other	f1e41bb0-f3af-4d67-9ffc-6acbf5678c67.png	Other	235859	filestorage/f1e41bb0-f3af-4d67-9ffc-6acbf5678c67.png
254	image.jpg	image/jpeg	90567	2025-08-22 10:15:55.395712	Residential Remodel - Bathroom	4cb7473b-6243-49be-95c7-ce193d845725.jpg	Other	605988	filestorage/4cb7473b-6243-49be-95c7-ce193d845725.jpg
255	image.jpg	image/jpeg	90567	2025-08-22 10:17:36.709703	Residential Remodel - Kitchen	ab5a22ab-b44d-4b4c-b2f5-0db5205d63dc.jpg	Template	142588	filestorage/ab5a22ab-b44d-4b4c-b2f5-0db5205d63dc.jpg
257	Screenshot from 2025-07-31 22-21-29.png	png	27632	2025-08-22 16:37:02.888284	Other	f42f2e0a-7f9e-4370-91ea-2b5edbec94c1.png	Other	692681	filestorage/f42f2e0a-7f9e-4370-91ea-2b5edbec94c1.png
\.





--
-- Data for Name: lead_associates; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.lead_associates (lead_id, associate_customer_id) FROM stdin;
153714	918287
379122	898055
334790	787375
764542	326700
788865	013482
573182	496853
022652	960451
304586	150019
301498	825054
973844	581012
078083	918814
884011	804615
554098	557090
701868	105420
858123	067404
467243	120226
253703	604561
688174	491637
870087	704761
535395	624754
616587	646089
516289	056713
043649	120480
076315	947079
822302	411972
660420	819745
914196	301041
888866	776072
629876	344611
556657	778225
652375	791659
225983	952419
580823	589276
235859	373889
\.


--
-- Data for Name: leads; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.leads (id, created_at, updated_at, deleted_at, lead_type, job_name, delivery_type, status, details, referrer, detailed_lead_status, payment_term, pricing_type, favorite, profile_pic, sales1_person_id, sales2_person_id, rejection_reason) FROM stdin;
278004	2025-08-11 15:24:03.906235+00	2025-08-11 15:49:14.556982+00	\N	WALK_IN	manohar	CUSTOMER_PICKUP	COMPLETED	{"Notes": "", "Phone": "9087654321", "TaxID": null, "Discount": {"ID": 1, "Name": "flash", "DiscountPercentage": 10}, "CountyTax": {"ID": 1, "Name": "Jackson", "TaxPercentage": 7.2}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "m@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 6, "IsPickUpDelivery": true, "ShippingAddressID": 5, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	NEW			659808	659808	
153714	2025-08-11 15:13:04.705662+00	2025-08-11 15:14:22.382219+00	\N	WALK_IN	durga	CUSTOMER_PICKUP	COMPLETED	{"Notes": "", "Phone": "1234567890", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 1, "Name": "Jackson", "TaxPercentage": 7.2}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 2, "IsPickUpDelivery": true, "ShippingAddressID": 1, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	NEW		a1a7921a-ecb2-4fcd-82da-428d9dea0bdf.png	659808	659808	
338697	2025-08-11 15:16:40.419402+00	2025-08-11 15:23:02.944768+00	\N	WALK_IN	sai nsg	CONTRACTOR_PICKUP	COMPLETED	{"Notes": "", "Phone": "1234567890", "TaxID": null, "Discount": {"ID": 1, "Name": "flash", "DiscountPercentage": 10}, "CountyTax": {"ID": 1, "Name": "Jackson", "TaxPercentage": 7.2}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 4, "IsPickUpDelivery": true, "ShippingAddressID": 3, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	NEW			659808	659808	
990977	2025-08-11 17:08:30.719943+00	2025-08-11 17:15:57.633843+00	\N	WALK_IN	nsg	CUSTOMER_PICKUP	NEW	{"Notes": "", "Phone": "1234567890", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 1, "Name": "Jackson", "TaxPercentage": 7.2}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 14, "IsPickUpDelivery": true, "ShippingAddressID": 13, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	REGULAR			910454	910454	
379122	2025-08-11 16:42:48.999023+00	2025-08-11 16:45:52.647391+00	\N	WALK_IN	aditya's leads	CUSTOMER_PICKUP	COMPLETED	{"Notes": "", "Phone": "1234567890", "TaxID": null, "Discount": {"ID": 1, "Name": "flash", "DiscountPercentage": 10}, "CountyTax": {"ID": 1, "Name": "Jackson", "TaxPercentage": 7.2}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 8, "IsPickUpDelivery": true, "ShippingAddressID": 7, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	NEW			659808	910454	
788865	2025-08-11 17:45:54.835994+00	2025-08-11 18:20:57.882175+00	\N	WEBSITE	Adhi	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "9998887771", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 7, "Name": "maincounty", "TaxPercentage": 7.7}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "adhi@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 20, "IsPickUpDelivery": false, "ShippingAddressID": 19, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CHEQUE	PRO			910454	659808	
087879	2025-08-11 17:33:35.831761+00	2025-08-11 17:37:11.52104+00	\N	WALK_IN	nsgg	CUSTOMER_PICKUP	NEW	{"Notes": "", "Phone": "1234588324", "TaxID": null, "Discount": {"ID": 1, "Name": "flash", "DiscountPercentage": 10}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 16, "IsPickUpDelivery": true, "ShippingAddressID": 15, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 2}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	NEW			659808	659808	
334790	2025-08-11 16:57:11.588707+00	2025-08-14 04:37:41.72712+00	\N	WEBSITE	Pradeep	INSTALLATION	NEW	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 10, "IsPickUpDelivery": false, "ShippingAddressID": 9, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_NOT_SENT	CHEQUE	BEST		0713ac23-6968-4a2b-ae47-37ed483228b0.png	910454	659808	
764542	2025-08-11 17:04:28.403934+00	2025-08-11 18:27:17.112694+00	\N	WEBSITE	Pradeep_yandamuri	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 12, "IsPickUpDelivery": false, "ShippingAddressID": 11, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	PRO		4396524e-5160-41e4-9690-6484dcc5a718.png	910454	659808	
198032	2025-08-11 19:23:33.097159+00	2025-08-11 19:23:33.424961+00	\N		gowtham		NEW	{"Notes": "", "Phone": "1287628761", "TaxID": null, "Discount": null, "CountyTax": null, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gm.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": null, "IsPickUpDelivery": false, "ShippingAddressID": null, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT							
893916	2025-08-11 17:44:06.856279+00	2025-08-11 17:52:10.864159+00	\N	NORMAL	nsggg	SHIPPING	COMPLETED	{"Notes": "", "Phone": "1234352345", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 1, "Name": "Jackson", "TaxPercentage": 7.2}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@c.om", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "sadfasdf", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 22, "IsPickUpDelivery": false, "ShippingAddressID": 21, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	NEW			910454	659808	
028735	2025-08-11 19:10:05.609607+00	2025-08-11 19:10:06.009562+00	\N		sai nsg		NEW	{"Notes": "", "Phone": "1234567890", "TaxID": null, "Discount": null, "CountyTax": null, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": null, "IsPickUpDelivery": false, "ShippingAddressID": null, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT							
945419	2025-08-11 19:12:52.367997+00	2025-08-11 19:12:52.728988+00	\N		bunny		NEW	{"Notes": "", "Phone": "1262934729", "TaxID": null, "Discount": null, "CountyTax": null, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "a@g.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": null, "IsPickUpDelivery": false, "ShippingAddressID": null, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT							
512161	2025-08-11 19:21:21.353483+00	2025-08-11 19:25:48.585743+00	\N	WALK_IN	kumar	INSTALLATION	NEW	{"Notes": "", "Phone": "9998887771", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "kumar@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 24, "IsPickUpDelivery": false, "ShippingAddressID": 23, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	PRO			910454	659808	
971784	2025-08-11 20:13:33.84842+00	2025-08-11 20:13:34.129901+00	\N		queen		NEW	{"Notes": "", "Phone": "1276128971", "TaxID": null, "Discount": null, "CountyTax": null, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "sd@#l.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": null, "IsPickUpDelivery": false, "ShippingAddressID": null, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT							
966384	2025-08-11 20:07:12.971077+00	2025-08-11 20:07:13.32549+00	\N		finisher		NEW	{"Notes": "", "Phone": "1234567890", "TaxID": null, "Discount": null, "CountyTax": null, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": null, "IsPickUpDelivery": false, "ShippingAddressID": null, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT							
553579	2025-08-11 20:10:15.36042+00	2025-08-11 20:10:15.677404+00	\N		king		NEW	{"Notes": "", "Phone": "3791782980", "TaxID": null, "Discount": null, "CountyTax": null, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@#mn.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": null, "IsPickUpDelivery": false, "ShippingAddressID": null, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT							
756697	2025-08-11 19:16:41.167155+00	2025-08-20 10:34:44.170823+00	\N		finisher		COMPLETED	{"Notes": "", "Phone": "1289763182", "TaxID": null, "Discount": null, "CountyTax": null, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@#g.co", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": null, "IsPickUpDelivery": false, "ShippingAddressID": null, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT							
920579	2025-08-11 20:40:58.566723+00	2025-08-11 20:40:58.926419+00	\N		safasdf		NEW	{"Notes": "", "Phone": "2342312312", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": null, "IsPickUpDelivery": false, "ShippingAddressID": null, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT							
078083	2025-08-11 20:48:45.120533+00	2025-08-13 00:14:56.844749+00	\N	WALK_IN	asdfasdf	CONTRACTOR_PICKUP	NEW	{"Notes": "", "Phone": "2341242143", "TaxID": null, "Discount": {"ID": 1, "Name": "flash", "DiscountPercentage": 10}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "12345", "DiscountID": null, "ShipToEmail": "ws2@AS.CO", "MobileNumber": "9876543210", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 36, "IsPickUpDelivery": true, "ShippingAddressID": 35, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	REGULAR	Caution		659808	910454	
571272	2025-08-11 20:51:19.628036+00	2025-08-12 03:47:55.302804+00	\N	WEBSITE	LEAD	CONTRACTOR_PICKUP	COMPLETED	{"Notes": "", "Phone": "1287623847", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "S@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 26, "IsPickUpDelivery": true, "ShippingAddressID": 25, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 2}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	PRO			659808	910454	
115148	2025-08-11 20:29:58.082976+00	2025-08-12 03:57:49.081826+00	\N	WALK_IN	sai, the gowtham	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "1237923872", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@g.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 28, "IsPickUpDelivery": false, "ShippingAddressID": 27, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	REGULAR			659808	910454	
684877	2025-08-12 05:23:02.440431+00	2025-08-12 06:20:50.365474+00	\N	WALK_IN	S	CUSTOMER_PICKUP	COMPLETED	{"Notes": "", "Phone": "9998887771", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "asdasd", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 32, "IsPickUpDelivery": true, "ShippingAddressID": 31, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	NEW			910454	910454	
973844	2025-08-12 09:01:16.424079+00	2025-08-12 10:37:06.618193+00	\N	WEBSITE	Test_Pradeep_1	INSTALLATION	REJECTED	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 3, "Name": "Madison", "TaxPercentage": 6.8}, "FaxNumber": "12345", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "8074604985", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 46, "IsPickUpDelivery": false, "ShippingAddressID": 45, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CHEQUE	PRO		af9d615f-8e34-4b44-ba26-fcccf434a8be.png	221472	910454	no material
573182	2025-08-12 04:48:26.482675+00	2025-08-12 06:24:51.490542+00	\N	WEBSITE	A	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "9998887771", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 7, "Name": "maincounty", "TaxPercentage": 7.7}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "a@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 30, "IsPickUpDelivery": false, "ShippingAddressID": 29, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	BEST			910454	659808	
022652	2025-08-12 06:32:16.867017+00	2025-08-12 06:59:50.502402+00	\N	WEBSITE	P	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 7, "Name": "maincounty", "TaxPercentage": 7.7}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 34, "IsPickUpDelivery": false, "ShippingAddressID": 33, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	PRO			910454	659808	
766166	2025-08-12 07:07:25.171709+00	2025-08-12 07:27:50.821823+00	\N	WALK_IN	pradeep yandamuri	CUSTOMER_PICKUP	COMPLETED	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 40, "IsPickUpDelivery": true, "ShippingAddressID": 39, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	REGULAR			659808	910454	
304586	2025-08-12 06:58:00.738074+00	2025-08-12 08:36:25.586477+00	\N	WEBSITE	Pradeep	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "12345", "DiscountID": null, "ShipToEmail": "pradeepyandamuri@gmail.com", "MobileNumber": "8074604985", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 38, "IsPickUpDelivery": false, "ShippingAddressID": 37, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CHEQUE	PRO		1f6a9d7b-4b6c-45e4-9900-7bb52bca10bc.png	910454	659808	
884011	2025-08-12 08:34:11.19793+00	2025-08-12 09:04:44.897633+00	\N	NORMAL	G	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "9998887771", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "123456", "DiscountID": null, "ShipToEmail": "g@gmail.com", "MobileNumber": "9998887771", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 42, "IsPickUpDelivery": false, "ShippingAddressID": 41, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CHEQUE	PRO			910454	659808	
301498	2025-08-12 08:59:37.795352+00	2025-08-12 09:10:33.19549+00	\N	WALK_IN	test_shankar_1	CUSTOMER_PICKUP	COMPLETED	{"Notes": "", "Phone": "5058148672", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 1, "Name": "Jackson", "TaxPercentage": 7.2}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "test@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "Yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 44, "IsPickUpDelivery": true, "ShippingAddressID": 43, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	REGULAR		46c11246-6b2e-490c-a319-1cdf4a487be5.pdf	910454	910454	
440604	2025-08-12 09:48:26.351242+00	2025-08-12 09:50:03.632723+00	\N	NORMAL	durga1	CUSTOMER_PICKUP	NEW	{"Notes": "", "Phone": "1234567890", "TaxID": null, "Discount": {"ID": 1, "Name": "flash", "DiscountPercentage": 10}, "CountyTax": {"ID": 1, "Name": "Jackson", "TaxPercentage": 7.2}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "asdasd", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 52, "IsPickUpDelivery": true, "ShippingAddressID": 51, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	NEW			221472	221472	
554098	2025-08-12 09:16:04.146658+00	2025-08-12 09:18:17.432035+00	\N	WEBSITE	shankar	CONTRACTOR_PICKUP	NEW	{"Notes": "", "Phone": "2648658378", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 1, "Name": "Jackson", "TaxPercentage": 7.2}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "shankar@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 48, "IsPickUpDelivery": true, "ShippingAddressID": 47, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 2}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	REGULAR			221472	221472	
297805	2025-08-12 09:37:36.100904+00	2025-08-12 09:37:36.624656+00	\N		asdfasdfsadf		NEW	{"Notes": "", "Phone": "3232534523", "TaxID": null, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "CountyTax": {"ID": 1, "Name": "Jackson", "TaxPercentage": 7.2}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": null, "IsPickUpDelivery": false, "ShippingAddressID": null, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT							
820700	2025-08-12 09:44:51.350162+00	2025-08-12 09:46:07.641705+00	\N	WALK_IN	killer1	CUSTOMER_PICKUP	NEW	{"Notes": "", "Phone": "1234567890", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 1, "Name": "Jackson", "TaxPercentage": 7.2}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "123", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 50, "IsPickUpDelivery": true, "ShippingAddressID": 49, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	NEW		10d9d789-e8d1-4157-951b-0f0ec3ece8a1.png	659808	910454	
467243	2025-08-12 23:11:54.978877+00	2025-08-13 00:17:27.505754+00	\N	NORMAL	Dheeraj123	SHIPPING	COMPLETED	{"Notes": "", "Phone": "1231239870", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "rin@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 60, "IsPickUpDelivery": false, "ShippingAddressID": 59, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	NEW	Urgent		351994	659808	
701868	2025-08-12 10:04:09.458507+00	2025-08-13 00:17:29.954909+00	\N	WEBSITE	Test_Pradeep_2	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 7, "Name": "maincounty", "TaxPercentage": 7.7}, "FaxNumber": "12345", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "1234567890", "ShipToMobile": "", "BillToCustomer": "1234", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 54, "IsPickUpDelivery": false, "ShippingAddressID": 53, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	PRO	Advisory		910454	221472	
487281	2025-08-12 10:41:53.582091+00	2025-08-12 13:11:38.593243+00	\N	NORMAL	Test Pradeep three	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 3, "Name": "Madison", "TaxPercentage": 6.8}, "FaxNumber": "123456", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "1234567890", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 56, "IsPickUpDelivery": false, "ShippingAddressID": 55, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	PRO			910454	221472	
858123	2025-08-12 12:55:29.994189+00	2025-08-13 01:04:45.39575+00	\N	WEBSITE	Test_Pradeep_4	CONTRACTOR_PICKUP	COMPLETED	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 10, "Name": "johnson", "TaxPercentage": 7.4}, "FaxNumber": "123456", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "1234567890", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 58, "IsPickUpDelivery": true, "ShippingAddressID": 57, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 3}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	BEST	Caution		910454	221472	
323585	2025-08-13 00:40:48.110866+00	2025-08-13 00:45:36.295478+00	\N	WEBSITE	David Shepherd	SHIPPING	COMPLETED	{"Notes": "", "Phone": "1232342211", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 7, "Name": "maincounty", "TaxPercentage": 7.6}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "dsf@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 68, "IsPickUpDelivery": false, "ShippingAddressID": 67, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CHEQUE	BEST			351994	351994	
317521	2025-08-13 00:07:05.242606+00	2025-08-13 00:08:03.647283+00	\N	NORMAL	sss ssss	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "4085904123", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "rinkelreddy@yahoo.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 64, "IsPickUpDelivery": false, "ShippingAddressID": 63, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	PRO			351994	910454	
253703	2025-08-12 23:50:15.13177+00	2025-08-13 00:17:25.64915+00	\N	WALK_IN	dheeraj_test_job_007	CONTRACTOR_PICKUP	COMPLETED	{"Notes": "", "Phone": "1232111133", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 3, "Name": "Madison", "TaxPercentage": 6.8}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "ya@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 62, "IsPickUpDelivery": true, "ShippingAddressID": 61, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	PRO	Caution	ef68c44d-b5b7-4c9d-97ea-e72902a6935a.png	351994	221472	
022137	2025-08-13 00:23:57.52238+00	2025-08-13 00:25:33.563875+00	\N	WALK_IN	test_dheeraj	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "4088399420", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "rinkelreddy@yahoo.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 66, "IsPickUpDelivery": false, "ShippingAddressID": 65, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	PRO			910454	351994	
688174	2025-08-13 01:09:54.232364+00	2025-08-14 13:15:05.629984+00	\N	WALK_IN	test_shankar_13	SHIPPING	COMPLETED	{"Notes": "", "Phone": "5058148672", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "shankar@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 72, "IsPickUpDelivery": false, "ShippingAddressID": 71, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	NEW	Advisory	4f1c3f20-92db-4a7a-b704-67b350f14409.pdf	659808	221472	
148872	2025-08-13 01:02:51.590242+00	2025-08-13 01:04:04.159673+00	\N	WEBSITE	test_dheeraj	CONTRACTOR_PICKUP	COMPLETED	{"Notes": "", "Phone": "0798127768", "TaxID": null, "Discount": {"ID": 1754998279239, "Name": "regular", "DiscountPercentage": 5}, "CountyTax": {"ID": 3, "Name": "Madison", "TaxPercentage": 6.8}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "wdewd@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 70, "IsPickUpDelivery": true, "ShippingAddressID": 69, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	REGULAR			221472	351994	
496909	2025-08-14 16:02:22.407627+00	2025-08-14 16:04:17.392412+00	\N	NORMAL	Dheerajreddy Pasham	CONTRACTOR_PICKUP	COMPLETED	{"Notes": "", "Phone": "7981277683", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 1, "Name": "Jackson", "TaxPercentage": 7.2}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "hjhj@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 94, "IsPickUpDelivery": true, "ShippingAddressID": 93, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	REGULAR			221472	659808	
870087	2025-08-13 03:41:57.72779+00	2025-08-14 13:14:57.194573+00	\N	WEBSITE	Test_Pradeep_5	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 10, "Name": "johnson", "TaxPercentage": 7.4}, "FaxNumber": "12345", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "1234567890", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 74, "IsPickUpDelivery": false, "ShippingAddressID": 73, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CHEQUE	PRO	Urgent	f8f2035e-bc01-42cf-8763-cff037d5bc93.png	910454	221472	
341276	2025-08-13 01:18:35.109274+00	2025-08-14 13:15:02.727878+00	\N		Shankar		NEW	{"Notes": "", "Phone": "1233444009", "TaxID": null, "Discount": {"ID": 1, "Name": "flash", "DiscountPercentage": 10}, "CountyTax": {"ID": 1, "Name": "Jackson", "TaxPercentage": 7.2}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "shankar@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": null, "IsPickUpDelivery": false, "ShippingAddressID": null, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT			Normal				
818324	2025-08-19 06:59:12.356738+00	2025-08-19 08:09:45.99034+00	\N	WEBSITE	Test_Pradeep_15	INSTALLATION	NEW	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 3, "Name": "Madison", "TaxPercentage": 6.8}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 124, "IsPickUpDelivery": false, "ShippingAddressID": 123, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CHEQUE	REGULAR			910454	659808	
484015	2025-08-14 16:16:34.952575+00	2025-08-14 16:18:50.326538+00	\N	WALK_IN	test_dheeraj	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "4088399420", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 3, "Name": "Madison", "TaxPercentage": 6.8}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "rinkelreddy@yahoo.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 96, "IsPickUpDelivery": false, "ShippingAddressID": 95, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	REGULAR		bf71ce3b-085e-430d-8e68-ad6299cb83d0.png	221472	221472	
249777	2025-08-14 02:22:29.117335+00	2025-08-14 02:25:10.367529+00	\N	NORMAL	Dheeraj123	CONTRACTOR_PICKUP	REJECTED	{"Notes": "", "Phone": "1233214523", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "we@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 78, "IsPickUpDelivery": true, "ShippingAddressID": 77, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	REGULAR			581600	910454	test123 fdf
255076	2025-08-14 16:30:40.295789+00	2025-08-14 16:32:36.032544+00	\N	NORMAL	Rishi Pasham	SHIPPING	NEW	{"Notes": "", "Phone": "8136506083", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 3, "Name": "Madison", "TaxPercentage": 6.8}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "rinkelreddy@yahoo.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 100, "IsPickUpDelivery": false, "ShippingAddressID": 99, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	REGULAR			910454	221472	
220509	2025-08-14 18:05:34.838835+00	2025-08-14 18:06:12.633347+00	\N	NORMAL	aditya_manohar	CUSTOMER_PICKUP	NEW	{"Notes": "", "Phone": "9985781999", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "adityamanohar20004@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "Yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 104, "IsPickUpDelivery": true, "ShippingAddressID": 103, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 2}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	REGULAR			659808	910454	
956125	2025-08-19 06:49:50.750043+00	2025-08-20 04:20:08.522881+00	\N	NORMAL	Test_Pradeep_11	INSTALLATION	NEW	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "CountyTax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 120, "IsPickUpDelivery": false, "ShippingAddressID": 119, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_NOT_SENT			Urgent		910454	659808	
673065	2025-08-19 06:55:06.92922+00	2025-08-20 10:34:40.791919+00	\N	WEBSITE	Test_Pradeep_12	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "CountyTax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 122, "IsPickUpDelivery": false, "ShippingAddressID": 121, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT					910454	659808	
822302	2025-08-20 11:28:31.046711+00	2025-08-20 11:28:31.046711+00	\N	WEBSITE	Pradeep456	INSTALLATION	NEW	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 7, "Name": "maincounty", "TaxPercentage": 7.6}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 126, "IsPickUpDelivery": false, "ShippingAddressID": 125, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_NOT_SENT	CASH	PRO		d7b10211-06b7-467f-964c-f7dd97f74bb0.png	910454	659808	
535395	2025-08-13 18:49:05.665796+00	2025-08-14 02:25:46.467244+00	\N	WEBSITE	Test_Pradeep_6	INSTALLATION	REJECTED	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 3, "Name": "Madison", "TaxPercentage": 6.8}, "FaxNumber": "123456", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "1234567890", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 76, "IsPickUpDelivery": false, "ShippingAddressID": 75, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	PRO		92ac850c-4124-4832-8d12-1113f9393d10.png	910454	221472	test
747298	2025-08-14 06:00:56.177747+00	2025-08-14 13:14:59.082806+00	\N	WEBSITE	Test_Pradeep_8	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 3, "Name": "Madison", "TaxPercentage": 6.8}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 82, "IsPickUpDelivery": false, "ShippingAddressID": 81, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CHEQUE	PRO	Caution	1c6ea98b-cbaa-4f70-8ac0-0851051aeaa2.png	910454	221472	
616587	2025-08-14 05:41:35.434963+00	2025-08-14 13:15:00.899714+00	\N	WEBSITE	Test_Pradeep_7	CONTRACTOR_PICKUP	COMPLETED	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 1, "Name": "Jackson", "TaxPercentage": 7.2}, "FaxNumber": "123456", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "1234567890", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 80, "IsPickUpDelivery": true, "ShippingAddressID": 79, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CHEQUE	PRO	Normal	ee57ced9-9ff1-4405-b954-98f498381246.png	910454	221472	
244629	2025-08-14 13:20:31.358591+00	2025-08-14 13:22:45.248914+00	\N	WALK_IN	sai nsg	CUSTOMER_PICKUP	COMPLETED	{"Notes": "", "Phone": "9640164778", "TaxID": null, "Discount": {"ID": 1, "Name": "flash", "DiscountPercentage": 10}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 86, "IsPickUpDelivery": true, "ShippingAddressID": 85, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	NEW			221472	910454	
516289	2025-08-14 08:11:30.312928+00	2025-08-14 21:04:53.006321+00	\N	WEBSITE	adiman	INSTALLATION	NEW	{"Notes": "", "Phone": "9121925025", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 7, "Name": "maincounty", "TaxPercentage": 7.6}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "adityamanohar@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 84, "IsPickUpDelivery": false, "ShippingAddressID": 83, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_NOT_SENT	CASH	PRO	Normal		910454	659808	
043649	2025-08-14 15:43:01.10091+00	2025-08-14 15:46:35.032815+00	\N	WALK_IN	test_shankar_3	CUSTOMER_PICKUP	COMPLETED	{"Notes": "", "Phone": "1324536758", "TaxID": null, "Discount": {"ID": 1, "Name": "flash", "DiscountPercentage": 10}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 90, "IsPickUpDelivery": true, "ShippingAddressID": 89, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	NEW		177b2ed3-3830-4a1a-b09e-fc7efe0cb8c7.pdf	910454	659808	
110545	2025-08-14 14:34:32.412733+00	2025-08-14 18:06:32.067224+00	\N	WALK_IN	durga	CUSTOMER_PICKUP	COMPLETED	{"Notes": "", "Phone": "9640164778", "TaxID": null, "Discount": {"ID": 1, "Name": "flash", "DiscountPercentage": 10}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 88, "IsPickUpDelivery": true, "ShippingAddressID": 87, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	NEW			659808	659808	
653535	2025-08-14 16:23:37.698427+00	2025-08-14 16:26:06.006248+00	\N	NORMAL	David Shepherd	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "4088399420", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 3, "Name": "Madison", "TaxPercentage": 6.8}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "rinkelreddy@yahoo.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 98, "IsPickUpDelivery": false, "ShippingAddressID": 97, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CHEQUE	REGULAR		95191787-d8b1-48f2-b4e9-36e9cb4bec2f.png	038230	659808	
806460	2025-08-14 15:56:55.176948+00	2025-08-14 18:06:29.973334+00	\N	WALK_IN	dheeraj_test_job_007	CONTRACTOR_PICKUP	COMPLETED	{"Notes": "", "Phone": "1234321123", "TaxID": null, "Discount": {"ID": 1, "Name": "flash", "DiscountPercentage": 10}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "dd@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 92, "IsPickUpDelivery": true, "ShippingAddressID": 91, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	REGULAR		ea98c18f-ca8d-412f-8280-84c98dc1d2e0.png	038230	910454	
076315	2025-08-14 18:24:51.685024+00	2025-08-14 18:26:25.195699+00	\N	NORMAL	dheeraj_test_job	CONTRACTOR_PICKUP	COMPLETED	{"Notes": "", "Phone": "1231234343", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "as@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 106, "IsPickUpDelivery": true, "ShippingAddressID": 105, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CHEQUE	REGULAR		7c3cdb0d-283e-4b1c-9348-49bed5dd95c2.png	038230	351994	
690103	2025-08-14 17:07:19.919054+00	2025-08-14 21:04:55.329626+00	\N	WALK_IN	test_final	CUSTOMER_PICKUP	NEW	{"Notes": "", "Phone": "1234567890", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 102, "IsPickUpDelivery": true, "ShippingAddressID": 101, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_NOT_SENT	CASH	NEW	Urgent	86819483-2a62-4461-a3e1-1ccdbbf626fb.png	659808	659808	
598530	2025-08-14 21:14:53.936072+00	2025-08-14 21:23:35.506787+00	\N	NORMAL	Rishi Pasham_001	SHIPPING	COMPLETED	{"Notes": "", "Phone": "4088399420", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "rinkelreddy@yahoo.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 108, "IsPickUpDelivery": false, "ShippingAddressID": 107, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	REGULAR		fb07f415-e7a9-4893-b974-0858683a0436.png	351994	581600	
605988	2025-08-14 21:29:32.264377+00	2025-08-14 21:31:42.29834+00	\N	NORMAL	Dheeraj Pasham	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "4088399420", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "rinkelreddy@yahoo.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 110, "IsPickUpDelivery": false, "ShippingAddressID": 109, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CHEQUE	REGULAR			221472	659808	
594353	2025-08-18 17:12:47.936302+00	2025-08-18 17:12:47.936302+00	\N	WEBSITE	Pradeep123	INSTALLATION	NEW	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 3, "Name": "Madison", "TaxPercentage": 6.8}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 112, "IsPickUpDelivery": false, "ShippingAddressID": 111, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_NOT_SENT	CHEQUE	PRO			910454	659808	
979672	2025-08-18 17:28:58.452562+00	2025-08-18 17:28:58.452562+00	\N	WEBSITE	Pradeep123	INSTALLATION	NEW	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 114, "IsPickUpDelivery": false, "ShippingAddressID": 113, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_NOT_SENT	CREDIT	REGULAR			910454	659808	
445169	2025-08-19 06:28:54.349342+00	2025-08-19 06:43:44.689672+00	\N	WEBSITE	Test_Pradeep_10	INSTALLATION	NEW	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 116, "IsPickUpDelivery": false, "ShippingAddressID": 115, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_NOT_SENT	CREDIT	PRO			910454	659808	
820904	2025-08-22 06:06:11.052296+00	2025-08-22 06:06:14.022512+00	\N		Pradeep		NEW	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "pradeepyandamuri532@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": null, "IsPickUpDelivery": false, "ShippingAddressID": null, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT							
418495	2025-08-19 06:46:11.449311+00	2025-08-20 07:21:18.387918+00	\N	NORMAL	Test_Pradeep_10	INSTALLATION	NEW	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "CountyTax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 118, "IsPickUpDelivery": false, "ShippingAddressID": 117, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_NOT_SENT	CREDIT		Caution		910454	659808	
629876	2025-08-22 02:55:02.983886+00	2025-08-22 02:55:20.588776+00	\N	WEBSITE	test_dheeraj	CONTRACTOR_PICKUP	REJECTED	{"Notes": "", "Phone": "4088399430", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "123@hhi.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 134, "IsPickUpDelivery": true, "ShippingAddressID": 133, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_NOT_SENT	CREDIT	PRO			910454	221472	test
660420	2025-08-21 09:46:46.334158+00	2025-08-21 14:45:43.730844+00	\N	WEBSITE	Test_Pradeep_16	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 7, "Name": "maincounty", "TaxPercentage": 7.6}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": null, "IsPickUpDelivery": false, "ShippingAddressID": null, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}		CASH	PRO			910454	659808	
914196	2025-08-21 13:20:55.889829+00	2025-08-21 13:20:55.889829+00	\N	WEBSITE	Pradeep000	INSTALLATION	NEW	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 7, "Name": "maincounty", "TaxPercentage": 7.6}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 130, "IsPickUpDelivery": false, "ShippingAddressID": 129, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_NOT_SENT	CASH	PRO			910454	659808	
888866	2025-08-22 02:42:23.908178+00	2025-08-22 02:47:24.641222+00	\N	FACEBOOK	raja bathroom	CONTRACTOR_PICKUP	COMPLETED	{"Notes": "", "Phone": "1239874567", "TaxID": null, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "CountyTax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "raj@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 132, "IsPickUpDelivery": true, "ShippingAddressID": 131, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CASH	REGULAR		baa02ecc-532d-43a1-81b2-9d6cb08e98de.png	877682	351994	
556657	2025-08-22 03:03:24.593484+00	2025-08-22 03:05:54.492369+00	\N	NORMAL	Dheeraj Pasham	CONTRACTOR_PICKUP	NEW	{"Notes": "", "Phone": "4088399420", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 3, "Name": "Madison", "TaxPercentage": 6.8}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "dheerajreddy83@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 136, "IsPickUpDelivery": true, "ShippingAddressID": 135, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	REGULAR			910454	221472	
652375	2025-08-22 03:11:25.691188+00	2025-08-22 03:15:42.011252+00	\N	WALK_IN	rishi 123123	INSTALLATION	COMPLETED	{"Notes": "", "Phone": "4088399420", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "rinkelreddy@yahoo.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 138, "IsPickUpDelivery": false, "ShippingAddressID": 137, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT	CREDIT	PRO		da8fb0c3-25b5-4926-b32c-6fa780d520f0.png	910454	063103	
225983	2025-08-22 03:20:46.127599+00	2025-08-22 03:20:46.127599+00	\N	NORMAL	David Shepherd	CONTRACTOR_PICKUP	NEW	{"Notes": "", "Phone": "1231234343", "TaxID": null, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "CountyTax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "as@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "yes", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 140, "IsPickUpDelivery": true, "ShippingAddressID": 139, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": 1}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_NOT_SENT	CREDIT	NEW			038230	351994	
635119	2025-08-22 06:06:13.100009+00	2025-08-22 06:06:18.02074+00	\N		Pradeep		NEW	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 1754932278832, "Name": "Best", "DiscountPercentage": 15}, "CountyTax": {"ID": 2, "Name": "Jefferson", "TaxPercentage": 8.5}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "pradeepyandamuri532@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": null, "IsPickUpDelivery": false, "ShippingAddressID": null, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT							
142588	2025-08-22 09:57:45.212871+00	2025-08-22 09:57:47.554626+00	\N		n sai gowtham		NEW	{"Notes": "", "Phone": "1234567890", "TaxID": null, "Discount": {"ID": 1, "Name": "flash", "DiscountPercentage": 10}, "CountyTax": {"ID": 7, "Name": "maincounty", "TaxPercentage": 7.6}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "s@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": null, "IsPickUpDelivery": false, "ShippingAddressID": null, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_SENT							
692681	2025-08-22 11:07:03.456989+00	2025-08-22 11:07:03.456989+00	\N	NORMAL	Pradeep1010	INSTALLATION	NEW	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 0, "Name": "", "DiscountPercentage": 0}, "CountyTax": {"ID": 0, "Name": "", "TaxPercentage": 0}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 146, "IsPickUpDelivery": false, "ShippingAddressID": 145, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_NOT_SENT				f42f2e0a-7f9e-4370-91ea-2b5edbec94c1.png	910454		
235859	2025-08-22 09:43:35.310936+00	2025-08-22 09:43:35.310936+00	\N	WEBSITE	Pradeep20	INSTALLATION	NEW	{"Notes": "", "Phone": "8074604985", "TaxID": null, "Discount": {"ID": 2, "Name": "promo", "DiscountPercentage": 20}, "CountyTax": {"ID": 7, "Name": "maincounty", "TaxPercentage": 7.6}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "yandamuripradeep@gmail.com", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "12345", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": 144, "IsPickUpDelivery": false, "ShippingAddressID": 143, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}	ESTIMATE_NOT_SENT	CASH	PRO		f1e41bb0-f3af-4d67-9ffc-6acbf5678c67.png	910454	659808	
580823	2025-08-22 08:46:06.442777+00	2025-08-22 09:47:34.20672+00	\N	WEBSITE	Pradeep001	INSTALLATION	NEW	{"Notes": "", "Phone": "", "TaxID": null, "Discount": {"ID": 1, "Name": "flash", "DiscountPercentage": 10}, "CountyTax": {"ID": 10, "Name": "johnson", "TaxPercentage": 7.4}, "FaxNumber": "", "DiscountID": null, "ShipToEmail": "", "MobileNumber": "", "ShipToMobile": "", "BillToCustomer": "", "BillingAddress": null, "ShippingAddress": null, "BillingAddressID": null, "IsPickUpDelivery": false, "ShippingAddressID": null, "PickupDeliveryAddress": null, "PickupDeliveryAddressID": null}	{"name": "", "email": "", "phone": "", "IsReferred": false}		CASH	PRO			910454	659808	
\.


--
-- Data for Name: manager_residential_entries; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.manager_residential_entries (id, residential_entry_name) FROM stdin;
1	Kitchen
2	Bathroom
3	Master Bathroom
4	Master Bedroom
5	Other
6	Other
\.


--
-- Data for Name: manager_residential_types; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.manager_residential_types (id, residential_entry_type) FROM stdin;
1	Residential Remodel
2	Residential New-Construction
3	Commercial
4	Apartment
\.


--
-- Data for Name: material_rows; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.material_rows (id, details, estimate_id, residential_entry_id) FROM stdin;
1754925203365	{"Total": 100, "IsCustom": false, "Material": "material1", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub1", "AdditionalDiscount": 0}	005640	1754925139227
1754925213112	{"Total": 100, "IsCustom": true, "Material": "asdfasdf", "Quantity": 10, "UnitPrice": 10, "Description": "asdfsadf", "SubCategory": "asdfasdf", "AdditionalDiscount": 0}	005640	1754925139227
1754925346006	{"Total": 100, "IsCustom": false, "Material": "material2", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	828551	1754925345946
1754925360944	{"Total": 100, "IsCustom": true, "Material": "sadfasdf", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "asdfasdf", "AdditionalDiscount": 0}	828551	1754925345946
1754925352005	{"Total": 100, "IsCustom": false, "Material": "material1", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub1", "AdditionalDiscount": 0}	828551	1754925352072
1754925378189	{"Total": 10, "IsCustom": true, "Material": "asdfasdf", "Quantity": 10, "UnitPrice": 1, "Description": "", "SubCategory": "asdfasdf", "AdditionalDiscount": 0}	828551	1754925352072
1754925822253	{"Total": 100, "IsCustom": false, "Material": "material2", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	138115	1754925822298
1754930592202	{"Total": 100, "IsCustom": false, "Material": "material1", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub1", "AdditionalDiscount": 0}	299571	1754930429838
1754932054527	{"Total": 100, "IsCustom": false, "Material": "material1", "Quantity": 10, "UnitPrice": 10, "Description": "sadfasdf", "SubCategory": "sub1", "AdditionalDiscount": 0}	449661	1754932054536
1754932065265	{"Total": 100, "IsCustom": true, "Material": "asdfasdf", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "asdfasdf", "AdditionalDiscount": 0}	449661	1754932054536
1754932059220	{"Total": 100, "IsCustom": false, "Material": "material2", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	449661	1754932059204
1754932082899	{"Total": 1, "IsCustom": true, "Material": "sadfsdf", "Quantity": 1, "UnitPrice": 1, "Description": "", "SubCategory": "asdfasdf", "AdditionalDiscount": 0}	449661	1754932059204
1754933591890	{"Total": 100, "IsCustom": false, "Material": "material1", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub1", "AdditionalDiscount": 0}	440111	1754933591839
1754936584805	{"Total": 900, "IsCustom": false, "Material": "Material-6", "Quantity": 30, "UnitPrice": 30, "Description": "", "SubCategory": "dgfhgj", "AdditionalDiscount": 0}	348917	1754931688619
1754936597093	{"Total": 200, "IsCustom": true, "Material": "steel", "Quantity": 20, "UnitPrice": 10, "Description": "", "SubCategory": "dhbs", "AdditionalDiscount": 0}	348917	1754931688619
1754934514431	{"Total": 300, "IsCustom": false, "Material": "Material-6", "Quantity": 10, "UnitPrice": 30, "Description": "", "SubCategory": "dgfhgj", "AdditionalDiscount": 0}	104900	1754934118795
1754934525682	{"Total": 400, "IsCustom": true, "Material": "steel", "Quantity": 10, "UnitPrice": 40, "Description": "", "SubCategory": "sub", "AdditionalDiscount": 0}	104900	1754934118795
1754934746914	{"Total": 200, "IsCustom": false, "Material": "material1", "Quantity": 20, "UnitPrice": 10, "Description": "", "SubCategory": "sub1", "AdditionalDiscount": 0}	104900	1754934746920
1754934760896	{"Total": 150, "IsCustom": false, "Material": "material2", "Quantity": 15, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	104900	1754934746920
1754934775935	{"Total": 600, "IsCustom": true, "Material": "wood", "Quantity": 20, "UnitPrice": 30, "Description": "", "SubCategory": "sub wood", "AdditionalDiscount": 0}	104900	1754934746920
1754934792569	{"Total": 1000, "IsCustom": true, "Material": "ceramic", "Quantity": 20, "UnitPrice": 50, "Description": "", "SubCategory": "sjkajksd", "AdditionalDiscount": 0}	104900	1754934746920
1754939980090	{"Total": 900, "IsCustom": false, "Material": "Material-6", "Quantity": 30, "UnitPrice": 30, "Description": "", "SubCategory": "dgfhgj", "AdditionalDiscount": 0}	947144	1754939980082
1754940004122	{"Total": 200, "IsCustom": true, "Material": "steel", "Quantity": 20, "UnitPrice": 10, "Description": "", "SubCategory": "asdfgh", "AdditionalDiscount": 0}	947144	1754939980082
1754934206592	{"Total": 100, "IsCustom": false, "Material": "material1", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub1", "AdditionalDiscount": 0}	761845	1754934206559
1754934214532	{"Total": 100, "IsCustom": true, "Material": "asdfasdf", "Quantity": 10, "UnitPrice": 10, "Description": "asdfasd", "SubCategory": "asdfasdf", "AdditionalDiscount": 0}	761845	1754934206559
1754937681948	{"Total": 200, "IsCustom": false, "Material": "material-7", "Quantity": 10, "UnitPrice": 20, "Description": "", "SubCategory": "kjjkdsjdk", "AdditionalDiscount": 0}	761845	1754934291375
1754937683990	{"Total": 100, "IsCustom": true, "Material": "sdfgsdfg", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "asdfasdf", "AdditionalDiscount": 0}	761845	1754934291375
1754976320216	{"Total": 200, "IsCustom": true, "Material": "wood", "Quantity": 10, "UnitPrice": 20, "Description": "", "SubCategory": "daa", "AdditionalDiscount": 0}	359478	1754976307915
1754976526775	{"Total": 100, "IsCustom": false, "Material": "material1", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub1", "AdditionalDiscount": 0}	359478	1754976459819
1754974758126	{"Total": 210, "IsCustom": false, "Material": "Material-3", "Quantity": 21, "UnitPrice": 10, "Description": "", "SubCategory": "aub", "AdditionalDiscount": 0}	846594	1754974597792
1754974783790	{"Total": 800, "IsCustom": true, "Material": "wood", "Quantity": 20, "UnitPrice": 40, "Description": "", "SubCategory": "skjd", "AdditionalDiscount": 0}	846594	1754974597792
1754974182951	{"Total": 300, "IsCustom": false, "Material": "Material-3", "Quantity": 30, "UnitPrice": 10, "Description": "", "SubCategory": "aub", "AdditionalDiscount": 0}	846594	1754973897203
1754974196004	{"Total": 200, "IsCustom": true, "Material": "steel", "Quantity": 10, "UnitPrice": 20, "Description": "", "SubCategory": "sub", "AdditionalDiscount": 0}	846594	1754973897203
1754980424134	{"Total": 200, "IsCustom": false, "Material": "Material-3", "Quantity": 20, "UnitPrice": 10, "Description": "", "SubCategory": "aub", "AdditionalDiscount": 0}	365344	1754980164865
1754980803983	{"Total": 200, "IsCustom": false, "Material": "Material-4", "Quantity": 10, "UnitPrice": 20, "Description": "", "SubCategory": "dhhds", "AdditionalDiscount": 0}	365344	1754980164865
1754982366887	{"Total": 200, "IsCustom": false, "Material": "Material-3", "Quantity": 20, "UnitPrice": 10, "Description": "dkjhsdah", "SubCategory": "aub", "AdditionalDiscount": 0}	849736	1754982366867
1754976097671	{"Total": 870, "IsCustom": false, "Material": "Material-6", "Quantity": 29, "UnitPrice": 30, "Description": "", "SubCategory": "dgfhgj", "AdditionalDiscount": 0}	359478	1754976097604
1754976111254	{"Total": 200, "IsCustom": true, "Material": "steel", "Quantity": 20, "UnitPrice": 10, "Description": "dasdsc", "SubCategory": "asd", "AdditionalDiscount": 0}	359478	1754976097604
1754976307935	{"Total": 400, "IsCustom": false, "Material": "Material-3", "Quantity": 40, "UnitPrice": 10, "Description": "", "SubCategory": "aub", "AdditionalDiscount": 0}	359478	1754976307915
1754980437292	{"Total": 300, "IsCustom": true, "Material": "steel", "Quantity": 10, "UnitPrice": 30, "Description": "", "SubCategory": "asjkd", "AdditionalDiscount": 0}	365344	1754980164865
1754980815066	{"Total": 100, "IsCustom": true, "Material": "iron", "Quantity": 20, "UnitPrice": 5, "Description": "", "SubCategory": "dsa", "AdditionalDiscount": 0}	365344	1754980164865
1754980523389	{"Total": 600, "IsCustom": false, "Material": "Material-6", "Quantity": 20, "UnitPrice": 30, "Description": "", "SubCategory": "dgfhgj", "AdditionalDiscount": 0}	365344	1754980523379
1754980543232	{"Total": 1000, "IsCustom": true, "Material": "wood", "Quantity": 20, "UnitPrice": 50, "Description": "", "SubCategory": "sdkjf", "AdditionalDiscount": 0}	365344	1754980523379
1754982599592	{"Total": 300, "IsCustom": false, "Material": "material2", "Quantity": 30, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	849736	1754982366867
1754980687538	{"Total": 300, "IsCustom": false, "Material": "Material-3", "Quantity": 30, "UnitPrice": 10, "Description": "zcddfnds", "SubCategory": "aub", "AdditionalDiscount": 0}	365344	1754980623793
1754980708657	{"Total": 800, "IsCustom": true, "Material": "steel & wood", "Quantity": 20, "UnitPrice": 40, "Description": "", "SubCategory": "sjhdbhds", "AdditionalDiscount": 0}	365344	1754980623793
1754982384761	{"Total": 600, "IsCustom": true, "Material": "steel", "Quantity": 20, "UnitPrice": 30, "Description": "", "SubCategory": "skdhsadjk", "AdditionalDiscount": 0}	849736	1754982366867
1754982609223	{"Total": 300, "IsCustom": true, "Material": "wood", "Quantity": 30, "UnitPrice": 10, "Description": "", "SubCategory": "sdjhsdh", "AdditionalDiscount": 0}	849736	1754982366867
1754982511175	{"Total": 200, "IsCustom": false, "Material": "Material-3", "Quantity": 20, "UnitPrice": 10, "Description": "", "SubCategory": "aub", "AdditionalDiscount": 0}	849736	1754982511168
1754982533420	{"Total": 600, "IsCustom": true, "Material": "wood", "Quantity": 20, "UnitPrice": 30, "Description": "", "SubCategory": "sdjkkshb", "AdditionalDiscount": 0}	849736	1754982511168
1754983741500	{"Total": 90, "IsCustom": false, "Material": "material2", "Quantity": 9, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	849736	1754983629610
1754987701306	{"Total": 300, "IsCustom": false, "Material": "Material-3", "Quantity": 30, "UnitPrice": 10, "Description": "", "SubCategory": "aub", "AdditionalDiscount": 0}	941746	1754981655078
1754987716652	{"Total": 200, "IsCustom": true, "Material": "steel", "Quantity": 20, "UnitPrice": 10, "Description": "", "SubCategory": "shff", "AdditionalDiscount": 0}	941746	1754981655078
1754989332300	{"Total": 100, "IsCustom": false, "Material": "material1", "Quantity": 10, "UnitPrice": 10, "Description": "this is good", "SubCategory": "sub1", "AdditionalDiscount": 0}	307848	1754989034701
1754989395443	{"Total": 100, "IsCustom": false, "Material": "Material-3", "Quantity": 10, "UnitPrice": 10, "Description": "hello", "SubCategory": "aub", "AdditionalDiscount": 0}	720330	1754989105243
1754989414110	{"Total": 200, "IsCustom": true, "Material": "steel", "Quantity": 10, "UnitPrice": 20, "Description": "how are you", "SubCategory": "sdk", "AdditionalDiscount": 0}	720330	1754989105243
1754989347212	{"Total": 50, "IsCustom": true, "Material": "hello", "Quantity": 5, "UnitPrice": 10, "Description": "this is good", "SubCategory": "", "AdditionalDiscount": 0}	307848	1754989034701
1754989332257	{"Total": 20, "IsCustom": false, "Material": "Material-4", "Quantity": 1, "UnitPrice": 20, "Description": "this is good one", "SubCategory": "dhhds", "AdditionalDiscount": 0}	307848	1754989043654
1754987876177	{"Total": 490, "IsCustom": false, "Material": "Material-3", "Quantity": 10, "UnitPrice": 49, "Description": "", "SubCategory": "aub", "AdditionalDiscount": 0}	153822	1754987522866
1754987889856	{"Total": 200, "IsCustom": true, "Material": "steel", "Quantity": 20, "UnitPrice": 10, "Description": "", "SubCategory": "sdhf", "AdditionalDiscount": 0}	153822	1754987522866
1754989395427	{"Total": 600, "IsCustom": false, "Material": "material1", "Quantity": 20, "UnitPrice": 10, "Description": "", "SubCategory": "sub1", "AdditionalDiscount": 0}	720330	1754989115005
1754988485427	{"Total": 600, "IsCustom": false, "Material": "Material-3", "Quantity": 60, "UnitPrice": 10, "Description": "", "SubCategory": "aub", "AdditionalDiscount": 0}	153822	1754988418112
1754988497683	{"Total": 200, "IsCustom": true, "Material": "wood ", "Quantity": 20, "UnitPrice": 10, "Description": "", "SubCategory": "sdjf", "AdditionalDiscount": 0}	153822	1754988418112
1754988712583	{"Total": 600, "IsCustom": false, "Material": "Material-6", "Quantity": 20, "UnitPrice": 30, "Description": "", "SubCategory": "dgfhgj", "AdditionalDiscount": 0}	153822	1754988712647
1754988730951	{"Total": 300, "IsCustom": true, "Material": "iron", "Quantity": 30, "UnitPrice": 10, "Description": "", "SubCategory": "hsdbsfd", "AdditionalDiscount": 0}	153822	1754988712647
1754993328026	{"Total": 800, "IsCustom": true, "Material": "wood", "Quantity": 20, "UnitPrice": 40, "Description": "good material", "SubCategory": "good", "AdditionalDiscount": 0}	561785	1754992910045
1754993316510	{"Total": 600, "IsCustom": false, "Material": "Material-4", "Quantity": 30, "UnitPrice": 20, "Description": "good", "SubCategory": "dhhds", "AdditionalDiscount": 0}	561785	1754993027495
1754993385025	{"Total": 600, "IsCustom": true, "Material": "steel", "Quantity": 20, "UnitPrice": 30, "Description": "nice", "SubCategory": "good", "AdditionalDiscount": 0}	561785	1754993027495
1754989504901	{"Total": 300, "IsCustom": true, "Material": "wood", "Quantity": 10, "UnitPrice": 30, "Description": "sds", "SubCategory": "shdbs", "AdditionalDiscount": 0}	720330	1754989115005
1754990000981	{"Total": 1000, "IsCustom": false, "Material": "material1", "Quantity": 100, "UnitPrice": 10, "Description": "this is good", "SubCategory": "sub1", "AdditionalDiscount": 0}	416334	1754990000977
1754990030325	{"Total": 1000, "IsCustom": true, "Material": "good", "Quantity": 10, "UnitPrice": 100, "Description": "this is good", "SubCategory": "Quartz", "AdditionalDiscount": 0}	416334	1754990000977
1754991948638	{"Total": 100, "IsCustom": false, "Material": "material2", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	640014	1754991834726
1754992141186	{"Total": 100, "IsCustom": false, "Material": "material2", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	436116	1754992010350
1754992147950	{"Total": 100, "IsCustom": true, "Material": "custom", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "custom", "AdditionalDiscount": 0}	436116	1754992010350
1754992172424	{"Total": 100, "IsCustom": false, "Material": "material2", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	436116	1754992172408
1754992182045	{"Total": 100, "IsCustom": true, "Material": "material", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "mat", "AdditionalDiscount": 0}	436116	1754992172408
1754993468060	{"Total": 300, "IsCustom": false, "Material": "Material-5", "Quantity": 30, "UnitPrice": 10, "Description": "sdafs", "SubCategory": "snd", "AdditionalDiscount": 0}	561785	1754993468042
1754993487726	{"Total": 1500, "IsCustom": true, "Material": "steel and wood", "Quantity": 30, "UnitPrice": 50, "Description": "nice", "SubCategory": "good", "AdditionalDiscount": 0}	561785	1754993468042
1754993707996	{"Total": 200, "IsCustom": false, "Material": "Material-3", "Quantity": 20, "UnitPrice": 10, "Description": "AD", "SubCategory": "aub", "AdditionalDiscount": 0}	561785	1754993602935
1754993720513	{"Total": 600, "IsCustom": true, "Material": "IRON", "Quantity": 20, "UnitPrice": 30, "Description": "SDJK", "SubCategory": "ASJKD", "AdditionalDiscount": 0}	561785	1754993602935
1754993316538	{"Total": 300, "IsCustom": false, "Material": "Material-5", "Quantity": 30, "UnitPrice": 10, "Description": "good", "SubCategory": "snd", "AdditionalDiscount": 0}	561785	1754992910045
1754995164958	{"Total": 300, "IsCustom": false, "Material": "Material-3", "Quantity": 30, "UnitPrice": 10, "Description": "", "SubCategory": "aub", "AdditionalDiscount": 0}	549052	1754995164876
1754995181114	{"Total": 300, "IsCustom": true, "Material": "iron", "Quantity": 10, "UnitPrice": 30, "Description": "dkjf", "SubCategory": "good", "AdditionalDiscount": 0}	549052	1754995164876
1754995507193	{"Total": 400, "IsCustom": false, "Material": "material-7", "Quantity": 20, "UnitPrice": 20, "Description": "gddgffgjf", "SubCategory": "kjjkdsjdk", "AdditionalDiscount": 0}	549052	1754995393376
1754995519974	{"Total": 1200, "IsCustom": true, "Material": "iron", "Quantity": 30, "UnitPrice": 40, "Description": "sadbf", "SubCategory": "good ", "AdditionalDiscount": 0}	549052	1754995393376
1755004311182	{"Total": 800, "IsCustom": false, "Material": "Material-4", "Quantity": 40, "UnitPrice": 20, "Description": "", "SubCategory": "dhhds", "AdditionalDiscount": 0}	549052	1755004271757
1755004323285	{"Total": 1200, "IsCustom": true, "Material": "iron", "Quantity": 40, "UnitPrice": 30, "Description": "klhad", "SubCategory": "good", "AdditionalDiscount": 0}	549052	1755004271757
1755004409943	{"Total": 300, "IsCustom": false, "Material": "material2", "Quantity": 30, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	549052	1755004409938
1755004425811	{"Total": 400, "IsCustom": true, "Material": "steel and iron", "Quantity": 20, "UnitPrice": 20, "Description": "sjk", "SubCategory": "good", "AdditionalDiscount": 0}	549052	1755004409938
1755040382333	{"Total": 100, "IsCustom": false, "Material": "material2", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	420632	1755040181565
1755047407567	{"Total": 50, "IsCustom": false, "Material": "material2", "Quantity": 5, "UnitPrice": 10, "Description": "sss", "SubCategory": "sub2", "AdditionalDiscount": 0}	296995	1755047241451
1755047407472	{"Total": 100, "IsCustom": false, "Material": "Material-4", "Quantity": 5, "UnitPrice": 20, "Description": "sss", "SubCategory": "dhhds", "AdditionalDiscount": 0}	296995	1755047247076
1755047850268	{"Total": 1000, "IsCustom": false, "Material": "material2", "Quantity": 100, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	221078	1755047850236
1755042725043	{"Total": 1000, "IsCustom": false, "Material": "Material-4", "Quantity": 50, "UnitPrice": 20, "Description": "", "SubCategory": "dhhds", "AdditionalDiscount": 0}	575388	1755042280869
1755043006298	{"Total": 240, "IsCustom": false, "Material": "Material-6", "Quantity": 8, "UnitPrice": 30, "Description": "", "SubCategory": "", "AdditionalDiscount": 0}	575388	1755042280869
1755044680260	{"Total": 300, "IsCustom": false, "Material": "Material-6", "Quantity": 10, "UnitPrice": 30, "Description": "", "SubCategory": "", "AdditionalDiscount": 0}	141628	1755044598688
1755044680292	{"Total": 200, "IsCustom": false, "Material": "Material-4", "Quantity": 10, "UnitPrice": 20, "Description": "", "SubCategory": "", "AdditionalDiscount": 0}	141628	1755044605893
1755045701771	{"Total": 1100, "IsCustom": false, "Material": "Material-5", "Quantity": 110, "UnitPrice": 10, "Description": "", "SubCategory": "snd", "AdditionalDiscount": 0}	367470	1755045602331
1755045722823	{"Total": 1110, "IsCustom": true, "Material": "hi", "Quantity": 10, "UnitPrice": 111, "Description": "", "SubCategory": "", "AdditionalDiscount": 0}	367470	1755045602331
1755047005197	{"Total": 100, "IsCustom": false, "Material": "material2", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	491094	1755046904924
1755047005171	{"Total": 2220, "IsCustom": false, "Material": "Material-4", "Quantity": 111, "UnitPrice": 20, "Description": "", "SubCategory": "dhhds", "AdditionalDiscount": 0}	491094	1755046911899
1755047065571	{"Total": 10, "IsCustom": false, "Material": "material1", "Quantity": 1, "UnitPrice": 10, "Description": "", "SubCategory": "sub1", "AdditionalDiscount": 0}	096722	1755003067973
1755056528393	{"Total": 800, "IsCustom": false, "Material": "Material-4", "Quantity": 40, "UnitPrice": 20, "Description": "", "SubCategory": "dhhds", "AdditionalDiscount": 0}	181659	1755056329204
1755056539994	{"Total": 600, "IsCustom": true, "Material": "iron", "Quantity": 30, "UnitPrice": 20, "Description": "", "SubCategory": "good", "AdditionalDiscount": 0}	181659	1755056329204
1755111478397	{"Total": 600, "IsCustom": true, "Material": "iron", "Quantity": 30, "UnitPrice": 20, "Description": "akjdsf", "SubCategory": "good", "AdditionalDiscount": 0}	215051	1755111461136
1755111628577	{"Total": 300, "IsCustom": false, "Material": "Material-3", "Quantity": 30, "UnitPrice": 10, "Description": "", "SubCategory": "aub", "AdditionalDiscount": 0}	215051	1755111628554
1755111646615	{"Total": 800, "IsCustom": true, "Material": "steel", "Quantity": 20, "UnitPrice": 40, "Description": "ashb", "SubCategory": "good", "AdditionalDiscount": 0}	215051	1755111628554
1755138174122	{"Total": 100, "IsCustom": false, "Material": "Material-5", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "snd", "AdditionalDiscount": 0}	764222	1755138100235
1755150310980	{"Total": 300, "IsCustom": false, "Material": "Material-3", "Quantity": 30, "UnitPrice": 10, "Description": "", "SubCategory": "aub", "AdditionalDiscount": 0}	276098	1755149396910
1755150320403	{"Total": 600, "IsCustom": true, "Material": "iron", "Quantity": 30, "UnitPrice": 20, "Description": "shd", "SubCategory": "good", "AdditionalDiscount": 0}	276098	1755149396910
1755151465119	{"Total": 200, "IsCustom": false, "Material": "Material-3", "Quantity": 20, "UnitPrice": 10, "Description": "", "SubCategory": "aub", "AdditionalDiscount": 0}	878573	1755150932103
1755151475083	{"Total": 300, "IsCustom": true, "Material": "iron", "Quantity": 30, "UnitPrice": 10, "Description": "", "SubCategory": "good", "AdditionalDiscount": 0}	878573	1755150932103
1755187042402	{"Total": 100, "IsCustom": false, "Material": "Material-5", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "snd", "AdditionalDiscount": 0}	422170	1755186887688
1755187055700	{"Total": 100, "IsCustom": true, "Material": "material t", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "yes", "AdditionalDiscount": 0}	422170	1755186887688
1755111102459	{"Total": 900, "IsCustom": false, "Material": "Material-6", "Quantity": 30, "UnitPrice": 30, "Description": "adfdsfsda", "SubCategory": "dgfhgj", "AdditionalDiscount": 0}	215051	1755110639937
1755111128664	{"Total": 200, "IsCustom": true, "Material": "wood", "Quantity": 10, "UnitPrice": 20, "Description": "kjsd", "SubCategory": "good", "AdditionalDiscount": 0}	215051	1755110639937
1755111184097	{"Total": 300, "IsCustom": true, "Material": "steel", "Quantity": 10, "UnitPrice": 30, "Description": "askjd", "SubCategory": "good", "AdditionalDiscount": 0}	215051	1755110639937
1755111713919	{"Total": 200, "IsCustom": false, "Material": "Material-3", "Quantity": 20, "UnitPrice": 10, "Description": "", "SubCategory": "aub", "AdditionalDiscount": 0}	215051	1755110649635
1755111723635	{"Total": 500, "IsCustom": true, "Material": "steel and wood", "Quantity": 50, "UnitPrice": 10, "Description": "", "SubCategory": "good", "AdditionalDiscount": 0}	215051	1755110649635
1755111461145	{"Total": 600, "IsCustom": false, "Material": "Material-4", "Quantity": 30, "UnitPrice": 20, "Description": "sjkdjk", "SubCategory": "dhhds", "AdditionalDiscount": 0}	215051	1755111461136
1755138220361	{"Total": 100, "IsCustom": true, "Material": "test123", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "", "AdditionalDiscount": 0}	764222	1755138100235
1755186198350	{"Total": 100, "IsCustom": false, "Material": "material1", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub1", "AdditionalDiscount": 0}	874354	1755185902660
1755186207848	{"Total": 1100, "IsCustom": true, "Material": "sdafasfd", "Quantity": 10, "UnitPrice": 110, "Description": "", "SubCategory": "sfasfdsdf", "AdditionalDiscount": 0}	874354	1755185902660
1755078794261	{"Total": 400, "IsCustom": false, "Material": "Material-3", "Quantity": 40, "UnitPrice": 10, "Description": "dc", "SubCategory": "aub", "AdditionalDiscount": 0}	181659	1755078794261
1755078807742	{"Total": 600, "IsCustom": true, "Material": "iron", "Quantity": 30, "UnitPrice": 20, "Description": "", "SubCategory": "good", "AdditionalDiscount": 0}	181659	1755078794261
1755078959958	{"Total": 600, "IsCustom": false, "Material": "Material-6", "Quantity": 20, "UnitPrice": 30, "Description": "zczzcx", "SubCategory": "dgfhgj", "AdditionalDiscount": 0}	181659	1755078960028
1755078980577	{"Total": 600, "IsCustom": true, "Material": "steel", "Quantity": 20, "UnitPrice": 30, "Description": "sda", "SubCategory": "good", "AdditionalDiscount": 0}	181659	1755078960028
1755182131869	{"Total": 100, "IsCustom": false, "Material": "material1", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub1", "AdditionalDiscount": 0}	836278	1755182017435
1755182131876	{"Total": 100, "IsCustom": false, "Material": "material2", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	836278	1755182036311
1755187556919	{"Total": 100, "IsCustom": false, "Material": "material2", "Quantity": 10, "UnitPrice": 10, "Description": "huhu", "SubCategory": "sub2", "AdditionalDiscount": 0}	034042	1755187347060
1755186573946	{"Total": 1000, "IsCustom": true, "Material": "asfsadf", "Quantity": 100, "UnitPrice": 10, "Description": "", "SubCategory": "sdfasdf", "AdditionalDiscount": 0}	874354	1755186567302
1755188223764	{"Total": 300, "IsCustom": false, "Material": "Material-6", "Quantity": 10, "UnitPrice": 30, "Description": "test", "SubCategory": "dgfhgj", "AdditionalDiscount": 0}	677087	1755188111178
1755188237854	{"Total": 100, "IsCustom": true, "Material": "test", "Quantity": 10, "UnitPrice": 10, "Description": "kjlkj", "SubCategory": "yty", "AdditionalDiscount": 0}	677087	1755188111178
1755188223811	{"Total": 200, "IsCustom": false, "Material": "material-7", "Quantity": 10, "UnitPrice": 20, "Description": "kjlj;ok;pl", "SubCategory": "kjjkdsjdk", "AdditionalDiscount": 0}	677087	1755188119760
1755188262867	{"Total": 868, "IsCustom": true, "Material": "test", "Quantity": 124, "UnitPrice": 7, "Description": "123", "SubCategory": "123", "AdditionalDiscount": 0}	677087	1755188119760
1755188642419	{"Total": 200, "IsCustom": false, "Material": "Material-4", "Quantity": 10, "UnitPrice": 20, "Description": "test", "SubCategory": "dhhds", "AdditionalDiscount": 0}	419955	1755188516761
1755188655756	{"Total": 780, "IsCustom": true, "Material": "test123", "Quantity": 10, "UnitPrice": 78, "Description": "uiui", "SubCategory": "gtg", "AdditionalDiscount": 0}	419955	1755188516761
1755188642406	{"Total": 1230, "IsCustom": false, "Material": "Material-3", "Quantity": 123, "UnitPrice": 10, "Description": "", "SubCategory": "aub", "AdditionalDiscount": 0}	419955	1755188525388
1755189245398	{"Total": 30, "IsCustom": false, "Material": "Material-5", "Quantity": 3, "UnitPrice": 10, "Description": "", "SubCategory": "", "AdditionalDiscount": 0}	542229	1755189048797
1755194755698	{"Total": 10, "IsCustom": false, "Material": "Material-3", "Quantity": 1, "UnitPrice": 10, "Description": "", "SubCategory": "aub", "AdditionalDiscount": 0}	009177	1755194697630
1755195918559	{"Total": 100, "IsCustom": false, "Material": "material2", "Quantity": 10, "UnitPrice": 10, "Description": "123123", "SubCategory": "sub2", "AdditionalDiscount": 0}	000848	1755195759973
1755195930158	{"Total": 90, "IsCustom": true, "Material": "123222", "Quantity": 10, "UnitPrice": 9, "Description": "2222", "SubCategory": "2222", "AdditionalDiscount": 0}	000848	1755195759973
1755206254894	{"Total": 20, "IsCustom": false, "Material": "material2", "Quantity": 2, "UnitPrice": 10, "Description": "test1", "SubCategory": "sub2", "AdditionalDiscount": 0}	152654	1755205625417
1755206310883	{"Total": 1000, "IsCustom": true, "Material": "name", "Quantity": 10, "UnitPrice": 100, "Description": "", "SubCategory": "", "AdditionalDiscount": 0}	152654	1755205625417
1755207058632	{"Total": 20, "IsCustom": false, "Material": "material2", "Quantity": 2, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	295267	1755207015294
1755590758895	{"Total": 200, "IsCustom": false, "Material": "Material-4", "Quantity": 10, "UnitPrice": 20, "Description": "", "SubCategory": "dhhds", "AdditionalDiscount": 0}	066615	1755586676753
1755590850869	{"Total": 100, "IsCustom": true, "Material": "iron", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "good", "AdditionalDiscount": 0}	066615	1755586676753
1755591112791	{"Total": 400, "IsCustom": false, "Material": "Material-4", "Quantity": 20, "UnitPrice": 20, "Description": "", "SubCategory": "dhhds", "AdditionalDiscount": 0}	045813	1755586444861
1755591141836	{"Total": 600, "IsCustom": true, "Material": "steel", "Quantity": 20, "UnitPrice": 30, "Description": "", "SubCategory": "good", "AdditionalDiscount": 0}	045813	1755586444861
1755769644365	{"Total": 600, "IsCustom": false, "Material": "Material-4", "Quantity": 30, "UnitPrice": 20, "Description": "", "SubCategory": "dhhds", "AdditionalDiscount": 0}	585455	1755769497398
1755769655913	{"Total": 300, "IsCustom": true, "Material": "steel", "Quantity": 10, "UnitPrice": 30, "Description": "", "SubCategory": "good", "AdditionalDiscount": 0}	585455	1755769497398
1755770088306	{"Total": 400, "IsCustom": false, "Material": "material-7", "Quantity": 20, "UnitPrice": 20, "Description": "", "SubCategory": "kjjkdsjdk", "AdditionalDiscount": 0}	585455	1755769919280
1755770098299	{"Total": 200, "IsCustom": true, "Material": "wood", "Quantity": 20, "UnitPrice": 10, "Description": "", "SubCategory": "good", "AdditionalDiscount": 0}	585455	1755769919280
1755830549364	{"Total": 100, "IsCustom": false, "Material": "Material-3", "Quantity": 10, "UnitPrice": 10, "Description": "test123", "SubCategory": "aub", "AdditionalDiscount": 0}	204729	1755830408667
1755830561012	{"Total": 100, "IsCustom": true, "Material": "test123", "Quantity": 10, "UnitPrice": 10, "Description": "descrip", "SubCategory": "hey", "AdditionalDiscount": 0}	204729	1755830408667
1755831706456	{"Total": 40, "IsCustom": false, "Material": "material-7", "Quantity": 2, "UnitPrice": 20, "Description": "test", "SubCategory": "kjjkdsjdk", "AdditionalDiscount": 0}	211806	1755831706472
1755831714973	{"Total": 100, "IsCustom": true, "Material": "test", "Quantity": 10, "UnitPrice": 10, "Description": "1234", "SubCategory": "test", "AdditionalDiscount": 0}	211806	1755831706472
1755832315997	{"Total": 200, "IsCustom": false, "Material": "Material-4", "Quantity": 10, "UnitPrice": 20, "Description": "test1234", "SubCategory": "dhhds", "AdditionalDiscount": 0}	106485	1755832197409
1755832318753	{"Total": 100, "IsCustom": true, "Material": "test", "Quantity": 10, "UnitPrice": 10, "Description": "1234", "SubCategory": "1234", "AdditionalDiscount": 0}	106485	1755832197409
1755832316025	{"Total": 1300, "IsCustom": false, "Material": "material-7", "Quantity": 10, "UnitPrice": 130, "Description": "test123", "SubCategory": "kjjkdsjdk", "AdditionalDiscount": 0}	106485	1755832205460
1755856420458	{"Total": 100, "IsCustom": false, "Material": "material2", "Quantity": 10, "UnitPrice": 10, "Description": "", "SubCategory": "sub2", "AdditionalDiscount": 0}	378895	1755856420394
\.


--
-- Data for Name: material_templates; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.material_templates (id, name, manager_residential_entry_id, manager_residential_type_id, category, sub_category) FROM stdin;
2	material2	0	0	Kitchen	[{"cost": "10", "name": "sub2", "pricing_type": "NEW"}, {"cost": "20", "name": "sub2", "pricing_type": "REGULAR"}, {"cost": "30", "name": "sub2", "pricing_type": "PRO"}, {"cost": "40", "name": "sub2", "pricing_type": "BEST"}]
4	Material-3	0	0	Master Bedroom	[{"cost": "10", "name": "aub", "pricing_type": "NEW"}, {"cost": "20", "name": "aub", "pricing_type": "REGULAR"}, {"cost": "30", "name": "aub", "pricing_type": "PRO"}, {"cost": "40", "name": "aub", "pricing_type": "BEST"}]
5	Material-4	0	0	New Corridor	[{"cost": "20", "name": "dhhds", "pricing_type": "NEW"}, {"cost": "30", "name": "dhhds", "pricing_type": "REGULAR"}, {"cost": "40", "name": "dhhds", "pricing_type": "PRO"}, {"cost": "50", "name": "dhhds", "pricing_type": "BEST"}]
6	Material-5	0	0	Bathroom	[{"cost": "10", "name": "snd", "pricing_type": "NEW"}, {"cost": "20", "name": "snd", "pricing_type": "REGULAR"}, {"cost": "30", "name": "snd", "pricing_type": "PRO"}, {"cost": "40", "name": "snd", "pricing_type": "BEST"}]
7	Material-6	0	0	Master Bathroom	[{"cost": "30", "name": "dgfhgj", "pricing_type": "NEW"}, {"cost": "40", "name": "dgfhgj", "pricing_type": "REGULAR"}, {"cost": "50", "name": "dgfhgj", "pricing_type": "PRO"}, {"cost": "60", "name": "dgfhgj", "pricing_type": "BEST"}]
8	material-7	0	0	Other	[{"cost": "20", "name": "kjjkdsjdk", "pricing_type": "NEW"}, {"cost": "30", "name": "kjjkdsjdk", "pricing_type": "REGULAR"}, {"cost": "40", "name": "kjjkdsjdk", "pricing_type": "PRO"}, {"cost": "50", "name": "kjjkdsjdk", "pricing_type": "BEST"}]
9	material-9	0	0	Bathroom	[{"cost": "20", "name": "good", "pricing_type": "NEW"}, {"cost": "40", "name": "good", "pricing_type": "REGULAR"}, {"cost": "60", "name": "good", "pricing_type": "PRO"}, {"cost": "80", "name": "good", "pricing_type": "BEST"}]
1	material1	0	0	Kitchen	[{"cost": "10", "name": "sub1", "pricing_type": "NEW"}, {"cost": "35", "name": "sub1", "pricing_type": "REGULAR"}, {"cost": "20", "name": "sub1", "pricing_type": "PRO"}, {"cost": "30", "name": "sub1", "pricing_type": "BEST"}]
\.


--
-- Data for Name: pickup_locations; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pickup_locations (id, name, street, unit_suite, city, state, zip, country, created_at, updated_at, deleted_at) FROM stdin;
1	Downtown store	456 Market Street	\N	Miami	Florida	62701	USA	\N	\N	\N
2	Main Warehouse	123 Warehouse Blvd	\N	Tampa	Florida	62704	USA	\N	\N	\N
3	East Side Location	789 East Avenue	\N	Orlando	Florida	62702	USA	\N	\N	\N
6	West Side Location	sun street	sdhbjsdbhjdshbj	miami	Illinois	12354	USA	0001-01-01 00:00:00+00	2025-08-12 10:59:58.543116+00	\N
7	abcse	dns	sadjkfn	xyz	Idaho	12345	USA	0001-01-01 00:00:00+00	2025-08-19 11:21:24.914299+00	2025-08-19 11:21:44.113443+00
\.


--
-- Data for Name: residential_entries; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.residential_entries (id, manager_residential_entry_id, manager_residential_type_id, details, lead_id, estimate_id) FROM stdin;
1754925139227	1	1	asdfasdf	153714	005640
1754974597792	4	2	asdasfa	573182	846594
1754925352073	1	1	asdfasdf	338697	828551
1754925352074	2	1	asdfasdf	338697	828551
1755042280869	2	2	Bathroom	253703	575388
1754989115005	2	3	dsb	973844	720330
1754925822301	1	1	asdfsadf	278004	138115
1754925822302	2	1	asdfasdf	278004	\N
1754976097604	3	2	sdsgf	684877	359478
1754976307915	4	2	sadfd	684877	359478
1754976459819	1	4	asdfasfd	684877	359478
1754930429840	2	3	asdf	379122	299571
1754931270237	3	2	asasd	334790	\N
1754989105243	4	2	kjd	973844	720330
1754932059205	1	1	asdfasdf	990977	449661
1754932059206	1	2	asdfasdf	990977	449661
1754990000977	1	2	hfasjfsdad	554098	416334
1754991678168	2	1	sfsadafd	297805	604666
1754933591840	2	1	wqerasdf	087879	440111
1754933591841	1	1	asdfasdf	087879	\N
1754933789174	2	2	asdfasdf	087879	\N
1754991834726	1	1	sdfasfd	820700	640014
1754991841419	2	2	asdfaf	820700	640014
1754934291375	1	1	asdfasdf	893916	\N
1754992010350	1	2	asdfasdf	440604	436116
1754934291376	3	2	sddsk	788865	104900
1754934291377	1	3	sadfg	788865	104900
1754934291378	5	4	dsfdfgd	788865	\N
1754931688619	3	2	ADSC	764542	348917
1754934206559	2	1	asdfasdf	893916	761845
1754939263846	2	1	asdfadsf	028735	602959
1754939552481	1	1	sadfad	945419	198231
1754939774446	1	1	asdfasdf	756697	466116
1754940197228	1	1	asdfasdf	198032	791963
1754940197229	3	2	hadhdajk	512161	947144
1754942775621	1	1	asdfasdf	966384	468618
1754942781800	2	2	asdfasdf	966384	468618
1754942998444	1	1	asdfsafd	553579	189986
1754943191639	1	1	asdfsadf	971784	260462
1754944823902	1	1	sfdasdf	920579	659728
1754945436867	1	1	SDFASDFSDF	571272	112352
1754944171822	1	1	asdfasdf	115148	790224
1754945304315	1	1	asdfsadf	078083	811282
1754973897203	3	2	adsdnklsa	573182	846594
1754989321905	2	1	asdfasdf	078083	\N
1754992172408	2	2	asdfaf	440604	436116
1754980164865	4	2	asdcdad	022652	365344
1754980523379	3	2	hdhj	022652	365344
1754980623793	4	3	jhsdfg	022652	365344
1755043530813	2	3	Bathroom	317521	800426
1755149396910	5	2	sersehdj	616587	276098
1754987522866	3	2	sdbashbkf	884011	153822
1754988418112	4	2	kskhdahs	884011	153822
1754988712647	4	3	skfshb	884011	153822
1754989461769	6	1	casdfsfd	884011	153822
1755044598688	3	2	MB	022137	141628
1755044605893	3	3	hey	022137	141628
1755045602331	3	3	MB	323585	367470
1754982366867	2	2	sjkdh	766166	849736
1754982511168	4	2	hsfdbsdhj	766166	849736
1754983629610	2	1	dasfasd	766166	849736
1754984525695	3	2	asdfasdf	766166	849736
1754981655078	3	2	sdkfsdhj	304586	941746
1754989034701	2	1	this is test	301498	307848
1754989043654	3	2	this is bathroom	301498	307848
1755046904924	4	2	MB	148872	491094
1755046911899	6	2	oather	148872	491094
1755003067973	3	2	zshghgvhg	858123	096722
1754992910045	2	3	sdfs	701868	561785
1754993027495	2	2	werwer	701868	561785
1754993468042	3	4	sdfsa	701868	561785
1754993602935	3	1	SDKJ	701868	561785
1755047241451	2	1	ssss	688174	296995
1755047247076	4	2	ssss	688174	296995
1754995164876	3	2	fdxgfchvh	487281	549052
1754995393376	6	2	dfxgchj	487281	549052
1755004271757	3	3	zdfsda	487281	549052
1755004409938	4	4	ksdfjks	487281	549052
1755040181565	1	1	Kitchen	467243	420632
1755159079427	3	2	asdf	516289	\N
1755047850236	1	2	gggg	341276	221078
1755150932103	3	3	efk	747298	878573
1755111461136	4	4	kajdfdsa	535395	215051
1755111628554	1	4	sjdhkl	535395	215051
1755138100235	3	3	test	249777	764222
1755110639937	3	3	dkhas	535395	215051
1755110649635	5	1	ahsbjd	535395	215051
1755177554853	2	1	sdafasdf	244629	328348
1755177558489	3	1	asdfsaf	244629	328348
1755056329204	3	3	shdhfbd	870087	181659
1755078794261	3	2	jhadb	870087	181659
1755078960028	3	4	ajkdjk	870087	181659
1755174329782	2	1	asdfasdf	870087	181659
1755186887688	1	2	testing	806460	422170
1755182017435	1	1	gjkhhkjh	110545	836278
1755182036311	2	2	hgjhgj	110545	836278
1755185902660	3	2	residential bathroom	043649	874354
1755186352303	2	2	sdfsdf	043649	874354
1755186567302	4	2	asdfsadf	043649	874354
1755187347060	2	2	test	496909	034042
1755188111178	2	2	bathroom	484015	677087
1755188119760	4	4	mb	484015	677087
1755188516761	3	2	mb	653535	419955
1755188525388	5	4	nc	653535	419955
1755189048797	3	2	mb	255076	542229
1755191154295	1	1	sdasdf	690103	\N
1755194697630	2	1	asdf	220509	009177
1755195759973	2	1	test	076315	000848
1755205625417	2	2	Bathroom	598530	152654
1755205655662	3	3	MB	598530	152654
1755207015294	2	1	test	605988	295267
1755537108021	3	2	jzchhjsa	594353	\N
1755538080609	4	2	kjsdhfs	979672	\N
1755584868899	3	2	hjbsad	445169	\N
1755585917372	4	3	sdsjk	418495	\N
1755586104268	4	2	skjdvsk	956125	\N
1755586676753	4	3	jdkf	818324	066615
1755689294052	3	2	jvkvgkvg	822302	\N
1755586444861	4	2	jbdhcsa	673065	045813
1755782380341	2	3	kjvsj	914196	\N
1755769497398	3	2	hjsgdsa	660420	585455
1755769919280	5	3	sadfdfh	660420	585455
1755830408667	2	2	Bathroom	888866	204729
1755831198898	2	2	test	629876	\N
1755831206626	3	3	test	629876	\N
1755831706472	3	3	test	556657	211806
1755832197409	3	3	CMB	652375	106485
1755832205460	3	2	RMB	652375	106485
1755832729219	3	4	amb	225983	\N
1755852182868	3	2	sadhfg	580823	100565
1755855741072	3	3	kjd	235859	\N
1755856420394	1	1	asasdasd	142588	378895
1755860786021	2	2	dcs	692681	\N
\.


--
-- Data for Name: service_categories; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.service_categories (id, name) FROM stdin;
1	Cleaning kit
2	Polishing kit
6	Cutting
\.


--
-- Data for Name: service_groups; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.service_groups (id, name) FROM stdin;
1	Cleaning
2	Polishing
6	cutting
\.




--
-- Data for Name: service_rows; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.service_rows (id, details, estimate_id, residential_entry_id) FROM stdin;
1754925203332	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	005640	1754925139227
1754925203357	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	005640	1754925139227
1754925220487	{"Name": "asdfasdf", "Total": 10, "IsCustom": true, "UnitPrice": 10, "AdditionalDiscount": 0}	005640	1754925139227
1754925345979	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	828551	1754925345946
1754925345990	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	828551	1754925345946
1754925366895	{"Name": "asdfasdf", "Total": 10, "IsCustom": true, "UnitPrice": 10, "AdditionalDiscount": 0}	828551	1754925345946
1754925352044	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	828551	1754925352072
1754925352068	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	828551	1754925352072
1754925384244	{"Name": "adsfasdf", "Total": 10, "IsCustom": true, "UnitPrice": 10, "AdditionalDiscount": 0}	828551	1754925352072
1754925822250	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	138115	1754925822298
1754925822341	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	138115	1754925822298
1754930592184	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	299571	1754930429838
1754930592228	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	299571	1754930429838
1754932054579	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	449661	1754932054536
1754932054522	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	449661	1754932054536
1754932072452	{"Name": "fasdfsd", "Total": 10, "IsCustom": true, "UnitPrice": 10, "AdditionalDiscount": 0}	449661	1754932054536
1754932059248	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	449661	1754932059204
1754932059269	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	449661	1754932059204
1754932087452	{"Name": "service1", "Total": 10, "IsCustom": false, "UnitPrice": 10, "AdditionalDiscount": 0}	449661	1754932059204
1754932092557	{"Name": "asdfasdf", "Total": 10, "IsCustom": true, "UnitPrice": 10, "AdditionalDiscount": 0}	449661	1754932059204
1754933591877	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	440111	1754933591839
1754933591831	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	440111	1754933591839
1754936584831	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	348917	1754931688619
1754936612443	{"Name": "cad", "Total": 20, "IsCustom": true, "UnitPrice": 20, "AdditionalDiscount": 0}	348917	1754931688619
1754934514395	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	104900	1754934118795
1754934550639	{"Name": "fitting", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	104900	1754934118795
1754934746875	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	104900	1754934746920
1754934746910	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	104900	1754934746920
1754934746885	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	104900	1754934746920
1754934957993	{"Name": "template", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	104900	1754934746920
1754934972828	{"Name": "cutting", "Total": 50, "IsCustom": true, "UnitPrice": 50, "AdditionalDiscount": 0}	104900	1754934746920
1754936584822	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	348917	1754931688619
1754936584804	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	348917	1754931688619
1754934206591	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	761845	1754934206559
1754934206580	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	761845	1754934206559
1754934206524	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	761845	1754934206559
1754934220866	{"Name": "asdfasdf", "Total": 10, "IsCustom": true, "UnitPrice": 10, "AdditionalDiscount": 0}	761845	1754934206559
1754937684959	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	761845	1754934291375
1754937686341	{"Name": "asdfasdf", "Total": 1, "IsCustom": true, "UnitPrice": 1, "AdditionalDiscount": 0}	761845	1754934291375
1754939263834	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	602959	1754939263846
1754939263816	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	602959	1754939263846
1754939263832	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	602959	1754939263846
1754939552520	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	198231	1754939552481
1754939552510	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	198231	1754939552481
1754939552482	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	198231	1754939552481
1754939774503	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	466116	1754939774446
1754939774443	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	466116	1754939774446
1754939774450	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	466116	1754939774446
1754939980020	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	947144	1754939980082
1754939980066	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	947144	1754939980082
1754940034301	{"Name": "CAD", "Total": 20, "IsCustom": true, "UnitPrice": 20, "AdditionalDiscount": 0}	947144	1754939980082
1754940197167	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	791963	1754940197228
1754940197174	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	791963	1754940197228
1754940197223	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	791963	1754940197228
1754942775623	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	468618	1754942775621
1754942775634	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	468618	1754942775621
1754942775581	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	468618	1754942775621
1754942781774	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	468618	1754942781800
1754942781773	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	468618	1754942781800
1754942781776	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	468618	1754942781800
1754942998410	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	189986	1754942998444
1754942998405	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	189986	1754942998444
1754942998460	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	189986	1754942998444
1754943191608	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	260462	1754943191639
1754943191604	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	260462	1754943191639
1754943191687	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	260462	1754943191639
1754944171749	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	790224	1754944171822
1754944171753	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	790224	1754944171822
1754944171802	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	790224	1754944171822
1754944823872	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	659728	1754944823902
1754944823923	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	659728	1754944823902
1754944823878	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	659728	1754944823902
1754945436942	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	112352	1754945436867
1754945436952	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	112352	1754945436867
1754945436943	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	112352	1754945436867
1754945304351	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	811282	1754945304315
1754945304350	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	811282	1754945304315
1754974798732	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	846594	1754974597792
1754987701265	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	941746	1754981655078
1754974803719	{"Name": "cutting", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	846594	1754974597792
1754974182999	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	846594	1754973897203
1754974220979	{"Name": "Template", "Total": 50, "IsCustom": true, "UnitPrice": 50, "AdditionalDiscount": 0}	846594	1754973897203
1754987701231	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	941746	1754981655078
1754987701294	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	941746	1754981655078
1754987761142	{"Name": "cutting", "Total": 20, "IsCustom": true, "UnitPrice": 20, "AdditionalDiscount": 0}	941746	1754981655078
1754980424138	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	365344	1754980164865
1754980835698	{"Name": "service1", "Total": 10, "IsCustom": false, "UnitPrice": 10, "AdditionalDiscount": 0}	365344	1754980164865
1754976097673	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	359478	1754976097604
1754976133142	{"Name": "cutting", "Total": 20, "IsCustom": true, "UnitPrice": 20, "AdditionalDiscount": 0}	359478	1754976097604
1754976307941	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	359478	1754976307915
1754976307935	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	359478	1754976307915
1754976338253	{"Name": "installation", "Total": 50, "IsCustom": true, "UnitPrice": 50, "AdditionalDiscount": 0}	359478	1754976307915
1754976532411	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	359478	1754976459819
1754980453676	{"Name": "cutting", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	365344	1754980164865
1754980841573	{"Name": "installation", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	365344	1754980164865
1754980523426	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	365344	1754980523379
1754980561692	{"Name": "fitting", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	365344	1754980523379
1754982366906	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	849736	1754982366867
1754982620349	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	849736	1754982366867
1754982401354	{"Name": "template", "Total": 10, "IsCustom": true, "UnitPrice": 10, "AdditionalDiscount": 0}	849736	1754982366867
1754982625077	{"Name": "cutting", "Total": 20, "IsCustom": true, "UnitPrice": 20, "AdditionalDiscount": 0}	849736	1754982366867
1754980730731	{"Name": "service1", "Total": 10, "IsCustom": false, "UnitPrice": 10, "AdditionalDiscount": 0}	365344	1754980623793
1754980739097	{"Name": "CAD & CUT", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	365344	1754980623793
1754982511169	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	849736	1754982511168
1754982511201	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	849736	1754982511168
1754982554761	{"Name": "fitting", "Total": 40, "IsCustom": true, "UnitPrice": 40, "AdditionalDiscount": 0}	849736	1754982511168
1754984525698	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	849736	1754984525695
1754984525693	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	849736	1754984525695
1754984525768	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	849736	1754984525695
1754987876166	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	153822	1754987522866
1754990067750	{"Name": "good service", "Total": 12, "IsCustom": true, "UnitPrice": 12, "AdditionalDiscount": 0}	416334	1754990000977
1754989332209	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	307848	1754989043654
1754989406938	{"Name": "good service", "Total": 20, "IsCustom": true, "UnitPrice": 20, "AdditionalDiscount": 0}	307848	1754989043654
1754989395448	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	720330	1754989115005
1754989395373	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	720330	1754989115005
1754989523976	{"Name": "fitting", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	720330	1754989115005
1754989395388	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	720330	1754989105243
1754991678199	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	604666	1754991678168
1754987876232	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	153822	1754987522866
1754987912080	{"Name": "cutting", "Total": 50, "IsCustom": true, "UnitPrice": 50, "AdditionalDiscount": 0}	153822	1754987522866
1754988521545	{"Name": "service1", "Total": 10, "IsCustom": false, "UnitPrice": 10, "AdditionalDiscount": 0}	153822	1754988418112
1754988712620	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	153822	1754988712647
1754988712595	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	153822	1754988712647
1754988712607	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	153822	1754988712647
1754988747267	{"Name": "fabrication", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	153822	1754988712647
1754989521018	{"Name": "service1", "Total": 10, "IsCustom": false, "UnitPrice": 10, "AdditionalDiscount": 0}	153822	1754989461769
1754989395407	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	720330	1754989105243
1754989921332	{"Name": "service1", "Total": 10, "IsCustom": false, "UnitPrice": 10, "AdditionalDiscount": 0}	720330	1754989105243
1754989439771	{"Name": "template", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	720330	1754989105243
1754990001030	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	416334	1754990000977
1754990000975	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	416334	1754990000977
1754990001021	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	416334	1754990000977
1754991678189	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	604666	1754991678168
1754991678228	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	604666	1754991678168
1754991948611	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	640014	1754991834726
1754991948603	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	640014	1754991834726
1754991948656	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	640014	1754991834726
1754991948698	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	640014	1754991841419
1754989395449	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	720330	1754989115005
1754989332288	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	307848	1754989034701
1754989332257	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	307848	1754989034701
1754989332262	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	307848	1754989043654
1754989332218	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	307848	1754989043654
1754991948666	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	640014	1754991841419
1754991948602	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	640014	1754991841419
1754992141166	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	436116	1754992010350
1754992141163	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	436116	1754992010350
1754992141226	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	436116	1754992010350
1754992158013	{"Name": "cusotm", "Total": 10, "IsCustom": true, "UnitPrice": 10, "AdditionalDiscount": 0}	436116	1754992010350
1754992172378	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	436116	1754992172408
1754992172426	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	436116	1754992172408
1754992172452	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	436116	1754992172408
1754992192920	{"Name": "hello", "Total": 10, "IsCustom": true, "UnitPrice": 10, "AdditionalDiscount": 0}	436116	1754992172408
1754993406967	{"Name": "fabrication", "Total": 50, "IsCustom": true, "UnitPrice": 50, "AdditionalDiscount": 0}	561785	1754993027495
1754993468088	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	561785	1754993468042
1754993468007	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	561785	1754993468042
1754993468037	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	561785	1754993468042
1754993510563	{"Name": "CAD", "Total": 40, "IsCustom": true, "UnitPrice": 40, "AdditionalDiscount": 0}	561785	1754993468042
1754993734523	{"Name": "service1", "Total": 10, "IsCustom": false, "UnitPrice": 10, "AdditionalDiscount": 0}	561785	1754993602935
1754993738162	{"Name": "installation", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	561785	1754993602935
1755043530802	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	800426	1755043530813
1754995164875	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	549052	1754995164876
1754995164921	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	549052	1754995164876
1754995209676	{"Name": "CAD", "Total": 50, "IsCustom": true, "UnitPrice": 50, "AdditionalDiscount": 0}	549052	1754995164876
1754995536731	{"Name": "service2", "Total": 10, "IsCustom": false, "UnitPrice": 10, "AdditionalDiscount": 0}	549052	1754995393376
1754995544552	{"Name": "fabrication", "Total": 60, "IsCustom": true, "UnitPrice": 60, "AdditionalDiscount": 0}	549052	1754995393376
1755043530799	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	800426	1755043530813
1755043530803	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	800426	1755043530813
1754993316602	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	561785	1754992910045
1754993316541	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	561785	1754992910045
1754993354019	{"Name": "installation", "Total": 20, "IsCustom": true, "UnitPrice": 20, "AdditionalDiscount": 0}	561785	1754992910045
1754993316538	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	561785	1754993027495
1755004350735	{"Name": "service1", "Total": 10, "IsCustom": false, "UnitPrice": 10, "AdditionalDiscount": 0}	549052	1755004271757
1755004355334	{"Name": "cutting", "Total": 40, "IsCustom": true, "UnitPrice": 40, "AdditionalDiscount": 0}	549052	1755004271757
1755004409957	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	549052	1755004409938
1755004409959	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	549052	1755004409938
1755004442727	{"Name": "CAD", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	549052	1755004409938
1755044680279	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	141628	1755044598688
1755042725026	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	575388	1755042280869
1755042725030	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	575388	1755042280869
1755044680257	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	141628	1755044598688
1755044680224	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	141628	1755044598688
1755042725031	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	575388	1755042280869
1755044680296	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	141628	1755044605893
1755044680316	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	141628	1755044605893
1755044680272	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	141628	1755044605893
1755045701758	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	367470	1755045602331
1755045701810	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	367470	1755045602331
1755045701812	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	367470	1755045602331
1755047005190	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	491094	1755046904924
1755047005209	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	491094	1755046904924
1755047005263	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	491094	1755046904924
1755047005213	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	491094	1755046911899
1755047005181	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	491094	1755046911899
1755047005200	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	491094	1755046911899
1755047065541	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	096722	1755003067973
1755047065569	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	096722	1755003067973
1755047065521	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	096722	1755003067973
1755047407474	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	296995	1755047241451
1755047407488	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	296995	1755047241451
1755047407521	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	296995	1755047241451
1755047407519	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	296995	1755047247076
1755047407534	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	296995	1755047247076
1755047407515	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	296995	1755047247076
1755047850256	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	221078	1755047850236
1755047850264	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	221078	1755047850236
1755047850262	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	221078	1755047850236
1755056528416	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	181659	1755056329204
1755111102416	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	215051	1755110639937
1755111146792	{"Name": "CAD", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	215051	1755110639937
1755111158887	{"Name": "cutting", "Total": 60, "IsCustom": true, "UnitPrice": 60, "AdditionalDiscount": 0}	215051	1755110639937
1755056528357	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	181659	1755056329204
1755056558364	{"Name": "cutting", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	181659	1755056329204
1755078794184	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	181659	1755078794261
1755078794199	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	181659	1755078794261
1755078823680	{"Name": "fabrication", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	181659	1755078794261
1755078959984	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	181659	1755078960028
1755079002311	{"Name": "installation", "Total": 50, "IsCustom": true, "UnitPrice": 50, "AdditionalDiscount": 0}	181659	1755078960028
1755111102391	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	215051	1755110649635
1755111102418	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	215051	1755110649635
1755111102462	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	215051	1755110649635
1755111743987	{"Name": "fabrication", "Total": 60, "IsCustom": true, "UnitPrice": 60, "AdditionalDiscount": 0}	215051	1755110649635
1755111461127	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	215051	1755111461136
1755111461166	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	215051	1755111461136
1755111497925	{"Name": "installation", "Total": 40, "IsCustom": true, "UnitPrice": 40, "AdditionalDiscount": 0}	215051	1755111461136
1755111628563	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	215051	1755111628554
1755111628538	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	215051	1755111628554
1755111667161	{"Name": "template", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	215051	1755111628554
1755138174210	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	764222	1755138100235
1755138174159	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	764222	1755138100235
1755138174129	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	764222	1755138100235
1755150310983	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	276098	1755149396910
1755150311065	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	276098	1755149396910
1755150341774	{"Name": "fabrication", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	276098	1755149396910
1755151465146	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	878573	1755150932103
1755151465156	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	878573	1755150932103
1755151492097	{"Name": "cutting", "Total": 100, "IsCustom": true, "UnitPrice": 100, "AdditionalDiscount": 0}	878573	1755150932103
1755174329843	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	181659	1755174329782
1755174329777	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	181659	1755174329782
1755174329811	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	181659	1755174329782
1755174362103	{"Name": "asdfsadf", "Total": 10, "IsCustom": true, "UnitPrice": 10, "AdditionalDiscount": 0}	181659	1755174329782
1755177554801	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	328348	1755177554853
1755177554810	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	328348	1755177554853
1755177554819	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	328348	1755177554853
1755186198351	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	874354	1755185902660
1755186198290	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	874354	1755185902660
1755186198337	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	874354	1755185902660
1755182131880	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	836278	1755182017435
1755182131822	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	836278	1755182017435
1755182131856	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	836278	1755182017435
1755182131872	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	836278	1755182036311
1755182131843	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	836278	1755182036311
1755182131820	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	836278	1755182036311
1755182194063	{"Name": "hfgjhfhgf", "Total": 11, "IsCustom": true, "UnitPrice": 11, "AdditionalDiscount": 0}	836278	1755182036311
1755194755785	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	009177	1755194697630
1755194755697	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	009177	1755194697630
1755186223383	{"Name": "sfasfdf", "Total": 1, "IsCustom": true, "UnitPrice": 1, "AdditionalDiscount": 0}	874354	1755185902660
1755186352321	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	874354	1755186352303
1755186352305	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	874354	1755186352303
1755186352241	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	874354	1755186352303
1755186555620	{"Name": "asdfsadf", "Total": 10, "IsCustom": true, "UnitPrice": 10, "AdditionalDiscount": 0}	874354	1755186352303
1755186567281	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	874354	1755186567302
1755186567349	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	874354	1755186567302
1755186567343	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	874354	1755186567302
1755187042363	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	422170	1755186887688
1755187042319	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	422170	1755186887688
1755187042347	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	422170	1755186887688
1755188223833	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	677087	1755188111178
1755188223749	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	677087	1755188111178
1755188223765	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	677087	1755188111178
1755188223769	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	677087	1755188119760
1755188223799	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	677087	1755188119760
1755188223774	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	677087	1755188119760
1755188642374	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	419955	1755188516761
1755188642390	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	419955	1755188516761
1755188642382	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	419955	1755188516761
1755188642416	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	419955	1755188525388
1755188642414	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	419955	1755188525388
1755188642438	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	419955	1755188525388
1755194755711	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	009177	1755194697630
1755195918549	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	000848	1755195759973
1755195918601	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	000848	1755195759973
1755195918543	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	000848	1755195759973
1755206254892	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	152654	1755205625417
1755206511204	{"Name": "new service", "Total": 1, "IsCustom": true, "UnitPrice": 1, "AdditionalDiscount": 0}	152654	1755205625417
1755206254881	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	152654	1755205655662
1755590758880	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	066615	1755586676753
1755590758817	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	066615	1755586676753
1755590921803	{"Name": "fabrication", "Total": 10, "IsCustom": true, "UnitPrice": 10, "AdditionalDiscount": 0}	066615	1755586676753
1755591112793	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	045813	1755586444861
1755591112868	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	045813	1755586444861
1755591160051	{"Name": "CAD", "Total": 20, "IsCustom": true, "UnitPrice": 20, "AdditionalDiscount": 0}	045813	1755586444861
1755830585868	{"Name": "service", "Total": 10, "IsCustom": true, "UnitPrice": 10, "AdditionalDiscount": 0}	204729	1755830408667
1755831706512	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	211806	1755831706472
1755831706502	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	211806	1755831706472
1755831706506	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	211806	1755831706472
1755831706423	{"Name": "Service 4", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	211806	1755831706472
1755831752842	{"Name": "test", "Total": 10, "IsCustom": true, "UnitPrice": 10, "AdditionalDiscount": 0}	211806	1755831706472
1755769644397	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	585455	1755769497398
1755769644312	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	585455	1755769497398
1755769696744	{"Name": "fabrication", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	585455	1755769497398
1755770188679	{"Name": "service1", "Total": 10, "IsCustom": false, "UnitPrice": 10, "AdditionalDiscount": 0}	585455	1755769919280
1755770194739	{"Name": "templating", "Total": 30, "IsCustom": true, "UnitPrice": 30, "AdditionalDiscount": 0}	585455	1755769919280
1755830549363	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	204729	1755830408667
1755830549419	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	204729	1755830408667
1755830549377	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	204729	1755830408667
1755830549406	{"Name": "Service 4", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	204729	1755830408667
1755832316015	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	106485	1755832197409
1755832316025	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	106485	1755832197409
1755832316030	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	106485	1755832197409
1755832316068	{"Name": "Service 4", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	106485	1755832197409
1755832316017	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	106485	1755832205460
1755832316029	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	106485	1755832205460
1755832316009	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	106485	1755832205460
1755832316000	{"Name": "Service 4", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	106485	1755832205460
1755853537713	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	100565	1755852182868
1755853537696	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	100565	1755852182868
1755853537754	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	100565	1755852182868
1755853537726	{"Name": "Service 4", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	100565	1755852182868
1755856420447	{"Name": "service1", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	378895	1755856420394
1755856420407	{"Name": "service3", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	378895	1755856420394
1755856420457	{"Name": "Service 4", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	378895	1755856420394
1755856420406	{"Name": "service2", "Total": 20, "IsCustom": false, "UnitPrice": 20, "AdditionalDiscount": 0}	378895	1755856420394
\.




--
-- Data for Name: services; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.services (id, details) FROM stdin;
2	{"SKU": "", "Name": "service2", "Notes": "", "NewAcc": 20, "ProAcc": 10, "BasicUOM": "", "IsTaxable": false, "ExpenseAcc": 20, "PriceRange": "", "RegularAcc": 30, "AvgUnitCost": 20, "SalesAccount": "", "service_type": {"ID": 2, "Name": "Polishing"}, "service_Group": {"ID": 1, "Name": "Cleaning"}, "service_category": {"ID": 2, "Name": "Polishing kit"}, "SpecialInstructions": ""}
3	{"SKU": "efgh", "Name": "service1", "Notes": "dhjldsh", "NewAcc": 20, "ProAcc": 10, "BasicUOM": "abcd", "IsTaxable": true, "ExpenseAcc": 12345, "PriceRange": "10-30", "RegularAcc": 30, "AvgUnitCost": 20, "SalesAccount": "1234567890", "service_type": {"ID": 2, "Name": "Polishing"}, "service_Group": {"ID": 2, "Name": "Polishing"}, "service_category": {"ID": 2, "Name": "Polishing kit"}, "SpecialInstructions": "kdsdhkshdlddnkcds\\ndshjhdsfkjdk"}
5	{"SKU": "zxcv", "Name": "service3", "Notes": "ksdkdjsjk", "NewAcc": 20, "ProAcc": 10, "BasicUOM": "asdf", "IsTaxable": true, "ExpenseAcc": 1234560, "PriceRange": "10-30", "RegularAcc": 30, "AvgUnitCost": 20, "SalesAccount": "1234567890", "service_type": {"ID": 6, "Name": "cutting"}, "service_Group": {"ID": 6, "Name": "cutting"}, "service_category": {"ID": 6, "Name": "Cutting"}, "SpecialInstructions": "sdbjiniuwe"}
6	{"SKU": "def", "Name": "Service 4", "Notes": "ad", "NewAcc": 20, "ProAcc": 10, "BasicUOM": "abc", "IsTaxable": true, "ExpenseAcc": 12456, "PriceRange": "10-30", "RegularAcc": 30, "AvgUnitCost": 20, "SalesAccount": "1234567890", "service_type": {"ID": 6, "Name": "cutting"}, "service_Group": {"ID": 6, "Name": "cutting"}, "service_category": {"ID": 6, "Name": "Cutting"}, "SpecialInstructions": "sdfsa"}
\.





--
-- Data for Name: stage_tasks; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.stage_tasks (id, stage_id, name, due_date, duration, completed, created_at, updated_at, deleted_at) FROM stdin;
2fd8d394-2777-4645-a504-9e3c5506334c	da8e664f-5cff-4f4d-bac8-6c4a60bcdda9	task2	1223-12-12 05:17:32+00	10	f	2025-08-13 08:25:20.515665+00	2025-08-13 08:25:20.515665+00	\N
be538b51-973a-4361-9ec0-ce9fb6cc446c	19db1509-6f0d-48b3-b5d2-edf1b15c6d7b	new sub task	2025-08-14 19:05:00+00	1	f	2025-08-13 19:05:17.391943+00	2025-08-13 19:05:17.391944+00	\N
86b70b29-070b-4d2e-a24c-5d92028e5245	6fa40aed-a891-4bf0-91af-e7605546c993	task1	1222-12-11 18:18:32+00	5	f	2025-08-13 08:24:55.100108+00	2025-08-14 11:27:38.299578+00	\N
c177e264-8785-42e6-80b1-6e0f28347026	3cef16ff-b603-4707-80be-d3df4e8fb406	task1	2004-12-11 18:42:00+00	5	f	2025-08-13 08:23:08.418941+00	2025-08-13 19:08:58.806501+00	\N
90bdc631-8669-42a7-9470-a533bc1c4d34	3cef16ff-b603-4707-80be-d3df4e8fb406	task1	2004-12-11 19:31:00+00	5	f	2025-08-13 08:23:39.092599+00	2025-08-13 19:08:58.876503+00	\N
584cb13b-a862-4e67-b9de-d1beb889ea02	3cef16ff-b603-4707-80be-d3df4e8fb406	new task	2025-08-14 09:45:00+00	1	f	2025-08-13 09:45:57.087673+00	2025-08-13 19:08:58.915453+00	\N
7f18615a-36f6-437e-b2a7-86eafa479133	4198c2e9-0bf7-450d-9fe6-529cd35744f9	sub_task_installation	2025-08-14 10:00:00+00	1	f	2025-08-13 10:00:38.032428+00	2025-08-13 10:00:38.032428+00	\N
018938ed-3155-43db-a059-257f5c837bab	afbfae76-6d15-4a17-8c89-25453499bbf4	task1	2003-12-11 18:42:00+00	5	f	2025-08-14 12:18:10.262589+00	2025-08-14 12:18:10.262589+00	\N
9b14d270-7200-4d43-9764-b5f2f3f6efe0	bdec0d5b-ec41-440e-a0f1-84b4aaa71600	dsafasdf	2003-12-11 18:42:00+00	5	f	2025-08-14 12:19:59.408034+00	2025-08-14 12:19:59.408034+00	\N
11ee4efc-6525-418d-9170-a3299a7d1f2c	afbfae76-6d15-4a17-8c89-25453499bbf4	sadfasd	1212-12-12 06:18:32+00	3	f	2025-08-14 12:21:05.556225+00	2025-08-14 12:21:05.556225+00	\N
1ab9d31f-5141-4d2d-8e29-f5aebb8ef501	afbfae76-6d15-4a17-8c89-25453499bbf4	sa	1212-12-11 18:18:32+00	3	f	2025-08-14 12:21:28.418873+00	2025-08-14 12:21:28.418873+00	\N
803d4cfd-0c8a-49ce-bbac-3c32e4ce2444	afbfae76-6d15-4a17-8c89-25453499bbf4	sadfasdf	2211-12-21 06:42:00+00	2	f	2025-08-14 12:22:27.763151+00	2025-08-14 12:22:27.763151+00	\N
8c334892-0e45-478a-9a24-0c1cda95f6d0	afbfae76-6d15-4a17-8c89-25453499bbf4	asdfasdf	1212-12-11 18:18:32+00	2	f	2025-08-14 12:23:27.603761+00	2025-08-14 12:23:27.603761+00	\N
b036b8b4-a918-424c-8532-43e3d39b013a	551b2e13-1c22-4e9d-8555-ec370a8acb9b	sadfasdf	2003-12-12 06:42:00+00	190	f	2025-08-14 12:24:00.707962+00	2025-08-14 12:24:00.707963+00	\N
250a0d8d-1167-4256-ba89-7ef3437e099d	2e2219ab-d25b-48d7-856c-afb415ecf047	test_task	2025-08-18 21:30:00+00	30	f	2025-08-14 21:36:30.312004+00	2025-08-14 21:36:30.312004+00	\N
27b1f682-dd35-46df-91e2-20e4d6033adc	fddc0b12-bf5e-4014-aedd-8e1680bd89a7	new task	2025-08-20 11:52:00+00	1	f	2025-08-19 11:52:28.081119+00	2025-08-19 11:52:28.081119+00	\N
b2d2f104-22eb-405a-a0a3-4358e75983c8	fddc0b12-bf5e-4014-aedd-8e1680bd89a7	new task 1	2025-08-20 12:05:00+00	1	f	2025-08-19 12:05:26.039215+00	2025-08-19 12:05:26.039215+00	\N
9f170b42-d1f2-43bd-a6fa-608b1a173cea	e82b5294-154d-4dec-aed9-273fd39fa78a	test	2025-08-21 11:50:00+00	1	f	2025-08-21 08:50:56.938299+00	2025-08-21 08:50:56.938299+00	\N
\.


--
-- Data for Name: stage_users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.stage_users (id, stage_id, user_id, name, created_at, updated_at, deleted_at) FROM stdin;
\.


--
-- Data for Name: stages; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.stages (id, lead_id, name, description, state, date, notes, team, created_at, updated_at, deleted_at, residential_entry_id) FROM stdin;
2129fd38-0219-4d10-add3-4b8a9577647a	153714	Template	Prepare and finalize the template for the project	Completed	2025-08-11 15:14:47.688005+00			2025-08-11 15:13:05.065239+00	2025-08-11 15:14:47.688049+00	\N	1754925139227
fa609dc2-43f0-4954-8c2b-7a9e553c571b	153714	CAD	Create CAD designs for the project	Completed	2025-08-11 15:14:49.622139+00			2025-08-11 15:13:05.06524+00	2025-08-11 15:14:49.622166+00	\N	1754925139227
99ed3433-275c-4b30-8278-92a1dabfc70c	153714	Cutting	Cut materials according to specifications	Completed	2025-08-11 15:14:51.267479+00			2025-08-11 15:13:05.06524+00	2025-08-11 15:14:51.267507+00	\N	1754925139227
c6be62b9-82c6-4dcf-b524-4c5444c2d00f	153714	Fabrication	Fabricate components for the project	Completed	2025-08-11 15:14:52.741686+00			2025-08-11 15:13:05.065243+00	2025-08-11 15:14:52.741715+00	\N	1754925139227
45ec4126-de41-42e5-83e0-8c4755fe91f8	153714	Installation	Install the project components on-site	Completed	2025-08-11 15:14:54.491806+00			2025-08-11 15:13:05.065243+00	2025-08-11 15:14:54.491834+00	\N	1754925139227
f7717665-71f2-4637-b561-aaaf371dd12c	153714	Service Call	Handle any service calls or maintenance	Completed	2025-08-11 15:15:32.368685+00			2025-08-11 15:13:05.065243+00	2025-08-11 15:15:32.368714+00	\N	1754925139227
c77704e3-096f-4121-adeb-cd958addc01e	338697	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 15:16:40.909211+00			2025-08-11 15:16:40.909211+00	2025-08-11 15:16:40.943424+00	\N	1754925345946
ac1c80cd-35ff-4c17-abde-f212791d79a9	338697	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 15:16:40.909212+00	2025-08-11 15:16:41.089921+00	\N	1754925345946
89cbe1d3-fa9b-46fa-9471-cb6902131d7a	338697	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 15:16:40.909212+00	2025-08-11 15:16:41.259052+00	\N	1754925345946
ba6a435b-fbe0-49c9-80d6-c1cc060dd253	338697	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 15:16:40.909212+00	2025-08-11 15:16:41.359738+00	\N	1754925345946
dbe28f84-d24b-4979-94c2-cc46f961d50e	338697	Installation	Install the project components on-site	Not Started	\N			2025-08-11 15:16:40.909212+00	2025-08-11 15:16:41.45978+00	\N	1754925345946
8aa8d037-bd76-4ee2-acec-775c90c6d81d	338697	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 15:16:40.909212+00	2025-08-11 15:16:41.565038+00	\N	1754925345946
f391a5fd-0c0f-4874-b510-954ecd5d102e	338697	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 15:16:41.810284+00			2025-08-11 15:16:41.810284+00	2025-08-11 15:16:41.843642+00	\N	1754925352072
29783967-5796-4a70-9e61-d9be1c0b22aa	338697	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 15:16:41.810285+00	2025-08-11 15:16:41.948086+00	\N	1754925352072
62de9616-d27a-4ec6-af05-ec4185f6e1d3	338697	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 15:16:41.810285+00	2025-08-11 15:16:42.050251+00	\N	1754925352072
1abe5b49-16e3-4e42-a535-c88ecb8d94fc	338697	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 15:16:41.810286+00	2025-08-11 15:16:42.174696+00	\N	1754925352072
0f058178-f6ab-452e-b3e5-01046666e1db	338697	Installation	Install the project components on-site	Not Started	\N			2025-08-11 15:16:41.810286+00	2025-08-11 15:16:42.274495+00	\N	1754925352072
102a3250-507e-44ae-abd3-2991674dc191	338697	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 15:16:41.810286+00	2025-08-11 15:16:42.395145+00	\N	1754925352072
9c24b43e-73e9-4188-a506-decf392839c2	278004	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 15:24:04.371519+00			2025-08-11 15:24:04.371518+00	2025-08-11 15:24:04.40493+00	\N	1754925822298
5e7fbfe4-6789-4b67-b876-a70cd21631d7	278004	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 15:24:04.371519+00	2025-08-11 15:24:04.540877+00	\N	1754925822298
c96ed351-7396-48b7-a6b0-f7d3fa6e6c76	278004	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 15:24:04.37152+00	2025-08-11 15:24:04.677031+00	\N	1754925822298
5d99027e-2a6d-4264-a26c-96ef46e285b4	278004	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 15:24:04.37152+00	2025-08-11 15:24:04.770337+00	\N	1754925822298
47be3109-89f0-4ae0-8021-2d14a224fc15	278004	Installation	Install the project components on-site	Not Started	\N			2025-08-11 15:24:04.37152+00	2025-08-11 15:24:04.870981+00	\N	1754925822298
a0ffee90-a289-4986-bb9f-7cf6fa52cbc2	278004	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 15:24:04.37152+00	2025-08-11 15:24:04.971116+00	\N	1754925822298
f154efde-6862-4615-8c0c-a168c73f91a4	379122	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 16:42:49.363979+00			2025-08-11 16:42:49.363979+00	2025-08-11 16:42:49.364001+00	\N	1754930429838
202f3f5c-31e5-40f7-95bc-fe32d098efb4	379122	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 16:42:49.363979+00	2025-08-11 16:42:49.444067+00	\N	1754930429838
e660b326-e772-4901-a41a-fce2e0df6680	379122	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 16:42:49.363979+00	2025-08-11 16:42:49.523344+00	\N	1754930429838
826844e1-76fb-4b2f-8c0b-fec813d636df	379122	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 16:42:49.36398+00	2025-08-11 16:42:49.563922+00	\N	1754930429838
1b11075a-df57-485a-beb2-d064d512f5bd	379122	Installation	Install the project components on-site	Not Started	\N			2025-08-11 16:42:49.36398+00	2025-08-11 16:42:49.598428+00	\N	1754930429838
184e32b1-d3c7-4d9e-acdc-f9d5c13b8232	379122	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 16:42:49.36398+00	2025-08-11 16:42:49.668258+00	\N	1754930429838
6ca2c068-d401-4d9f-b9f6-11f3a7d2dd53	334790	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 16:57:11.982308+00			2025-08-11 16:57:11.982307+00	2025-08-11 16:57:11.982352+00	\N	1754931270237
ee4ac892-bdbb-44e1-b686-94001f17f358	334790	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 16:57:11.982309+00	2025-08-11 16:57:12.05258+00	\N	1754931270237
de2b9c6e-10c8-44e5-a970-be150d424acb	334790	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 16:57:11.982309+00	2025-08-11 16:57:12.122936+00	\N	1754931270237
848acb8a-a8aa-458e-9691-fb30f4754dde	334790	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 16:57:11.98231+00	2025-08-11 16:57:12.15657+00	\N	1754931270237
641a8eb3-23b3-4a3c-8e91-16e5586b3acc	334790	Installation	Install the project components on-site	Not Started	\N			2025-08-11 16:57:11.98231+00	2025-08-11 16:57:12.192061+00	\N	1754931270237
361f334a-e880-4198-853b-dab657c31e51	334790	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 16:57:11.982311+00	2025-08-11 16:57:12.227842+00	\N	1754931270237
fdf73542-6317-454e-a256-b69eddf9d2d5	990977	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 17:08:31.193619+00			2025-08-11 17:08:31.193619+00	2025-08-11 17:08:31.233621+00	\N	1754932054536
c3fe99b7-2ab5-4c3e-9daf-49c7915c256b	990977	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 17:08:31.193619+00	2025-08-11 17:08:31.383569+00	\N	1754932054536
950b86b6-bd87-44bd-a097-7373079b793a	990977	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 17:08:31.19362+00	2025-08-11 17:08:31.548959+00	\N	1754932054536
f0fb80fb-3d62-4e6e-8ecf-70300d9f73f7	990977	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 17:08:31.19362+00	2025-08-11 17:08:31.654574+00	\N	1754932054536
bd05b370-0093-41bf-8c84-ceca3da80427	990977	Installation	Install the project components on-site	Not Started	\N			2025-08-11 17:08:31.19362+00	2025-08-11 17:08:31.759491+00	\N	1754932054536
f2cead69-1c12-43f3-87dc-24438bc68138	990977	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 17:08:31.19362+00	2025-08-11 17:08:31.864671+00	\N	1754932054536
d9148c4d-8149-4801-9647-8d4f0a529aac	990977	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 17:08:32.143562+00			2025-08-11 17:08:32.143562+00	2025-08-11 17:08:32.178285+00	\N	1754932059204
77ebbfac-41d0-4963-b842-147942af6d3c	990977	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 17:08:32.143563+00	2025-08-11 17:08:32.298279+00	\N	1754932059204
86f69102-d070-453e-a7ab-ecb475c939b3	990977	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 17:08:32.143563+00	2025-08-11 17:08:32.410736+00	\N	1754932059204
2a5b113d-6870-42fa-8913-ba49a2c2a700	990977	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 17:08:32.143563+00	2025-08-11 17:08:32.52474+00	\N	1754932059204
bcd61705-2ab7-45da-9743-b2753a7e7848	990977	Installation	Install the project components on-site	Not Started	\N			2025-08-11 17:08:32.143563+00	2025-08-11 17:08:32.6348+00	\N	1754932059204
8d760edc-0bdc-4667-8f84-41d226498cd3	990977	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 17:08:32.143564+00	2025-08-11 17:08:32.73986+00	\N	1754932059204
bbd93374-ff0f-4a6f-8fcd-42a370b0a286	087879	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 17:33:36.327576+00			2025-08-11 17:33:36.327576+00	2025-08-11 17:33:36.356327+00	\N	1754933591839
3a2290a4-b434-4b4e-a9de-82a0c5a602b3	087879	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 17:33:36.327577+00	2025-08-11 17:33:36.51875+00	\N	1754933591839
8ea05ebb-a783-4937-aee5-a1b5f104aaa2	087879	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 17:33:36.327577+00	2025-08-11 17:33:36.651764+00	\N	1754933591839
f85d8bca-a1ac-46a7-8ae6-6a3d9b893230	087879	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 17:33:36.327577+00	2025-08-11 17:33:37.138499+00	\N	1754933591839
afe972ed-7b79-477c-8c1a-f4b5d091cd7d	087879	Installation	Install the project components on-site	Not Started	\N			2025-08-11 17:33:36.327578+00	2025-08-11 17:33:37.25875+00	\N	1754933591839
a574347f-a1e0-4cff-9d4d-4b55f49c0634	087879	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 17:33:36.327578+00	2025-08-11 17:33:37.391799+00	\N	1754933591839
286c8e24-5adf-4d4e-95fb-81cc368d57fe	893916	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 17:44:07.340444+00			2025-08-11 17:44:07.340443+00	2025-08-11 17:44:07.375321+00	\N	1754934206559
76e40b2f-1fa8-4d4c-840a-9dfddeeec87f	893916	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 17:44:07.340444+00	2025-08-11 17:44:07.542536+00	\N	1754934206559
763650d6-2a6d-4c9b-acdf-4d35191a72b8	893916	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 17:44:07.340444+00	2025-08-11 17:44:07.68289+00	\N	1754934206559
1869f729-d64a-480a-9a0d-b268f9e9cfd2	893916	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 17:44:07.340445+00	2025-08-11 17:44:07.781428+00	\N	1754934206559
65c5d5ab-32dd-47e7-9a9d-6b3e0db851ed	893916	Installation	Install the project components on-site	Not Started	\N			2025-08-11 17:44:07.340445+00	2025-08-11 17:44:07.880784+00	\N	1754934206559
1a92c545-bab6-4a8b-bbea-55ff345319ea	893916	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 17:44:07.340445+00	2025-08-11 17:44:07.980906+00	\N	1754934206559
6ec35ac5-9687-4909-9f66-c60299a6e08b	788865	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 17:45:55.188144+00			2025-08-11 17:45:55.188144+00	2025-08-11 17:45:55.188186+00	\N	1754934118795
d3923326-f733-4f3c-8c36-2f3f3b7cbac0	788865	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 17:45:55.188145+00	2025-08-11 17:45:55.267835+00	\N	1754934118795
eb9c8cfd-288f-4bda-8a5b-08d54f6b3a6e	788865	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 17:45:55.188145+00	2025-08-11 17:45:55.348552+00	\N	1754934118795
1f923c46-f81f-4a02-bf64-d74f054284ee	788865	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 17:45:55.188146+00	2025-08-11 17:45:55.388014+00	\N	1754934118795
62bba5e4-2a9f-4424-b2ee-4c9c28d4a0ee	788865	Installation	Install the project components on-site	Not Started	\N			2025-08-11 17:45:55.188146+00	2025-08-11 17:45:55.423328+00	\N	1754934118795
fcbcde77-ac3b-456a-9d73-d388a8594c30	788865	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 17:45:55.188148+00	2025-08-11 17:45:55.458544+00	\N	1754934118795
86f29ddc-00bb-49b6-b8ea-50909c4cbf42	788865	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 17:56:27.024098+00			2025-08-11 17:56:27.024097+00	2025-08-11 17:56:27.024136+00	\N	1754934746920
3dd10275-6281-44a6-9bad-c59d74099e7d	788865	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 17:56:27.024098+00	2025-08-11 17:56:27.093537+00	\N	1754934746920
bbdd1f76-244b-43d1-9498-9692d0c08c53	788865	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 17:56:27.024098+00	2025-08-11 17:56:27.163468+00	\N	1754934746920
bc919d60-8135-4430-8318-d2a912b6f0b6	788865	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 17:56:27.024099+00	2025-08-11 17:56:27.198857+00	\N	1754934746920
d42cb5a4-22e7-4cb6-b136-32c29d48a10e	788865	Installation	Install the project components on-site	Not Started	\N			2025-08-11 17:56:27.024099+00	2025-08-11 17:56:27.235237+00	\N	1754934746920
c6c2b3f3-a4e8-4c34-a70e-cbbb97d8227e	788865	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 17:56:27.024099+00	2025-08-11 17:56:27.268984+00	\N	1754934746920
08447251-46a2-49a2-bdda-df144b12674b	764542	Template	Prepare and finalize the template for the project	Completed	2025-08-11 18:27:45.504205+00			2025-08-11 17:04:28.75109+00	2025-08-11 18:27:45.50426+00	\N	1754931688619
80920ca5-e8f6-41da-9527-1768bb9b6d25	764542	Cutting	Cut materials according to specifications	Completed	2025-08-11 18:31:58.264761+00			2025-08-11 17:04:28.751092+00	2025-08-11 18:31:58.264814+00	\N	1754931688619
4812a74c-3ea1-45bd-8ad8-7ffb05364c20	764542	CAD	Create CAD designs for the project	Completed	2025-08-11 18:27:49.629071+00			2025-08-11 17:04:28.751092+00	2025-08-11 18:27:49.629143+00	\N	1754931688619
5781358f-96c4-49b0-aabb-f22c54e9fc8d	764542	Fabrication	Fabricate components for the project	Completed	2025-08-11 18:32:01.748898+00			2025-08-11 17:04:28.751093+00	2025-08-11 18:32:01.748942+00	\N	1754931688619
7b7019a2-3213-42af-9df3-344a709d30e6	764542	Installation	Install the project components on-site	Completed	2025-08-11 18:32:04.771573+00			2025-08-11 17:04:28.751094+00	2025-08-11 18:32:04.771629+00	\N	1754931688619
c50f99b8-756f-4442-a931-b12fe1a89c7b	764542	Service Call	Handle any service calls or maintenance	Completed	2025-08-11 18:32:08.193993+00			2025-08-11 17:04:28.751094+00	2025-08-11 18:32:08.194063+00	\N	1754931688619
933aaa4d-7f0d-4f5a-af92-ccb952c7bb70	028735	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 19:10:06.236284+00			2025-08-11 19:10:06.236284+00	2025-08-11 19:10:06.29662+00	\N	1754939263846
dc52c2b3-034d-46c7-91e4-e963105cc180	028735	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 19:10:06.236284+00	2025-08-11 19:10:06.474128+00	\N	1754939263846
81f1e57e-65c3-4480-8a01-b3b2bc8c18e1	028735	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 19:10:06.236285+00	2025-08-11 19:10:06.616977+00	\N	1754939263846
da3f3149-75f8-4e7b-9635-64dd8b9f44fd	028735	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 19:10:06.236285+00	2025-08-11 19:10:06.749061+00	\N	1754939263846
5bf8cd71-da91-400d-b730-fd855a5251e6	028735	Installation	Install the project components on-site	Not Started	\N			2025-08-11 19:10:06.236286+00	2025-08-11 19:10:06.869353+00	\N	1754939263846
2eb2bcf3-3295-4e00-a8be-be623a0610fe	028735	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 19:10:06.236286+00	2025-08-11 19:10:06.988662+00	\N	1754939263846
abe93e85-1cf4-40e5-b477-9712ad57b25f	945419	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 19:12:52.928986+00			2025-08-11 19:12:52.928985+00	2025-08-11 19:12:52.962216+00	\N	1754939552481
ee905d17-9b8e-4e34-a176-6b6078d54231	945419	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 19:12:52.928986+00	2025-08-11 19:12:53.103859+00	\N	1754939552481
39e8573c-69c2-4d39-bd74-180949d4da02	945419	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 19:12:52.928986+00	2025-08-11 19:12:53.254312+00	\N	1754939552481
935fe7b7-cccd-4cfd-bf69-f5362da08e8e	945419	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 19:12:52.928987+00	2025-08-11 19:12:53.379012+00	\N	1754939552481
805d208e-fa92-41aa-b183-a2146918741a	945419	Installation	Install the project components on-site	Not Started	\N			2025-08-11 19:12:52.928987+00	2025-08-11 19:12:53.484029+00	\N	1754939552481
9f28484f-fc5a-4c63-8e95-fb2fe56e872c	945419	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 19:12:52.928987+00	2025-08-11 19:12:53.591213+00	\N	1754939552481
fb3b3a2a-2dc7-4110-a469-95eeb7638efe	756697	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 19:16:41.760361+00			2025-08-11 19:16:41.76036+00	2025-08-11 19:16:41.804306+00	\N	1754939774446
6fc8e9df-97ac-41c0-9c99-68575a7b3a9b	756697	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 19:16:41.760361+00	2025-08-11 19:16:41.992826+00	\N	1754939774446
3bfb1482-c225-4245-a979-21133742b8c1	756697	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 19:16:41.760362+00	2025-08-11 19:16:42.16665+00	\N	1754939774446
f881e1e0-4b73-449b-8568-55d180749ab5	756697	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 19:16:41.760362+00	2025-08-11 19:16:42.28453+00	\N	1754939774446
f1b10e97-f3c5-43a6-90cb-9a16fdc817e2	756697	Installation	Install the project components on-site	Not Started	\N			2025-08-11 19:16:41.760362+00	2025-08-11 19:16:42.409217+00	\N	1754939774446
212ab044-0ad6-47ae-8260-0cf03d87cc78	756697	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 19:16:41.760362+00	2025-08-11 19:16:42.518058+00	\N	1754939774446
aa0f97db-6548-4f2b-9489-8e01178fe6a9	512161	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 19:21:21.855534+00			2025-08-11 19:21:21.855533+00	2025-08-11 19:21:21.889427+00	\N	1754939980082
1c4ec449-9bd8-485a-a453-a1ac1a14e2e3	512161	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 19:21:21.855534+00	2025-08-11 19:21:22.029719+00	\N	1754939980082
9240dd0c-0478-4543-9dc2-9fedf5bf7eb5	512161	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 19:21:21.855535+00	2025-08-11 19:21:22.169487+00	\N	1754939980082
80b6dccc-4b86-4068-b8cf-b0a2f7843d2b	512161	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 19:21:21.855536+00	2025-08-11 19:21:22.274934+00	\N	1754939980082
bbf72729-a8c8-4127-ac6f-f1f7a1e4dc7b	512161	Installation	Install the project components on-site	Not Started	\N			2025-08-11 19:21:21.855536+00	2025-08-11 19:21:22.383995+00	\N	1754939980082
7b005391-c439-4afa-8098-713f682f4d17	512161	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 19:21:21.855537+00	2025-08-11 19:21:22.494976+00	\N	1754939980082
279fa318-2ceb-40a6-adb5-cbaf7b31c164	198032	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 19:23:33.630932+00			2025-08-11 19:23:33.630931+00	2025-08-11 19:23:33.672989+00	\N	1754940197228
d3a91efd-5c9d-406a-9324-cf385e740a11	198032	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 19:23:33.630932+00	2025-08-11 19:23:33.827265+00	\N	1754940197228
9758cf67-4583-4a91-9d27-2664318d2398	198032	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 19:23:33.630935+00	2025-08-11 19:23:33.9674+00	\N	1754940197228
28092659-d0e2-40fb-8a5c-25a756e2d889	198032	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 19:23:33.630936+00	2025-08-11 19:23:34.072064+00	\N	1754940197228
789bbc28-5a13-4820-85c9-73f64e27310f	198032	Installation	Install the project components on-site	Not Started	\N			2025-08-11 19:23:33.630936+00	2025-08-11 19:23:34.177021+00	\N	1754940197228
9201a0ad-60fc-4432-bfc3-ef2fd90f04fe	198032	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 19:23:33.630936+00	2025-08-11 19:23:34.281945+00	\N	1754940197228
379c0620-c018-4b1b-ae21-09663c76a381	966384	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 20:07:13.524893+00			2025-08-11 20:07:13.524893+00	2025-08-11 20:07:13.563765+00	\N	1754942775621
4c1cfd38-4ea3-4ccc-8e3b-983648a98b40	966384	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 20:07:13.524894+00	2025-08-11 20:07:13.710596+00	\N	1754942775621
5c3c1672-7991-4dd0-aa48-b863e9fa088b	966384	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 20:07:13.524894+00	2025-08-11 20:07:13.856081+00	\N	1754942775621
50b5806f-56cf-45d9-9da4-68d9fad96f32	966384	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 20:07:13.524895+00	2025-08-11 20:07:13.969718+00	\N	1754942775621
750603d6-4976-4777-9607-479622279b0f	966384	Installation	Install the project components on-site	Not Started	\N			2025-08-11 20:07:13.524895+00	2025-08-11 20:07:14.085279+00	\N	1754942775621
770ba3b3-7e04-4596-a5c9-962bd3eff987	966384	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 20:07:13.524895+00	2025-08-11 20:07:14.206018+00	\N	1754942775621
e8b25e49-8231-431d-b645-93b7749041c5	966384	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 20:07:14.460517+00			2025-08-11 20:07:14.460517+00	2025-08-11 20:07:14.49669+00	\N	1754942781800
b0267c7d-df6d-428f-957d-d2cb0debe0dc	966384	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 20:07:14.460518+00	2025-08-11 20:07:14.605824+00	\N	1754942781800
1250dd09-9c98-4c2b-9e01-e3c8ffa11d4c	966384	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 20:07:14.460518+00	2025-08-11 20:07:14.71634+00	\N	1754942781800
77a7e2a6-11d6-420a-a4c4-0d202d982054	966384	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 20:07:14.460518+00	2025-08-11 20:07:14.830769+00	\N	1754942781800
1227d4a3-35a3-42ef-a21b-3f8b32ddb4ee	966384	Installation	Install the project components on-site	Not Started	\N			2025-08-11 20:07:14.460519+00	2025-08-11 20:07:14.940859+00	\N	1754942781800
f52c88be-68b2-4a83-8673-d730e4481f26	966384	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 20:07:14.460519+00	2025-08-11 20:07:15.045967+00	\N	1754942781800
8e921c7f-dee7-4c51-9d1e-c382c279d9b4	553579	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 20:10:15.891269+00			2025-08-11 20:10:15.891268+00	2025-08-11 20:10:15.924553+00	\N	1754942998444
86c212ce-6893-4f0e-bdbc-bb479f3f7295	553579	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 20:10:15.891269+00	2025-08-11 20:10:16.064374+00	\N	1754942998444
9fb7a564-ef13-429d-805c-8005ee6eab3e	553579	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 20:10:15.891269+00	2025-08-11 20:10:16.203863+00	\N	1754942998444
01584ca0-82af-4c41-b96b-1a3d69580689	553579	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 20:10:15.89127+00	2025-08-11 20:10:16.314188+00	\N	1754942998444
12733e2a-8937-4480-873f-cddb99e27dee	553579	Installation	Install the project components on-site	Not Started	\N			2025-08-11 20:10:15.89127+00	2025-08-11 20:10:16.414572+00	\N	1754942998444
15a0ca20-8288-4fee-9f9e-b8129e710a76	553579	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 20:10:15.891271+00	2025-08-11 20:10:16.517226+00	\N	1754942998444
c90d5702-c88e-4afb-b3a9-261625cb9d7c	971784	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 20:13:34.313035+00			2025-08-11 20:13:34.313035+00	2025-08-11 20:13:34.348384+00	\N	1754943191639
f4403acc-91be-4d4d-bf63-77e4d840aa1c	971784	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 20:13:34.313036+00	2025-08-11 20:13:34.48422+00	\N	1754943191639
35e45935-89d5-4266-b35e-4555821eefae	971784	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 20:13:34.313036+00	2025-08-11 20:13:34.623693+00	\N	1754943191639
351ce04c-5624-4925-9591-3f590f51620a	971784	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 20:13:34.313037+00	2025-08-11 20:13:34.730156+00	\N	1754943191639
aaca24ba-5b0e-481f-aeb9-6561f108c0b5	971784	Installation	Install the project components on-site	Not Started	\N			2025-08-11 20:13:34.313037+00	2025-08-11 20:13:34.828481+00	\N	1754943191639
e3a36b6f-db56-4bf2-9075-8c59de79504c	971784	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 20:13:34.313039+00	2025-08-11 20:13:34.931543+00	\N	1754943191639
7b7983b9-958a-4915-b97e-cb2187c00058	115148	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 20:29:58.593089+00	2025-08-11 20:29:58.930235+00	\N	1754944171822
1271e57e-0414-47fe-8e35-562b4d3531f6	115148	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 20:29:58.593089+00	2025-08-11 20:29:59.045177+00	\N	1754944171822
4ad510b1-b10b-45dd-ba6e-6cc115186e09	115148	Installation	Install the project components on-site	Not Started	\N			2025-08-11 20:29:58.59309+00	2025-08-11 20:29:59.160995+00	\N	1754944171822
e963fda7-ecfc-4887-a293-ff1475853152	115148	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 20:29:58.59309+00	2025-08-11 20:29:59.277172+00	\N	1754944171822
f73b4731-4e79-46ea-9c89-02752b0887b6	920579	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 20:40:59.117615+00			2025-08-11 20:40:59.117614+00	2025-08-11 20:40:59.151812+00	\N	1754944823902
022a430f-84a0-4ea7-8c13-ad007328c9d1	920579	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 20:40:59.117615+00	2025-08-11 20:40:59.303344+00	\N	1754944823902
aa51dcee-6492-4172-9e81-a51ef877a1e8	920579	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 20:40:59.117616+00	2025-08-11 20:40:59.438853+00	\N	1754944823902
38108722-7f97-465e-8369-2722009ca721	920579	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 20:40:59.117616+00	2025-08-11 20:40:59.54858+00	\N	1754944823902
a7074720-39ce-48b3-a99a-8a29667a3248	920579	Installation	Install the project components on-site	Not Started	\N			2025-08-11 20:40:59.117616+00	2025-08-11 20:40:59.663521+00	\N	1754944823902
4987789f-3a28-4e7d-b78d-359a65526c60	920579	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 20:40:59.117616+00	2025-08-11 20:40:59.784783+00	\N	1754944823902
fda0eb43-7ce5-4bef-b534-e59dc9885565	078083	Template	Prepare and finalize the template for the project	In Progress	2025-08-11 20:48:45.601074+00			2025-08-11 20:48:45.601073+00	2025-08-11 20:48:45.634885+00	\N	1754945304315
02b4f72f-0838-4cda-ae9e-b08d6fc067a1	078083	CAD	Create CAD designs for the project	Not Started	\N			2025-08-11 20:48:45.601074+00	2025-08-11 20:48:45.7662+00	\N	1754945304315
af7bc930-c4a9-41b8-bcae-28da076c395f	078083	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-11 20:48:45.601075+00	2025-08-11 20:48:45.906308+00	\N	1754945304315
9868bbc3-ce14-4580-92db-d6f5977a7128	078083	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 20:48:45.601075+00	2025-08-11 20:48:46.010038+00	\N	1754945304315
bfbe183b-0cc3-4993-99ac-c5835819fe57	078083	Installation	Install the project components on-site	Not Started	\N			2025-08-11 20:48:45.601075+00	2025-08-11 20:48:46.105601+00	\N	1754945304315
9ec38a16-35fa-4235-9cfd-3e183063804f	078083	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 20:48:45.601076+00	2025-08-11 20:48:46.220241+00	\N	1754945304315
b8a14ac5-e101-4576-bc71-580e3503a0b9	571272	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-11 20:51:20.105997+00	2025-08-11 20:51:20.514967+00	\N	1754945436867
e2267606-239e-4797-98b4-dc4ab497e38c	571272	Installation	Install the project components on-site	Not Started	\N			2025-08-11 20:51:20.105997+00	2025-08-11 20:51:20.61573+00	\N	1754945436867
936c675f-b383-4354-86e1-5a63a3dc6222	571272	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-11 20:51:20.105997+00	2025-08-11 20:51:20.713318+00	\N	1754945436867
41a9844f-1339-4708-8334-68bd4c55a303	115148	Template	Prepare and finalize the template for the project	Completed	2025-08-12 03:58:15.749859+00			2025-08-11 20:29:58.593088+00	2025-08-12 03:58:15.749921+00	\N	1754944171822
496e4d2a-48bb-4927-9016-20f8f5780c08	115148	CAD	Create CAD designs for the project	In Progress	2025-08-12 03:58:15.814509+00			2025-08-11 20:29:58.593089+00	2025-08-12 03:58:15.814563+00	\N	1754944171822
fadb2da9-5efa-44b7-a34d-486ee84df7e7	573182	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 04:48:26.864707+00			2025-08-12 04:48:26.864706+00	2025-08-12 04:48:26.864759+00	\N	1754973897203
277f1fad-94df-423e-9825-a84b84cf4254	573182	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 04:48:26.864708+00	2025-08-12 04:48:26.976574+00	\N	1754973897203
d29eea2c-f05c-4d79-a354-81e40312e963	573182	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 04:48:26.864708+00	2025-08-12 04:48:27.05465+00	\N	1754973897203
f07e9494-0152-42f2-89c8-de4960fadef1	573182	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 04:48:26.864709+00	2025-08-12 04:48:27.09181+00	\N	1754973897203
a7cca7f9-3747-4e3e-be88-789ca667c73e	573182	Installation	Install the project components on-site	Not Started	\N			2025-08-12 04:48:26.864709+00	2025-08-12 04:48:27.140786+00	\N	1754973897203
7037dbb4-6027-499f-9096-07455aced667	573182	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 04:48:26.86471+00	2025-08-12 04:48:27.176657+00	\N	1754973897203
6fd9c338-d6ff-4eff-b433-916739b4c794	571272	Template	Prepare and finalize the template for the project	Completed	2025-08-12 05:01:00.65875+00			2025-08-11 20:51:20.105995+00	2025-08-12 05:01:00.658793+00	\N	1754945436867
476a50d0-2cc8-4ba2-b012-12aeba0fecae	684877	Template	Prepare and finalize the template for the project	Completed	2025-08-12 06:23:48.711465+00			2025-08-12 05:23:03.049241+00	2025-08-12 06:23:48.71161+00	\N	1754976097604
17493f8e-977d-4283-8e4e-8689b60d6c1e	571272	CAD	Create CAD designs for the project	Completed	2025-08-12 05:01:05.639317+00			2025-08-11 20:51:20.105996+00	2025-08-12 05:01:05.639386+00	\N	1754945436867
70c44e8f-19fb-483d-89ce-88049662edab	571272	Cutting	Cut materials according to specifications	In Progress	2025-08-12 05:01:05.739096+00			2025-08-11 20:51:20.105996+00	2025-08-12 05:01:05.739274+00	\N	1754945436867
b61ebbbf-cbf6-479b-af69-53e825d81ca8	684877	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 05:23:03.049243+00	2025-08-12 05:23:03.242301+00	\N	1754976097604
a97bb4b1-46f4-46d0-83bb-9d0b805cf2d2	684877	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 05:23:03.049244+00	2025-08-12 05:23:03.419366+00	\N	1754976097604
417d4489-8cf9-4e1d-8c8e-12a6710f9499	684877	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 05:23:03.049244+00	2025-08-12 05:23:03.568951+00	\N	1754976097604
743a152e-60d6-41bf-80e5-b23503252bc4	684877	Installation	Install the project components on-site	Not Started	\N			2025-08-12 05:23:03.049245+00	2025-08-12 05:23:03.694754+00	\N	1754976097604
ea4dac2e-55cb-404d-88e1-44231964a024	684877	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 05:23:03.049246+00	2025-08-12 05:23:03.797174+00	\N	1754976097604
5744ddf8-549f-4180-baf0-25e6a1858269	684877	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 05:25:53.364721+00			2025-08-12 05:25:53.364721+00	2025-08-12 05:25:53.364816+00	\N	1754976307915
f9f848ec-e511-43ac-a35e-f785df2f650f	684877	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 05:25:53.364722+00	2025-08-12 05:25:53.480901+00	\N	1754976307915
b1f2fe3f-63e1-4010-a0ee-c6c5e1350a49	684877	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 05:25:53.364723+00	2025-08-12 05:25:53.5559+00	\N	1754976307915
b8197cb9-90d6-448d-87cb-198c08e37799	684877	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 05:25:53.364724+00	2025-08-12 05:25:53.598844+00	\N	1754976307915
7522144a-bd48-43d4-87c5-7d5ccc3d9181	684877	Installation	Install the project components on-site	Not Started	\N			2025-08-12 05:25:53.364724+00	2025-08-12 05:25:53.634668+00	\N	1754976307915
5e190261-0f6a-4bba-8c53-06cb2d302ce7	684877	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 05:25:53.364724+00	2025-08-12 05:25:53.686589+00	\N	1754976307915
3232aff9-d336-4e49-a68e-fd797bac8bfe	684877	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 05:28:59.32733+00			2025-08-12 05:28:59.32733+00	2025-08-12 05:28:59.32738+00	\N	1754976097604
20eb7222-4610-4469-9434-ab9e8bbc4e0f	684877	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 05:28:59.32733+00	2025-08-12 05:28:59.406794+00	\N	1754976097604
ce513a2a-59d7-4cc5-acbf-ac9e4509a161	684877	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 05:28:59.327331+00	2025-08-12 05:28:59.476081+00	\N	1754976097604
74c4fd22-9e04-4cc1-919f-ffb6e96a0fe5	684877	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 05:28:59.327332+00	2025-08-12 05:28:59.521077+00	\N	1754976097604
e35101bc-6fe8-4826-bc4d-8af44d27453e	684877	Installation	Install the project components on-site	Not Started	\N			2025-08-12 05:28:59.327332+00	2025-08-12 05:28:59.555168+00	\N	1754976097604
4346fdb9-1959-4252-a560-2abf5ed37946	684877	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 05:28:59.327332+00	2025-08-12 05:28:59.590973+00	\N	1754976097604
9a8d7102-482d-45d6-8c7c-8e365bcedcbf	684877	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 05:28:59.714637+00			2025-08-12 05:28:59.714637+00	2025-08-12 05:28:59.714682+00	\N	1754976307915
2a72084e-e960-4303-8c2d-97941fb1b282	684877	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 05:28:59.714637+00	2025-08-12 05:28:59.75+00	\N	1754976307915
ce1da4d2-70c5-4b98-b746-bff38fa0fe47	684877	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 05:28:59.714638+00	2025-08-12 05:28:59.787129+00	\N	1754976307915
c9094d7a-8511-49df-a178-f7c71a732625	684877	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 05:28:59.714639+00	2025-08-12 05:28:59.827574+00	\N	1754976307915
509f488e-aa2b-4ab0-ad08-c045f687ffbf	684877	Installation	Install the project components on-site	Not Started	\N			2025-08-12 05:28:59.714639+00	2025-08-12 05:28:59.881233+00	\N	1754976307915
cdfb17b5-2f0f-4418-9f96-50555913da4a	684877	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 05:28:59.714639+00	2025-08-12 05:28:59.939404+00	\N	1754976307915
bb93cad0-9fa9-45b7-bdd0-f2488c397b4e	684877	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 05:29:00.070278+00			2025-08-12 05:29:00.070278+00	2025-08-12 05:29:00.070338+00	\N	1754976459819
c90873f1-1b45-4f43-bafb-000fa1851ad5	684877	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 05:29:00.070279+00	2025-08-12 05:29:00.104652+00	\N	1754976459819
3ca1ed2d-87c3-4166-b4ad-7c7d20327dbb	684877	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 05:29:00.070279+00	2025-08-12 05:29:00.140912+00	\N	1754976459819
70aa3577-016d-4397-9b36-b4f201438f86	684877	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 05:29:00.07028+00	2025-08-12 05:29:00.197456+00	\N	1754976459819
dbae8dd6-2750-476e-a311-1633c7dd9220	684877	Installation	Install the project components on-site	Not Started	\N			2025-08-12 05:29:00.07028+00	2025-08-12 05:29:00.237915+00	\N	1754976459819
703764f7-3eae-4088-9d63-c60aca7bd015	684877	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 05:29:00.070281+00	2025-08-12 05:29:00.279423+00	\N	1754976459819
839d191f-e3f6-4904-adfd-4b3830e6fcf4	022652	Template	Prepare and finalize the template for the project	Completed	2025-08-12 07:00:19.596987+00			2025-08-12 06:32:17.299635+00	2025-08-12 07:00:19.597044+00	\N	1754980164865
a899d21d-b679-45e6-b447-6c3b8d1eab7b	022652	CAD	Create CAD designs for the project	Completed	2025-08-12 07:23:10.795652+00			2025-08-12 06:32:17.299639+00	2025-08-12 07:23:10.795709+00	\N	1754980164865
022ea4ae-55d9-4487-b3a3-7182d69f3615	022652	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 08:36:01.064531+00			2025-08-12 06:36:18.571275+00	2025-08-12 08:36:01.064559+00	\N	1754980523379
73c0315c-2808-4443-a317-c0e16c543f6f	766166	Cutting	Cut materials according to specifications	Completed	2025-08-12 08:28:31.201523+00			2025-08-12 07:07:25.861331+00	2025-08-12 08:28:31.201599+00	\N	1754982366867
cd22ea65-feaa-49ae-9bba-17a18c36f84b	766166	Template	Prepare and finalize the template for the project	Completed	2025-08-12 07:15:09.772575+00			2025-08-12 07:07:25.861327+00	2025-08-12 07:15:09.772609+00	\N	1754982366867
09efc49a-6eab-4fd8-97d2-03ef0ab39f26	022652	Fabrication	Fabricate components for the project	Completed	2025-08-12 07:23:15.940978+00			2025-08-12 06:32:17.299642+00	2025-08-12 07:23:15.941049+00	\N	1754980164865
74eb3212-3562-4e30-956f-240e71d376d4	766166	CAD	Create CAD designs for the project	Completed	2025-08-12 07:15:51.409915+00			2025-08-12 07:07:25.86133+00	2025-08-12 07:15:51.409987+00	\N	1754982366867
818fb540-8aee-487c-88fd-2ebb34267d18	022652	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 07:23:21.390251+00			2025-08-12 06:32:17.299644+00	2025-08-12 07:23:21.390293+00	\N	1754980164865
5ebd0de1-980c-4a46-8199-f7055270deff	022652	CAD	Create CAD designs for the project	Completed	2025-08-12 08:35:51.685689+00			2025-08-12 06:36:18.571273+00	2025-08-12 08:35:51.68572+00	\N	1754980523379
9354e5fc-0832-45c6-94e0-81b416bcb8e9	022652	Installation	Install the project components on-site	Completed	2025-08-12 08:35:56.73373+00			2025-08-12 06:36:18.571274+00	2025-08-12 08:35:56.733764+00	\N	1754980523379
8e0ffa32-4298-4b98-aa50-9db19aed6b41	304586	Template	Prepare and finalize the template for the project	Completed	2025-08-12 08:36:34.792348+00			2025-08-12 06:58:01.107801+00	2025-08-12 08:36:34.792383+00	\N	1754981655078
738a5114-ffcf-4ed1-8a86-495a2144dfec	304586	Cutting	Cut materials according to specifications	Completed	2025-08-12 08:36:42.203245+00			2025-08-12 06:58:01.107803+00	2025-08-12 08:36:42.203291+00	\N	1754981655078
6925450c-9deb-4aa9-983d-4a28622bfe80	304586	CAD	Create CAD designs for the project	Completed	2025-08-12 08:36:39.0127+00			2025-08-12 06:58:01.107802+00	2025-08-12 08:36:39.012755+00	\N	1754981655078
6ec4f196-334e-480d-a301-91b5f53bc714	304586	Fabrication	Fabricate components for the project	Completed	2025-08-12 08:36:46.281411+00			2025-08-12 06:58:01.107804+00	2025-08-12 08:36:46.281467+00	\N	1754981655078
5ae2aec0-9b2e-441e-836d-82d58d3e4fc1	304586	Installation	Install the project components on-site	Completed	2025-08-12 08:36:49.677762+00			2025-08-12 06:58:01.107804+00	2025-08-12 08:36:49.677805+00	\N	1754981655078
6f90dc8d-e1ea-46ac-8619-615bd89d13ce	304586	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 10:50:57.820874+00			2025-08-12 06:58:01.107805+00	2025-08-12 10:50:57.820906+00	\N	1754981655078
887014d4-ef28-4201-a256-c0ade7c40846	022652	Template	Prepare and finalize the template for the project	Completed	2025-08-12 07:13:38.088072+00			2025-08-12 06:36:18.571272+00	2025-08-12 07:13:38.08818+00	\N	1754980523379
ea86ce3f-bf9c-4363-bc05-7f574f9fb107	766166	Template	Prepare and finalize the template for the project	Completed	2025-08-12 07:15:41.622707+00			2025-08-12 07:09:28.446181+00	2025-08-12 07:15:41.622753+00	\N	1754982511168
cbd8c012-e3c2-4fcc-a018-63086c42cc0b	022652	Cutting	Cut materials according to specifications	Completed	2025-08-12 07:23:13.604053+00			2025-08-12 06:32:17.299641+00	2025-08-12 07:23:13.604117+00	\N	1754980164865
df5b099c-0ff7-4ee6-a6b9-05e46552175d	884011	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 08:52:42.872843+00			2025-08-12 08:52:42.872843+00	2025-08-12 08:52:42.872914+00	\N	1754988712647
fcc22586-e244-4bfd-9256-c207bad4bab1	022652	Installation	Install the project components on-site	Completed	2025-08-12 07:23:18.364887+00			2025-08-12 06:32:17.299643+00	2025-08-12 07:23:18.364945+00	\N	1754980164865
d9713559-0430-4fcd-bee8-926abb447828	766166	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 07:42:14.026878+00			2025-08-12 07:42:14.026878+00	2025-08-12 07:42:14.026959+00	\N	1754984525695
e0e0ad5a-bb81-4453-8b40-578351a542eb	766166	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 07:42:14.02688+00	2025-08-12 07:42:14.198213+00	\N	1754984525695
6db2914a-d1ae-4734-98e4-c37618569ed5	766166	Installation	Install the project components on-site	Not Started	\N			2025-08-12 07:42:14.026881+00	2025-08-12 07:42:14.278281+00	\N	1754984525695
fe2d56a5-c56f-459b-9535-41095cd0edd5	766166	Fabrication	Fabricate components for the project	Completed	2025-08-12 08:28:35.133495+00			2025-08-12 07:07:25.861331+00	2025-08-12 08:28:35.133555+00	\N	1754982366867
df494813-7560-4d54-8242-243b827ae981	766166	Installation	Install the project components on-site	Completed	2025-08-12 08:28:38.369091+00			2025-08-12 07:07:25.861332+00	2025-08-12 08:28:38.369241+00	\N	1754982366867
0e8e3d5b-fa6b-409b-b502-62ab97678a8d	766166	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 08:28:41.147702+00			2025-08-12 07:07:25.861332+00	2025-08-12 08:28:41.147748+00	\N	1754982366867
210ff69c-588d-49f9-95b1-e5a4817f2bde	766166	CAD	Create CAD designs for the project	Completed	2025-08-12 08:29:24.10652+00			2025-08-12 07:09:28.446183+00	2025-08-12 08:29:24.106565+00	\N	1754982511168
f0625a7a-9a1f-489f-90d3-ac55b51817ae	766166	CAD	Create CAD designs for the project	In Progress	2025-08-12 08:29:24.193726+00			2025-08-12 07:42:14.02688+00	2025-08-12 08:29:24.193786+00	\N	1754984525695
4a275045-ef1b-4caa-8f2c-4958b0a7fd00	766166	Cutting	Cut materials according to specifications	Completed	2025-08-12 08:29:27.116857+00			2025-08-12 07:09:28.446184+00	2025-08-12 08:29:27.116913+00	\N	1754982511168
1c56b355-d653-460f-b13e-073371da7e1a	766166	Fabrication	Fabricate components for the project	In Progress	2025-08-12 08:29:27.199728+00			2025-08-12 07:42:14.026881+00	2025-08-12 08:29:27.199832+00	\N	1754984525695
15845647-4a30-4d80-877f-5a3636dec273	766166	Fabrication	Fabricate components for the project	Completed	2025-08-12 08:29:29.841019+00			2025-08-12 07:09:28.446185+00	2025-08-12 08:29:29.841073+00	\N	1754982511168
a100910d-b4c8-4f29-aabe-5a080f15a8b5	766166	Installation	Install the project components on-site	Completed	2025-08-12 08:29:32.466461+00			2025-08-12 07:09:28.446185+00	2025-08-12 08:29:32.466542+00	\N	1754982511168
9594b25a-7e11-446c-95f4-5fd98354e97b	766166	Service Call	Handle any service calls or maintenance	In Progress	2025-08-12 08:29:32.503704+00			2025-08-12 07:42:14.026881+00	2025-08-12 08:29:32.503755+00	\N	1754984525695
0c37958c-b694-43ad-bfe1-00d0c74a3447	766166	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 08:29:35.191066+00			2025-08-12 07:09:28.446186+00	2025-08-12 08:29:35.191123+00	\N	1754982511168
01e5fb4a-0e60-4d33-8524-c1da9993513d	884011	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 08:52:42.872844+00	2025-08-12 08:52:42.973999+00	\N	1754988712647
908e7073-a24b-4007-b0a5-f48ace82a0e6	022652	Cutting	Cut materials according to specifications	Completed	2025-08-12 08:35:53.820153+00			2025-08-12 06:36:18.571274+00	2025-08-12 08:35:53.820183+00	\N	1754980523379
ab85c1d0-ad0e-4733-aa81-f92709c2deff	022652	Fabrication	Fabricate components for the project	Completed	2025-08-12 08:35:55.307973+00			2025-08-12 06:36:18.571274+00	2025-08-12 08:35:55.308005+00	\N	1754980523379
4d6e0f6d-8f2a-48e7-83cc-1f503fe628cc	884011	Template	Prepare and finalize the template for the project	Completed	2025-08-12 08:39:47.050613+00			2025-08-12 08:34:11.716391+00	2025-08-12 08:39:47.050639+00	\N	1754987522866
68863f1a-9926-4a97-aa7a-6b8a0c1bd27f	884011	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 08:52:42.872845+00	2025-08-12 08:52:43.060064+00	\N	1754988712647
b437224d-abba-4bda-a3f4-a1aa4f733cf6	884011	CAD	Create CAD designs for the project	Completed	2025-08-12 08:39:49.346372+00			2025-08-12 08:34:11.716392+00	2025-08-12 08:39:49.346398+00	\N	1754987522866
26c0c970-2fc5-472f-8d60-44f3e22edcaf	884011	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 08:52:42.872845+00	2025-08-12 08:52:43.116181+00	\N	1754988712647
b629985e-ccb5-4b8d-93f5-cac3ce3d728f	884011	Cutting	Cut materials according to specifications	Completed	2025-08-12 08:39:51.304235+00			2025-08-12 08:34:11.716393+00	2025-08-12 08:39:51.304265+00	\N	1754987522866
f3b348d2-1079-4215-8a7a-845b55216738	884011	Installation	Install the project components on-site	Not Started	\N			2025-08-12 08:52:42.872846+00	2025-08-12 08:52:43.161476+00	\N	1754988712647
e40b4d1b-6ff0-4ad3-9aba-4a073e0cfc31	884011	Fabrication	Fabricate components for the project	Completed	2025-08-12 08:39:53.793298+00			2025-08-12 08:34:11.716393+00	2025-08-12 08:39:53.793327+00	\N	1754987522866
40453613-e31e-4b38-9b88-6148cf37058a	884011	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 08:52:42.872846+00	2025-08-12 08:52:43.235134+00	\N	1754988712647
7db69618-b58e-4b16-8074-3d3853535d3a	884011	Installation	Install the project components on-site	Completed	2025-08-12 08:39:55.798595+00			2025-08-12 08:34:11.716394+00	2025-08-12 08:39:55.798623+00	\N	1754987522866
84dbcc85-5c51-41f7-8756-272d3dcac1fd	884011	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 08:39:57.497895+00			2025-08-12 08:34:11.716396+00	2025-08-12 08:39:57.497917+00	\N	1754987522866
d927b2e9-f408-4342-8f32-1a8fb18a5f6b	301498	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 08:59:38.243352+00	2025-08-12 08:59:38.457839+00	\N	1754989034701
2edc1a69-46ae-45b6-af79-714a722f4a2a	301498	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 08:59:38.243354+00	2025-08-12 08:59:38.497969+00	\N	1754989034701
910aa00e-4d3f-4e82-b77b-e966faa27a0a	301498	Installation	Install the project components on-site	Not Started	\N			2025-08-12 08:59:38.243355+00	2025-08-12 08:59:38.540738+00	\N	1754989034701
92d48c56-ceb1-4d3d-8492-a6b0db7ddc0e	301498	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 08:59:38.243357+00	2025-08-12 08:59:38.578487+00	\N	1754989034701
5e162d83-d47e-4f40-9627-dd0b22a8e8a2	301498	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 08:59:38.72806+00	2025-08-12 08:59:38.762469+00	\N	1754989043654
25e0dade-18cb-49b5-b14d-b34d489f87d5	301498	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 08:59:38.728061+00	2025-08-12 08:59:38.809172+00	\N	1754989043654
a166dc39-48b5-44ff-81e7-8d29557f8a46	301498	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 08:59:38.728063+00	2025-08-12 08:59:38.858928+00	\N	1754989043654
f2792f6c-03c7-4593-a945-a68353fd2a41	301498	Installation	Install the project components on-site	Not Started	\N			2025-08-12 08:59:38.728064+00	2025-08-12 08:59:38.898043+00	\N	1754989043654
66dfeb0e-2cc2-4c51-968d-a8eb88484f08	301498	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 08:59:38.728066+00	2025-08-12 08:59:38.957632+00	\N	1754989043654
f85dd802-6252-4335-a39e-58e16179ebd8	301498	Template	Prepare and finalize the template for the project	Completed	2025-08-12 09:19:43.073762+00			2025-08-12 08:59:38.243349+00	2025-08-12 09:19:43.073826+00	\N	1754989034701
41d2f096-15c6-4a4d-9874-9d46612e4254	973844	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 09:01:16.915069+00			2025-08-12 09:01:16.915068+00	2025-08-12 09:01:16.915109+00	\N	1754989105243
dfa251a3-f324-496e-9a6d-0f19017cce27	973844	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 09:01:16.915069+00	2025-08-12 09:01:16.998584+00	\N	1754989105243
f9ff5e0c-f30a-475f-b27a-0217f5e593d9	973844	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 09:01:16.91507+00	2025-08-12 09:01:17.081239+00	\N	1754989105243
c3fec955-1c53-44a9-8280-c53f26db6cb9	973844	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 09:01:16.91507+00	2025-08-12 09:01:17.112663+00	\N	1754989105243
3533e1ed-1de5-4a16-bc10-ee0266767268	973844	Installation	Install the project components on-site	Not Started	\N			2025-08-12 09:01:16.915071+00	2025-08-12 09:01:17.152462+00	\N	1754989105243
2fdcb0ee-98b2-43ff-b8ec-1468a6e1e001	973844	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 09:01:16.915071+00	2025-08-12 09:01:17.212416+00	\N	1754989105243
5a614114-e213-4a81-8054-615d86f4e7f1	973844	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 09:01:17.417018+00			2025-08-12 09:01:17.417018+00	2025-08-12 09:01:17.417076+00	\N	1754989115005
555f159b-45d9-4eb6-976f-e29c500f1b53	973844	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 09:01:17.417019+00	2025-08-12 09:01:17.462206+00	\N	1754989115005
ff6fb0ec-8b06-4bee-a2cb-3cee54fb6de1	973844	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 09:01:17.417019+00	2025-08-12 09:01:17.498361+00	\N	1754989115005
10a48d2c-5430-4c67-b848-361c5ab3557c	973844	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 09:01:17.41702+00	2025-08-12 09:01:17.538713+00	\N	1754989115005
e0bd3e9f-c973-404c-9975-238b43ad6006	973844	Installation	Install the project components on-site	Not Started	\N			2025-08-12 09:01:17.41702+00	2025-08-12 09:01:17.574596+00	\N	1754989115005
9a50d1f9-2611-45a9-b20e-42ee7d6376e4	973844	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 09:01:17.417021+00	2025-08-12 09:01:17.615483+00	\N	1754989115005
d0aab0ac-d504-4cff-a149-d89533e830c8	078083	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 09:02:12.544063+00			2025-08-12 09:02:12.544063+00	2025-08-12 09:02:12.544099+00	\N	1754989321905
4dda5c7b-5dd7-4ca4-930c-141b0a26a87e	078083	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 09:02:12.544063+00	2025-08-12 09:02:12.649758+00	\N	1754989321905
a1e0c343-b55c-44e4-b436-04592d8daca4	078083	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 09:02:12.544063+00	2025-08-12 09:02:12.737304+00	\N	1754989321905
a866a4ba-f15f-4756-9890-e5fe23e7f017	078083	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 09:02:12.544064+00	2025-08-12 09:02:12.791968+00	\N	1754989321905
321324d5-baf7-4f9e-a8d3-67252aa06862	078083	Installation	Install the project components on-site	Not Started	\N			2025-08-12 09:02:12.544065+00	2025-08-12 09:02:12.830288+00	\N	1754989321905
b00ffb33-3fa3-4599-a0f9-7e75c62cf536	078083	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 09:02:12.544065+00	2025-08-12 09:02:12.877966+00	\N	1754989321905
e9ea53ca-db4c-4329-9b85-66ea933ffe5e	884011	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 09:04:43.998633+00			2025-08-12 09:04:43.998631+00	2025-08-12 09:04:43.998669+00	\N	1754989461769
c7f2b6b3-472a-481b-a255-e9475af32e3a	884011	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 09:04:43.998634+00	2025-08-12 09:04:44.078689+00	\N	1754989461769
76f264fa-db7d-443d-92ef-6f418743e5b7	884011	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 09:04:43.998634+00	2025-08-12 09:04:44.151067+00	\N	1754989461769
85ec6048-0ac4-468c-bccf-06596befea53	884011	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 09:04:43.998635+00	2025-08-12 09:04:44.205898+00	\N	1754989461769
f8d35763-3bd6-4f28-84ab-6fdc4ef312f9	884011	Installation	Install the project components on-site	Not Started	\N			2025-08-12 09:04:43.998635+00	2025-08-12 09:04:44.264277+00	\N	1754989461769
224931ec-226f-4ece-8b74-8f363e555463	884011	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 09:04:43.998636+00	2025-08-12 09:04:44.303664+00	\N	1754989461769
975764ed-5e4c-49d7-8350-d1c7409b8002	301498	Template	Prepare and finalize the template for the project	Completed	2025-08-12 09:10:44.276171+00			2025-08-12 08:59:38.728057+00	2025-08-12 09:10:44.276231+00	\N	1754989043654
499c75ea-834d-4fd7-b7c6-d11214f20d34	554098	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 09:16:04.743102+00			2025-08-12 09:16:04.743102+00	2025-08-12 09:16:04.786676+00	\N	1754990000977
7389b1c5-b394-474f-88e7-0c042b44c394	554098	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 09:16:04.743104+00	2025-08-12 09:16:04.968684+00	\N	1754990000977
b65d0cb0-69c7-452b-8e80-394be8239138	554098	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 09:16:04.743106+00	2025-08-12 09:16:05.127353+00	\N	1754990000977
d108949e-8c7c-458c-baf2-cadc9a8536bf	554098	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 09:16:04.743107+00	2025-08-12 09:16:05.237745+00	\N	1754990000977
9bf50629-d8fb-4cb3-927c-05d38f5ad0b8	554098	Installation	Install the project components on-site	Not Started	\N			2025-08-12 09:16:04.743108+00	2025-08-12 09:16:05.372116+00	\N	1754990000977
1e3d551d-65b2-4177-93f3-1f50882fe4ec	554098	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 09:16:04.74311+00	2025-08-12 09:16:05.497469+00	\N	1754990000977
c9f6071c-d68d-449c-9ca5-1ef15abb6651	301498	CAD	Create CAD designs for the project	In Progress	2025-08-12 09:19:48.484585+00			2025-08-12 08:59:38.243351+00	2025-08-12 09:19:48.484636+00	\N	1754989034701
342131ba-7713-4940-8aa2-faaec324a498	297805	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 09:41:42.524232+00			2025-08-12 09:41:42.524232+00	2025-08-12 09:41:42.524292+00	\N	1754991678168
c63755c3-5ccc-40c6-97aa-795baf0795d5	297805	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 09:41:42.524233+00	2025-08-12 09:41:42.595281+00	\N	1754991678168
16bd5b88-5189-48fe-b7ee-bdcce8c31382	297805	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 09:41:42.524234+00	2025-08-12 09:41:42.664996+00	\N	1754991678168
cfb28da1-76d6-4b0c-b44f-38db70fcea1b	297805	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 09:41:42.524235+00	2025-08-12 09:41:42.700215+00	\N	1754991678168
2a02c5e3-a763-4401-bfff-af3170459fd7	297805	Installation	Install the project components on-site	Not Started	\N			2025-08-12 09:41:42.524235+00	2025-08-12 09:41:42.733955+00	\N	1754991678168
111092bc-174f-4fcd-b709-779985e19a19	297805	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 09:41:42.524235+00	2025-08-12 09:41:42.76974+00	\N	1754991678168
f771ae3d-bc5a-4bbc-9e72-6d2635d632cf	820700	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 09:44:51.849387+00			2025-08-12 09:44:51.849387+00	2025-08-12 09:44:51.849423+00	\N	1754991834726
553aa847-246d-44da-a792-04c7b509f2bd	820700	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 09:44:51.849388+00	2025-08-12 09:44:51.920142+00	\N	1754991834726
d79bcc55-2e80-404f-b0c1-817d35d223eb	820700	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 09:44:51.849388+00	2025-08-12 09:44:52.030967+00	\N	1754991834726
f5c95edc-d3bd-4d66-847c-a21309060da0	820700	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 09:44:51.849388+00	2025-08-12 09:44:52.089808+00	\N	1754991834726
6c763a92-2d75-4a54-baf4-1b26b9199a52	820700	Installation	Install the project components on-site	Not Started	\N			2025-08-12 09:44:51.849388+00	2025-08-12 09:44:52.149797+00	\N	1754991834726
4572dd91-a7e7-4570-a46e-e9cb54700bf6	820700	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 09:44:51.849389+00	2025-08-12 09:44:52.231154+00	\N	1754991834726
8b0de601-7fed-46fd-a461-3464da07c664	820700	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 09:44:52.434648+00			2025-08-12 09:44:52.434648+00	2025-08-12 09:44:52.434692+00	\N	1754991841419
1456aed3-a565-4e65-abd9-286acdd8c084	820700	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 09:44:52.434649+00	2025-08-12 09:44:52.470397+00	\N	1754991841419
52a8a5d2-73bc-4e80-812e-69bcc3befefc	820700	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 09:44:52.434649+00	2025-08-12 09:44:52.509694+00	\N	1754991841419
afb4fb22-fe9b-4a3c-ad48-4d39354fed64	820700	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 09:44:52.434649+00	2025-08-12 09:44:52.546395+00	\N	1754991841419
8cf66019-72ee-41dc-8014-3f23a2661f1c	820700	Installation	Install the project components on-site	Not Started	\N			2025-08-12 09:44:52.43465+00	2025-08-12 09:44:52.610625+00	\N	1754991841419
119705f9-2f95-4b21-8bd5-c11e4490ec6e	820700	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 09:44:52.434651+00	2025-08-12 09:44:52.651338+00	\N	1754991841419
3b4679c0-e5d8-4d7f-82aa-3ef08d26a0da	440604	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 09:48:26.668145+00			2025-08-12 09:48:26.668145+00	2025-08-12 09:48:26.66818+00	\N	1754992010350
7e441bbc-231b-439c-95d5-5d8fc2d41ac6	440604	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 09:48:26.668146+00	2025-08-12 09:48:26.755671+00	\N	1754992010350
883bcef7-2a05-4f17-9789-8443382c1b16	440604	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 09:48:26.668146+00	2025-08-12 09:48:26.827628+00	\N	1754992010350
9f7c2de7-d3e5-4cac-ba2a-94bdf6cb6cf7	440604	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 09:48:26.668146+00	2025-08-12 09:48:26.892325+00	\N	1754992010350
06176c35-4815-4511-bfdb-fc8e574fb410	440604	Installation	Install the project components on-site	Not Started	\N			2025-08-12 09:48:26.668147+00	2025-08-12 09:48:26.929212+00	\N	1754992010350
08458a1f-7753-43ea-8690-47ea82432577	440604	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 09:48:26.668147+00	2025-08-12 09:48:26.98724+00	\N	1754992010350
611a6a89-2bda-4cbe-beba-edab01ef07d9	440604	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 09:50:04.156686+00			2025-08-12 09:50:04.156686+00	2025-08-12 09:50:04.200195+00	\N	1754992172408
27ca40bf-7454-4765-90ab-5527a6120f56	440604	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 09:50:04.156687+00	2025-08-12 09:50:04.386505+00	\N	1754992172408
01cb203d-9a66-46c8-ab00-e8335626f671	440604	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 09:50:04.156687+00	2025-08-12 09:50:04.536851+00	\N	1754992172408
3458c094-450e-4887-9a74-b4377de78706	440604	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 09:50:04.156687+00	2025-08-12 09:50:04.650646+00	\N	1754992172408
4c03f50a-9b2b-4840-a8fa-702adc961a9d	440604	Installation	Install the project components on-site	Not Started	\N			2025-08-12 09:50:04.156687+00	2025-08-12 09:50:04.809732+00	\N	1754992172408
cad7b9de-7b12-4964-b217-8f7b270b3660	440604	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 09:50:04.156687+00	2025-08-12 09:50:04.948226+00	\N	1754992172408
f7f9d51d-a2cc-49c3-a5b6-e7b38c078516	701868	Template	Prepare and finalize the template for the project	Completed	2025-08-12 10:19:45.608894+00			2025-08-12 10:04:09.885137+00	2025-08-12 10:19:45.608952+00	\N	1754992910045
5d391faa-4855-47bb-9efb-e7a43ec5d77d	701868	CAD	Create CAD designs for the project	Completed	2025-08-12 10:19:48.96226+00			2025-08-12 10:04:09.885138+00	2025-08-12 10:19:48.962342+00	\N	1754992910045
fd9c6af2-19ac-4e84-97e6-48f6b7318b19	701868	Cutting	Cut materials according to specifications	Completed	2025-08-12 10:20:21.194466+00			2025-08-12 10:04:10.33901+00	2025-08-12 10:20:21.194512+00	\N	1754993027495
7002490a-d7f5-474a-8bc8-42e4e83327ec	701868	Cutting	Cut materials according to specifications	Completed	2025-08-12 10:19:52.098969+00			2025-08-12 10:04:09.885138+00	2025-08-12 10:19:52.099011+00	\N	1754992910045
9e13109f-7731-4376-b725-c3a72e6b9f43	701868	Fabrication	Fabricate components for the project	Completed	2025-08-12 10:20:24.155122+00			2025-08-12 10:04:10.33901+00	2025-08-12 10:20:24.155194+00	\N	1754993027495
3fbf5b1e-6156-4494-a16d-652724cec213	701868	Fabrication	Fabricate components for the project	Completed	2025-08-12 10:19:56.11381+00			2025-08-12 10:04:09.885139+00	2025-08-12 10:19:56.11387+00	\N	1754992910045
759b01a3-47ad-4e9f-b907-0866cf286096	701868	Installation	Install the project components on-site	Completed	2025-08-12 10:20:26.859901+00			2025-08-12 10:04:10.33901+00	2025-08-12 10:20:26.859946+00	\N	1754993027495
142534a4-1020-4532-8fa8-d178c8400e3d	701868	Installation	Install the project components on-site	Completed	2025-08-12 10:20:00.241509+00			2025-08-12 10:04:09.885139+00	2025-08-12 10:20:00.241578+00	\N	1754992910045
793f653e-9e19-4e2c-a444-7223daa082e0	701868	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 10:35:15.561825+00			2025-08-12 10:04:09.88514+00	2025-08-12 10:35:15.561853+00	\N	1754992910045
c4c620ef-d4e7-402c-8967-e62942f05774	701868	Template	Prepare and finalize the template for the project	Completed	2025-08-12 10:20:36.635203+00			2025-08-12 10:12:06.143098+00	2025-08-12 10:20:36.635237+00	\N	1754993468042
2985d581-2f0e-431a-950b-3c2da48925b4	701868	Template	Prepare and finalize the template for the project	Completed	2025-08-12 10:20:15.438864+00			2025-08-12 10:04:10.339008+00	2025-08-12 10:20:15.438914+00	\N	1754993027495
7d487417-1c7a-4b22-936f-a619658e07b0	701868	CAD	Create CAD designs for the project	Completed	2025-08-12 10:20:18.019802+00			2025-08-12 10:04:10.339009+00	2025-08-12 10:20:18.019854+00	\N	1754993027495
f610b574-be6a-42f9-b70f-f2522982ff3a	701868	Cutting	Cut materials according to specifications	Completed	2025-08-12 10:20:41.960372+00			2025-08-12 10:12:06.1431+00	2025-08-12 10:20:41.960455+00	\N	1754993468042
230f139f-fd89-4973-9c8d-00f36ba3263e	701868	Fabrication	Fabricate components for the project	Completed	2025-08-12 10:20:44.358511+00			2025-08-12 10:12:06.1431+00	2025-08-12 10:20:44.358564+00	\N	1754993468042
17c1d14a-29a7-4238-aeaa-1a7eef659947	701868	Installation	Install the project components on-site	Completed	2025-08-12 10:20:47.059844+00			2025-08-12 10:12:06.143101+00	2025-08-12 10:20:47.059902+00	\N	1754993468042
1e04bf6a-d1c0-421c-aadd-9a3d3e73d080	701868	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 10:36:13.565158+00			2025-08-12 10:04:10.339011+00	2025-08-12 10:36:13.565189+00	\N	1754993027495
a5b31e93-2858-46b9-a62d-cd24ca0931c0	701868	CAD	Create CAD designs for the project	Completed	2025-08-12 10:20:39.208292+00			2025-08-12 10:12:06.143099+00	2025-08-12 10:20:39.208428+00	\N	1754993468042
68066585-0379-4278-b001-e446d354f825	701868	Fabrication	Fabricate components for the project	Completed	2025-08-12 10:21:05.089653+00			2025-08-12 10:14:24.246189+00	2025-08-12 10:21:05.089714+00	\N	1754993602935
63950017-511d-4a88-a5d6-c3a2287cd7f3	701868	Template	Prepare and finalize the template for the project	Completed	2025-08-12 10:20:56.840022+00			2025-08-12 10:14:24.246186+00	2025-08-12 10:20:56.840108+00	\N	1754993602935
e9fe11a9-1ced-4afe-abc7-2f5d2150b414	701868	CAD	Create CAD designs for the project	Completed	2025-08-12 10:20:59.613962+00			2025-08-12 10:14:24.246187+00	2025-08-12 10:20:59.61401+00	\N	1754993602935
4de7c350-ae29-48fc-8923-258519a64cc1	701868	Cutting	Cut materials according to specifications	Completed	2025-08-12 10:21:02.201013+00			2025-08-12 10:14:24.246188+00	2025-08-12 10:21:02.201069+00	\N	1754993602935
96e7388f-4f66-41a6-86d4-cae800ed77ed	701868	Installation	Install the project components on-site	Completed	2025-08-12 10:21:07.880977+00			2025-08-12 10:14:24.24619+00	2025-08-12 10:21:07.881015+00	\N	1754993602935
16900605-abf1-45e6-86c0-814869531601	701868	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 10:36:03.208908+00			2025-08-12 10:14:24.24619+00	2025-08-12 10:36:03.209274+00	\N	1754993602935
3100bd0d-a0ce-485d-8814-3df6938f3c3b	701868	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 10:36:08.42665+00			2025-08-12 10:12:06.143101+00	2025-08-12 10:36:08.426674+00	\N	1754993468042
e04f9825-eb85-45e9-a21e-68d99431517b	487281	Template	Prepare and finalize the template for the project	Completed	2025-08-12 11:29:20.947441+00			2025-08-12 10:41:53.723808+00	2025-08-12 11:29:20.947477+00	\N	1754995164876
0d74826b-da90-429f-88ef-bface448e503	487281	CAD	Create CAD designs for the project	Completed	2025-08-12 12:47:51.721935+00			2025-08-12 10:41:53.723809+00	2025-08-12 12:47:51.721966+00	\N	1754995164876
740172c8-4caa-4628-94f8-4ed0960b3afa	487281	Fabrication	Fabricate components for the project	Completed	2025-08-12 13:18:09.379277+00			2025-08-12 10:44:28.038935+00	2025-08-12 13:18:09.379302+00	\N	1754995393376
1266b7a7-2715-4ec3-9a90-0bd4fe3ad284	858123	Template	Prepare and finalize the template for the project	In Progress	2025-08-12 12:55:30.010594+00			2025-08-12 12:55:30.010593+00	2025-08-12 12:55:30.010663+00	\N	1755003067973
29194a2a-dde2-40aa-99dd-38cb7a10dbb6	858123	CAD	Create CAD designs for the project	Not Started	\N			2025-08-12 12:55:30.010595+00	2025-08-12 12:55:30.013604+00	\N	1755003067973
54c5b4fd-f0e5-420f-aaf2-a2216da0b132	858123	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-12 12:55:30.010597+00	2025-08-12 12:55:30.0157+00	\N	1755003067973
69da4525-47a7-4b8e-b53a-303d5644c546	858123	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-12 12:55:30.010599+00	2025-08-12 12:55:30.017481+00	\N	1755003067973
24e6cd6d-be7f-4bc7-b8ec-d44ed8c80339	858123	Installation	Install the project components on-site	Not Started	\N			2025-08-12 12:55:30.0106+00	2025-08-12 12:55:30.018498+00	\N	1755003067973
cc0fedac-e9d1-4fc9-832d-f042fe5536fb	858123	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-12 12:55:30.010601+00	2025-08-12 12:55:30.019506+00	\N	1755003067973
80d1cb48-11f0-4952-889a-45b8b09adddf	487281	Installation	Install the project components on-site	Completed	2025-08-12 13:17:46.494582+00			2025-08-12 10:41:53.723813+00	2025-08-12 13:17:46.494617+00	\N	1754995164876
747d920a-564b-48c2-8a89-eb379a112792	487281	Cutting	Cut materials according to specifications	Completed	2025-08-12 13:17:40.80699+00			2025-08-12 10:41:53.72381+00	2025-08-12 13:17:40.807023+00	\N	1754995164876
fade50ae-fc11-47a9-98e9-5dd4394576d8	487281	Template	Prepare and finalize the template for the project	Completed	2025-08-12 13:18:53.946967+00			2025-08-12 13:14:24.98478+00	2025-08-12 13:18:53.947039+00	\N	1755004409938
60da81a7-6063-4262-b309-af8d1d9faf76	487281	Fabrication	Fabricate components for the project	Completed	2025-08-12 13:17:43.099131+00			2025-08-12 10:41:53.723812+00	2025-08-12 13:17:43.09916+00	\N	1754995164876
bd2224f4-85a3-421e-85a2-8b7258d6cd2b	487281	Cutting	Cut materials according to specifications	Completed	2025-08-12 13:18:28.733788+00			2025-08-12 13:11:38.58438+00	2025-08-12 13:18:28.733814+00	\N	1755004271757
502ea663-d3ed-4227-baa0-b80a53e1d9e4	467243	Template	Prepare and finalize the template for the project	Completed	2025-08-12 23:19:29.518985+00			2025-08-12 23:11:54.987517+00	2025-08-12 23:19:29.519014+00	\N	1755040181565
b91af1a8-51f5-40b3-a167-8c3caf7f4f38	487281	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 13:17:48.724214+00			2025-08-12 10:41:53.723814+00	2025-08-12 13:17:48.724245+00	\N	1754995164876
8f754fba-a4a1-4c21-b31f-3e9870cd3640	487281	Fabrication	Fabricate components for the project	Completed	2025-08-12 13:18:30.9592+00			2025-08-12 13:11:38.584382+00	2025-08-12 13:18:30.959255+00	\N	1755004271757
a1903abf-3ba5-466d-b629-3047efcb538a	487281	Template	Prepare and finalize the template for the project	Completed	2025-08-12 13:17:56.626641+00			2025-08-12 10:44:28.038931+00	2025-08-12 13:17:56.62667+00	\N	1754995393376
315e2e67-a3ea-4807-9524-96aaf273a1fc	487281	CAD	Create CAD designs for the project	Completed	2025-08-12 13:18:02.064704+00			2025-08-12 10:44:28.038933+00	2025-08-12 13:18:02.064739+00	\N	1754995393376
7008b5c1-2161-47f4-b97c-edb0b2a87e2f	467243	Cutting	Cut materials according to specifications	Completed	2025-08-12 23:20:57.207661+00			2025-08-12 23:11:54.98752+00	2025-08-12 23:20:57.207712+00	\N	1755040181565
50fb21e0-26f1-49cd-a881-3eafe62f6d0a	487281	Cutting	Cut materials according to specifications	Completed	2025-08-12 13:18:04.780509+00			2025-08-12 10:44:28.038934+00	2025-08-12 13:18:04.780549+00	\N	1754995393376
e9c3310d-e5fd-4cea-8357-dc1b97f2179f	487281	CAD	Create CAD designs for the project	Completed	2025-08-12 13:18:56.33402+00			2025-08-12 13:14:24.984781+00	2025-08-12 13:18:56.33405+00	\N	1755004409938
81f3040c-62b3-47bc-8e92-8a4f3286a571	487281	Installation	Install the project components on-site	Completed	2025-08-12 13:18:12.579738+00			2025-08-12 10:44:28.038936+00	2025-08-12 13:18:12.57977+00	\N	1754995393376
484d78b6-8127-49ae-a931-7a2a1ba83faa	487281	Cutting	Cut materials according to specifications	Completed	2025-08-12 13:18:59.038744+00			2025-08-12 13:14:24.984782+00	2025-08-12 13:18:59.038769+00	\N	1755004409938
ca334bb4-5400-4a64-8bd1-6b53a4ee169f	487281	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 13:18:17.784383+00			2025-08-12 10:44:28.038937+00	2025-08-12 13:18:17.784408+00	\N	1754995393376
3e03802e-a104-468a-ae8d-458cf53413b5	487281	Fabrication	Fabricate components for the project	Completed	2025-08-12 13:19:01.285677+00			2025-08-12 13:14:24.984784+00	2025-08-12 13:19:01.285709+00	\N	1755004409938
d8770b44-49da-4666-9e75-7f6e93bb0ca8	487281	Template	Prepare and finalize the template for the project	Completed	2025-08-12 13:18:23.15894+00			2025-08-12 13:11:38.584377+00	2025-08-12 13:18:23.158965+00	\N	1755004271757
01215f2a-4ed5-4646-b50a-69715207a28c	487281	CAD	Create CAD designs for the project	Completed	2025-08-12 13:18:26.458749+00			2025-08-12 13:11:38.584379+00	2025-08-12 13:18:26.458775+00	\N	1755004271757
2f8dc159-7435-4efc-a07b-f6701c0f2b1e	467243	CAD	Create CAD designs for the project	Completed	2025-08-12 23:19:33.092565+00			2025-08-12 23:11:54.987519+00	2025-08-12 23:19:33.092591+00	\N	1755040181565
98eff946-5d8d-4872-82f3-43041622e60d	487281	Installation	Install the project components on-site	Completed	2025-08-12 13:18:33.249584+00			2025-08-12 13:11:38.584384+00	2025-08-12 13:18:33.249614+00	\N	1755004271757
b925527d-fd4f-4d6c-a41d-ed9b8ed629ca	467243	Fabrication	Fabricate components for the project	Completed	2025-08-12 23:21:53.851531+00			2025-08-12 23:11:54.987521+00	2025-08-12 23:21:53.851632+00	\N	1755040181565
e3cadf85-93dc-4783-afef-c43fbbf97ec4	487281	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 13:18:35.463306+00			2025-08-12 13:11:38.584385+00	2025-08-12 13:18:35.463335+00	\N	1755004271757
c99feb3c-8653-4148-914b-2ce1137ee3b2	467243	Installation	Install the project components on-site	Completed	2025-08-12 23:21:59.139767+00			2025-08-12 23:11:54.987522+00	2025-08-12 23:21:59.139797+00	\N	1755040181565
e34bb30b-0003-45be-a5d5-64f85ae9c863	487281	Installation	Install the project components on-site	Completed	2025-08-12 13:19:04.298705+00			2025-08-12 13:14:24.984785+00	2025-08-12 13:19:04.298726+00	\N	1755004409938
a8579203-b1d9-4d0a-aace-f6bbed45b28b	487281	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 13:19:11.259009+00			2025-08-12 13:14:24.984786+00	2025-08-12 13:19:11.259036+00	\N	1755004409938
bb64ad6c-02bd-415b-9c65-1cd241db7a84	467243	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 23:42:35.291972+00			2025-08-12 23:11:54.987523+00	2025-08-12 23:42:35.291998+00	\N	1755040181565
664c356a-e444-4318-9bf1-4326f7ee9087	253703	Template	Prepare and finalize the template for the project	Completed	2025-08-12 23:57:59.796767+00			2025-08-12 23:50:15.140404+00	2025-08-12 23:57:59.796794+00	\N	1755042280869
f86847e2-2066-441b-b245-f87c599695ee	253703	Cutting	Cut materials according to specifications	Completed	2025-08-12 23:58:19.063479+00			2025-08-12 23:50:15.140408+00	2025-08-12 23:58:19.063502+00	\N	1755042280869
01bbcf8e-9be1-4930-9a5a-675bf92d2de9	253703	Installation	Install the project components on-site	Completed	2025-08-12 23:58:24.489563+00			2025-08-12 23:50:15.140411+00	2025-08-12 23:58:24.489607+00	\N	1755042280869
e182a799-2726-4b2e-b463-99689b272368	253703	CAD	Create CAD designs for the project	Completed	2025-08-12 23:58:12.248232+00			2025-08-12 23:50:15.140406+00	2025-08-12 23:58:12.248256+00	\N	1755042280869
e83088b4-907e-4606-86ff-b7a3b8051427	253703	Fabrication	Fabricate components for the project	Completed	2025-08-12 23:58:22.272799+00			2025-08-12 23:50:15.140409+00	2025-08-12 23:58:22.272835+00	\N	1755042280869
c06a5815-fa33-4deb-822e-36629c2ccbc2	323585	Template	Prepare and finalize the template for the project	Completed	2025-08-13 00:47:28.034824+00			2025-08-13 00:40:48.117738+00	2025-08-13 00:47:28.034858+00	\N	1755045602331
b8ecec96-4e8d-404d-9b18-a1ec236270db	253703	Service Call	Handle any service calls or maintenance	Completed	2025-08-12 23:58:26.947335+00			2025-08-12 23:50:15.140413+00	2025-08-12 23:58:26.947357+00	\N	1755042280869
407d1483-9d0d-4e59-9db5-ebe6cca8926b	317521	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-13 00:07:05.521852+00	2025-08-13 00:07:05.530059+00	\N	1755043530813
b785e8fd-c20f-4839-b2d6-db880a07e892	317521	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-13 00:07:05.521853+00	2025-08-13 00:07:05.533025+00	\N	1755043530813
6a346b9a-84db-4126-9243-74de646cdaba	317521	Installation	Install the project components on-site	Not Started	\N			2025-08-13 00:07:05.521854+00	2025-08-13 00:07:05.536108+00	\N	1755043530813
5e13cc10-9516-49b7-beda-87b43cfd8228	317521	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-13 00:07:05.521856+00	2025-08-13 00:07:05.538919+00	\N	1755043530813
6cdbe368-ab15-4381-a735-e11846fe390d	317521	Template	Prepare and finalize the template for the project	Completed	2025-08-13 00:13:03.918662+00			2025-08-13 00:07:05.521848+00	2025-08-13 00:13:03.918688+00	\N	1755043530813
6b422252-5f79-49ac-8a26-fe4335686594	317521	CAD	Create CAD designs for the project	In Progress	2025-08-13 00:13:03.921039+00			2025-08-13 00:07:05.52185+00	2025-08-13 00:13:03.921067+00	\N	1755043530813
00bd35c4-02aa-4b9f-ae76-eb1ee4efb3bf	022137	Installation	Install the project components on-site	Not Started	\N			2025-08-13 00:23:57.530504+00	2025-08-13 00:23:57.537106+00	\N	1755044598688
899893c3-0155-4fe9-9ddf-84affa1be3c7	022137	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-13 00:23:57.530505+00	2025-08-13 00:23:57.538193+00	\N	1755044598688
cde1c72a-938c-4cad-b255-def36af19121	022137	Installation	Install the project components on-site	Not Started	\N			2025-08-13 00:23:57.542093+00	2025-08-13 00:23:57.546398+00	\N	1755044605893
b379caaa-5429-42e6-8a6e-320d54f9762f	022137	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-13 00:23:57.542094+00	2025-08-13 00:23:57.547342+00	\N	1755044605893
cff949f8-4d68-47e5-8ea3-40a1f651c44a	022137	Template	Prepare and finalize the template for the project	Completed	2025-08-13 00:26:00.156063+00			2025-08-13 00:23:57.530499+00	2025-08-13 00:26:00.1561+00	\N	1755044598688
1543050a-e7af-44d4-bc73-c5d3460d7e12	022137	Template	Prepare and finalize the template for the project	Completed	2025-08-13 00:26:14.501647+00			2025-08-13 00:23:57.542053+00	2025-08-13 00:26:14.501676+00	\N	1755044605893
e2f332eb-bb0a-4010-82d4-9cba1c425afc	022137	CAD	Create CAD designs for the project	Completed	2025-08-13 00:26:22.218625+00			2025-08-13 00:23:57.5305+00	2025-08-13 00:26:22.218652+00	\N	1755044598688
d59f53e5-447b-4612-8152-4b2a3e477376	022137	CAD	Create CAD designs for the project	In Progress	2025-08-13 00:26:22.219476+00			2025-08-13 00:23:57.542054+00	2025-08-13 00:26:22.219496+00	\N	1755044605893
f30d8116-d323-43d1-b32e-2986e2f30ca6	022137	Cutting	Cut materials according to specifications	Completed	2025-08-13 00:26:31.450819+00			2025-08-13 00:23:57.530502+00	2025-08-13 00:26:31.450857+00	\N	1755044598688
d205a932-617a-4b7d-9f55-1f171606829e	022137	Cutting	Cut materials according to specifications	In Progress	2025-08-13 00:26:31.451972+00			2025-08-13 00:23:57.542056+00	2025-08-13 00:26:31.452007+00	\N	1755044605893
6fa2b636-3096-48fa-a896-77ad37b6fd4e	022137	Fabrication	Fabricate components for the project	Completed	2025-08-13 00:26:34.320037+00			2025-08-13 00:23:57.530503+00	2025-08-13 00:26:34.320061+00	\N	1755044598688
1599a18c-6ac2-4801-b635-a1b215241a3a	022137	Fabrication	Fabricate components for the project	In Progress	2025-08-13 00:26:34.320869+00			2025-08-13 00:23:57.542091+00	2025-08-13 00:26:34.320891+00	\N	1755044605893
08845447-c840-4a77-88d8-083bd1aeae14	323585	Installation	Install the project components on-site	Not Started	\N			2025-08-13 00:40:48.117743+00	2025-08-13 00:40:48.123291+00	\N	1755045602331
d234d50d-c78d-4562-86bd-0f24bbf71e6a	323585	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-13 00:40:48.117744+00	2025-08-13 00:40:48.124259+00	\N	1755045602331
6485d5f8-f5f4-4d28-8fd7-b39092f7161e	323585	CAD	Create CAD designs for the project	Completed	2025-08-13 00:47:32.475208+00			2025-08-13 00:40:48.117739+00	2025-08-13 00:47:32.475239+00	\N	1755045602331
7756ed6b-a1d7-41d7-bc12-7d8e1bce3446	323585	Cutting	Cut materials according to specifications	Completed	2025-08-13 00:50:33.618806+00			2025-08-13 00:40:48.117741+00	2025-08-13 00:50:33.61883+00	\N	1755045602331
4f13c0d2-aae2-48e1-9d9a-85084cc3d998	323585	Fabrication	Fabricate components for the project	In Progress	2025-08-13 00:50:33.620963+00			2025-08-13 00:40:48.117742+00	2025-08-13 00:50:33.620989+00	\N	1755045602331
83873ff9-6584-412e-b5e0-9e43143e0d61	148872	Template	Prepare and finalize the template for the project	Completed	2025-08-14 11:38:40.195537+00			2025-08-13 01:02:51.597481+00	2025-08-14 11:38:40.195573+00	\N	1755046904924
a6d85a15-4d8a-4d08-9daf-0b323999fe5c	688174	CAD	Create CAD designs for the project	Completed	2025-08-14 11:28:24.011787+00			2025-08-13 01:09:54.242865+00	2025-08-14 11:28:24.01183+00	\N	1755047241451
2ce6dd08-545f-47a3-bfb7-31e9d03ae6a7	688174	Installation	Install the project components on-site	Completed	2025-08-14 11:34:48.665752+00			2025-08-13 01:09:54.242869+00	2025-08-14 11:34:48.66582+00	\N	1755047241451
83ee3e04-cc2a-4e0a-99e5-222ddb2fb3cf	688174	Service Call	Handle any service calls or maintenance	In Progress	2025-08-14 11:34:48.740206+00			2025-08-13 01:09:54.24287+00	2025-08-14 11:34:48.740259+00	\N	1755047241451
10d85f7f-b165-4293-98c9-8e6a8202b965	148872	Template	Prepare and finalize the template for the project	Completed	2025-08-14 12:08:40.264365+00			2025-08-13 01:02:51.607826+00	2025-08-14 12:08:40.2644+00	\N	1755046911899
fc0cb528-7eff-4029-bcc5-8e4e57482f23	148872	CAD	Create CAD designs for the project	Completed	2025-08-14 11:38:43.205363+00			2025-08-13 01:02:51.597483+00	2025-08-14 11:38:43.205394+00	\N	1755046904924
5e2e8ebd-cc4e-4fac-a3ca-66d26f723f5d	148872	Cutting	Cut materials according to specifications	Completed	2025-08-14 12:08:48.206203+00			2025-08-13 01:02:51.607829+00	2025-08-14 12:08:48.206243+00	\N	1755046911899
5de64379-863a-4b25-a774-7c64bed48312	148872	CAD	Create CAD designs for the project	Completed	2025-08-14 12:08:46.256181+00			2025-08-13 01:02:51.607827+00	2025-08-14 12:08:46.256224+00	\N	1755046911899
41ec0891-212e-4d75-837c-20924fdf15e1	148872	Fabrication	Fabricate components for the project	Completed	2025-08-14 12:08:50.104712+00			2025-08-13 01:02:51.607831+00	2025-08-14 12:08:50.104751+00	\N	1755046911899
4618b9bf-e21e-41d7-9001-0ffd85cd643b	148872	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-13 01:02:51.597485+00	2025-08-14 12:08:28.450583+00	\N	1755046904924
819a7abe-8596-43d2-abd1-7d4d907efc07	148872	Installation	Install the project components on-site	Not Started	\N			2025-08-13 01:02:51.597487+00	2025-08-14 12:08:28.492004+00	\N	1755046904924
7e57cf57-da6e-4b1b-a299-9c27f8408861	148872	Installation	Install the project components on-site	Completed	2025-08-14 12:08:51.976919+00			2025-08-13 01:02:51.607833+00	2025-08-14 12:08:51.97696+00	\N	1755046911899
db75c1cb-57e7-4e06-98a6-01514607412c	148872	Service Call	Handle any service calls or maintenance	Completed	2025-08-14 12:08:54.360837+00			2025-08-13 01:02:51.607834+00	2025-08-14 12:08:54.360874+00	\N	1755046911899
10d4f2f1-3af0-4952-b5e8-284b30df4270	148872	Cutting	Cut materials according to specifications	In Progress	2025-08-14 12:09:24.610965+00			2025-08-13 01:02:51.597484+00	2025-08-14 12:09:24.610997+00	\N	1755046904924
08f686a1-abeb-4687-82e3-40ea03127c83	688174	Installation	Install the project components on-site	Not Started	\N			2025-08-13 01:09:54.255156+00	2025-08-13 01:09:54.259623+00	\N	1755047247076
a0b14c20-e032-4fbf-a5fa-b30dd60add77	688174	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-13 01:09:54.255157+00	2025-08-13 01:09:54.260705+00	\N	1755047247076
14f326dc-c309-47d4-8e06-9c748936f99f	688174	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-13 01:09:54.255155+00	2025-08-14 11:34:23.779775+00	\N	1755047247076
afbfae76-6d15-4a17-8c89-25453499bbf4	870087	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 12:17:53.431008+00			2025-08-13 03:41:57.73521+00	2025-08-14 12:17:53.431053+00	\N	1755056329204
febb05aa-4644-4dac-bfda-6ddb0235e1f8	341276	Template	Prepare and finalize the template for the project	In Progress	2025-08-13 01:18:35.232026+00			2025-08-13 01:18:35.232026+00	2025-08-13 01:18:35.232653+00	\N	1755047850236
4fc62501-e29a-4071-a10f-f85d7cd08cd8	341276	CAD	Create CAD designs for the project	Not Started	\N			2025-08-13 01:18:35.232028+00	2025-08-13 01:18:35.236267+00	\N	1755047850236
8c099b3f-aada-423e-b268-fb24b7bebbc6	341276	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-13 01:18:35.232029+00	2025-08-13 01:18:35.239289+00	\N	1755047850236
26ccc13e-950e-4cdb-a008-8032584e0bf9	341276	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-13 01:18:35.23203+00	2025-08-13 01:18:35.241761+00	\N	1755047850236
1f3dba63-a0c4-47c0-a28c-b756d06c391b	341276	Installation	Install the project components on-site	Not Started	\N			2025-08-13 01:18:35.232031+00	2025-08-13 01:18:35.244548+00	\N	1755047850236
c54f0888-1733-4400-a467-e24f20b42e0c	341276	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-13 01:18:35.232032+00	2025-08-13 01:18:35.24701+00	\N	1755047850236
096a1395-707a-4dbb-a540-c585c1490b8b	870087	Template	Prepare and finalize the template for the project	Completed	2025-08-14 12:19:33.829036+00			2025-08-13 09:56:59.851681+00	2025-08-14 12:19:33.829074+00	\N	1755078960028
b05b7378-8e41-4551-8413-48837ae309ab	870087	CAD	Create CAD designs for the project	Not Started	\N			2025-08-13 03:41:57.735212+00	2025-08-14 12:17:53.517307+00	\N	1755056329204
f7538e4c-e69b-4a61-aa04-2b79c8cf20dc	870087	CAD	Create CAD designs for the project	Completed	2025-08-14 12:19:35.609535+00			2025-08-13 09:56:59.851682+00	2025-08-14 12:19:35.609577+00	\N	1755078960028
075b798c-0528-49a2-869b-2ca4d8fbd653	535395	CAD	Create CAD designs for the project	In Progress	2025-08-13 19:05:00.967569+00			2025-08-13 18:49:06.091039+00	2025-08-13 19:05:00.967697+00	\N	1755110639937
c3643d53-8b57-4d3d-b5fc-49e50d681cb9	870087	Cutting	Cut materials according to specifications	Completed	2025-08-14 12:19:37.848007+00			2025-08-13 09:56:59.851683+00	2025-08-14 12:19:37.848059+00	\N	1755078960028
da8e664f-5cff-4f4d-bac8-6c4a60bcdda9	688174	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 15:20:11.115015+00			2025-08-13 01:09:54.255151+00	2025-08-14 15:20:11.11506+00	\N	1755047247076
9a0cc04d-fe75-4f82-8470-1f6a9d70fbc2	870087	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-13 03:41:57.735213+00	2025-08-14 12:17:53.557683+00	\N	1755056329204
6bef024b-004f-4b79-bee1-fc62dc904d83	535395	Template	Prepare and finalize the template for the project	In Progress	2025-08-13 18:49:06.091038+00			2025-08-13 18:49:06.091037+00	2025-08-13 18:49:06.091099+00	\N	1755110639937
a884a9cd-91c3-4243-a8b4-85da105aa2a2	535395	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-13 18:49:06.091039+00	2025-08-13 18:49:06.247677+00	\N	1755110639937
86180999-59b9-4dda-86ae-7de245191b21	535395	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-13 18:49:06.09104+00	2025-08-13 18:49:06.287613+00	\N	1755110639937
6c32c563-845f-4290-a36f-f3a5a6d7a2a0	535395	Installation	Install the project components on-site	Not Started	\N			2025-08-13 18:49:06.091041+00	2025-08-13 18:49:06.324384+00	\N	1755110639937
3f3efb48-4909-49c1-b5e0-f78dfeadd7d2	535395	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-13 18:49:06.091043+00	2025-08-13 18:49:06.366712+00	\N	1755110639937
6beb9e8c-8fd4-44c7-8932-6349c437fc11	535395	Template	Prepare and finalize the template for the project	In Progress	2025-08-13 18:49:06.527965+00			2025-08-13 18:49:06.527964+00	2025-08-13 18:49:06.528065+00	\N	1755110649635
551b2e13-1c22-4e9d-8555-ec370a8acb9b	870087	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 12:18:24.303867+00			2025-08-13 09:53:57.528132+00	2025-08-14 12:18:24.303914+00	\N	1755078794261
0bf68364-661c-412a-b068-7ab4e24a6ef3	688174	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-13 01:09:54.255154+00	2025-08-14 11:28:15.977065+00	\N	1755047247076
54863d82-3e7a-434c-9556-02413e6f2146	688174	Template	Prepare and finalize the template for the project	Completed	2025-08-14 11:28:21.716124+00			2025-08-13 01:09:54.242864+00	2025-08-14 11:28:21.716169+00	\N	1755047241451
0b03a013-f41a-491e-b74d-6ac88606f1e7	870087	CAD	Create CAD designs for the project	Not Started	\N			2025-08-13 09:53:57.528133+00	2025-08-14 12:18:24.339319+00	\N	1755078794261
dbbfc742-f4fd-4f6c-8dfe-6a87aba14bee	870087	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-13 09:53:57.528133+00	2025-08-14 12:18:24.374632+00	\N	1755078794261
b45d876b-3023-4236-8fbb-c4e57fd88e39	688174	CAD	Create CAD designs for the project	Not Started	\N			2025-08-13 01:09:54.255153+00	2025-08-14 11:34:23.74388+00	\N	1755047247076
fbf71480-60d0-4596-8f1f-ae028bd18d90	535395	CAD	Create CAD designs for the project	Not Started	\N			2025-08-13 18:49:06.527967+00	2025-08-13 18:49:06.567315+00	\N	1755110649635
cebddf35-7f2c-4bfb-8b3f-b823aaa9df0f	535395	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-13 18:49:06.527969+00	2025-08-13 18:49:06.60642+00	\N	1755110649635
2c4cc25e-bb49-4536-88bf-0f42803cfad3	535395	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-13 18:49:06.52797+00	2025-08-13 18:49:06.643646+00	\N	1755110649635
78b5249d-4c0c-4430-808f-cf0b98f96a31	535395	Installation	Install the project components on-site	Not Started	\N			2025-08-13 18:49:06.527971+00	2025-08-13 18:49:06.682626+00	\N	1755110649635
616c9c56-202d-41fa-91c4-8566513facb4	535395	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-13 18:49:06.527976+00	2025-08-13 18:49:06.726578+00	\N	1755110649635
f1478462-ffa4-4efe-97c6-ca997d9e9bac	535395	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-13 18:58:49.58745+00	2025-08-13 18:58:49.726642+00	\N	1755111461136
b4ec9bbc-01c4-4bdc-a812-d6ee782b0f1d	535395	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-13 18:58:49.587451+00	2025-08-13 18:58:49.76328+00	\N	1755111461136
579e3021-b6a9-45e8-8997-3cc5be067b30	535395	Installation	Install the project components on-site	Not Started	\N			2025-08-13 18:58:49.587451+00	2025-08-13 18:58:49.799045+00	\N	1755111461136
8fb97b48-767e-4d75-bcb0-2119b072031e	535395	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-13 18:58:49.587452+00	2025-08-13 18:58:49.827631+00	\N	1755111461136
911dc9fa-fab6-4c57-bc69-8e27116f2e3f	535395	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-13 19:02:50.653865+00	2025-08-13 19:02:50.798383+00	\N	1755111628554
19c84da0-9442-4e64-9699-56917f4a23ce	535395	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-13 19:02:50.653866+00	2025-08-13 19:02:50.836754+00	\N	1755111628554
624d8bc2-ccc4-47bc-a4cd-c122408fc32e	535395	Installation	Install the project components on-site	Not Started	\N			2025-08-13 19:02:50.653867+00	2025-08-13 19:02:50.878035+00	\N	1755111628554
1464b987-07a3-476b-a45c-09a72d4830df	535395	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-13 19:02:50.653868+00	2025-08-13 19:02:50.917642+00	\N	1755111628554
3824e1c9-adbf-42b1-8137-15cad4ee5e4c	535395	Template	Prepare and finalize the template for the project	Completed	2025-08-13 19:03:24.893174+00			2025-08-13 18:58:49.587449+00	2025-08-13 19:03:24.893314+00	\N	1755111461136
0a0a1d43-0e5c-4d45-b82c-579fcfba9d46	535395	CAD	Create CAD designs for the project	Completed	2025-08-13 19:03:34.51459+00			2025-08-13 18:58:49.58745+00	2025-08-13 19:03:34.514631+00	\N	1755111461136
19db1509-6f0d-48b3-b5d2-edf1b15c6d7b	535395	CAD	Create CAD designs for the project	In Progress	2025-08-13 19:03:34.549617+00			2025-08-13 19:02:50.653862+00	2025-08-13 19:03:34.549706+00	\N	1755111628554
83630a3a-283e-48b8-915a-afc4ba873890	535395	Template	Prepare and finalize the template for the project	Completed	2025-08-13 19:05:00.898724+00			2025-08-13 19:02:50.653861+00	2025-08-13 19:05:00.898801+00	\N	1755111628554
c8fbb572-16fa-4f86-a0d0-e1b505e50b0f	870087	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-13 09:53:57.528134+00	2025-08-14 12:18:24.409068+00	\N	1755078794261
6bcc2839-827d-46e8-957a-b30772837e5d	870087	Installation	Install the project components on-site	Completed	2025-08-14 12:19:41.248902+00			2025-08-13 09:56:59.851684+00	2025-08-14 12:19:41.248954+00	\N	1755078960028
bfbc15f6-e01f-492e-9655-0d394c5c0b1d	870087	Fabrication	Fabricate components for the project	Completed	2025-08-14 12:19:39.480162+00			2025-08-13 09:56:59.851684+00	2025-08-14 12:19:39.480201+00	\N	1755078960028
feec3431-0030-499c-a04c-28c924a0a54e	747298	Template	Prepare and finalize the template for the project	Completed	2025-08-14 11:17:15.836615+00			2025-08-14 06:00:56.498365+00	2025-08-14 11:17:15.836652+00	\N	1755150932103
bdec0d5b-ec41-440e-a0f1-84b4aaa71600	870087	Service Call	Handle any service calls or maintenance	In Progress	2025-08-14 12:19:41.282518+00			2025-08-13 09:56:59.851685+00	2025-08-14 12:19:41.282573+00	\N	1755078960028
75895b51-5ce2-4a68-8a4d-8f8b159c3d4b	870087	Installation	Install the project components on-site	Not Started	\N			2025-08-13 09:53:57.528135+00	2025-08-14 12:18:24.444651+00	\N	1755078794261
0bfbc2d9-1a6a-4e50-90e9-b2bd2fd53e10	870087	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-13 09:53:57.528135+00	2025-08-14 12:18:24.483967+00	\N	1755078794261
80649aed-002d-4bec-b58c-dbf7122b3e7d	249777	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 02:22:29.123178+00	2025-08-14 02:22:29.126582+00	\N	1755138100235
75ff9fe9-cec9-4294-b6e7-2cec68d798f4	249777	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 02:22:29.123179+00	2025-08-14 02:22:29.127393+00	\N	1755138100235
a2b3faa1-2972-4ae2-bb39-55b37e04d053	249777	Installation	Install the project components on-site	Not Started	\N			2025-08-14 02:22:29.12318+00	2025-08-14 02:22:29.128215+00	\N	1755138100235
e17e1e18-9aa9-429d-a0d8-74821921b311	249777	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 02:22:29.123181+00	2025-08-14 02:22:29.128913+00	\N	1755138100235
1d22a72f-7ecf-42c0-8001-7cff57f356bb	249777	Template	Prepare and finalize the template for the project	Completed	2025-08-14 02:24:14.098629+00			2025-08-14 02:22:29.123175+00	2025-08-14 02:24:14.098657+00	\N	1755138100235
8f3d4da5-6637-448a-b02c-2822ecd481b5	249777	CAD	Create CAD designs for the project	In Progress	2025-08-14 02:24:14.100267+00			2025-08-14 02:22:29.123177+00	2025-08-14 02:24:14.100292+00	\N	1755138100235
c8f5a4e1-8aa1-4727-8e74-60f1127dbea0	616587	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 05:41:35.824491+00			2025-08-14 05:41:35.824491+00	2025-08-14 05:41:35.824536+00	\N	1755149396910
4894c93f-e504-4f03-8071-06fc301232aa	616587	CAD	Create CAD designs for the project	Not Started	\N			2025-08-14 05:41:35.824492+00	2025-08-14 05:41:35.920015+00	\N	1755149396910
32bce690-7c4f-4957-9467-69ffe804b75d	616587	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 05:41:35.824492+00	2025-08-14 05:41:36.02601+00	\N	1755149396910
9c5a63d5-a524-485f-91c2-88185968f523	616587	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 05:41:35.824493+00	2025-08-14 05:41:36.064097+00	\N	1755149396910
b4aada1f-8844-41a9-adef-229d15ceeccc	616587	Installation	Install the project components on-site	Not Started	\N			2025-08-14 05:41:35.824493+00	2025-08-14 05:41:36.106414+00	\N	1755149396910
7568cf56-7754-45ed-ad96-a1f4bf462ea9	616587	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 05:41:35.824494+00	2025-08-14 05:41:36.144776+00	\N	1755149396910
51700928-7769-464b-ac0b-ca0ff9694c42	747298	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 06:00:56.498369+00	2025-08-14 06:00:56.701674+00	\N	1755150932103
3cef16ff-b603-4707-80be-d3df4e8fb406	870087	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-13 03:41:57.735214+00	2025-08-14 12:17:53.597665+00	\N	1755056329204
e8c23660-5725-4678-b840-190aa474442e	747298	CAD	Create CAD designs for the project	Completed	2025-08-14 11:17:50.175018+00			2025-08-14 06:00:56.498367+00	2025-08-14 11:17:50.175058+00	\N	1755150932103
3b1402ec-a0f2-43d6-bda3-5cc4f6a7386e	747298	Cutting	Cut materials according to specifications	In Progress	2025-08-14 11:17:50.25004+00			2025-08-14 06:00:56.498368+00	2025-08-14 11:17:50.250085+00	\N	1755150932103
4198c2e9-0bf7-450d-9fe6-529cd35744f9	870087	Installation	Install the project components on-site	Not Started	\N			2025-08-13 03:41:57.735215+00	2025-08-14 12:17:53.711032+00	\N	1755056329204
bab50200-42c8-4710-bdf0-bc04574538a4	870087	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-13 03:41:57.735216+00	2025-08-14 12:17:53.82222+00	\N	1755056329204
9d33d6e7-58b3-4e9b-9987-9dceba86e613	747298	Installation	Install the project components on-site	Not Started	\N			2025-08-14 06:00:56.498369+00	2025-08-14 06:00:56.737422+00	\N	1755150932103
d79d5f24-a523-460b-9f21-762fa4c435d2	747298	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 06:00:56.49837+00	2025-08-14 06:00:56.777738+00	\N	1755150932103
18e50665-6581-480c-af78-c5266af2d12e	516289	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 08:11:30.682002+00			2025-08-14 08:11:30.682001+00	2025-08-14 08:11:30.682049+00	\N	1755159079427
5e981e34-3318-4a62-abf7-c02a46de4998	516289	CAD	Create CAD designs for the project	Not Started	\N			2025-08-14 08:11:30.682002+00	2025-08-14 08:11:30.752868+00	\N	1755159079427
a0a82deb-da5c-457c-828c-db990bbf459c	516289	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 08:11:30.682003+00	2025-08-14 08:11:30.832611+00	\N	1755159079427
753cc379-00e9-4724-bd55-b2ab3c4cab0b	516289	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 08:11:30.682003+00	2025-08-14 08:11:30.873355+00	\N	1755159079427
89c866a8-52b6-4b23-aab0-2b1c24777634	516289	Installation	Install the project components on-site	Not Started	\N			2025-08-14 08:11:30.682003+00	2025-08-14 08:11:30.906272+00	\N	1755159079427
66db2f46-6b4b-44f6-9323-8bfdda97b26a	516289	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 08:11:30.682003+00	2025-08-14 08:11:30.941443+00	\N	1755159079427
ccd39dfb-8ebb-4bf8-9a6d-946a6a650f1c	688174	Cutting	Cut materials according to specifications	Completed	2025-08-14 11:28:26.603269+00			2025-08-13 01:09:54.242867+00	2025-08-14 11:28:26.603314+00	\N	1755047241451
6fa40aed-a891-4bf0-91af-e7605546c993	688174	Fabrication	Fabricate components for the project	Completed	2025-08-14 11:34:33.354736+00			2025-08-13 01:09:54.242868+00	2025-08-14 11:34:33.354777+00	\N	1755047241451
c9cfdeba-06b6-437f-81a4-786df4088fdb	148872	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-13 01:02:51.597488+00	2025-08-14 12:08:28.526324+00	\N	1755046904924
d6d747e2-6576-48b8-b1b3-d90b02c5fdfe	870087	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 12:26:34.301625+00			2025-08-14 12:26:34.301625+00	2025-08-14 12:26:34.301678+00	\N	1755174329782
a0126a96-e26e-4d0b-81a8-66527249d073	870087	CAD	Create CAD designs for the project	Not Started	\N			2025-08-14 12:26:34.301627+00	2025-08-14 12:26:34.371402+00	\N	1755174329782
977c1e61-6975-4e8a-9464-c7c370130c9e	870087	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 12:26:34.301627+00	2025-08-14 12:26:34.440531+00	\N	1755174329782
0f71feda-86c5-4290-ba5c-8d6b039aec93	870087	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 12:26:34.301627+00	2025-08-14 12:26:34.471128+00	\N	1755174329782
b9ef1348-4c46-4343-a0c5-6e78cb4e7106	870087	Installation	Install the project components on-site	Not Started	\N			2025-08-14 12:26:34.301627+00	2025-08-14 12:26:34.510686+00	\N	1755174329782
6bc138c0-62e1-4699-941b-b01560004e1b	870087	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 12:26:34.301627+00	2025-08-14 12:26:34.541791+00	\N	1755174329782
8787cd75-ba7e-4faf-96cb-fdacaaf9f6c3	244629	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 13:20:31.848683+00			2025-08-14 13:20:31.848683+00	2025-08-14 13:20:31.877593+00	\N	1755177554853
3128ab76-9e0e-4118-baa5-0534a78784c7	244629	CAD	Create CAD designs for the project	Not Started	\N			2025-08-14 13:20:31.848684+00	2025-08-14 13:20:32.027673+00	\N	1755177554853
c2c23668-c738-4f72-a632-4b3af5bd231a	244629	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 13:20:31.848684+00	2025-08-14 13:20:32.17277+00	\N	1755177554853
15bea71c-41b0-4394-8ec9-57d2ac887326	244629	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 13:20:31.848684+00	2025-08-14 13:20:32.277878+00	\N	1755177554853
556d9565-eb82-4b56-8c33-3e16de72d297	244629	Installation	Install the project components on-site	Not Started	\N			2025-08-14 13:20:31.848685+00	2025-08-14 13:20:32.393994+00	\N	1755177554853
503abdb3-e24a-4d5f-8613-b5ace8641d55	244629	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 13:20:31.848685+00	2025-08-14 13:20:32.492311+00	\N	1755177554853
d191fed3-0e35-41d9-9b7a-f88d8650b156	244629	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 13:20:32.749193+00			2025-08-14 13:20:32.749193+00	2025-08-14 13:20:32.777566+00	\N	1755177558489
f539ab43-c3c3-46b0-b1ec-e7508496a906	244629	CAD	Create CAD designs for the project	Not Started	\N			2025-08-14 13:20:32.749194+00	2025-08-14 13:20:32.898294+00	\N	1755177558489
76ca4376-f3de-487a-b4ed-8635569f0c38	244629	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 13:20:32.749194+00	2025-08-14 13:20:33.007681+00	\N	1755177558489
bd39af34-20d9-4089-86cf-fe7f134a1044	244629	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 13:20:32.749194+00	2025-08-14 13:20:33.113363+00	\N	1755177558489
f0da5339-6a6b-4fb0-ac84-249001a56d6b	244629	Installation	Install the project components on-site	Not Started	\N			2025-08-14 13:20:32.749195+00	2025-08-14 13:20:33.217118+00	\N	1755177558489
89d5a548-3dd0-4a79-8029-951b9e5e06c3	244629	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 13:20:32.749195+00	2025-08-14 13:20:33.317615+00	\N	1755177558489
dca675ae-30db-4f2a-9b88-9b65759fdb57	110545	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 14:34:32.784164+00			2025-08-14 14:34:32.784164+00	2025-08-14 14:34:32.784182+00	\N	1755182017435
363e935a-f397-4761-83b4-b0d7346c27a4	110545	CAD	Create CAD designs for the project	Not Started	\N			2025-08-14 14:34:32.784164+00	2025-08-14 14:34:32.863723+00	\N	1755182017435
74972c52-393d-484a-a86b-64b04a8a75b5	110545	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 14:34:32.784165+00	2025-08-14 14:34:32.988713+00	\N	1755182017435
cdfebf50-0e94-4679-8928-88acb9370bdc	110545	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 14:34:32.784165+00	2025-08-14 14:34:33.025722+00	\N	1755182017435
7942aece-73e7-41cc-baed-f10c9ff569e8	110545	Installation	Install the project components on-site	Not Started	\N			2025-08-14 14:34:32.784165+00	2025-08-14 14:34:33.084169+00	\N	1755182017435
3a2f25a8-b8ee-4f05-821a-2680ab615737	110545	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 14:34:32.784165+00	2025-08-14 14:34:33.118405+00	\N	1755182017435
4af648f8-6fa1-4ed3-8d7f-8909d225a1a2	110545	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 14:34:33.258707+00			2025-08-14 14:34:33.258707+00	2025-08-14 14:34:33.258779+00	\N	1755182036311
2a8ac1b0-b226-4c89-87ae-43ac74c08d54	110545	CAD	Create CAD designs for the project	Not Started	\N			2025-08-14 14:34:33.258708+00	2025-08-14 14:34:33.299738+00	\N	1755182036311
5b2e6f00-ac77-4ac1-a174-bfdc5b3614ee	110545	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 14:34:33.258708+00	2025-08-14 14:34:33.334902+00	\N	1755182036311
27cfbd6c-2ed4-4432-8cea-a384ed974e9a	110545	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 14:34:33.258709+00	2025-08-14 14:34:33.424571+00	\N	1755182036311
0fcbf9ed-db33-4bb6-bb60-4b12d17b4b33	110545	Installation	Install the project components on-site	Not Started	\N			2025-08-14 14:34:33.258709+00	2025-08-14 14:34:33.458722+00	\N	1755182036311
07730a73-a232-4a0d-b395-c2535c310753	110545	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 14:34:33.258709+00	2025-08-14 14:34:33.497542+00	\N	1755182036311
537adf38-dbc2-4d34-9a1f-a884a79ab35b	043649	Installation	Install the project components on-site	Not Started	\N			2025-08-14 15:43:01.104917+00	2025-08-14 15:43:01.109369+00	\N	1755185902660
5d41108b-445f-4624-9e3f-b01b9944fdbc	043649	Cutting	Cut materials according to specifications	Completed	2025-08-14 15:46:47.283189+00			2025-08-14 15:43:01.104914+00	2025-08-14 15:46:47.283219+00	\N	1755185902660
27fab65f-23d6-4349-a82f-7d23fed8562b	043649	Fabrication	Fabricate components for the project	In Progress	2025-08-14 15:46:47.283671+00			2025-08-14 15:43:01.104916+00	2025-08-14 15:46:47.283699+00	\N	1755185902660
08040c43-d740-4292-9953-493e60ad7336	043649	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 15:43:01.104919+00	2025-08-14 15:43:01.109848+00	\N	1755185902660
ac580f3b-e896-4249-848e-04be35c6d1f3	043649	Installation	Install the project components on-site	Not Started	\N			2025-08-14 15:45:57.549297+00	2025-08-14 15:45:57.551967+00	\N	1755186352303
62d2b9c3-91f7-4325-b94b-2f8714c14b6b	043649	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 15:45:57.549298+00	2025-08-14 15:45:57.55235+00	\N	1755186352303
fe6eec82-fa55-421b-bde3-936fdd6d6a4f	043649	Template	Prepare and finalize the template for the project	Completed	2025-08-14 15:46:44.434483+00			2025-08-14 15:43:01.104911+00	2025-08-14 15:46:44.43452+00	\N	1755185902660
b0e983a7-2767-4aa0-ac6a-ae9acda05f71	043649	CAD	Create CAD designs for the project	Completed	2025-08-14 15:46:45.972895+00			2025-08-14 15:43:01.104913+00	2025-08-14 15:46:45.972921+00	\N	1755185902660
a108b358-aa5c-4369-af80-733a8e845974	043649	Template	Prepare and finalize the template for the project	Completed	2025-08-14 15:46:58.793233+00			2025-08-14 15:45:57.549291+00	2025-08-14 15:46:58.79326+00	\N	1755186352303
8c77529b-ee95-4f55-82f0-b122022063af	496909	Cutting	Cut materials according to specifications	Completed	2025-08-14 16:11:34.295754+00			2025-08-14 16:03:55.945715+00	2025-08-14 16:11:34.295832+00	\N	1755187347060
fa6af64d-a307-4412-a82f-000da7be9911	043649	CAD	Create CAD designs for the project	Completed	2025-08-14 15:47:00.100249+00			2025-08-14 15:45:57.549293+00	2025-08-14 15:47:00.100287+00	\N	1755186352303
2263cbcd-ba1b-4e59-9281-7a53da94b30d	043649	Cutting	Cut materials according to specifications	Completed	2025-08-14 15:47:01.294748+00			2025-08-14 15:45:57.549294+00	2025-08-14 15:47:01.294798+00	\N	1755186352303
d915bd58-a2ae-4644-9bf9-8954412039de	043649	Fabrication	Fabricate components for the project	In Progress	2025-08-14 15:47:01.295586+00			2025-08-14 15:45:57.549296+00	2025-08-14 15:47:01.295609+00	\N	1755186352303
a5f9a73a-e2d4-41c6-b451-7dce07043302	043649	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 15:49:43.889659+00			2025-08-14 15:49:43.889658+00	2025-08-14 15:49:43.889684+00	\N	1755186567302
955ee079-3a5d-40ba-b656-244d93a799d1	043649	CAD	Create CAD designs for the project	Not Started	\N			2025-08-14 15:49:43.88966+00	2025-08-14 15:49:43.890876+00	\N	1755186567302
443c814e-fd7f-44fb-9378-41fd121b7029	043649	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 15:49:43.889661+00	2025-08-14 15:49:43.891536+00	\N	1755186567302
898173ad-e5f6-429e-b2a9-37611282746d	043649	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 15:49:43.889662+00	2025-08-14 15:49:43.892126+00	\N	1755186567302
a88b9773-c4c0-4a96-8c26-7f48ba7ce73b	043649	Installation	Install the project components on-site	Not Started	\N			2025-08-14 15:49:43.889663+00	2025-08-14 15:49:43.892444+00	\N	1755186567302
58563edc-bf3e-4887-a650-a39b7d93d82b	043649	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 15:49:43.889664+00	2025-08-14 15:49:43.892723+00	\N	1755186567302
09b535b8-428c-4aec-9d37-8cac6cd6481e	806460	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 15:56:55.179772+00	2025-08-14 15:56:55.181831+00	\N	1755186887688
6df528b6-f710-4976-b0ba-44a7ca543e15	806460	Installation	Install the project components on-site	Not Started	\N			2025-08-14 15:56:55.179773+00	2025-08-14 15:56:55.182153+00	\N	1755186887688
50c6170c-d170-4f8c-adc9-8d985542083c	806460	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 15:56:55.179774+00	2025-08-14 15:56:55.18245+00	\N	1755186887688
014ab35d-6b75-4683-87b6-6edde32fa289	806460	Template	Prepare and finalize the template for the project	Completed	2025-08-14 16:00:07.335909+00			2025-08-14 15:56:55.179768+00	2025-08-14 16:00:07.335938+00	\N	1755186887688
02143ea9-a25e-49b7-9a1a-54db02beeef7	496909	Fabrication	Fabricate components for the project	Completed	2025-08-14 16:11:37.389584+00			2025-08-14 16:03:55.945716+00	2025-08-14 16:11:37.389606+00	\N	1755187347060
53b4d190-a6d2-4e97-b81e-c15cbd96ee55	806460	CAD	Create CAD designs for the project	Completed	2025-08-14 16:00:40.827084+00			2025-08-14 15:56:55.17977+00	2025-08-14 16:00:40.82711+00	\N	1755186887688
729fc05f-c506-4c9f-98ec-c0fcf5822439	806460	Cutting	Cut materials according to specifications	In Progress	2025-08-14 16:00:40.828059+00			2025-08-14 15:56:55.179771+00	2025-08-14 16:00:40.828083+00	\N	1755186887688
8a2b18ef-e044-4fc4-9977-e375fea3a925	496909	Template	Prepare and finalize the template for the project	Completed	2025-08-14 16:11:27.142881+00			2025-08-14 16:03:55.945713+00	2025-08-14 16:11:27.142913+00	\N	1755187347060
c6d98383-0f77-4679-8da5-db59864acc00	496909	CAD	Create CAD designs for the project	Completed	2025-08-14 16:11:30.85992+00			2025-08-14 16:03:55.945714+00	2025-08-14 16:11:30.859946+00	\N	1755187347060
2c829819-4d8a-49fb-b03e-a114d7b413eb	496909	Installation	Install the project components on-site	Completed	2025-08-14 16:11:40.222601+00			2025-08-14 16:03:55.945717+00	2025-08-14 16:11:40.222632+00	\N	1755187347060
ac5e83b7-a7a6-4014-8e84-12b1d1a285f2	496909	Service Call	Handle any service calls or maintenance	Completed	2025-08-14 16:11:43.01185+00			2025-08-14 16:03:55.945718+00	2025-08-14 16:11:43.01187+00	\N	1755187347060
3b5f38b8-545e-4007-bc42-0083385ad047	484015	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 16:16:34.955332+00			2025-08-14 16:16:34.955332+00	2025-08-14 16:16:34.955379+00	\N	1755188111178
f57456ac-865e-4189-b89e-05caf1ad4f9e	484015	CAD	Create CAD designs for the project	Not Started	\N			2025-08-14 16:16:34.955333+00	2025-08-14 16:16:34.956468+00	\N	1755188111178
87eacaa7-f738-41cc-8260-72bb4e712238	484015	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 16:16:34.955334+00	2025-08-14 16:16:34.95704+00	\N	1755188111178
5cc6f32c-2410-403f-9187-3b63b95c6742	484015	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 16:16:34.955344+00	2025-08-14 16:16:34.957494+00	\N	1755188111178
37890042-0cb5-4a1f-9ddb-9f0fe70201cb	484015	Installation	Install the project components on-site	Not Started	\N			2025-08-14 16:16:34.955345+00	2025-08-14 16:16:34.958797+00	\N	1755188111178
99cd963f-6342-416c-a5b7-29c8f40e59ca	484015	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 16:16:34.955346+00	2025-08-14 16:16:34.959564+00	\N	1755188111178
bc0bd6cf-98a8-44f1-af18-215511d4f8ee	484015	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 16:16:34.961005+00			2025-08-14 16:16:34.961004+00	2025-08-14 16:16:34.961024+00	\N	1755188119760
c31e76ec-9926-47d6-9781-90bfb7b0684f	484015	CAD	Create CAD designs for the project	Not Started	\N			2025-08-14 16:16:34.961006+00	2025-08-14 16:16:34.961474+00	\N	1755188119760
46648401-fd79-4f64-bc93-db759cb720ea	484015	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 16:16:34.961007+00	2025-08-14 16:16:34.961818+00	\N	1755188119760
3d15fa4b-5c46-4560-a4d3-94dc2c9af0bf	484015	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 16:16:34.961008+00	2025-08-14 16:16:34.962492+00	\N	1755188119760
7557a225-91a9-4c00-b1de-5e6074759fb5	484015	Installation	Install the project components on-site	Not Started	\N			2025-08-14 16:16:34.961009+00	2025-08-14 16:16:34.962842+00	\N	1755188119760
6324bffe-4c70-465e-9908-4079c39bef8c	484015	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 16:16:34.96101+00	2025-08-14 16:16:34.963186+00	\N	1755188119760
038b649f-d389-4ed9-b53a-220c6ef02a5c	653535	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 16:23:37.701773+00			2025-08-14 16:23:37.701773+00	2025-08-14 16:23:37.701837+00	\N	1755188516761
eaa7bee8-276b-4b1b-9784-862e665fa5bf	653535	CAD	Create CAD designs for the project	Not Started	\N			2025-08-14 16:23:37.701774+00	2025-08-14 16:23:37.702891+00	\N	1755188516761
aff4d710-cd00-42db-9c64-af38dbdfef86	653535	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 16:23:37.701775+00	2025-08-14 16:23:37.703804+00	\N	1755188516761
93b31370-06ef-4e40-9b9e-7dd9bccb7d1f	653535	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 16:23:37.701776+00	2025-08-14 16:23:37.704288+00	\N	1755188516761
d718c68f-f52b-470a-ab6e-d30a4b07d2ec	653535	Installation	Install the project components on-site	Not Started	\N			2025-08-14 16:23:37.701777+00	2025-08-14 16:23:37.704823+00	\N	1755188516761
55a834b4-be10-4321-840a-e3eb2f4edfbb	653535	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 16:23:37.701778+00	2025-08-14 16:23:37.705204+00	\N	1755188516761
2b8b045b-ad32-45e0-8c6b-9ea10246c58f	653535	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 16:23:37.706865+00			2025-08-14 16:23:37.706864+00	2025-08-14 16:23:37.706889+00	\N	1755188525388
52914ce3-979b-4906-a339-740085f72f85	653535	CAD	Create CAD designs for the project	Not Started	\N			2025-08-14 16:23:37.706866+00	2025-08-14 16:23:37.707321+00	\N	1755188525388
8f1b26d9-f7c6-4fe5-8885-fd78fd55c366	653535	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 16:23:37.706868+00	2025-08-14 16:23:37.707771+00	\N	1755188525388
c297ecb5-4878-4faf-a067-2fabac7d401b	653535	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 16:23:37.706869+00	2025-08-14 16:23:37.708092+00	\N	1755188525388
5027ec7d-786b-4ce2-b1d5-b73749c1a4b6	653535	Installation	Install the project components on-site	Not Started	\N			2025-08-14 16:23:37.706871+00	2025-08-14 16:23:37.708506+00	\N	1755188525388
8f77f0ce-b941-4fd5-9279-f455b41d39a7	653535	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 16:23:37.706872+00	2025-08-14 16:23:37.708858+00	\N	1755188525388
be6d1fd2-debe-487a-9978-cebcbba646c0	255076	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 16:32:36.02881+00			2025-08-14 16:32:36.02881+00	2025-08-14 16:32:36.028842+00	\N	1755189048797
70b57a90-ad45-4779-b86b-28f02b8cb34c	255076	CAD	Create CAD designs for the project	Not Started	\N			2025-08-14 16:32:36.028811+00	2025-08-14 16:32:36.029917+00	\N	1755189048797
f9927fb4-2da4-4c58-89b9-6118bc8c2a6e	255076	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 16:32:36.028812+00	2025-08-14 16:32:36.030798+00	\N	1755189048797
31eb2cae-b80b-4d6e-a61e-f189c2772236	255076	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 16:32:36.028814+00	2025-08-14 16:32:36.031251+00	\N	1755189048797
d944a832-b41e-4659-9c85-9dfc9c5cb3a0	255076	Installation	Install the project components on-site	Not Started	\N			2025-08-14 16:32:36.028815+00	2025-08-14 16:32:36.031758+00	\N	1755189048797
f06124dc-d91a-4c7b-b5c7-d1d5d68216ce	255076	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 16:32:36.028816+00	2025-08-14 16:32:36.032117+00	\N	1755189048797
f0523208-1bfb-451c-a48a-8d4e4d7e55eb	690103	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 17:07:19.92178+00			2025-08-14 17:07:19.921779+00	2025-08-14 17:07:19.921801+00	\N	1755191154295
b795cb61-6dad-458c-8f85-25cf76df636f	690103	CAD	Create CAD designs for the project	Not Started	\N			2025-08-14 17:07:19.921781+00	2025-08-14 17:07:19.922873+00	\N	1755191154295
80064f53-c680-4f8c-961a-4667829d0a2c	690103	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 17:07:19.921782+00	2025-08-14 17:07:19.923514+00	\N	1755191154295
1c70f7ff-8665-4764-b846-c5d09e852b32	690103	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 17:07:19.921783+00	2025-08-14 17:07:19.924036+00	\N	1755191154295
8a416b45-bb47-44b5-8436-ede507bf8fd2	690103	Installation	Install the project components on-site	Not Started	\N			2025-08-14 17:07:19.921784+00	2025-08-14 17:07:19.924484+00	\N	1755191154295
d0929f54-efb6-4df5-93f4-142d302e1ad8	690103	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 17:07:19.921785+00	2025-08-14 17:07:19.924908+00	\N	1755191154295
b46f4c80-0752-4377-ae8b-3fbe8ba14d10	220509	Template	Prepare and finalize the template for the project	In Progress	2025-08-14 18:05:34.841803+00			2025-08-14 18:05:34.841803+00	2025-08-14 18:05:34.841847+00	\N	1755194697630
50c283f5-f4dd-451e-9b36-890ca0ecde45	220509	CAD	Create CAD designs for the project	Not Started	\N			2025-08-14 18:05:34.841805+00	2025-08-14 18:05:34.842901+00	\N	1755194697630
77fe64b3-cec5-4bc9-88f2-2450aaf3366a	220509	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 18:05:34.841806+00	2025-08-14 18:05:34.843766+00	\N	1755194697630
dabd243a-e9d9-4e54-bfdc-a81e7e5cae27	220509	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 18:05:34.841807+00	2025-08-14 18:05:34.844231+00	\N	1755194697630
237f608c-7954-4f86-a877-2890cc88a981	220509	Installation	Install the project components on-site	Not Started	\N			2025-08-14 18:05:34.841808+00	2025-08-14 18:05:34.844695+00	\N	1755194697630
f794ac8c-36b2-4c68-b285-0b74df1e0ba1	220509	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 18:05:34.84181+00	2025-08-14 18:05:34.845232+00	\N	1755194697630
181e83d2-63fa-4eec-97b3-53114e03092e	076315	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 18:24:51.690872+00	2025-08-14 18:24:51.693063+00	\N	1755195759973
6ca0f562-4b27-4bdc-b8d3-c7293384e9f3	076315	Installation	Install the project components on-site	Not Started	\N			2025-08-14 18:24:51.690873+00	2025-08-14 18:24:51.693405+00	\N	1755195759973
dde61c8d-bc41-4eba-897e-7337d0fdb513	076315	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 18:24:51.690874+00	2025-08-14 18:24:51.693845+00	\N	1755195759973
965d690d-dc04-451b-a44e-2cc573a927b2	076315	Template	Prepare and finalize the template for the project	Completed	2025-08-14 18:27:36.634547+00			2025-08-14 18:24:51.690868+00	2025-08-14 18:27:36.634577+00	\N	1755195759973
22efa7b5-b30a-4890-b441-4cb50b43c3b0	598530	Template	Prepare and finalize the template for the project	Completed	2025-08-20 04:39:05.611181+00			2025-08-14 21:14:53.93918+00	2025-08-20 04:39:05.61121+00	\N	1755205625417
1b43750c-dbd1-41f1-9fba-453dd39d671a	076315	CAD	Create CAD designs for the project	Completed	2025-08-14 18:27:40.422178+00			2025-08-14 18:24:51.69087+00	2025-08-14 18:27:40.422212+00	\N	1755195759973
043b9a07-6e2a-4d26-8061-9dde1f526e9c	076315	Cutting	Cut materials according to specifications	In Progress	2025-08-14 18:27:40.422625+00			2025-08-14 18:24:51.690871+00	2025-08-14 18:27:40.422645+00	\N	1755195759973
4d629c1b-77d0-4051-8ea1-ad8e0acf541d	598530	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-14 21:14:53.939183+00	2025-08-14 21:14:53.941265+00	\N	1755205625417
46a72619-6666-4bbd-a115-f00261207622	598530	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 21:14:53.939185+00	2025-08-14 21:14:53.94186+00	\N	1755205625417
6e16d387-1e95-4c48-af7a-228e5a91284a	598530	Installation	Install the project components on-site	Not Started	\N			2025-08-14 21:14:53.939186+00	2025-08-14 21:14:53.943804+00	\N	1755205625417
6e39d92f-2b22-42a9-8acc-9488816987a7	598530	CAD	Create CAD designs for the project	In Progress	2025-08-20 04:39:05.612683+00			2025-08-14 21:14:53.939182+00	2025-08-20 04:39:05.612706+00	\N	1755205625417
7c3ca567-fdc9-47e9-a543-f6c69c36848a	598530	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 21:14:53.939188+00	2025-08-14 21:14:53.944276+00	\N	1755205625417
defa9cdb-370d-4108-ab8d-a636228b6531	605988	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-14 21:30:24.437433+00	2025-08-14 21:30:24.439558+00	\N	1755207015294
2e1013df-d092-4d2b-9365-33d68fdcc9a4	605988	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-14 21:30:24.437435+00	2025-08-14 21:30:24.440258+00	\N	1755207015294
bbe99083-32c3-425a-a0f6-a08a2a577a51	605988	Template	Prepare and finalize the template for the project	Completed	2025-08-14 21:32:09.115316+00			2025-08-14 21:30:24.437429+00	2025-08-14 21:32:09.115341+00	\N	1755207015294
51939157-ba04-461b-abcc-901c9ae7157d	598530	Template	Prepare and finalize the template for the project	Completed	2025-08-20 06:46:10.196763+00			2025-08-14 21:14:53.945701+00	2025-08-20 06:46:10.196827+00	\N	1755205655662
a95a67b7-9dd1-4bca-a74d-a9103f1bdfc5	598530	CAD	Create CAD designs for the project	Completed	2025-08-21 06:39:02.882337+00			2025-08-14 21:14:53.945703+00	2025-08-21 06:39:02.882366+00	\N	1755205655662
c86e1cfd-f286-4c97-8b80-e9e90bfccd2e	598530	Cutting	Cut materials according to specifications	Completed	2025-08-21 06:39:05.518571+00			2025-08-14 21:14:53.945704+00	2025-08-21 06:39:05.518598+00	\N	1755205655662
a5ef8703-b23a-4c0d-acd3-7fe5da4d7943	598530	Fabrication	Fabricate components for the project	Completed	2025-08-21 06:39:08.020667+00			2025-08-14 21:14:53.945705+00	2025-08-21 06:39:08.020685+00	\N	1755205655662
1c9db3ee-ea55-4126-ba11-4eceef422a4b	598530	Installation	Install the project components on-site	Completed	2025-08-21 06:39:10.164211+00			2025-08-14 21:14:53.945707+00	2025-08-21 06:39:10.164242+00	\N	1755205655662
5091b9b2-dfe6-4d5b-8934-34c252311741	598530	Service Call	Handle any service calls or maintenance	Completed	2025-08-21 06:39:12.522297+00			2025-08-14 21:14:53.945709+00	2025-08-21 06:39:12.522327+00	\N	1755205655662
d17cad34-bc63-466c-ad43-5bf7d12f5ef9	605988	Cutting	Cut materials according to specifications	Not Started	\N	asdf		2025-08-14 21:30:24.437432+00	2025-08-18 13:10:10.115799+00	\N	1755207015294
340aa901-a133-429f-800b-41a3111c705e	605988	Installation	Install the project components on-site	Not Started	\N	asdfasdf		2025-08-14 21:30:24.437434+00	2025-08-18 13:10:13.693234+00	\N	1755207015294
2e2219ab-d25b-48d7-856c-afb415ecf047	605988	CAD	Create CAD designs for the project	In Progress	2025-08-14 21:32:09.116156+00	This is the notes section.		2025-08-14 21:30:24.437431+00	2025-08-18 14:03:54.775923+00	\N	1755207015294
43ae7407-5168-493e-8043-d6d8d5b024e6	594353	Template	Prepare and finalize the template for the project	In Progress	2025-08-18 17:12:48.589955+00			2025-08-18 17:12:48.589955+00	2025-08-18 17:12:48.591089+00	\N	1755537108021
74f1a45f-fdb7-4d5c-90b4-1cecbc5b6423	594353	CAD	Create CAD designs for the project	Not Started	\N			2025-08-18 17:12:48.589956+00	2025-08-18 17:12:48.666201+00	\N	1755537108021
db77df36-8464-49ec-9d91-0e13a12fc80f	594353	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-18 17:12:48.589957+00	2025-08-18 17:12:48.735111+00	\N	1755537108021
34586129-e68d-4ef3-b080-3f25e2c7cebd	594353	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-18 17:12:48.589957+00	2025-08-18 17:12:48.773098+00	\N	1755537108021
95688fa5-e48a-4950-9ead-5188d9126714	594353	Installation	Install the project components on-site	Not Started	\N			2025-08-18 17:12:48.589958+00	2025-08-18 17:12:48.803846+00	\N	1755537108021
19c6b0f0-6746-4fdd-bd51-a73ee1ec9bdc	594353	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-18 17:12:48.589958+00	2025-08-18 17:12:48.839664+00	\N	1755537108021
bddda183-22f8-43cf-adda-c581dbefc9bc	979672	Template	Prepare and finalize the template for the project	In Progress	2025-08-18 17:28:58.806874+00			2025-08-18 17:28:58.806873+00	2025-08-18 17:28:58.806924+00	\N	1755538080609
1283cbc8-daad-4242-8898-51fe87e8f712	979672	CAD	Create CAD designs for the project	Not Started	\N			2025-08-18 17:28:58.806875+00	2025-08-18 17:28:58.886762+00	\N	1755538080609
cc899635-70de-4614-8a54-afee268f98e0	979672	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-18 17:28:58.806875+00	2025-08-18 17:28:58.948575+00	\N	1755538080609
59d36b0e-993f-4cdd-a0bc-6ea40f0ff648	979672	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-18 17:28:58.806876+00	2025-08-18 17:28:58.992214+00	\N	1755538080609
fc420f38-cfb9-4c4a-af51-92fd66c63c55	979672	Installation	Install the project components on-site	Not Started	\N			2025-08-18 17:28:58.806876+00	2025-08-18 17:28:59.033351+00	\N	1755538080609
60842967-0c4b-460f-98af-5503ead35f46	979672	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-18 17:28:58.806877+00	2025-08-18 17:28:59.087085+00	\N	1755538080609
39a36c7a-8688-4d29-8740-186b661c9fea	445169	Template	Prepare and finalize the template for the project	In Progress	2025-08-19 06:28:54.705601+00			2025-08-19 06:28:54.705601+00	2025-08-19 06:28:54.705663+00	\N	1755584868899
8c861a6f-b67a-4ad6-b818-0062dbd45a85	445169	CAD	Create CAD designs for the project	Not Started	\N			2025-08-19 06:28:54.705602+00	2025-08-19 06:28:54.802185+00	\N	1755584868899
b0881818-7948-4559-a210-f2cfbcf340ae	445169	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-19 06:28:54.705603+00	2025-08-19 06:28:54.887959+00	\N	1755584868899
77716d64-5f64-4223-8249-7d868cfa8e6b	445169	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-19 06:28:54.705604+00	2025-08-19 06:28:54.927183+00	\N	1755584868899
03ce2a99-6330-44f4-9712-9b4d1a04f75c	445169	Installation	Install the project components on-site	Not Started	\N			2025-08-19 06:28:54.705605+00	2025-08-19 06:28:54.972503+00	\N	1755584868899
40db1c9f-706b-4966-b712-2f90d08c917d	445169	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-19 06:28:54.705605+00	2025-08-19 06:28:55.019943+00	\N	1755584868899
99f8b4d4-3968-43f3-a20d-971f7516b3bb	418495	Template	Prepare and finalize the template for the project	In Progress	2025-08-19 06:46:11.784967+00			2025-08-19 06:46:11.784967+00	2025-08-19 06:46:11.785015+00	\N	1755585917372
0f3ab21f-f263-4dfa-a54f-840fbf3dcefa	418495	CAD	Create CAD designs for the project	Not Started	\N			2025-08-19 06:46:11.784968+00	2025-08-19 06:46:11.86869+00	\N	1755585917372
663f35c2-5dc1-4c42-87fd-032a4bd887f8	418495	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-19 06:46:11.784969+00	2025-08-19 06:46:11.943787+00	\N	1755585917372
200b379c-af22-4336-9269-e9f1313eb7bf	418495	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-19 06:46:11.784969+00	2025-08-19 06:46:11.990241+00	\N	1755585917372
f44587de-b3e5-48f9-a636-19ba7c0412a4	418495	Installation	Install the project components on-site	Not Started	\N			2025-08-19 06:46:11.78497+00	2025-08-19 06:46:12.030547+00	\N	1755585917372
84153ec6-611e-4cd2-86a3-902edb820993	418495	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-19 06:46:11.78497+00	2025-08-19 06:46:12.063827+00	\N	1755585917372
d6debe48-183e-4e46-8685-4b21a74eddca	956125	Template	Prepare and finalize the template for the project	In Progress	2025-08-19 06:49:51.110436+00			2025-08-19 06:49:51.110435+00	2025-08-19 06:49:51.110498+00	\N	1755586104268
cc8a1dcc-bdda-4393-bf03-6b67141e3871	956125	CAD	Create CAD designs for the project	Not Started	\N			2025-08-19 06:49:51.110437+00	2025-08-19 06:49:51.203916+00	\N	1755586104268
9dd9d00d-591a-4e2d-a1bd-95bc689ab5f0	956125	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-19 06:49:51.110438+00	2025-08-19 06:49:51.295238+00	\N	1755586104268
5c136a48-5d33-4290-ad5c-b768e184324c	956125	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-19 06:49:51.110439+00	2025-08-19 06:49:51.343603+00	\N	1755586104268
b5f93bcf-5e63-4d8b-96bf-55c5aa5d0583	956125	Installation	Install the project components on-site	Not Started	\N			2025-08-19 06:49:51.110439+00	2025-08-19 06:49:51.390514+00	\N	1755586104268
8103e070-5897-40d5-849b-f97d333bc15b	956125	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-19 06:49:51.11044+00	2025-08-19 06:49:51.428277+00	\N	1755586104268
fbba306e-1437-4731-b14e-3c17fb39823e	818324	Template	Prepare and finalize the template for the project	In Progress	2025-08-19 06:59:12.695754+00			2025-08-19 06:59:12.695753+00	2025-08-19 06:59:12.695807+00	\N	1755586676753
79bb8866-4e7f-4ea2-9e28-973aafc8cbd1	818324	CAD	Create CAD designs for the project	Not Started	\N			2025-08-19 06:59:12.695755+00	2025-08-19 06:59:12.778602+00	\N	1755586676753
20a71402-844d-415e-a979-0da3e5c9fa0d	818324	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-19 06:59:12.695756+00	2025-08-19 06:59:12.855402+00	\N	1755586676753
599834b1-599c-45df-b757-8daad112abe1	818324	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-19 06:59:12.695756+00	2025-08-19 06:59:12.918728+00	\N	1755586676753
f4993eee-d5b6-4583-97c7-f03f2aeefd39	818324	Installation	Install the project components on-site	Not Started	\N			2025-08-19 06:59:12.695757+00	2025-08-19 06:59:12.961443+00	\N	1755586676753
de516f07-76a7-404e-a6e6-9e5a85b6f30d	818324	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-19 06:59:12.695758+00	2025-08-19 06:59:13.003786+00	\N	1755586676753
fddc0b12-bf5e-4014-aedd-8e1680bd89a7	673065	Template	Prepare and finalize the template for the project	Completed	2025-08-20 04:06:59.287754+00			2025-08-19 06:55:07.283827+00	2025-08-20 04:06:59.287789+00	\N	1755586444861
70b72dc4-f8a7-4c0e-8ca9-56123ac158ff	673065	CAD	Create CAD designs for the project	Completed	2025-08-20 04:10:40.816297+00			2025-08-19 06:55:07.283829+00	2025-08-20 04:10:40.816328+00	\N	1755586444861
633cb84b-8e23-4c80-889b-bee3a8ddcd8d	822302	Template	Prepare and finalize the template for the project	In Progress	2025-08-20 11:28:31.517765+00			2025-08-20 11:28:31.517765+00	2025-08-20 11:28:31.517827+00	\N	1755689294052
442b0ae5-cab3-4356-8090-2b8f040eab21	822302	CAD	Create CAD designs for the project	Not Started	\N			2025-08-20 11:28:31.517766+00	2025-08-20 11:28:31.599125+00	\N	1755689294052
67bab12f-06c9-4852-b958-1198f61d9d91	822302	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-20 11:28:31.517767+00	2025-08-20 11:28:31.676584+00	\N	1755689294052
3c910b0f-966a-4985-9a3d-2036ec87b365	822302	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-20 11:28:31.517768+00	2025-08-20 11:28:31.712205+00	\N	1755689294052
3d8a6323-5a92-464b-beee-4454caedc12c	822302	Installation	Install the project components on-site	Not Started	\N			2025-08-20 11:28:31.517769+00	2025-08-20 11:28:31.746935+00	\N	1755689294052
c6d9c888-1b72-4bf4-9acf-bff303787161	822302	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-20 11:28:31.517769+00	2025-08-20 11:28:31.785939+00	\N	1755689294052
e82b5294-154d-4dec-aed9-273fd39fa78a	673065	Cutting	Cut materials according to specifications	Completed	2025-08-22 10:10:04.984454+00			2025-08-19 06:55:07.28383+00	2025-08-22 10:10:04.984764+00	\N	1755586444861
c34e9d9d-f807-4b88-afe7-6d838c6c8402	673065	Fabrication	Fabricate components for the project	Completed	2025-08-22 10:10:07.878677+00			2025-08-19 06:55:07.283832+00	2025-08-22 10:10:07.878709+00	\N	1755586444861
3d4a0147-b570-4d8d-bb49-2e7a6840971b	673065	Installation	Install the project components on-site	Completed	2025-08-22 10:10:12.008041+00			2025-08-19 06:55:07.283835+00	2025-08-22 10:10:12.008077+00	\N	1755586444861
25a47f29-de4a-4d6b-a212-29970815442c	673065	Service Call	Handle any service calls or maintenance	Completed	2025-08-22 10:10:14.898695+00			2025-08-19 06:55:07.283837+00	2025-08-22 10:10:14.89875+00	\N	1755586444861
473704cc-3b7c-4e94-961b-b8ce7fe80340	660420	Template	Prepare and finalize the template for the project	Completed	2025-08-21 10:01:32.262787+00			2025-08-21 09:46:46.337767+00	2025-08-21 10:01:32.262817+00	\N	1755769497398
75d8b1dc-e835-4cab-b919-ce31cb2f48a4	888866	Template	Prepare and finalize the template for the project	Completed	2025-08-22 02:52:13.204806+00			2025-08-22 02:42:23.91231+00	2025-08-22 02:52:13.204836+00	\N	1755830408667
866f359f-d04d-4469-9152-5599e55c5ef5	660420	CAD	Create CAD designs for the project	Completed	2025-08-21 11:17:48.346868+00			2025-08-21 09:46:46.337769+00	2025-08-21 11:17:48.346903+00	\N	1755769497398
60298dc9-5452-4f95-8566-da0b7bf87e94	660420	Cutting	Cut materials according to specifications	Completed	2025-08-21 11:17:51.730845+00			2025-08-21 09:46:46.33777+00	2025-08-21 11:17:51.730894+00	\N	1755769497398
958f657b-bac8-4102-ac61-345d37a0af40	888866	CAD	Create CAD designs for the project	Completed	2025-08-22 02:52:15.139958+00			2025-08-22 02:42:23.912311+00	2025-08-22 02:52:15.139995+00	\N	1755830408667
3a15433f-7fa0-4a50-9296-606491068fe1	660420	Fabrication	Fabricate components for the project	Completed	2025-08-21 11:17:57.62561+00			2025-08-21 09:46:46.337771+00	2025-08-21 11:17:57.625635+00	\N	1755769497398
c498bfe2-e7ba-4703-9785-3b232c08bef0	660420	Installation	Install the project components on-site	Completed	2025-08-21 11:18:02.265337+00			2025-08-21 09:46:46.337773+00	2025-08-21 11:18:02.265376+00	\N	1755769497398
577d8bf2-79a7-41e6-96a0-f8926bf9f55c	888866	Cutting	Cut materials according to specifications	Completed	2025-08-22 02:52:17.049237+00			2025-08-22 02:42:23.912312+00	2025-08-22 02:52:17.049264+00	\N	1755830408667
87d216e4-0aa7-477c-8eb1-a658abd06700	660420	Service Call	Handle any service calls or maintenance	Completed	2025-08-21 11:19:49.176422+00			2025-08-21 09:46:46.337774+00	2025-08-21 11:19:49.176447+00	\N	1755769497398
1a819e49-1978-4443-90c1-65925d3a7677	660420	Template	Prepare and finalize the template for the project	Completed	2025-08-21 11:21:46.119797+00			2025-08-21 09:53:06.076205+00	2025-08-21 11:21:46.119821+00	\N	1755769919280
71b0102b-d4ae-4abb-9e93-10c6e85fe756	660420	CAD	Create CAD designs for the project	Completed	2025-08-21 11:21:49.99929+00			2025-08-21 09:53:06.076207+00	2025-08-21 11:21:49.99932+00	\N	1755769919280
591e2fe7-229a-4959-af12-a0f7eec03d17	888866	Fabrication	Fabricate components for the project	Completed	2025-08-22 02:52:18.662994+00			2025-08-22 02:42:23.912313+00	2025-08-22 02:52:18.663017+00	\N	1755830408667
bcc6252a-6003-411e-8251-daea25bac9f4	660420	Cutting	Cut materials according to specifications	Completed	2025-08-21 11:21:53.91992+00			2025-08-21 09:53:06.076209+00	2025-08-21 11:21:53.919948+00	\N	1755769919280
01ad0761-0fb0-4cf0-bacc-ceb9310098aa	660420	Fabrication	Fabricate components for the project	Completed	2025-08-21 11:21:57.489578+00			2025-08-21 09:53:06.07621+00	2025-08-21 11:21:57.489602+00	\N	1755769919280
2afa647c-c597-4756-bac9-1ba8b3f81c68	888866	Installation	Install the project components on-site	Completed	2025-08-22 02:52:20.214982+00			2025-08-22 02:42:23.912314+00	2025-08-22 02:52:20.215008+00	\N	1755830408667
c4a62bb2-3dbf-41a7-aa1b-b79eef183293	660420	Installation	Install the project components on-site	Completed	2025-08-21 11:22:05.570485+00			2025-08-21 09:53:06.076211+00	2025-08-21 11:22:05.570512+00	\N	1755769919280
1f03cb2d-82d6-4290-9964-4d2773c956e0	660420	Service Call	Handle any service calls or maintenance	Completed	2025-08-21 11:22:09.67063+00			2025-08-21 09:53:06.076213+00	2025-08-21 11:22:09.670674+00	\N	1755769919280
c12e4a50-dd60-47d1-bb6a-e8f7d2ff09da	914196	Template	Prepare and finalize the template for the project	In Progress	2025-08-21 13:20:58.859357+00			2025-08-21 13:20:58.859357+00	2025-08-21 13:20:58.85941+00	\N	1755782380341
f82590c6-0764-4d82-af1c-8fa450e4511b	914196	CAD	Create CAD designs for the project	Not Started	\N			2025-08-21 13:20:58.859358+00	2025-08-21 13:20:59.377008+00	\N	1755782380341
716916ad-bc5c-4085-9feb-52c8f2720bd6	914196	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-21 13:20:58.859359+00	2025-08-21 13:20:59.986671+00	\N	1755782380341
2e17e5cb-3737-4a16-bc32-226d6b7bba02	914196	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-21 13:20:58.85936+00	2025-08-21 13:21:00.292672+00	\N	1755782380341
ad2f7e3f-9cbb-4a6d-838c-77544809eee3	914196	Installation	Install the project components on-site	Not Started	\N			2025-08-21 13:20:58.859362+00	2025-08-21 13:21:00.600186+00	\N	1755782380341
fd06efc0-de8a-41e2-bd56-5dbdddb48144	914196	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-21 13:20:58.859363+00	2025-08-21 13:21:00.906728+00	\N	1755782380341
8b1d5fc1-df8e-4e94-a742-ec2190fbe5c3	652375	Installation	Install the project components on-site	Completed	2025-08-22 03:17:38.651958+00			2025-08-22 03:11:25.699762+00	2025-08-22 03:17:38.651992+00	\N	1755832205460
74431ea3-004c-4e8e-8d49-5eaa670ba7be	888866	Service Call	Handle any service calls or maintenance	Completed	2025-08-22 02:52:22.070984+00			2025-08-22 02:42:23.912315+00	2025-08-22 02:52:22.07101+00	\N	1755830408667
27c6c46d-cbbf-4687-801f-b7ece7b3ad48	629876	Template	Prepare and finalize the template for the project	In Progress	2025-08-22 02:55:02.996626+00			2025-08-22 02:55:02.996625+00	2025-08-22 02:55:02.996703+00	\N	1755831198898
4ae6a0c8-ff79-4514-bd1a-04cbe7c48de7	629876	CAD	Create CAD designs for the project	Not Started	\N			2025-08-22 02:55:02.996627+00	2025-08-22 02:55:02.998486+00	\N	1755831198898
d175aae2-5a90-4f6b-a89a-4391e3df5dd8	629876	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-22 02:55:02.996629+00	2025-08-22 02:55:02.999621+00	\N	1755831198898
ed044542-ae3f-4308-afeb-71436e4bac96	629876	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-22 02:55:02.99663+00	2025-08-22 02:55:03.001906+00	\N	1755831198898
3ae10adc-99ac-4d23-b8d9-b1e29270d966	629876	Installation	Install the project components on-site	Not Started	\N			2025-08-22 02:55:02.996631+00	2025-08-22 02:55:03.003646+00	\N	1755831198898
1ec83467-9059-4f30-8111-1379edbdd0ca	629876	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-22 02:55:02.996632+00	2025-08-22 02:55:03.006047+00	\N	1755831198898
49d53ac9-b196-410b-b94b-6e5537f3da29	629876	Template	Prepare and finalize the template for the project	In Progress	2025-08-22 02:55:03.008953+00			2025-08-22 02:55:03.008953+00	2025-08-22 02:55:03.009038+00	\N	1755831206626
4ba26730-4965-4545-b66e-f7e39668a126	629876	CAD	Create CAD designs for the project	Not Started	\N			2025-08-22 02:55:03.008954+00	2025-08-22 02:55:03.011282+00	\N	1755831206626
dae479df-1205-4e1b-bd6c-30360b7a3e18	629876	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-22 02:55:03.008956+00	2025-08-22 02:55:03.014434+00	\N	1755831206626
75158a29-6c00-42ad-abdf-de952058998b	629876	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-22 02:55:03.008957+00	2025-08-22 02:55:03.0157+00	\N	1755831206626
8a4b7342-71eb-4809-a569-4ddd5e284a8a	629876	Installation	Install the project components on-site	Not Started	\N			2025-08-22 02:55:03.008959+00	2025-08-22 02:55:03.020147+00	\N	1755831206626
8c995b35-52ed-4e3d-8039-8338cc601d1b	629876	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-22 02:55:03.00896+00	2025-08-22 02:55:03.023324+00	\N	1755831206626
16cb677b-781a-41b5-ba34-4d0a844759bb	556657	Template	Prepare and finalize the template for the project	In Progress	2025-08-22 03:03:24.639876+00			2025-08-22 03:03:24.639876+00	2025-08-22 03:03:24.640246+00	\N	1755831706472
748f5b09-e8c4-4397-86eb-841f6577068a	556657	CAD	Create CAD designs for the project	Not Started	\N			2025-08-22 03:03:24.639877+00	2025-08-22 03:03:24.642942+00	\N	1755831706472
433b81db-0de4-4086-bedb-96100e77ddc8	556657	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-22 03:03:24.639878+00	2025-08-22 03:03:24.645181+00	\N	1755831706472
15198c5f-9e24-430e-97e3-c2e3b49b72b7	556657	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-22 03:03:24.639879+00	2025-08-22 03:03:24.647128+00	\N	1755831706472
47a800d5-dbe1-4174-8eb7-e9dd09869d13	556657	Installation	Install the project components on-site	Not Started	\N			2025-08-22 03:03:24.639881+00	2025-08-22 03:03:24.649088+00	\N	1755831706472
65af3c2b-d168-4398-bfc9-4e9e7404bebd	556657	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-22 03:03:24.639882+00	2025-08-22 03:03:24.65072+00	\N	1755831706472
4c2c2490-f5ce-40aa-a468-2ddd2e7f0f7f	652375	Template	Prepare and finalize the template for the project	Completed	2025-08-22 03:16:40.882122+00			2025-08-22 03:11:25.694707+00	2025-08-22 03:16:40.882146+00	\N	1755832197409
012785dc-11d3-484f-93ce-dc84f7528c25	652375	Service Call	Handle any service calls or maintenance	Completed	2025-08-22 03:17:11.941015+00			2025-08-22 03:11:25.694736+00	2025-08-22 03:17:11.94104+00	\N	1755832197409
b19e9ebc-7691-4854-ae06-7d03dc13cdbe	652375	CAD	Create CAD designs for the project	Completed	2025-08-22 03:16:42.881707+00			2025-08-22 03:11:25.694709+00	2025-08-22 03:16:42.881752+00	\N	1755832197409
11c70517-14bb-4c12-a58e-1d362ff441a7	652375	Template	Prepare and finalize the template for the project	Completed	2025-08-22 03:17:32.197584+00			2025-08-22 03:11:25.699741+00	2025-08-22 03:17:32.197614+00	\N	1755832205460
2b4803c3-8597-4a21-96e6-895ad7646639	652375	Cutting	Cut materials according to specifications	Completed	2025-08-22 03:16:45.066071+00			2025-08-22 03:11:25.694711+00	2025-08-22 03:16:45.066161+00	\N	1755832197409
b6f84230-1838-4a78-aaec-19237658632c	652375	Fabrication	Fabricate components for the project	Completed	2025-08-22 03:16:46.655774+00			2025-08-22 03:11:25.694733+00	2025-08-22 03:16:46.655803+00	\N	1755832197409
7d0b7d97-0514-413f-a9db-51a7a5da9944	652375	Service Call	Handle any service calls or maintenance	Completed	2025-08-22 03:17:56.490168+00			2025-08-22 03:11:25.699763+00	2025-08-22 03:17:56.490196+00	\N	1755832205460
4dbe2e6e-df0e-44f9-bca0-56c7c5038675	652375	Installation	Install the project components on-site	Completed	2025-08-22 03:16:48.061428+00			2025-08-22 03:11:25.694734+00	2025-08-22 03:16:48.061454+00	\N	1755832197409
9628a077-081c-42bc-bdcd-9a9bd7ad4f92	652375	CAD	Create CAD designs for the project	Completed	2025-08-22 03:17:34.012783+00			2025-08-22 03:11:25.699742+00	2025-08-22 03:17:34.012809+00	\N	1755832205460
b73bc9f4-1cb7-4608-975e-abd5ccfc04ba	225983	Template	Prepare and finalize the template for the project	In Progress	2025-08-22 03:20:46.131597+00			2025-08-22 03:20:46.131596+00	2025-08-22 03:20:46.131684+00	\N	1755832729219
40b2dbf2-3632-431e-aaf5-d429f63445ac	652375	Cutting	Cut materials according to specifications	Completed	2025-08-22 03:17:35.49277+00			2025-08-22 03:11:25.699743+00	2025-08-22 03:17:35.492795+00	\N	1755832205460
e537ed09-1f49-4ced-82ae-f1fc78eb8b1a	225983	CAD	Create CAD designs for the project	Not Started	\N			2025-08-22 03:20:46.131598+00	2025-08-22 03:20:46.132968+00	\N	1755832729219
f625129f-0479-43ad-8448-c8b19e679fa4	652375	Fabrication	Fabricate components for the project	Completed	2025-08-22 03:17:37.072567+00			2025-08-22 03:11:25.699745+00	2025-08-22 03:17:37.072596+00	\N	1755832205460
8c9bc908-9895-4812-8012-ffb1ac3d4b73	225983	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-22 03:20:46.1316+00	2025-08-22 03:20:46.134914+00	\N	1755832729219
f125aba3-d15c-4c5b-99d9-5fd23aad1417	225983	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-22 03:20:46.131601+00	2025-08-22 03:20:46.135511+00	\N	1755832729219
e06a6a9d-81e7-4c0e-a34a-96cfa7989a8c	225983	Installation	Install the project components on-site	Not Started	\N			2025-08-22 03:20:46.131602+00	2025-08-22 03:20:46.135853+00	\N	1755832729219
eb3929da-896f-4617-9245-499cd43b8e18	225983	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-22 03:20:46.131604+00	2025-08-22 03:20:46.136274+00	\N	1755832729219
7db61ebb-12b9-4719-97c5-9f5bdf7eea7a	580823	Template	Prepare and finalize the template for the project	In Progress	2025-08-22 08:46:09.570203+00			2025-08-22 08:46:09.570201+00	2025-08-22 08:46:09.571507+00	\N	1755852182868
f00751d0-f649-4867-b0b7-593b94fda91c	580823	CAD	Create CAD designs for the project	Not Started	\N			2025-08-22 08:46:09.570204+00	2025-08-22 08:46:10.184737+00	\N	1755852182868
aba7371c-1019-4648-906e-82b1966b99de	580823	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-22 08:46:09.570205+00	2025-08-22 08:46:10.726871+00	\N	1755852182868
9c40f445-eace-4d53-8c28-4245963b0a10	580823	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-22 08:46:09.570205+00	2025-08-22 08:46:11.008459+00	\N	1755852182868
f642aad6-0327-43d1-830b-25bf2602e290	580823	Installation	Install the project components on-site	Not Started	\N			2025-08-22 08:46:09.570206+00	2025-08-22 08:46:11.312437+00	\N	1755852182868
62d620a7-15f4-4402-b59d-78451fa3ebe4	580823	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-22 08:46:09.570206+00	2025-08-22 08:46:11.576795+00	\N	1755852182868
c496065e-998f-4ecd-b524-6a977b0c7cea	235859	Template	Prepare and finalize the template for the project	In Progress	2025-08-22 09:43:35.314547+00			2025-08-22 09:43:35.314546+00	2025-08-22 09:43:35.314577+00	\N	1755855741072
7d613a7e-223d-4a72-9b8f-1790860b31c4	235859	CAD	Create CAD designs for the project	Not Started	\N			2025-08-22 09:43:35.314548+00	2025-08-22 09:43:35.315768+00	\N	1755855741072
4e7fa6a8-e8c9-4bfb-8372-016f1340177f	235859	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-22 09:43:35.31455+00	2025-08-22 09:43:35.3167+00	\N	1755855741072
952100e1-aee7-4702-9e74-02323767e256	235859	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-22 09:43:35.314552+00	2025-08-22 09:43:35.31714+00	\N	1755855741072
676d62e9-7574-4d82-84d7-11ac7d252d7e	235859	Installation	Install the project components on-site	Not Started	\N			2025-08-22 09:43:35.314553+00	2025-08-22 09:43:35.317583+00	\N	1755855741072
575867aa-647c-49e5-b20a-f26d8effd08e	235859	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-22 09:43:35.314554+00	2025-08-22 09:43:35.318054+00	\N	1755855741072
a6cd1371-f1d8-4de8-858c-a49bba951f7c	142588	Template	Prepare and finalize the template for the project	In Progress	2025-08-22 09:57:49.106258+00			2025-08-22 09:57:49.106258+00	2025-08-22 09:57:49.35696+00	\N	1755856420394
abeafc73-2035-4057-a2a9-effd19a31b91	142588	CAD	Create CAD designs for the project	Not Started	\N			2025-08-22 09:57:49.106259+00	2025-08-22 09:57:50.642533+00	\N	1755856420394
b6252758-8d76-431e-a6e4-2b2f22e82f9b	142588	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-22 09:57:49.106259+00	2025-08-22 09:57:51.964452+00	\N	1755856420394
1656b7dd-e082-4a7f-aa05-9e2039db518e	142588	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-22 09:57:49.106259+00	2025-08-22 09:57:53.034197+00	\N	1755856420394
bb606411-8570-4b81-9368-8ee90125ee06	142588	Installation	Install the project components on-site	Not Started	\N			2025-08-22 09:57:49.106259+00	2025-08-22 09:57:54.049297+00	\N	1755856420394
74e05539-63ea-42dd-ae31-21b71c5ee29c	142588	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-22 09:57:49.106259+00	2025-08-22 09:57:55.080967+00	\N	1755856420394
4e9394ab-897d-49f6-879b-d2d855ffd9d7	692681	Template	Prepare and finalize the template for the project	In Progress	2025-08-22 11:07:05.858871+00			2025-08-22 11:07:05.858871+00	2025-08-22 11:07:05.8596+00	\N	1755860786021
b7986363-8dba-4855-987a-42ad804753be	692681	CAD	Create CAD designs for the project	Not Started	\N			2025-08-22 11:07:05.858871+00	2025-08-22 11:07:06.574707+00	\N	1755860786021
f26d42f6-c546-43b1-8605-7013b283408e	692681	Cutting	Cut materials according to specifications	Not Started	\N			2025-08-22 11:07:05.858872+00	2025-08-22 11:07:07.136655+00	\N	1755860786021
54f2c373-a53a-451c-bfbf-5272e630fe66	692681	Fabrication	Fabricate components for the project	Not Started	\N			2025-08-22 11:07:05.858873+00	2025-08-22 11:07:07.49619+00	\N	1755860786021
43d9ef6c-a6b3-438c-a40a-16df374000a5	692681	Installation	Install the project components on-site	Not Started	\N			2025-08-22 11:07:05.858873+00	2025-08-22 11:07:07.803636+00	\N	1755860786021
584834c3-62c1-46d3-a126-2f1883a920fd	692681	Service Call	Handle any service calls or maintenance	Not Started	\N			2025-08-22 11:07:05.858873+00	2025-08-22 11:07:08.110741+00	\N	1755860786021
\.


--
-- Data for Name: tasks; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tasks (id, created_at, updated_at, deleted_at, title, description, due_date, completed_at, priority, status, task_type) FROM stdin;
1	2025-08-12 05:54:35.715735+00	2025-08-12 05:54:35.715735+00	\N	new task	asdsadffasdfsad	2025-08-12T05:53:52.571Z	\N	low	pending	meeting
2	2025-08-12 13:22:47.656736+00	2025-08-12 13:22:47.656736+00	\N	Pradeep_task	sdksd	2025-08-12T18:30:00.000Z	\N	low	pending	email
3	2025-08-14 13:00:05.237601+00	2025-08-14 13:00:05.237601+00	\N	asfasdf	asdfasdf	2025-08-14T11:37:22.072Z	\N	medium	pending	follow_up
4	2025-08-14 13:02:04.816316+00	2025-08-14 13:02:04.816316+00	\N	sadfasdfasdfasdf	asdfasdf	2025-08-19T18:30:00.000Z	\N	medium	pending	follow_up
5	2025-08-14 13:04:41.951263+00	2025-08-14 13:04:41.951263+00	\N	sadfasdf	asdfasdf	2025-08-14T13:04:19.988Z	\N	medium	pending	follow_up
6	2025-08-14 15:18:57.130759+00	2025-08-14 15:18:57.130759+00	\N	safsadf	asdfasfsadf	2025-08-14T15:18:43.965Z	\N	medium	pending	follow_up
\.


--
-- Name: addresses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.addresses_id_seq', 146, true);


--
-- Name: county_taxes_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.county_taxes_id_seq', 12, true);






--
-- Name: files_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.files_id_seq', 257, true);





--
-- Name: manager_residential_entries_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.manager_residential_entries_id_seq', 5, true);


--
-- Name: manager_residential_types_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.manager_residential_types_id_seq', 5, true);


--
-- Name: material_rows_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.material_rows_id_seq', 1, false);


--
-- Name: material_templates_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.material_templates_id_seq', 9, true);


--
-- Name: pickup_locations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.pickup_locations_id_seq', 7, true);


--
-- Name: residential_entries_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.residential_entries_id_seq', 1755856420394, true);


--
-- Name: service_categories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.service_categories_id_seq', 6, true);


--
-- Name: service_groups_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.service_groups_id_seq', 6, true);




--
-- Name: service_rows_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.service_rows_id_seq', 1, false);



--
-- Name: services_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.services_id_seq', 6, true);


--
-- Name: tasks_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.tasks_id_seq', 6, true);


--
-- Name: addresses addresses_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.addresses
    ADD CONSTRAINT addresses_pkey PRIMARY KEY (id);


--
-- Name: agents agents_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.agents
    ADD CONSTRAINT agents_pkey PRIMARY KEY (id);


--
-- Name: associate_customers associate_customers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.associate_customers
    ADD CONSTRAINT associate_customers_pkey PRIMARY KEY (id);


--
-- Name: county_taxes county_taxes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.county_taxes
    ADD CONSTRAINT county_taxes_pkey PRIMARY KEY (id);


--
-- Name: customers customers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);



--
-- Name: estimates estimates_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.estimates
    ADD CONSTRAINT estimates_pkey PRIMARY KEY (id);


--
-- Name: files files_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_pkey PRIMARY KEY (id);




--
-- Name: lead_associates lead_associates_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lead_associates
    ADD CONSTRAINT lead_associates_pkey PRIMARY KEY (lead_id, associate_customer_id);


--
-- Name: leads leads_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.leads
    ADD CONSTRAINT leads_pkey PRIMARY KEY (id);


--
-- Name: manager_residential_entries manager_residential_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.manager_residential_entries
    ADD CONSTRAINT manager_residential_entries_pkey PRIMARY KEY (id);


--
-- Name: manager_residential_types manager_residential_types_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.manager_residential_types
    ADD CONSTRAINT manager_residential_types_pkey PRIMARY KEY (id);


--
-- Name: material_rows material_rows_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.material_rows
    ADD CONSTRAINT material_rows_pkey PRIMARY KEY (id);


--
-- Name: material_templates material_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.material_templates
    ADD CONSTRAINT material_templates_pkey PRIMARY KEY (id);


--
-- Name: pickup_locations pickup_locations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pickup_locations
    ADD CONSTRAINT pickup_locations_pkey PRIMARY KEY (id);


--
-- Name: residential_entries residential_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.residential_entries
    ADD CONSTRAINT residential_entries_pkey PRIMARY KEY (id);


--
-- Name: service_categories service_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.service_categories
    ADD CONSTRAINT service_categories_pkey PRIMARY KEY (id);


--
-- Name: service_groups service_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.service_groups
    ADD CONSTRAINT service_groups_pkey PRIMARY KEY (id);




--
-- Name: service_rows service_rows_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.service_rows
    ADD CONSTRAINT service_rows_pkey PRIMARY KEY (id);



--
-- Name: services services_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.services
    ADD CONSTRAINT services_pkey PRIMARY KEY (id);




--
-- Name: stage_tasks stage_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stage_tasks
    ADD CONSTRAINT stage_tasks_pkey PRIMARY KEY (id);


--
-- Name: stage_users stage_users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stage_users
    ADD CONSTRAINT stage_users_pkey PRIMARY KEY (id);


--
-- Name: stages stages_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stages
    ADD CONSTRAINT stages_pkey PRIMARY KEY (id);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);


--
-- Name: agents uni_agents_email; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.agents
    ADD CONSTRAINT uni_agents_email UNIQUE (email);


--
-- Name: customers uni_customers_name; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT uni_customers_name UNIQUE (name);


--
-- Name: idx_agents_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_agents_deleted_at ON public.agents USING btree (deleted_at);


--
-- Name: idx_agents_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_agents_email ON public.agents USING btree (email);


--
-- Name: idx_county_taxes_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_county_taxes_name ON public.county_taxes USING btree (name);



--
-- Name: idx_estimates_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_estimates_deleted_at ON public.estimates USING btree (deleted_at);





--
-- Name: idx_leads_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_leads_deleted_at ON public.leads USING btree (deleted_at);


--
-- Name: idx_material_templates_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_material_templates_name ON public.material_templates USING btree (name);


--
-- Name: idx_pickup_locations_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_pickup_locations_deleted_at ON public.pickup_locations USING btree (deleted_at);


--
-- Name: idx_service_groups_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_service_groups_name ON public.service_groups USING btree (name);






--
-- Name: idx_stage_tasks_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_stage_tasks_deleted_at ON public.stage_tasks USING btree (deleted_at);


--
-- Name: idx_stage_tasks_stage_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_stage_tasks_stage_id ON public.stage_tasks USING btree (stage_id);


--
-- Name: idx_stage_users_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_stage_users_deleted_at ON public.stage_users USING btree (deleted_at);


--
-- Name: idx_stage_users_stage_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_stage_users_stage_id ON public.stage_users USING btree (stage_id);


--
-- Name: idx_stages_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_stages_deleted_at ON public.stages USING btree (deleted_at);


--
-- Name: idx_stages_lead_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_stages_lead_id ON public.stages USING btree (lead_id);


--
-- Name: idx_tasks_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_tasks_deleted_at ON public.tasks USING btree (deleted_at);


--
-- Name: idx_tasks_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_tasks_status ON public.tasks USING btree (status);


--
-- Name: idx_tasks_title; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_tasks_title ON public.tasks USING btree (title);





INSERT INTO manager_residential_entries (residential_entry_name)
VALUES 
  ('Kitchen'),
  ('Bathroom'),
  ('Master Bathroom'),
  ('Master Bedroom'),
  ('Other')
ON CONFLICT (residential_entry_name) DO NOTHING;

-- manager_residential_types
INSERT INTO manager_residential_types (residential_entry_type)
VALUES
  ('Residential Remodel'),
  ('Residential New-Construction'),
  ('Commercial'),
  ('Apartment')
ON CONFLICT (residential_entry_type) DO NOTHING;

-- county_taxes
INSERT INTO county_taxes (name, tax_percentage)
VALUES 
  ('Jackson', 7.2),
  ('Jefferson', 8.5),
  ('Madison', 6.8)
ON CONFLICT (name) DO UPDATE SET tax_percentage = EXCLUDED.tax_percentage;


-- pickup_locations
INSERT INTO pickup_locations (name, street, unit_suite, city, state, zip, country)
VALUES
  ('Downtown store', '456 Market Street', NULL, 'Miami', 'Florida', '62701', 'USA'),
  ('Main Warehouse', '123 Warehouse Blvd', NULL, 'Tampa', 'Florida', '62704', 'USA'),
  ('East Side Location', '789 East Avenue', NULL, 'Orlando', 'Florida', '62702', 'USA')
ON CONFLICT (name) DO UPDATE SET 
    street = EXCLUDED.street,
    unit_suite = EXCLUDED.unit_suite,
    city = EXCLUDED.city,
    state = EXCLUDED.state,
    zip = EXCLUDED.zip,
    country = EXCLUDED.country;



-- service_groups
INSERT INTO service_groups (name)
VALUES ('Cleaning'), ('Polishing')
ON CONFLICT (name) DO NOTHING;

-- service_categories
INSERT INTO service_categories (name)
VALUES ('Cleaning kit'), ('Polishing kit')
ON CONFLICT (name) DO NOTHING;




--
-- PostgreSQL database dump complete
--
