package models

// User represents an individual who participates in the coffee runs.
// It tracks their payment and consumption history along with order preferences.
type User struct {
	Id        int     `json:"id"`                  // Unique identifier for the user
	FirstName string  `json:"firstName"`           // User's first name (used for display)
	LastName  string  `json:"lastName"`            // User's last name (used for display)
	Paid      float64 `json:"paid"`                // Total amount the user has paid
	Consumed  float64 `json:"consumed"`            // Total value of coffee consumed by the user
	OrderDesc string  `json:"orderDesc,omitempty"` // Description of user's preferred coffee order (optional)
	OrderCost float64 `json:"orderCost,omitempty"` // Default cost of the user's usual order (optional)
}
