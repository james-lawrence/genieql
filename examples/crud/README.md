# Example Installation Instructions
```bash
# From GOPATH, replace database connection information as needed
USERNAME=postgres
HOST=localhost
PORT=5432
pushd src/bitbucket.org/jatone/genieql/examples/crud
createdb -p $PORT -U $USERNAME genie_crud_example "genieql"
cat structure.sql | psql -p $PORT -U $USERNAME -d genie_crud_example
genieql bootstrap postgres://$USERNAME@$HOST:$PORT/genie_crud_example?sslmode=disable
popd
go generate bitbucket.org/jatone/genieql/examples/crud
dropdb genie_crud_example
```
