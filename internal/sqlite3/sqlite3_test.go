package sqlite3_test

import (
	"database/sql"
	"os"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/internal/drivers"
	. "bitbucket.org/jatone/genieql/internal/sqlite3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sqlite3", func() {
	var (
		dbfile  *os.File
		db      *sql.DB
		dialect genieql.Dialect
		driver  = genieql.MustLookupDriver(drivers.StandardLib)
	)

	BeforeEach(func() {
		var (
			err error
		)

		dbfile, err = os.CreateTemp(".sqllite", "")
		Expect(err).ToNot(HaveOccurred())

		db, err = sql.Open("sqlite3", dbfile.Name())
		Expect(err).ToNot(HaveOccurred())

		_, err = db.Exec("CREATE TABLE example (id integer not null primary key, name text)")
		Expect(err).ToNot(HaveOccurred())

		dialect = NewDialect(db)
	})

	AfterEach(func() {
		Expect(os.Remove(dbfile.Name())).ToNot(HaveOccurred())
		Expect(dbfile.Close()).ToNot(HaveOccurred())
	})

	DescribeTable("Insert",
		func(table string, columns, defaults []string, query string) {
			Expect(dialect.Insert(1, table, columns, defaults)).To(Equal(query))
		},
		Entry("example 1", "MyTable1", []string{"col1", "col2", "col3"}, []string{}, "INSERT INTO MyTable1 (col1,col2,col3) VALUES ($1,$2,$3)"),
		Entry("example 2", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col4"}, "INSERT INTO MyTable2 (col1,col2,col3,col4) VALUES ($1,$2,$3,DEFAULT)"),
		Entry("example 3", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col3"}, "INSERT INTO MyTable2 (col1,col2,col3,col4) VALUES (DEFAULT,$1,DEFAULT,$2)"),
	)

	DescribeTable("Select",
		func(table string, columns, predicates []string, query string) {
			Expect(dialect.Select(table, columns, predicates)).To(Equal(query))
		},
		Entry("example 1", "MyTable1", []string{"col1", "col2", "col3"}, []string{"col1"}, "SELECT col1,col2,col3 FROM MyTable1 WHERE col1 = $1"),
		Entry("example 2", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col2"}, "SELECT col1,col2,col3,col4 FROM MyTable2 WHERE col1 = $1 AND col2 = $2"),
	)

	DescribeTable("Update",
		func(table string, columns, predicates []string, query string) {
			Expect(dialect.Update(table, columns, predicates, columns)).To(Equal(query))
		},
		Entry("example 1", "MyTable1", []string{"col1", "col2", "col3"}, []string{"col1"}, "UPDATE MyTable1 SET col1 = $1, col2 = $2, col3 = $3 WHERE col1 = $4"),
		Entry("example 2", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col2"}, "UPDATE MyTable2 SET col1 = $1, col2 = $2, col3 = $3, col4 = $4 WHERE col1 = $5 AND col2 = $6"),
		Entry("example 3", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{}, "UPDATE MyTable2 SET col1 = $1, col2 = $2, col3 = $3, col4 = $4 WHERE 't'"),
	)

	DescribeTable("Delete",
		func(table string, columns, predicates []string, query string) {
			Expect(dialect.Delete(table, columns, predicates)).To(Equal(query))
		},
		Entry("example 1", "MyTable1", []string{"col1", "col2", "col3"}, []string{"col1"}, "DELETE FROM MyTable1 WHERE col1 = $1"),
		Entry("example 2", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col2"}, "DELETE FROM MyTable2 WHERE col1 = $1 AND col2 = $2"),
		Entry("example 3", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{}, "DELETE FROM MyTable2 WHERE 't'"),
	)

	Describe("ColumnInformationForTable", func() {
		PIt("should return an array of genieql.ColumnInfo", func() {
			columnInfo, err := dialect.ColumnInformationForTable(driver, "example")
			Expect(err).ToNot(HaveOccurred())
			Expect(columnInfo).To(ConsistOf(
				genieql.ColumnInfo{Name: "id", Definition: genieql.ColumnDefinition{Nullable: false, PrimaryKey: true, Type: "integer"}},
				genieql.ColumnInfo{Name: "name", Definition: genieql.ColumnDefinition{Nullable: true, PrimaryKey: false, Type: "text"}},
			))
		})
	})

	Describe("ColumnInformationForQuery", func() {
		PIt("should return an array of genieql.ColumnInfo", func() {
			columnInfo, err := dialect.ColumnInformationForQuery(driver, "SELECT id,name FROM example")
			Expect(err).ToNot(HaveOccurred())
			Expect(columnInfo).To(ConsistOf(
				genieql.ColumnInfo{Name: "id", Definition: genieql.ColumnDefinition{Nullable: true, PrimaryKey: false, Type: "INT"}},
				genieql.ColumnInfo{Name: "name", Definition: genieql.ColumnDefinition{Nullable: true, PrimaryKey: false, Type: "TEXT"}},
			))
		})
	})
})
