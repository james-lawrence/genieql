```bash
[jatone@jambli genieql]$ gilo-shim $PWD/bin/sqlmap --package=bitbucket.org/jatone/sso --type=Identity --table=identity lowersnakecase
2016/01/17 20:19:23 Package bitbucket.org/jatone/sso
2016/01/17 20:19:23 Type Identity
2016/01/17 20:19:23 Table identity
2016/01/17 20:19:23 Strategy lowersnakecase
2016/01/17 20:19:23 Custom Aliases []
--- m dump:
type: sso.Identity
fields:
- type: string
  structfield: ID
  aliases:
  - id
  - identity_id
- type: string
  structfield: Email
  aliases:
  - email
  - identity_email
- type: time.Time
  structfield: Created
  aliases:
  - created
  - identity_created
- type: time.Time
  structfield: Updated
  aliases:
  - updated
  - identity_updated
[jatone@jambli genieql]$ gilo-shim $PWD/bin/sqlmap --package=bitbucket.org/jatone/sso --type=Identity --table=identity uppersnakecase
2016/01/17 20:19:25 Package bitbucket.org/jatone/sso
2016/01/17 20:19:25 Type Identity
2016/01/17 20:19:25 Table identity
2016/01/17 20:19:25 Strategy uppersnakecase
2016/01/17 20:19:25 Custom Aliases []
--- m dump:
type: sso.Identity
fields:
- type: string
  structfield: ID
  aliases:
  - ID
  - IDENTITY_ID
- type: string
  structfield: Email
  aliases:
  - EMAIL
  - IDENTITY_EMAIL
- type: time.Time
  structfield: Created
  aliases:
  - CREATED
  - IDENTITY_CREATED
- type: time.Time
  structfield: Updated
  aliases:
  - UPDATED
  - IDENTITY_UPDATED
 ```