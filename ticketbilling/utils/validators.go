package utils

import (
	"regexp"
	"strings"
)

// Normalize limpia espacios y convierte a mayúsculas
func Normalize(input string) string {
	return strings.ToUpper(strings.TrimSpace(input))
}
func Space(input string) string {
	return strings.TrimSpace(input)
}

// HasSpecialChars devuelve true si contiene caracteres que NO sean letras, números o guiones
func HasSpecialChars(text string) bool {
	regex := regexp.MustCompile(`[^A-Z0-9\-]`)
	return regex.MatchString(text)
}

// IsValidEmail valida el formato básico de un correo
func IsValidEmail(email string) bool {
	regex := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	return regex.MatchString(email)
}

// IsValidRFC valida RFC de México (Física: 13 chars, Moral: 12 chars)
// Estructura: [Letras]{3,4} [Fecha: YYMMDD] [Homoclave: 3]

func IsValidRFC(rfc string) bool {
	regex := regexp.MustCompile(`^[A-ZÑ&]{3,4}[0-9]{6}[A-Z0-9]{3}$`)
	return regex.MatchString(rfc)
}
