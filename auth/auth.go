package auth

import (

	"github.com/alexedwards/argon2id"
	
)


func HashPassword(senha string) ([]byte, error) {
	

	hashSenha, err:= argon2id.CreateHash(senha,argon2id.DefaultParams)
	if err != nil {
		return  nil, err
	}

	return []byte(hashSenha), nil

}

// funcao que ira servir para comparar a senha que o usuario digitar, com a que esta salva no banco de dados
func HashCompare(hash []byte, senha string) (bool, error) {

	ok, err := argon2id.ComparePasswordAndHash(senha, string(hash))
	if err != nil {
		return false, err
	}

	return ok, nil
}
