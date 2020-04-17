package model

import "errors"

type Trade struct {
	Order      int     `json:"ticket"`
	Login      int     `json:"login"`
	Symbol     string  `json:"symbol"`
	Digits     int     `json:"digits"`
	Cmd        int     `json:"cmd"`
	Volume     int     `json:"volume"`
	OpenTime   int     `json:"open_time"`
	OpenPrice  float64 `json:"open_price"`
	CloseTime  int     `json:"close_time"`
	ClosePrice float64 `json:"close_price"`
	Sl         float64 `json:"sl"`
	Tp         float64 `json:"tp"`
	Comment    string  `json:"comment"`
	Expiration int     `json:"expiration"`
	Profit     float64 `json:"profit"`
	Magic      int     `json:"magic"`
}

var (
	ErrCmdInvalid = errors.New("command value from 0 to 5")
)

type AddTrade struct {
	Command int     `json:"command" example:"0" binding:"required"`
	Symbol  string  `json:"symbol" example:"EURUSD" binding:"required"`
	Volume  float64 `json:"volume" example:"0.1" binding:"required"`
	Price   float64 `json:"price" example:"1.1456" binding:"required"`
	Sl      float64 `json:"sl" example:"1.1456"`
	Tp      float64 `json:"tp" example:"1.1456"`
}

// Validation example
func (a AddTrade) Validation() error {
	switch {
	case a.Command < 0 || a.Command > 5:
		return ErrCmdInvalid
	default:
		return nil
	}
}

type UpdateTrade struct {
	Ticket int     `json:"ticket" example:"101" binding:"required"`
	Price  float64 `json:"price" example:"1.1456"`
	Sl     float64 `json:"sl" example:"1.1456"`
	Tp     float64 `json:"tp" example:"1.1456"`
}

// Validation example
func (a UpdateTrade) Validation() error {
	switch {
	default:
		return nil
	}
}

type CloseTrade struct {
	Ticket int     `json:"ticket" example:"101" binding:"required"`
	Volume float64 `json:"volume" example:"0.1"`
}

// Validation example
func (a CloseTrade) Validation() error {
	switch {
	default:
		return nil
	}
}
