package auth

import (
	"errors"

	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func GerarJWT(id int32) (string, error) {


	secret := os.Getenv("JWT_SECRET")

	if secret  == ""{
		return  "", errors.New("JWT nao configurado")
	}

	claim:= jwt.MapClaims{

		"sub": id, //id do usuario
		"exp": time.Now().Add(time.Hour * 24).Unix(), // 24 para o token expirar
		"iat": time.Now().Unix(), // data da criação
	}



	token:= jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	return  token.SignedString([]byte(secret))


}
