package genieql

import (
	"go/ast"

	"bitbucket.org/jatone/genieql/astutil"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("genieql", func() {
	DescribeTable("mapColumns",
		func(matched, unmatched []ColumnInfo, fields []*ast.Field, aliaser Aliaser) {
			columns := append(matched, unmatched...)
			mcolumns, munmatched := mapInfo(columns, fields, aliaser)
			Expect(mcolumns).To(Equal(matched))
			Expect(munmatched).To(Equal(unmatched))
		},
		Entry(
			"no matches",
			[]ColumnInfo{},
			[]ColumnInfo{{Name: "column1"}, {Name: "column2"}, {Name: "column3"}},
			[]*ast.Field{},
			AliasStrategyLowercase,
		),
		Entry(
			"single exact match",
			[]ColumnInfo{{Name: "column1"}},
			[]ColumnInfo{{Name: "column2"}, {Name: "column3"}},
			[]*ast.Field{
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("column1")),
			},
			AliasStrategyLowercase,
		),
		Entry(
			"single aliased match",
			[]ColumnInfo{{Name: "created_at"}},
			[]ColumnInfo{{Name: "column2"}, {Name: "column3"}},
			[]*ast.Field{
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("CreatedAt")),
			},
			AliasStrategyCamelcase,
		),
	)
	// Describe("mapColumns", func() {
	// 	It("should filter out columns that do not match the provided fields", func() {
	// 		columns := []ColumnInfo{{Name: "column1"}, {Name: "column2"}, {Name: "column3"}}
	// 		filtered, unmmatched := mapColumns(columns, []*ast.Field{}, AliasStrategyLowercase)
	// 		Expect(filtered).To(BeEmpty())
	// 		Expect(unmmatched).To(Equal(columns))
	//
	// 		fields := []*ast.Field{
	// 			{Names: []*ast.Ident{&ast.Ident{Name: "column1"}}},
	// 		}
	//
	// 		filtered, unmmatched = mapColumns(columns, fields, AliasStrategyLowercase)
	// 		Expect(filtered).To(Equal([]ColumnInfo{{Name: "column1"}}))
	// 		Expect(unmmatched).To(Equal([]ColumnInfo{{Name: "column2"}, {Name: "column3"}}))
	// 	})
	// })
})
