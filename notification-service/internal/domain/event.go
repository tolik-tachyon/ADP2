package domain

type Event struct {
	EventID       string
	CustomerEmail string
	Amount        int64
	Status        string
	OrderID       string
}
