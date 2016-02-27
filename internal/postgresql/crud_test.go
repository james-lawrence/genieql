package postgresql_test

import (
	"fmt"

	. "bitbucket.org/jatone/genieql/internal/postgresql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Crud", func() {
	Describe("Insert", func() {
		var insertQueryTable = []struct {
			table    string
			columns  []string
			defaults []string
			query    string
		}{
			{"MyTable1", []string{"col1", "col2", "col3"}, []string{}, "INSERT INTO MyTable1 (col1,col2,col3) VALUES ($1,$2,$3) RETURNING col1,col2,col3"},
			{"MyTable2", []string{"col1", "col2", "col3"}, []string{"col4"}, "INSERT INTO MyTable2 (col1,col2,col3,col4) VALUES ($1,$2,$3,DEFAULT) RETURNING col1,col2,col3,col4"},
		}
		It("should create insert queries", func() {
			for _, tt := range insertQueryTable {
				Expect(Insert(tt.table, tt.columns, tt.defaults)).To(Equal(tt.query))
			}
		})
	})

	Describe("Select", func() {
		var selectQueryTable = []struct {
			table      string
			columns    []string
			predicates []string
			query      string
		}{
			{"MyTable1", []string{"col1", "col2", "col3"}, []string{"col1"}, "SELECT col1,col2,col3 FROM MyTable1 WHERE col1 = $1"},
			{"MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col2"}, "SELECT col1,col2,col3,col4 FROM MyTable2 WHERE col1 = $1 AND col2 = $2"},
		}

		It("should create select queries", func() {
			for idx, tt := range selectQueryTable {
				Expect(Select(tt.table, tt.columns, tt.predicates)).To(Equal(tt.query), fmt.Sprintf("select query test %d failed", idx))
			}
		})
	})

	Describe("Update", func() {
		var updateQueryTable = []struct {
			table      string
			columns    []string
			predicates []string
			query      string
		}{
			{"MyTable1", []string{"col1", "col2", "col3"}, []string{"col1"}, "UPDATE MyTable1 SET (col1 = $1, col2 = $2, col3 = $3) WHERE col1 = $4 RETURNING col1,col2,col3"},
			{"MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col2"}, "UPDATE MyTable2 SET (col1 = $1, col2 = $2, col3 = $3, col4 = $4) WHERE col1 = $5 AND col2 = $6 RETURNING col1,col2,col3,col4"},
		}

		It("should create update queries", func() {
			for idx, tt := range updateQueryTable {
				Expect(Update(tt.table, tt.columns, tt.predicates)).To(Equal(tt.query), fmt.Sprintf("update query test %d failed", idx))
			}
		})
	})

	Describe("Delete", func() {
		var deleteQueryTable = []struct {
			table      string
			columns    []string
			predicates []string
			query      string
		}{
			{"MyTable1", []string{"col1", "col2", "col3"}, []string{"col1"}, "DELETE FROM MyTable1 WHERE col1 = $1 RETURNING col1,col2,col3"},
			{"MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col2"}, "DELETE FROM MyTable2 WHERE col1 = $1 AND col2 = $2 RETURNING col1,col2,col3,col4"},
		}

		It("should create select queries", func() {
			for idx, tt := range deleteQueryTable {
				Expect(Delete(tt.table, tt.columns, tt.predicates)).To(Equal(tt.query), fmt.Sprintf("delete query test %d failed", idx))
			}
		})
	})
})
