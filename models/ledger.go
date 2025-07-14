package models

// CoffeeRun represents a single day's coffee run.
// It tracks who paid and the cost of each user's order for that day.
type CoffeeRun struct {
	Id      int             `json:"id"`      // Unique identifier for the coffee run
	Day     int             `json:"day"`     // Sequential day number of the run
	PayerId int             `json:"payerId"` // ID of the user who paid for this run
	Orders  map[int]float64 `json:"orders"`  // Mapping of user ID to the cost of their drink (userId → drink cost)
}

// LedgerEntry represents a recorded coffee run for reporting purposes.
// It contains both user IDs and resolved names for display, along with total cost.
type LedgerEntry struct {
	Id        int             `json:"id"`        // Unique identifier for the ledger entry
	Day       int             `json:"day"`       // Sequential day number of the run
	PayerId   int             `json:"payerId"`   // ID of the user who paid
	PayerName string          `json:"payerName"` // Full name of the payer (for display purposes)
	TotalCost float64         `json:"totalCost"` // Total amount spent on this coffee run
	Orders    map[int]float64 `json:"orders"`    // Mapping of user ID to the cost of their drink (userId → drink cost)
}
