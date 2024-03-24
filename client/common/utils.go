package common

import "fmt"

const DELIMITER = ";"
const BET_MSG_FORMAT = "{PlayerName:%s,PlayerSurname:%s,PlayerDocID:%d,PlayerDateOfBirth:%s,WageredNumber:%d,AgencyID:%d}"

type Bet struct {
	PlayerName        string
	PlayerSurname     string
	PlayerDocID       int
	PlayerDateOfBirth string
	WageredNumber     int
	AgencyID          int
}

func (bet *Bet) ToString() string {
	return fmt.Sprintf(BET_MSG_FORMAT,
		bet.PlayerName,
		bet.PlayerSurname,
		bet.PlayerDocID,
		bet.PlayerDateOfBirth,
		bet.WageredNumber,
		bet.AgencyID,
	)
}
