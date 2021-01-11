// Package stdlib provides wrappers of standard library packages to be imported natively in Yaegi.
package stdlib

import "reflect"

// Symbols variable stores the map of stdlib symbols per package
var Symbols = map[string]map[string]reflect.Value{}

func init() {
	Symbols["github.com/containous/yaegi/stdlib"] = map[string]reflect.Value{
		"Symbols": reflect.ValueOf(Symbols),
	}
}

// Provide access to go standard library (http://golang.org/pkg/)
// go list std | grep -v internal | grep -v '\.' | grep -v unsafe | grep -v syscall

//go:generate goexports archive/tar archive/zip
//go:generate goexports bufio bytes
//go:generate goexports compress/bzip2 compress/flate compress/gzip compress/lzw compress/zlib
//go:generate goexports container/heap container/list container/ring
//go:generate goexports context crypto crypto/aes crypto/cipher crypto/des crypto/dsa crypto/ecdsa
//go:generate goexports crypto/ed25519 crypto/elliptic crypto/hmac crypto/md5 crypto/rand
//go:generate goexports crypto/rc4 crypto/rsa crypto/sha1 crypto/sha256 crypto/sha512
//go:generate goexports crypto/subtle crypto/tls crypto/x509 crypto/x509/pkix
//go:generate goexports database/sql database/sql/driver
//go:generate goexports debug/dwarf debug/elf debug/gosym debug/macho debug/pe debug/plan9obj
//go:generate goexports encoding encoding/ascii85 encoding/asn1 encoding/base32
//go:generate goexports encoding/base64 encoding/binary encoding/csv encoding/gob
//go:generate goexports encoding/hex encoding/json encoding/pem encoding/xml
//go:generate goexports errors expvar flag fmt
//go:generate goexports go/ast go/build go/constant go/doc go/format go/importer
//go:generate goexports go/parser go/printer go/scanner go/token go/types
//go:generate goexports hash hash/adler32 hash/crc32 hash/crc64 hash/fnv hash/maphash
//go:generate goexports html html/template
//go:generate goexports image image/color image/color/palette
//go:generate goexports image/draw image/gif image/jpeg image/png index/suffixarray
//go:generate goexports io io/ioutil log log/syslog
//go:generate goexports math math/big math/bits math/cmplx math/rand
//go:generate goexports mime mime/multipart mime/quotedprintable
//go:generate goexports net net/http net/http/cgi net/http/cookiejar net/http/fcgi
//go:generate goexports net/http/httptest net/http/httptrace net/http/httputil net/http/pprof
//go:generate goexports net/mail net/rpc net/rpc/jsonrpc net/smtp net/textproto net/url
//go:generate goexports os os/exec os/signal os/user
//go:generate goexports path path/filepath reflect regexp regexp/syntax
//go:generate goexports runtime runtime/debug runtime/pprof runtime/trace
//go:generate goexports sort strconv strings sync sync/atomic
//go:generate goexports text/scanner text/tabwriter text/template text/template/parse
//go:generate goexports time unicode unicode/utf16 unicode/utf8
