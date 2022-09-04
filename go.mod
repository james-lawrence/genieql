module bitbucket.org/jatone/genieql

go 1.16

require (
	github.com/alecthomas/kingpin v2.2.6+incompatible
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/containous/yaegi v0.8.14
	github.com/davecgh/go-spew v1.1.1
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/jackc/pgtype v1.11.0
	github.com/jackc/pgx/v4 v4.16.1
	github.com/mattn/go-sqlite3 v1.14.14
	github.com/onsi/ginkgo/v2 v2.1.4
	github.com/onsi/gomega v1.19.0
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/serenize/snaker v0.0.0-20201027110005-a7ad2135616e
	github.com/zieckey/goini v0.0.0-20180118150432-0da17d361d26
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	golang.org/x/text v0.3.7
	golang.org/x/tools v0.1.11
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/containous/yaegi => github.com/james-lawrence/yaegi v0.8.8-modules-enh
