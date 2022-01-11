# k3s-dump-bootstrap

Decrypts and dumps K3s bootstrap data read from stdin.

Note: `<token>` parameter should be just the bare passphrase, not a full K10-format token including the cluster CA hash.

Example usage (after running `go build`):
```bash
mysql --host=dbhost --user=root --password=password --silent --skip-column-names k3s -e 'SELECT value FROM kine WHERE name LIKE "/bootstrap/%" LIMIT 1' | ./k3s-dump-bootstrap token
```
