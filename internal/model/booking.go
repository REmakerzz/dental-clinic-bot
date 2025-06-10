package model

type Booking struct {
    ID         int
    Name       string
    Phone      string
    Service    string
    DateTime   string
    Step       int // номер шага сценария
}
