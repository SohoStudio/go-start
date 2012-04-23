package model

import "github.com/ungerik/go-mail"

type Email string

func (self *Email) IsDefault() bool {
	return *self == ""
}

func (self *Email) String() string {
	return string(*self)
}

func (self *Email) SetString(value string) ValidationErrors {
	if value != "" {
		if value, err = email.ValidateAddress(value); err != nil {
			return New(err.Error(), nil)
		}
	}
	*self = Email(value)
	return nil
}

func (self *Email) FixValue(metaData *MetaData) {
}

func (self *Email) Validate(metaData *MetaData) ValidationErrors {
	str := self.Get()
	if self.Required(metaData) || str != "" {
		if _, err := email.ValidateAddress(str); err != nil {
			return Err(err, metaData)
		}
	}
	return nil
}

func (self *Email) Required(metaData *MetaData) bool {
	return metaData.BoolAttrib("required")
}
