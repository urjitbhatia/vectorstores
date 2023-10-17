package pgvector

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type dbMap map[string]any

func (d *dbMap) Scan(src any) error {
	var bytes []byte
	switch v := src.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to unmarshal JSONB value from pg: %s", src)
	}
	err := json.Unmarshal(bytes, d)
	return err
}

func (d dbMap) Value() (driver.Value, error) {
	bytes, err := json.Marshal(d)
	return string(bytes), err
}
