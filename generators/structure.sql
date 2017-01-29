CREATE TABLE IF NOT EXISTS type1 (
	field1 text PRIMARY KEY DEFAULT ''::text NOT NULL,
	field2 text,
	field3 boolean DEFAULT FALSE NOT NULL,
	field4 boolean,
	field5 int DEFAULT 0 NOT NULL,
	field6 int,
	field7 timestamp with time zone DEFAULT (now() at time zone 'utc') NOT NULL,
	field8 timestamp with time zone,
	unmappedField int DEFAULT 1 NOT NULL
);

CREATE TABLE IF NOT EXISTS example1 (
	id BIGSERIAL PRIMARY KEY,
	text_field text DEFAULT '',
	uuid_field uuid NOT NULL,
	created_at timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
	updated_at timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp
);
