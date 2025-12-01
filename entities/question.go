type Question struct {
	ID            int // AUTOINCREMENT
	SessionID     string
	Text          string
	OrderNum      int
	AllowMultiple bool
	MaxChoices    int
	CreatedAt     Timestamp
}