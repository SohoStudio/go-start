package model

import (
	"bytes"

// "fmt"
// "github.com/ungerik/go-start/errs"
)

type ValidationError struct {
	Description string
	MetaData    *MetaData
}

func (self *ValidationError) Error() string {
	if self.MetaData == nil {
		return fmt.Sprintf("ValidationError: %s", self.Description)
	}
	return fmt.Sprintf("ValidationError %s: %s", self.MetaData.Selector(), self.Description)
}

type ValidationErrors []ValidationError

func (self ValidationErrors) Error() string {
	var buf bytes.Buffer
	for i := range self {
		buf.WriteString(self[i].Error())
		buf.WriteByte('\n')
	}
	return buf.String()
}

func Err(description string, metaData *MetaData) ValidationErrors {
	return ValidationErrors{{Description: description, MetaData: metaData}}
}
