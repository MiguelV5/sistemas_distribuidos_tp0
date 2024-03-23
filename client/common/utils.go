package common

import "fmt"

type Bet struct {
	PlayerName        string
	PlayerSurname     string
	PlayerDocID       int
	PlayerDateOfBirth string
	WageredNumber     int
	AgencyID          int
}

func (bet *Bet) ToString() string {
	return fmt.Sprintf("{PlayerName:%s,PlayerSurname:%s,PlayerDocID:%d,PlayerDateOfBirth:%s,WageredNumber:%d,AgencyID:%d};",
		bet.PlayerName,
		bet.PlayerSurname,
		bet.PlayerDocID,
		bet.PlayerDateOfBirth,
		bet.WageredNumber,
		bet.AgencyID,
	)
}
