package entity

type Currency string

const (
	CurrencyUSD Currency = "USD"
)

var CurrencyStrToCurrency = map[string]Currency{
	"USD": CurrencyUSD,
}
