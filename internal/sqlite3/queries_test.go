package sqlite3_test

import (
	"database/sql"
	"os"

	// load in the sqllite driver.
	_ "github.com/mattn/go-sqlite3"

	. "bitbucket.org/jatone/genieql/internal/sqlite3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("queries", func() {
	DescribeTable("Insert",
		func(table string, columns, defaults []string, query string) {
			Expect(Insert(1, table, columns, defaults)).To(Equal(query))
		},
		Entry("example 1", "MyTable1", []string{"col1", "col2", "col3"}, []string{}, "INSERT INTO MyTable1 (col1,col2,col3) VALUES ($1,$2,$3)"),
		Entry("example 2", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col4"}, "INSERT INTO MyTable2 (col1,col2,col3,col4) VALUES ($1,$2,$3,DEFAULT)"),
		Entry("example 3", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col3"}, "INSERT INTO MyTable2 (col1,col2,col3,col4) VALUES (DEFAULT,$1,DEFAULT,$2)"),
	)

	DescribeTable("Select",
		func(table string, columns, predicates []string, query string) {
			Expect(Select(table, columns, predicates)).To(Equal(query))
		},
		Entry("example 1", "MyTable1", []string{"col1", "col2", "col3"}, []string{"col1"}, "SELECT col1,col2,col3 FROM MyTable1 WHERE col1 = $1"),
		Entry("example 2", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col2"}, "SELECT col1,col2,col3,col4 FROM MyTable2 WHERE col1 = $1 AND col2 = $2"),
	)

	DescribeTable("Update",
		func(table string, columns, predicates []string, query string) {
			Expect(Update(table, columns, predicates)).To(Equal(query))
		},
		Entry("example 1", "MyTable1", []string{"col1", "col2", "col3"}, []string{"col1"}, "UPDATE MyTable1 SET col1 = $1, col2 = $2, col3 = $3 WHERE col1 = $4"),
		Entry("example 2", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col2"}, "UPDATE MyTable2 SET col1 = $1, col2 = $2, col3 = $3, col4 = $4 WHERE col1 = $5 AND col2 = $6"),
		Entry("example 3", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{}, "UPDATE MyTable2 SET col1 = $1, col2 = $2, col3 = $3, col4 = $4 WHERE 't'"),
	)

	DescribeTable("Delete",
		func(table string, columns, predicates []string, query string) {
			Expect(Delete(table, columns, predicates)).To(Equal(query))
		},
		Entry("example 1", "MyTable1", []string{"col1", "col2", "col3"}, []string{"col1"}, "DELETE FROM MyTable1 WHERE col1 = $1"),
		Entry("example 2", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col2"}, "DELETE FROM MyTable2 WHERE col1 = $1 AND col2 = $2"),
		Entry("example 3", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{}, "DELETE FROM MyTable2 WHERE 't'"),
	)

	Describe("queries should be valid", func() {
		var (
			dbfile *os.File
			db     *sql.DB
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
		})

		AfterEach(func() {
			Expect(os.Remove(dbfile.Name())).ToNot(HaveOccurred())
			Expect(dbfile.Close()).ToNot(HaveOccurred())
		})

		It("should be able to select", func() {
			var (
				id int
			)
			query := Select("example", []string{"id"}, []string{"name"})
			Expect(db.QueryRow(query, "foo").Scan(&id)).To(MatchError(sql.ErrNoRows))
		})

		It("should be able to insert", func() {
			var (
				err   error
				query string
				id    int
			)

			query = Insert(1, "example", []string{"id", "name"}, []string{})
			_, err = db.Exec(query, 1, "foo")
			Expect(err).ToNot(HaveOccurred())
			_, err = db.Exec(query, 2, "bar")
			Expect(err).ToNot(HaveOccurred())
			query = Select("example", []string{"id"}, []string{"name"})
			Expect(db.QueryRow(query, "foo").Scan(&id)).ToNot(HaveOccurred())
			Expect(id).To(Equal(1))
		})

		It("should be able to update", func() {
			var (
				err   error
				query string
				name  string
			)

			query = Insert(1, "example", []string{"id", "name"}, []string{})
			_, err = db.Exec(query, 1, "foo")
			Expect(err).ToNot(HaveOccurred())

			query = Select("example", []string{"name"}, []string{"id"})
			Expect(db.QueryRow(query, 1).Scan(&name)).ToNot(HaveOccurred())
			Expect(name).To(Equal("foo"))

			query = Update("example", []string{"name"}, []string{"id"})
			_, err = db.Exec(query, "bar", 1)
			Expect(err).ToNot(HaveOccurred())

			query = Select("example", []string{"name"}, []string{"id"})
			Expect(db.QueryRow(query, 1).Scan(&name)).ToNot(HaveOccurred())
			Expect(name).To(Equal("bar"))
		})

		It("should be able to delete", func() {
			var (
				err   error
				query string
				id    int
			)

			query = Insert(1, "example", []string{"id", "name"}, []string{})
			_, err = db.Exec(query, 1, "foo")
			Expect(err).ToNot(HaveOccurred())

			query = Select("example", []string{"id"}, []string{"id"})
			Expect(db.QueryRow(query, 1).Scan(&id)).ToNot(HaveOccurred())
			Expect(id).To(Equal(1))

			query = Delete("example", []string{}, []string{"id"})
			_, err = db.Exec(query, 1)
			Expect(err).ToNot(HaveOccurred())

			query = Select("example", []string{"id"}, []string{"id"})
			Expect(db.QueryRow(query, 1).Scan(&id)).To(MatchError(sql.ErrNoRows))
		})
	})

})
