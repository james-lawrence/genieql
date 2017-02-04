package generators

import (
	"fmt"
	"go/ast"
	"go/types"
	"log"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/x/stringsx"
)

// mappedParam converts a *ast.Field that represents a struct into an array
// of ColumnMap.
func mappedParam(ctx Context, param *ast.Field) (genieql.MappingConfig, []genieql.ColumnInfo, error) {
	var (
		err   error
		infos []genieql.ColumnInfo
		m     genieql.MappingConfig
	)

	if err = ctx.Configuration.ReadMap(packageName(ctx.CurrentPackage, param.Type), types.ExprString(param.Type), "default", &m); err != nil {
		return m, infos, err
	}

	infos, _, err = m.MappedColumnInfo(ctx.Dialect, ctx.FileSet, ctx.CurrentPackage)
	return m, infos, err
}

// converts a *ast.Field that represents a struct into a list of fields that map
// to columns.
func mappedFields(ctx Context, param *ast.Field, ignoreSet ...string) ([]*ast.Field, error) {
	var (
		err   error
		infos []*ast.Field
		m     genieql.MappingConfig
	)

	if err = ctx.Configuration.ReadMap(packageName(ctx.CurrentPackage, param.Type), types.ExprString(param.Type), "default", &m); err != nil {
		return infos, err
	}

	infos, _, err = m.MappedFields(ctx.Dialect, ctx.FileSet, ctx.CurrentPackage, ignoreSet...)
	return infos, err
}

func mappedStructure(ctx Context, param *ast.Field, ignoreSet ...string) ([]genieql.ColumnInfo, []*ast.Field, error) {
	var (
		err     error
		infos   []*ast.Field
		columns []genieql.ColumnInfo
		m       genieql.MappingConfig
	)

	if err = ctx.Configuration.ReadMap(packageName(ctx.CurrentPackage, param.Type), types.ExprString(unwrapExpr(param.Type)), "default", &m); err != nil {
		return columns, infos, err
	}

	if columns, err = m.ColumnInfo(ctx.Dialect); err != nil {
		return columns, infos, err
	}

	infos, _, err = m.MapFieldsToColumns(
		ctx.FileSet,
		ctx.CurrentPackage,
		genieql.ColumnInfoSet(columns).Filter(genieql.ColumnInfoFilterIgnore(ignoreSet...))...,
	)

	return columns, infos, err
}

func mapFields(ctx Context, params []*ast.Field, ignoreSet ...string) ([]genieql.ColumnMap, error) {
	result := make([]genieql.ColumnMap, 0, len(params))
	for _, param := range params {
		var (
			err     error
			columns []genieql.ColumnMap
		)

		if columns, err = mapColumns(ctx, param, ignoreSet...); err != nil {
			return result, err
		}

		result = append(result, columns...)
	}

	return result, nil
}

func mapColumns(ctx Context, param *ast.Field, ignoreSet ...string) ([]genieql.ColumnMap, error) {
	if builtinType(param.Type) {
		return builtinParam(param)
	}
	return mapParam(ctx, param, ignoreSet...)
}

// mappedParam converts a *ast.Field that represents a struct into an array
// of ColumnMap.
func mapParam(ctx Context, param *ast.Field, ignoreSet ...string) ([]genieql.ColumnMap, error) {
	var (
		err     error
		m       genieql.MappingConfig
		columns []genieql.ColumnInfo
		cMap    []genieql.ColumnMap
	)

	if m, columns, err = mappedParam(ctx, param); err != nil {
		return cMap, err
	}
	aliaser := m.Aliaser()

	fmt.Printf("type: %T\n", param.Type)
	fields, err := genieql.ResolveTypeFields(param.Type, ctx.FileSet, ctx.CurrentPackage)
	log.Println(astutil.MapFieldsToNameExpr(fields...), err)
	// ctx.Dialect.
	for _, arg := range param.Names {
		for _, column := range columns {
			if stringsx.Contains(column.Name, ignoreSet...) {
				continue
			}

			c, err := column.MapColumn(&ast.SelectorExpr{
				Sel: ast.NewIdent(aliaser.Alias(column.Name)),
				X:   arg,
			})
			if err != nil {
				return []genieql.ColumnMap{}, err
			}
			cMap = append(cMap, c)
		}
	}

	return cMap, nil
}
