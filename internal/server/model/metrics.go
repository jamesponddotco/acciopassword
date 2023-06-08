package model

// Metrics represents the access metrics of a given resource.
type Metrics struct {
	// Random is the number of random passwords generated since the last reset.
	Random uint64 `json:"random"`

	// Diceware is the number of diceware passwords generated since the last reset.
	Diceware uint64 `json:"diceware"`

	// PIN is the number of PINs generated since the last reset.
	PIN uint64 `json:"pin"`

	// Total is the total number of passwords generated since the last reset.
	Total uint64 `json:"total"`
}

// NewMetrics creates a new Metrics instance with each counter set to their given value.
func NewMetrics(random, diceware, pin, total uint64) *Metrics {
	return &Metrics{
		Random:   random,
		Diceware: diceware,
		PIN:      pin,
		Total:    total,
	}
}
