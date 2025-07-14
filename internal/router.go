package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ynnekF/bean-counter/models"
)

var logger *zap.SugaredLogger = CreateLogger()

var users = map[int]models.User{
	0: {Id: 0, FirstName: "Alice", LastName: "A", Paid: 25, Consumed: 20, OrderDesc: "Latte", OrderCost: 5},
	1: {Id: 1, FirstName: "Bob", LastName: "B", Paid: 5, Consumed: 25, OrderDesc: "Espresso", OrderCost: 3},
	2: {Id: 2, FirstName: "Charlie", LastName: "C", Paid: 25, Consumed: 15, OrderDesc: "Mocha", OrderCost: 4},
}

var (
	// mu is used to synchronize access to shared in-memory data structures
	// to prevent race conditions in concurrent handler requests.
	mu sync.Mutex

	// nextID holds the next available unique ID for a new user or run.
	nextID = 1

	// runs stores a slice of all recorded coffee runs, used to track daily activity.
	runs = []models.CoffeeRun{}

	// ledger is a user-friendly, display-ready log of all coffee runs,
	// enriched with payer names and total cost for front-end rendering.
	ledger = []models.LedgerEntry{}

	// currentDay tracks the simulated "day" number in the system.
	// It increments after each confirmed coffee run.
	currentDay = 1
)

// total calculates the sum of all values in a map[int]float64.
// It is used to compute the total cost of coffee runs.
func total(m map[int]float64) float64 {
	var sum float64
	for _, v := range m {
		sum += v
	}
	return sum
}

// CreateUser handles a HTTP POST request to create a new user.
// Response codes:
//   - 201 CREATED if the user is successfully created
//   - 400 BAD REQUEST if the request body is invalid or cannot be parsed
//   - 500 INTERNAL SERVER ERROR if there is an unexpected error during processing
//
// It expects a JSON body matching the models.User struct. On success, it assigns a unique ID to the user,
// stores them in memory, and returns the created user. The user is added to the global users slice, which
// is protected by a mutex for thread safety.
func CreateUser(c *gin.Context) {
	logger.Debug("createUser() called")

	var newUser models.User

	// Bind the JSON body to the newUser struct
	// If binding fails, return a 400 BAD REQUEST error
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	mu.Lock()

	// Check if the nextID already exists in the users slice
	if len(users) > nextID {
		logger.Warn("nextID is not incremented, user with this ID already exists")

		nextID = len(users)
	}

	// Assign the next available ID to the new user
	newUser.Id = nextID
	nextID++

	users[newUser.Id] = newUser
	mu.Unlock()

	// Log the creation of the new user and return the user data
	logger.Info("user created: ", newUser.FirstName, newUser.LastName, " with ID: ", newUser.Id)
	c.JSON(http.StatusCreated, newUser)
}

// DeleteUser removes a user from the system based on their ID.
// Response codes:
//   - 200 OK if the user is successfully deleted
//   - 400 BAD REQUEST if the ID is missing or invalid
//   - 404 NOT FOUND if the user with the specified ID does not exist
//   - 500 INTERNAL SERVER ERROR if there is an unexpected error during processing
//
// On success it deletes the user from the global users map
func DeleteUser(c *gin.Context) {
	rawId := c.Query("id")

	logger.Info("deleteUser() called with id: ", rawId)
	if rawId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user id"})
		return
	}

	// Convert the ID from string to int
	id, err := strconv.Atoi(rawId)
	if err != nil {
		logger.Error("error converting string to int:", err)

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if _, ok := users[id]; !ok {
		logger.Warn("user with id: ", id, " not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	delete(users, id)
	logger.Info("user with ID: ", id, " deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// ResetUser resets a user's paid and consumed amounts to zero.
// Response codes:
//   - 204 NO CONTENT if the user is successfully reset
//   - 400 BAD REQUEST if the ID is missing or invalid
//   - 404 NOT FOUND if the user with the specified ID does not exist
//   - 500 INTERNAL SERVER ERROR if there is an unexpected error during processing
//
// It expects a query parameter "id" to identify the user.
// On success it resets the user's Paid and Consumed fields to zero.
func ResetUser(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	user, exists := users[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Reset the user's Paid and Consumed amounts to zero
	user.Paid = 0
	user.Consumed = 0
	users[id] = user

	c.Status(http.StatusNoContent)
}

// UpdateUserField updates a specific field of a user based on the provided ID and field name.
// Response codes:
//   - 204 NO CONTENT if the user field is successfully updated
//   - 400 BAD REQUEST if the request body is invalid or the field is not recognized
//   - 404 NOT FOUND if the user with the specified ID does not exist
//   - 500 INTERNAL SERVER ERROR if there is an unexpected error during processing
//
// It expects a JSON body with the user ID, field name, and new value.
func UpdateUserField(c *gin.Context) {
	var req struct {
		Id    int    `json:"id"`
		Field string `json:"field"`
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	user, exists := users[req.Id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	switch req.Field {
	case "firstName":
		user.FirstName = req.Value
	case "lastName":
		user.LastName = req.Value
	case "paid":
		if f, err := strconv.ParseFloat(req.Value, 64); err == nil {
			user.Paid = f
		}
	case "consumed":
		if f, err := strconv.ParseFloat(req.Value, 64); err == nil {
			user.Consumed = f
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid field"})
		return
	}

	users[req.Id] = user
	c.Status(http.StatusNoContent)
}

// WhoPays determines which user should pay next based on their balance
// Response codes:
//   - 200 OK if a user is found to pay next
//   - 400 BAD REQUEST if there are no users available
//   - 500 INTERNAL SERVER ERROR if there is an unexpected error during processing
//
// It sorts the users by their balance (Paid - Consumed) and returns the user with the lowest balance.
func WhoPays(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	if len(users) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no users available"})
		return
	}

	// Sort users by their balance (Paid - Consumed)
	// The user with the lowest balance will be selected to pay next
	logger.Debug("whoPaysNext() called, sorting users by balance")

	// Convert map to slice
	userList := make([]models.User, 0, len(users))
	for _, u := range users {
		userList = append(userList, u)
	}

	// Sort the slice
	sort.Slice(userList, func(i, j int) bool {
		return (userList[i].Paid - userList[i].Consumed) < (userList[j].Paid - userList[j].Consumed)
	})

	fmt.Println("Sorted users by balance: ", userList)

	// Return the user with the lowest balance
	lowest := userList[0]
	c.JSON(http.StatusOK, gin.H{
		"firstName": lowest.FirstName,
		"lastName":  lowest.LastName,
		"balance":   lowest.Paid - lowest.Consumed,
		"userId":    lowest.Id,
	})
}

// GetBalances returns the current balances of all users
// Response codes:
//   - 200 OK with the list of users and their balances
//   - 500 INTERNAL SERVER ERROR if there is an unexpected error during processing
//
// It locks the mutex to ensure thread safety while accessing the shared users slice.
// The balances are returned as a JSON response with an indented format for readability.
func GetBalances(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	logger.Info("getBalances() called")
	c.IndentedJSON(http.StatusOK, users)
}

// GetCurrentDay returns the current simulated day number
// Response codes:
//   - 200 OK with the current day number
//   - 500 INTERNAL SERVER ERROR if there is an unexpected error during processing
func GetCurrentDay(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	logger.Info("getCurrentDay() called")
	c.JSON(http.StatusOK, gin.H{"currentDay": currentDay})
}

// RecordRun records a coffee run, updating user balances and storing the run details
// Response codes:
//   - 201 CREATED if the run is successfully recorded
//   - 400 BAD REQUEST if the run data is invalid or a user does not exist
//   - 500 INTERNAL SERVER ERROR if there is an unexpected error during processing
//
// It expects a JSON body with the run details, including the payer ID and orders.
func RecordRun(c *gin.Context) {
	logger.Debug("recordRun() called")
	var run models.CoffeeRun
	if err := json.NewDecoder(c.Request.Body).Decode(&run); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run data"})
		return
	}

	if _, ok := users[run.PayerId]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("user with ID %d does not exist", run.PayerId)})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	runTotal := total(run.Orders)
	// Update user Paid/Consumed values
	for userId, cost := range run.Orders {
		user, exists := users[userId]

		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("user with ID %d does not exist", userId)})
			return
		}

		user.Consumed += cost

		if run.PayerId == userId {
			user.Paid += runTotal
		}

		users[userId] = user
	}

	run.Day = currentDay
	currentDay++
	runs = append(runs, run)

	var ledgerEntry models.LedgerEntry
	ledgerEntry.Id = len(runs) - 1
	ledgerEntry.Day = run.Day
	ledgerEntry.PayerId = run.PayerId
	ledgerEntry.PayerName = users[run.PayerId].FirstName + " " + users[run.PayerId].LastName
	ledgerEntry.TotalCost = runTotal
	ledgerEntry.Orders = run.Orders

	ledger = append(ledger, ledgerEntry)
	c.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

// GetLedger returns the ledger of all coffee runs
// Response codes:
//   - 200 OK with the ledger entries
//   - 500 INTERNAL SERVER ERROR if there is an unexpected error during processing
func GetLedger(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	c.JSON(http.StatusOK, ledger)
}
