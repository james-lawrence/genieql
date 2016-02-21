CREATE TABLE IF NOT EXISTS query_literal (
  id BIGSERIAL PRIMARY KEY,
  email   text DEFAULT '',
  created timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  updated timestamp WITH TIME ZONE NOT NULL DEFAULT current_timestamp
);
