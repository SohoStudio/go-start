package mongo

import (
	"fmt"
	"github.com/ungerik/go-start/errs"
	"github.com/ungerik/go-start/model"
	"github.com/ungerik/go-start/mgo/bson"

	
)

/*
Ref holds a mongo.ID and the name of the collection of the referenced document.
Only the ID will be saved in MongoDB, the collection name is for validation
and convenience functions only.
*/
type Ref struct {
	ID             bson.ObjectId `gostart:"-"`
	CollectionName string        `gostart:"-"`
}

func (self *Ref) String() string {
	return self.ID.Hex()
}

func (self *Ref) SetString(str string) error {
	switch len(str) {
	case 0, 12:
		self.ID = bson.ObjectId(str)
	case 24:
		self.ID = bson.ObjectIdHex(str)
	default:
		return errs.Format("Invalid string for bson.ObjectId: '%s'", str)
	}
	return nil
}

// Implements bson.Getter
func (self Ref) GetBSON() (interface{}, error) {
	if !self.ID.Valid() {
		return nil, nil
	}
	return self.ID, nil
}

// Implements bson.Setter
func (self *Ref) SetBSON(raw bson.Raw) error {
	var id *bson.ObjectId
	err := raw.Unmarshal(&id)
	if err != nil {
		return err
	}
	if id == nil {
		self.ID = ""
	} else {
		self.ID = *id
	}
	return nil
}

func (self *Ref) Validate(metaData *model.MetaData) error {
	if self.CollectionName == "" {
		return errs.Format("Missing CollectionName")
	}
	length := len(self.ID)
	if length != 0 && length != 12 {
		return errs.Format("Invalid ObjectId length %d", length)
	}
	if self.Required(metaData) && self.IsEmpty() {
		return model.NewRequiredError(metaData)
	}
	return nil
}

func (self *Ref) Required(metaData *model.MetaData) bool {
	return metaData.BoolAttrib(model.StructTagKey, "required")
}

// Dummy function to implement model.Reference
func (self *Ref) Reference() {
}

func (self *Ref) Collection() (collection *Collection, err error) {
	if self.CollectionName == "" {
		return nil, errs.Format("Missing collection name. Did you call mongo.Document.Init()?")
	}
	collection, ok := collections[self.CollectionName]
	if !ok {
		return nil, errs.Format("Collection '" + self.CollectionName + "' not registered")
	}
	return collection, nil
}

func (self *Ref) Get() (doc interface{}, err error) {
	collection, err := self.Collection()
	if err != nil {
		return nil, err
	}
	return collection.DocumentWithID(self.ID)
}

func (self *Ref) TryGet() (doc interface{}, ok bool, err error) {
	collection, err := self.Collection()
	if err != nil {
		return nil, false, err
	}
	return collection.TryDocumentWithID(self.ID)
}

func (self *Ref) Set(document Document) {
	if document.Collection() == nil {
		panic("model.Document.Collection() == nil")
	}
	if document.Collection().Name != self.CollectionName {
		panic(fmt.Sprintf("Can't set document from different collection. Expected collection '%s', got '%s'", self.CollectionName, document.Collection().Name))
	}
	self.ID = document.ObjectId()
}

func (self *Ref) IsEmpty() bool {
	return self.ID == ""
}

func (self *Ref) SetEmpty() {
	self.ID = ""
}

func (self *Ref) References(doc Document) (ok bool, err error) {
	collection := doc.Collection()
	if collection == nil {
		return false, errs.Format("Document is not initialized")
	}
	if collection.Name != self.CollectionName {
		return false, errs.Format("mongo.Ref to collection '%s' can't reference document of collection '%s'", self.CollectionName, collection.Name)
	}
	id := doc.ObjectId()
	if !id.Valid() {
		return false, errs.Format("Can't reference document with empty ID")
	}
	return self.ID == id, nil
}
