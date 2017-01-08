CREATE TABLE IF NOT EXISTS example1 (
  id BIGSERIAL PRIMARY KEY,
  text_field text DEFAULT '',
  uuid_field uuid NOT NULL,
  created_at timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  updated_at timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS example2 (
  uuid_field uuid PRIMARY KEY,
  text_field text NOT NULL DEFAULT '',
  bool_field boolean NOT NULL DEFAULT 't',
  created_at timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  updated_at timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS example3 (
  id BIGSERIAL PRIMARY KEY,
  email   text DEFAULT '',
  created timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  updated timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp
);
