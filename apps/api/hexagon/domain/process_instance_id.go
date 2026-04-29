package domain

// ProcessInstanceID identifiziert eine `apps/api`-Prozessinstanz
// (plan-0.1.0.md §5.1). Ein Cursor kapselt die ID, mit der er erzeugt
// wurde; weicht sie beim Folge-Request von der aktuellen Prozess-ID ab,
// liefert der Use Case ErrCursorInvalid — der Client muss vom Anfang
// neu paginieren.
//
// Erzeugung erfolgt in main.go via crypto/rand; das hexagon-Package
// hält bewusst keinen Konstruktor, damit Domain-Typen rein bleiben.
type ProcessInstanceID string
