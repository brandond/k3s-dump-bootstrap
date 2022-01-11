# k3s-dump-bootstrap

Decrypts and dumps K3s bootstrap data read from stdin.

Note: `<token>` parameter should be just the bare passphrase, not a full K10-format token including the cluster CA hash.

Example usage:
1. `go get github.com/brandond/k3s-dump-bootstrap`
2. `mysql --host=dbhost --user=root --password=password --silent --skip-column-names k3s -e 'SELECT CONVERT(value USING utf8) FROM kine WHERE deleted=0 AND name LIKE "/bootstrap/%" ORDER BY id DESC LIMIT 1' | k3s-dump-bootstrap token`
