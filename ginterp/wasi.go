package ginterp

import (
	"go/build"

	"bitbucket.org/jatone/genieql/internal/envx"
)

func WasiPackage() *build.Package {
	return &build.Package{
		Dir:           envx.String("", "GENIEQL_WASI_PACKAGE_DIR"),
		Name:          envx.String("", "GENIEQL_WASI_PACKAGE_NAME"),
		ImportComment: envx.String("", "GENIEQL_WASI_PACKAGE_IMPORT_COMMENT"),
		Doc:           envx.String("", "GENIEQL_WASI_PACKAGE_DOC"),
		ImportPath:    envx.String("", "GENIEQL_WASI_PACKAGE_IMPORT_PATH"),
		Root:          envx.String("", "GENIEQL_WASI_PACKAGE_ROOT"),
		SrcRoot:       envx.String("", "GENIEQL_WASI_PACKAGE_SRC_ROOT"),
		PkgRoot:       envx.String("", "GENIEQL_WASI_PACKAGE_PKG_ROOT"),
		PkgTargetRoot: envx.String("", "GENIEQL_WASI_PACKAGE_PKG_TARGET_ROOT"),
		BinDir:        envx.String("", "GENIEQL_WASI_PACKAGE_BIN_DIR"),
		Goroot:        envx.Boolean(false, "GENIEQL_WASI_PACKAGE_GO_ROOT"),
		PkgObj:        envx.String("", "GENIEQL_WASI_PACKAGE_PKG_OBJ"),
		AllTags:       envx.Strings(nil, "GENIEQL_WASI_PACKAGE_ALL_TAGS"),
		ConflictDir:   envx.String("", "GENIEQL_WASI_PACKAGE_CONFLICT_DIR"),
		BinaryOnly:    envx.Boolean(false, "GENIEQL_WASI_PACKAGE_BINARY_ONLY"),
	}
}