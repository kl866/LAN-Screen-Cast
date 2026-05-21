package session

type Sender interface {
	ID() string
	Activate()
	Deactivate()
	IsActive() bool
}
