package main

import (
	"database/sql"
	"time"

	"bitbucket.org/jatone/sso"
)

type IdentityScanner interface {
	Scan(arg0 *sso.Identity) error
}

func NewMaybeIdentityScanner(r *sql.Rows, err error) IdentityScanner {
	return identityScanner{
		err:  err,
		rows: r,
	}
}

func NewIdentityScanner(r *sql.Rows) IdentityScanner {
	return identityScanner{
		rows: r,
	}
}

type identityScanner struct {
	err  error
	rows *sql.Rows
}

func (t identityScanner) Scan(arg0 *sso.Identity) error {
	if t.err != nil {
		return t.err
	}

	var c0 time.Time
	var c1 string
	var c2 string

	if err := t.rows.Scan(&c0, &c1, &c2); err != nil {
		return err
	}

	arg0.Created = c0
	arg0.Email = c1
	arg0.ID = c2

	return t.rows.Err()
}
