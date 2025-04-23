package model

import (
	"github.com/lestrrat-go/jwx/v2/jwe"
)

type Receipt struct {
	Receipt string `json:"receipt"`
}

func (r *Receipt) CreateReceipt(msg *jwe.Message) *Receipt {
	if msg != nil {

		receipt, err := jwe.Compact(msg)

		if err == nil {
			return &Receipt{
				Receipt: string(receipt),
			}
		}
	}
	return nil
}
