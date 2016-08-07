DROP TABLE IF EXISTS type1;
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
)
