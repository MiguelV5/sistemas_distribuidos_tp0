package common

import (
	"encoding/csv"
	"fmt"
	"strconv"
)

const DELIMITER = ";"
const BET_MSG_FORMAT = "{PlayerName:%s,PlayerSurname:%s,PlayerDocID:%d,PlayerDateOfBirth:%s,WageredNumber:%d,AgencyID:%d}"
const CHUNK_ACK_MSG_FORMAT_FROM_SV = "ACK_CHUNK;"

const DEFAULT_BETS_PER_CHUNK = 10
const KiB = 1024
const MAX_CHUNK_SIZE_PER_MSG = 8 * KiB

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

func ReadBetFromCsvRecord(agencyFileReader *csv.Reader, rawAgencyID string) (Bet, error) {
	record, err := agencyFileReader.Read()
	if err != nil {
		return Bet{}, err
	}

	playerDocID, err := strconv.Atoi(record[2])
	if err != nil {
		return Bet{}, err
	}
	wageredNumber, err := strconv.Atoi(record[4])
	if err != nil {
		return Bet{}, err
	}
	agencyID, err := strconv.Atoi(rawAgencyID)
	if err != nil {
		return Bet{}, err
	}

	recordAsBet := Bet{
		PlayerName:        record[0],
		PlayerSurname:     record[1],
		PlayerDocID:       playerDocID,
		PlayerDateOfBirth: record[3],
		WageredNumber:     wageredNumber,
		AgencyID:          agencyID,
	}

	return recordAsBet, nil

}
