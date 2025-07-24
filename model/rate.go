package model

const (
	getRateQuery = `
SELECT rate_id, currency_from, currency_to, rate, ctime from rates
WHERE currency_from = $1 AND currency_to = $2 ORDER BY ctime DESC`
)

// swagger:model
type Rate struct {
	RateID       string  `db:"rate_id"`
	CurrencyFrom string  `db:"currency_from"`
	CurrencyTo   string  `db:"currency_to"`
	Rate         float64 `db:"rate"`
	CTime        string  `db:"ctime"`
}

type Rates struct {
	RubToIls Rate
	IlsToRub Rate
}

func GetCurrentRates(db DBWrapper) (Rates, error) {
	var rates Rates
	err := db.Get(&rates.RubToIls, getRateQuery, "RUB", "ILS")
	if err != nil {
		return rates, err
	}
	err = db.Get(&rates.IlsToRub, getRateQuery, "ILS", "RUB")
	return rates, err
}
