package model

// Password represents a randomly generated password.
type Password struct {
	// Diceware is a password generated using the Diceware method.
	Diceware string `json:"diceware,omitempty"`

	// Random is a password generated using a cryptographically secure random number generator.
	Random string `json:"random,omitempty"`

	// PIN is a password generated using a cryptographically secure random
	// number generator, but with only digits.
	PIN string `json:"pin,omitempty"`
}

// NewDicewarePassword creates a new password using the Diceware method.
func NewDicewarePassword(diceware string) *Password {
	return &Password{
		Diceware: diceware,
	}
}

// NewRandomPassword creates a new password using a cryptographically secure
// random number generator.
func NewRandomPassword(random string) *Password {
	return &Password{
		Random: random,
	}
}

// NewPIN creates a new password using a cryptographically secure random
// number generator, but with only digits.
func NewPIN(pin string) *Password {
	return &Password{
		PIN: pin,
	}
}
