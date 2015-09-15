package locket

type CellEvent interface {
	EventType() CellEventType
	CellIDs() []string
}

type CellEventType int

const (
	CellEventTypeInvalid CellEventType = iota
	CellDisappeared
)

type CellDisappearedEvent struct {
	IDs []string
}

func (CellDisappearedEvent) EventType() CellEventType {
	return CellDisappeared
}

func (e CellDisappearedEvent) CellIDs() []string {
	return e.IDs
}
