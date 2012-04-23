package model

type Validator interface {
	Validate(metaData *MetaData) ValidationErrors
}
