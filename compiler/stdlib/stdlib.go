// Package stdlib provides wrappers of standard library packages to be imported natively in Yaegi.
package stdlib

import "reflect"

// Symbols variable stores the map of stdlib symbols per package
var Symbols = map[string]map[string]reflect.Value{}

func init() {
	Symbols["github.com/traefik/yaegi/stdlib"] = map[string]reflect.Value{
		"Symbols": reflect.ValueOf(Symbols),
	}
}

// Provide access to go standard library (http://golang.org/pkg/)
// build yaegi-extract in the yaegi repository
// go build -o yaegi-extract ./internal/cmd/extract
// go list std | grep -v internal | grep -v '\.' | grep -v unsafe | grep -v syscall

// go:generate yaegi-extract archive/tar archive/zip
// go:generate yaegi-extract bufio bytes
// go:generate yaegi-extract compress/bzip2 compress/flate compress/gzip compress/lzw compress/zlib
// go:generate yaegi-extract container/heap container/list container/ring
// go:generate yaegi-extract context crypto crypto/aes crypto/cipher crypto/des crypto/dsa crypto/ecdsa
// go:generate yaegi-extract crypto/ed25519 crypto/elliptic crypto/hmac crypto/md5 crypto/rand
// go:generate yaegi-extract crypto/rc4 crypto/rsa crypto/sha1 crypto/sha256 crypto/sha512
// go:generate yaegi-extract crypto/subtle crypto/tls crypto/x509 crypto/x509/pkix
// go:generate yaegi-extract database/sql database/sql/driver
// go:generate yaegi-extract debug/dwarf debug/elf debug/gosym debug/macho debug/pe debug/plan9obj
// go:generate yaegi-extract encoding encoding/ascii85 encoding/asn1 encoding/base32
// go:generate yaegi-extract encoding/base64 encoding/binary encoding/csv encoding/gob
// go:generate yaegi-extract encoding/hex encoding/json encoding/pem encoding/xml
// go:generate yaegi-extract errors expvar flag fmt
// go:generate yaegi-extract go/ast go/build go/constant go/doc go/format go/importer
// go:generate yaegi-extract go/parser go/printer go/scanner go/token go/types
// go:generate yaegi-extract hash hash/adler32 hash/crc32 hash/crc64 hash/fnv hash/maphash
// go:generate yaegi-extract html html/template
// go:generate yaegi-extract image image/color image/color/palette
// go:generate yaegi-extract image/draw image/gif image/jpeg image/png index/suffixarray
// go:generate yaegi-extract io io/ioutil log log/syslog
// go:generate yaegi-extract math math/big math/bits math/cmplx math/rand
// go:generate yaegi-extract mime mime/multipart mime/quotedprintable
// go:generate yaegi-extract net net/http net/http/cgi net/http/cookiejar net/http/fcgi
// go:generate yaegi-extract net/http/httptest net/http/httptrace net/http/httputil net/http/pprof
// go:generate yaegi-extract net/mail net/rpc net/rpc/jsonrpc net/smtp net/textproto net/url
// go:generate yaegi-extract os os/exec os/signal os/user
// go:generate yaegi-extract path path/filepath reflect regexp regexp/syntax
// go:generate yaegi-extract runtime runtime/debug runtime/pprof runtime/trace
// go:generate yaegi-extract sort strconv strings sync sync/atomic
// go:generate yaegi-extract text/scanner text/tabwriter text/template text/template/parse
// go:generate yaegi-extract time unicode unicode/utf16 unicode/utf8
