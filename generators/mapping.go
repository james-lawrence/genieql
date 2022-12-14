package generators

import (
	"go/ast"
	"go/build"
	"go/types"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/internal/stringsx"
	"bitbucket.org/jatone/genieql/internal/transformx"
)

// mappedParam converts a *ast.Field that represents a struct into an array
// of ColumnInfo.
func mappedParam(ctx Context, param *ast.Field) (m genieql.MappingConfig, infos []genieql.ColumnInfo, err error) {
	var (
		pkg *build.Package
	)

	ipath := importPath(ctx, astutil.UnwrapExpr(param.Type))

	if ipath == ctx.CurrentPackage.ImportPath {
		pkg = ctx.CurrentPackage
	} else {
		if pkg, err = genieql.LocatePackage(ipath, ".", ctx.Build, genieql.StrictPackageImport(ipath)); err != nil {
			return m, infos, err
		}
	}

	if err = ctx.Configuration.ReadMap(&m, genieql.MCOPackage(pkg), genieql.MCOType(types.ExprString(determineType(param.Type)))); err != nil {
		return m, infos, err
	}

	infos, _, err = m.MappedColumnInfo(ctx.Driver, ctx.Dialect, ctx.FileSet, pkg)
	return m, infos, err
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
	} else if pkg, err = genieql.LocatePackage(ipath, ".", ctx.Build, genieql.StrictPackageName(filepath.Base(ipath))); err != nil {
		return columns, infos, err
	}

	if err = ctx.Configuration.ReadMap(&m, genieql.MCOPackage(pkg), genieql.MCOType(types.ExprString(astutil.UnwrapExpr(param.Type)))); err != nil {
		return columns, infos, err
	}

	infos, _, err = m.MapColumnsToFields(
		ctx.FileSet,
		ctx.CurrentPackage,
		genieql.ColumnInfoSet(m.Columns).Filter(genieql.ColumnInfoFilterIgnore(ignoreSet...))...,
	)

	return m.Columns, infos, err
}

func MapFields(ctx Context, params []*ast.Field, ignoreSet ...string) ([]genieql.ColumnMap, error) {
	result := make([]genieql.ColumnMap, 0, len(params))
	for _, param := range params {
		var (
			err     error
			columns []genieql.ColumnMap
		)

		if columns, err = MapField(ctx, param, ignoreSet...); err != nil {
			return result, err
		}

		result = append(result, columns...)
	}

	return result, nil
}

func MapField(ctx Context, param *ast.Field, ignoreSet ...string) ([]genieql.ColumnMap, error) {
	x := removeEllipsis(param.Type)
	if builtinType(x) {
		return builtinParam(ctx, param)
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

			cMap = append(cMap, column.MapColumn(&ast.SelectorExpr{
				Sel: ast.NewIdent(transformx.String(column.Name, aliaser)),
				X:   arg,
			}))
		}
	}

	return cMap, nil
}
