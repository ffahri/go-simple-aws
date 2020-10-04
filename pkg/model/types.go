package model

import (
	"encoding/json"
	"errors"
)

type TestModel struct {
	Id    string
	Name  string
	Value int
}

func MarshallModel(model TestModel) ([]byte, error) {
	arr, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	return arr, nil
}
func UnmarshallModel(model *string) (TestModel, error) {
	mdl := TestModel{}
	if err := json.Unmarshal([]byte(*model), &mdl); err != nil {
		return mdl, err
	}
	return mdl, nil
}
func CheckModelFields(model TestModel) error {
	if model.Id == "" && model.Name == "" {
		return errors.New("id & name cannot be blank")
	}
	return nil
}
