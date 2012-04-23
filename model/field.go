package model

type Field interface {
	IsDefault() bool
	SetDefault()
	String() string
	SetString(value string) ValidationErrors
	Validator
}
