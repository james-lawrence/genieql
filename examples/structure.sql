CREATE TABLE IF NOT EXISTS example1 (
  uuid_field uuid PRIMARY KEY,
  smallint_field smallint NOT NULL DEFAULT 1,
  int_field integer NOT NULL DEFAULT 1,
  bigint_field bigint NOT NULL DEFAULT 1,
  decimal_field decimal NOT NULL DEFAULT 1.0,
  numeric_field numeric NOT NULL DEFAULT 1.0,
  real_field real NOT NULL DEFAULT 1.0,
  double_precision_field double precision NOT NULL DEFAULT 1.0,
  character_field varchar(10) NOT NULL DEFAULT '',
  character_fixed_field char(10) NOT NULL DEFAULT '',
  byte_array_field bytea NOT NULL DEFAULT ''::bytea,
  interval_field interval NOT NULL DEFAULT INTERVAL '1 seconds',
  cidr_field cidr NOT NULL DEFAULT '0.0.0.0/8'::cidr,
  inet_field inet NOT NULL DEFAULT '0.0.0.0'::inet,
  macaddr_field macaddr NOT NULL DEFAULT '00:00:00:00:00:00'::macaddr,
  bit_field bit(8) NOT NULL DEFAULT '00000000'::bit(8),
  bit_varying_field bit varying (8) NOT NULL DEFAULT '0'::bit(1),
  json_field json NOT NULL DEFAULT '{}'::json,
  jsonb_field jsonb NOT NULL DEFAULT '{}'::jsonb,
  text_field text NOT NULL DEFAULT '',
  bool_field boolean NOT NULL DEFAULT 'f',
  uuid_array uuid[] not null default '{}'::uuid[],
  int2_array int2[] not null default '{}'::int2[],
  int4_array int4[] not null default '{}'::int4[],
  int8_array int8[] not null default '{}'::int8[],
  timestamp_field timestamptz NOT NULL DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS example2 (
  uuid_field uuid PRIMARY KEY,
  text_field text NOT NULL DEFAULT '',
  bool_field boolean NOT NULL DEFAULT 'f',
  uuid_array uuid[] not null default '{}'::uuid[],
  int4_array int4[] not null default '{}'::int4[],
  int8_array int8[] not null default '{}'::int8[],
  timestamp_field timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS example3 (
  id BIGSERIAL PRIMARY KEY,
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

CREATE TABLE IF NOT EXISTS timestamp_examples (
  uuid uuid PRIMARY KEY,
  "timestamp" timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  "timestamptz" timestamptz NOT NULL DEFAULT current_timestamp,
  "timestamptz_nullable" timestamptz DEFAULT current_timestamp
);