package main

import (
	"flag"
	"log"
	"time"

	"github.com/godyy/ggs/internal/core/auth/jwt"
)

func main() {
	priPemFile := flag.String("pri-pem", "", "private key pem file")
	userInfo := flag.String("user-info", "", "user info of json")
	issuer := flag.String("issuer", "dev", "issuer")
	exp := flag.Duration("exp", 30*time.Minute, "expire time")
	flag.Parse()

	priKey, err := jwt.LoadPrivKey(*priPemFile)
	if err != nil {
		log.Fatal(err)
	}

	signedToken, err := jwt.SignToken(priKey, *issuer, string(*userInfo), *exp, time.Now())
	if err != nil {
		log.Fatal(err)
	}

	log.Println(signedToken)
}
