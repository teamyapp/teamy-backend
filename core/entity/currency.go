package entity

type Currency string

const (
	CurrencyUSD Currency = "USD"
)

var StringToCurrency = map[string]Currency{
	"USD": CurrencyUSD,
}
