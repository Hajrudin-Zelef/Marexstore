// cmd/hashpass : génère un hash bcrypt (cost 12) pour ADMIN_PASS_HASH.
// Usage : printf '%s' 'mot-de-passe' | go run ./cmd/hashpass
// Le mot de passe ne doit jamais figurer en argument (visible via ps),
// ni être commité : seul le hash va dans le .env (non versionné).
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	var pw string
	if len(os.Args) > 1 {
		log.Fatal("refusé : passez le mot de passe via stdin, pas en argument")
	}
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	pw = strings.TrimRight(string(b), "\r\n")
	if pw == "" {
		log.Fatal("mot de passe vide sur stdin")
	}
	h, err := bcrypt.GenerateFromPassword([]byte(pw), 12)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(h))
}
