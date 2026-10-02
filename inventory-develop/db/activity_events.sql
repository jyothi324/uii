-- Activity events table and trigger for realtime notifications

CREATE TABLE IF NOT EXISTS public.activity_events (
    id bigserial NOT NULL,
    entity_type varchar(50) NOT NULL,
    entity_id varchar(255) NOT NULL,
    "action" varchar(50) NOT NULL,
    user_id varchar(50) NULL,
    user_name varchar(100) NULL,
    payload jsonb NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT activity_events_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_activity_events_entity ON public.activity_events (entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_activity_events_created_at ON public.activity_events (created_at DESC);

CREATE OR REPLACE FUNCTION public.notify_activity_event()
RETURNS trigger AS $$
BEGIN
    PERFORM pg_notify(
        'activity_events_channel',
        row_to_json(NEW)::text
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS activity_events_trigger ON public.activity_events;

CREATE TRIGGER activity_events_trigger
AFTER INSERT ON public.activity_events
FOR EACH ROW
EXECUTE FUNCTION public.notify_activity_event();
