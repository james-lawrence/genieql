package generators

import (
	"go/ast"
	"go/build"
	"go/types"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/internal/x/stringsx"
)

// mappedParam converts a *ast.Field that represents a struct into an array
// of ColumnInfo.
func mappedParam(ctx Context, param *ast.Field) (m genieql.MappingConfig, infos []genieql.ColumnInfo, err error) {
	var (
		pkg *build.Package
	)

	ipath := importPath(ctx, unwrapExpr(param.Type))

	if ipath == ctx.CurrentPackage.ImportPath {
		pkg = ctx.CurrentPackage
	} else {
		if pkg, err = genieql.LocatePackage(ipath, build.Default, genieql.StrictPackageImport(ipath)); err != nil {
			return m, infos, err
		}
	}

	if err = ctx.Configuration.ReadMap("default", &m, genieql.MCOPackage(pkg), genieql.MCOType(types.ExprString(determineType(param.Type)))); err != nil {
		return m, infos, err
	}

	infos, _, err = m.MappedColumnInfo(ctx.Driver, ctx.Dialect, ctx.FileSet, pkg)
	return m, infos, err
}

// converts a *ast.Field that represents a struct into a list of fields that map
// to columns.
func mappedFields(ctx Context, param *ast.Field, ignoreSet ...string) ([]*ast.Field, error) {
	var (
		err   error
		infos []*ast.Field
		m     genieql.MappingConfig
		pkg   *build.Package
	)

	ipath := importPath(ctx, param.Type)

	if ipath == ctx.CurrentPackage.ImportPath {
		pkg = ctx.CurrentPackage
	} else {
		if pkg, err = genieql.LocatePackage(ipath, build.Default, genieql.StrictPackageName(filepath.Base(ipath))); err != nil {
			return infos, err
		}
	}

	if err = ctx.Configuration.ReadMap("default", &m, genieql.MCOPackage(pkg), genieql.MCOType(types.ExprString(determineType(param.Type)))); err != nil {
		return infos, err
	}

	infos, _, err = m.MappedFields(ctx.Dialect, ctx.FileSet, pkg, ignoreSet...)
	return infos, err
}

func mappedStructure(ctx Context, param *ast.Field, ignoreSet ...string) ([]genieql.ColumnInfo, []*ast.Field, error) {
	var (
		err     error
		infos   []*ast.Field
		columns []genieql.ColumnInfo
		m       genieql.MappingConfig
		pkg     *build.Package
	)

	ipath := importPath(ctx, param.Type)
	if ipath == ctx.CurrentPackage.ImportPath {
		pkg = ctx.CurrentPackage
	} else if pkg, err = genieql.LocatePackage(ipath, build.Default, genieql.StrictPackageName(filepath.Base(ipath))); err != nil {
		return columns, infos, err
	}

	if err = ctx.Configuration.ReadMap("default", &m, genieql.MCOPackage(pkg), genieql.MCOType(types.ExprString(unwrapExpr(param.Type)))); err != nil {
		return columns, infos, err
	}

	infos, _, err = m.MapFieldsToColumns(
		ctx.FileSet,
		ctx.CurrentPackage,
		genieql.ColumnInfoSet(m.Columns).Filter(genieql.ColumnInfoFilterIgnore(ignoreSet...))...,
	)

	return m.Columns, infos, err
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
	x := removeEllipsis(param.Type)
	if builtinType(x) {
		return builtinParam(param)
	}

	return mapParam(ctx, param, ignoreSet...)
}

func removeEllipsis(e ast.Expr) ast.Expr {
	if e, ellipsis := e.(*ast.Ellipsis); ellipsis {
		return e.Elt
	}

	return e
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
