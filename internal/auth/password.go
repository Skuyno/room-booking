package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword возвращает bcrypt-хэш пароля с дефолтной cost. Хэш
// самодостаточен: содержит соль и параметры внутри строки, поэтому при
// CheckPassword дополнительная информация не нужна.
func HashPassword(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword сверяет пароль с ранее посчитанным bcrypt-хэшем. Возвращает
// nil при совпадении и bcrypt.ErrMismatchedHashAndPassword при расхождении.
func CheckPassword(hash, raw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw))
}
