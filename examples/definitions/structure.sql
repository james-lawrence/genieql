CREATE TABLE IF NOT EXISTS example1 (
  id BIGSERIAL PRIMARY KEY,
  textField text NOT NULL DEFAULT '',
  uuidField uuid NOT NULL,
  created timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  updated timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS example2 (
  uuidField uuid PRIMARY KEY,
  textField text NOT NULL DEFAULT '',
  boolField boolean NOT NULL DEFAULT 't',
  created timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  updated timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp
);
