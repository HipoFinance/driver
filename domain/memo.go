package domain

import "encoding/json"

type Memorable interface {
	ToJson() string
	FromJson(jstr string) error
}

type Memo struct {
	Key  string `json:"key"`
	Memo string `json:"memo"`
}

type ExtractionMemo struct {
	LatestProcessedHash string `json:"latest_processed_hash"`
}

func (obj *ExtractionMemo) ToJson() string {
	jstr, err := json.Marshal(obj)
	if err != nil {
		return err.Error()
	}
	return string(jstr)
}

func (obj *ExtractionMemo) FromJson(jstr string) error {
	err := json.Unmarshal([]byte(jstr), obj)
	return err
}
