module bitbucket.org/jatone/genieql

go 1.14

require (
	github.com/alecthomas/kingpin v2.2.6+incompatible
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/containous/yaegi v0.8.14
	github.com/davecgh/go-spew v1.1.1
	github.com/jackc/pgtype v1.3.0
	github.com/jackc/pgx/v4 v4.6.0 // indirect
	github.com/lib/pq v1.7.0
	github.com/mattn/go-sqlite3 v1.14.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/serenize/snaker v0.0.0-20171204205717-a683aaf2d516
	github.com/zieckey/goini v0.0.0-20180118150432-0da17d361d26
	golang.org/x/tools v0.0.0-20200619210111-0f592d2728bb
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)

// replace github.com/containous/yaegi => github.com/james-lawrence/yaegi v0.8.9-0.20210220084843-ca146ece0f2f
replace github.com/containous/yaegi => ../yaegi
