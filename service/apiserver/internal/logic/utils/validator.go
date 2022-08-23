package utils

const (
	minSymbolLength = 3
)

func ValidateSymbol(symbol string) bool {
	return len(symbol) >= minSymbolLength
}
