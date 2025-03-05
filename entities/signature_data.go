package entities

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SignatureData struct {
	Type      PubKeyType `json:"ty,omitempty"`
	PublicKey PublicKeyString     `json:"pubK,omitempty"`
	Signature HexString     `json:"sig"`
}

func (sD SignatureData) GormDataType() string {
	return "json"
}
func (sD SignatureData) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	asJson, _ := json.Marshal(sD)
	return clause.Expr{
		SQL:  "?",
		Vars: []interface{}{asJson},
	}
}

func (sD *SignatureData) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Value not instance of string:", value))
	}

	result := SignatureData{}
	err := json.Unmarshal(data, &result)
	*sD = SignatureData(result)
	return err
}

// Value return json value, implement driver.Valuer interface
func (sD *SignatureData) Value() (driver.Value, error) {
	if len(sD.Signature) == 0 {
		return nil, nil
	}
	b, _ := json.Marshal(sD)
	return string(b), nil
}
