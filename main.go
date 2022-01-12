package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"

	"github.com/brandond/k3s-dump-bootstrap/pkg/bootstrap"
	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := decryptBootstrap(ctx, os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func decryptBootstrap(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return errors.New("exactly one argument required: <token>")
	}
	token := args[1]

	ciphertext, err := io.ReadAll(os.Stdin)
	if err != nil {
		return errors.Wrap(err, "failed to read ciphertext")
	}

	rawData, err := decrypt(token, ciphertext)
	if err != nil {
		return errors.Wrap(err, "failed to decrypt ciphertext")
	}

	buf := bytes.NewReader(rawData)
	files := make(bootstrap.PathsDataformat)
	if !isMigrated(buf, &files) {
		if err := migrateBootstrapData(buf, files); err != nil {
			return errors.Wrap(err, "failed to migrate bootstrap data")
		}
	}

	for pathKey, fileData := range files {
		digest := sha256.Sum256(fileData.Content)
		hexDigest := hex.EncodeToString(digest[:])
		fmt.Printf("\n\n%s\t%s\t%v\n", hexDigest, pathKey, fileData.Timestamp)
		binary.Write(os.Stdout, binary.LittleEndian, fileData.Content)
	}

	return nil
}

func decrypt(passphrase string, ciphertext []byte) ([]byte, error) {
	parts := strings.SplitN(string(ciphertext), ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cipher text, not : delimited")
	}

	clearKey := pbkdf2.Key([]byte(passphrase), []byte(parts[0]), 4096, 32, sha1.New)
	key, err := aes.NewCipher(clearKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(key)
	if err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], nil)
}

func isMigrated(buf io.ReadSeeker, files *bootstrap.PathsDataformat) bool {
	buf.Seek(0, 0)
	defer buf.Seek(0, 0)

	if err := json.NewDecoder(buf).Decode(files); err != nil {
		// This will fail if data is being pulled from old an cluster since
		// older clusters used a map[string][]byte for the data structure.
		// Therefore, we need to perform a migration to the newer bootstrap
		// format; bootstrap.BootstrapFile.
		return false
	}

	return true
}

func migrateBootstrapData(data io.Reader, files bootstrap.PathsDataformat) error {
	var oldBootstrapData map[string][]byte
	if err := json.NewDecoder(data).Decode(&oldBootstrapData); err != nil {
		// if this errors here, we can assume that the error being thrown
		// is not related to needing to perform a migration.
		return err
	}

	// iterate through the old bootstrap data structure
	// and copy into the new bootstrap data structure
	for k, v := range oldBootstrapData {
		files[k] = bootstrap.File{
			Content: v,
		}
	}

	return nil
}
