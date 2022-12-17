package postgresql_test

import (
	"bitbucket.org/jatone/genieql"

	"bitbucket.org/jatone/genieql/internal/drivers"
	. "bitbucket.org/jatone/genieql/internal/postgresql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("postgresql", func() {
	Describe("dialect", func() {
		driver := genieql.MustLookupDriver(drivers.PGX)
		It("should return the columns in the query in sorted order", func() {
			info, err := NewDialect(DB).ColumnInformationForQuery(
				driver,
				"SELECT xact_rollback, conflicts, blks_read, blks_hit FROM pg_stat_database",
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(genieql.ColumnInfoSet(info).ColumnNames()).To(Equal([]string{"blks_hit", "blks_read", "conflicts", "xact_rollback"}))
		})

		It("should return the columns in the table in the sorted order", func() {
			info, err := NewDialect(DB).ColumnInformationForTable(driver, "pg_stat_database")
			Expect(err).ToNot(HaveOccurred())
			Expect(
				genieql.ColumnInfoSet(info).ColumnNames(),
			).To(
				Equal([]string{
					"active_time",
					"blk_read_time",
					"blk_write_time",
					"blks_hit",
					"blks_read",
					"checksum_failures",
					"checksum_last_failure",
					"conflicts",
					"datid",
					"datname",
					"deadlocks",
					"idle_in_transaction_time",
					"numbackends",
					"session_time",
					"sessions",
					"sessions_abandoned",
					"sessions_fatal",
					"sessions_killed",
					"stats_reset",
					"temp_bytes",
					"temp_files",
					"tup_deleted",
					"tup_fetched",
					"tup_inserted",
					"tup_returned",
					"tup_updated",
					"xact_commit",
					"xact_rollback",
				}),
			)
		})

		It("should support insert queries", func() {
			q := NewDialect(DB).Insert(1, 0, "table", "", []string{"c1", "c2", "c2"}, []string{"c1", "c2", "c2"}, []string{"c1"})
			Expect(q).To(Equal(`INSERT INTO table ("c1","c2","c2") VALUES (DEFAULT,$1,$2) RETURNING "c1","c2","c2"`))
		})

		It("should support select queries", func() {
			q := NewDialect(DB).Select("table", []string{"c1", "c2", "c2"}, []string{"c1"})
			Expect(q).To(Equal(`SELECT "c1","c2","c2" FROM table WHERE "c1" = $1`))
		})

		It("should support update queries", func() {
			q := NewDialect(DB).Update("table", []string{"c1", "c2", "c2"}, []string{"c1"}, []string{"c1", "c2", "c2"})
			Expect(q).To(Equal(`UPDATE table SET "c1" = $1, "c2" = $2, "c2" = $3 WHERE "c1" = $4 RETURNING "c1","c2","c2"`))
		})

		It("should support delete queries", func() {
			q := NewDialect(DB).Delete("table", []string{"c1", "c2", "c2"}, []string{"c1"})
			Expect(q).To(Equal(`DELETE FROM table WHERE "c1" = $1 RETURNING "c1","c2","c2"`))
		})
	})
})
