-- +goose Up
-- +goose StatementBegin
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
  uuid_field uuid PRIMARY KEY,
  smallint_field smallint NOT NULL DEFAULT 1,
  int_field integer NOT NULL DEFAULT 1,
  bigint_field bigint NOT NULL DEFAULT 1,
  decimal_field decimal NOT NULL DEFAULT 1.0,
  numeric_field numeric NOT NULL DEFAULT 1.0,
  real_field real NOT NULL DEFAULT 1.0,
  uinteger_field uinteger NOT NULL DEFAULT 1,
  double_precision_field double precision NOT NULL DEFAULT 1.0,
  -- character_field varchar(10) NOT NULL DEFAULT '',
  -- character_fixed_field char(10) NOT NULL DEFAULT '',
  byte_array_field bytea NOT NULL DEFAULT ''::bytea,
  interval_field interval NOT NULL DEFAULT INTERVAL '1 seconds',
  inet_field inet NOT NULL DEFAULT '0.0.0.0'::inet,
  -- json_field json NOT NULL DEFAULT '{}'::json,
  text_field text NOT NULL DEFAULT '',
  bool_field boolean NOT NULL DEFAULT 'f',
  -- uuid_array uuid[] not null default '{}'::uuid[],
  -- int2_array int2[] not null default '{}'::int2[],
  -- int4_array int4[] not null default '{}'::int4[],
  -- int8_array int8[] not null default '{}'::int8[],
  timestamp_field timestamptz NOT NULL DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS example2 (
  uuid_field uuid PRIMARY KEY,
  text_field text NOT NULL DEFAULT '',
  bool_field boolean NOT NULL DEFAULT 'f',
  -- uuid_array uuid[] not null default '{}'::uuid[],
  -- int4_array int4[] not null default '{}'::int4[],
  -- int8_array int8[] not null default '{}'::int8[],
  timestamp_field timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS example3 (
  uid uuid PRIMARY KEY,
  email   text DEFAULT '',
  created timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  updated timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS example4 (
  id uuid PRIMARY KEY,
  email   text NOT NULL DEFAULT '',
  created timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  updated timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS type1;
DROP TABLE IF EXISTS example1;
DROP TABLE IF EXISTS example2;
DROP TABLE IF EXISTS example3;
DROP TABLE IF EXISTS example4;
-- +goose StatementEnd
