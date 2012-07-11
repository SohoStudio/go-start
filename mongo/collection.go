package mongo

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"github.com/ungerik/go-start/errs"
	"github.com/ungerik/go-start/model"
	"github.com/ungerik/go-start/utils"
	"github.com/ungerik/go-start/debug"
)

///////////////////////////////////////////////////////////////////////////////
// NewCollection

func NewCollection(name string, documentPrototype interface{}) *Collection {
	if _, ok := collections[name]; ok {
		panic(fmt.Sprintf("Collection %s already created", name))
	}

	t := reflect.TypeOf(documentPrototype)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	collection := &Collection{Name: name, DocumentType: t}

	collection.Init()
	collections[name] = collection

	return collection
}

///////////////////////////////////////////////////////////////////////////////
// ForeignRef

type ForeignRef struct {
	Collection *Collection
	Selector   string
}

///////////////////////////////////////////////////////////////////////////////
// Collection

/*
Collection represents a MongoDB collection and implements mongo.Query for all
documents in the collection.

Example for creating, modifying and saving a document:

	user := models.Users.NewDocument().(*models.User)

	user.Name.First.Set("Erik")
	user.Name.Last.Set("Unger")

	err := user.Save()

*/
type Collection struct {
	queryBase
	Name          string
	DocumentType  reflect.Type
	mgoCollection *mgo.Collection
	// foreignRefs  []ForeignRef
}

func (self *Collection) Init() {
	self.thisQuery = self
	//self.mgoCollection = database.C(self.Name)
	// self.foreignRefs = []ForeignRef{}
}

func (self *Collection) checkDBConnection() {
	if self == nil {
		panic("mongo.Collection is nil")
	}
	if self.mgoCollection.Database.Session == nil {
		panic("mongo.Collection '" + self.Name + "' not initialized. Have you called mongo.AddCollection(" + self.Name + ")?")
	}
}

func (self *Collection) Selector() string {
	return ""
}

func (self *Collection) bsonSelector() bson.M {
	return bson.M{}
}

func (self *Collection) mongoQuery() (q *mgo.Query, err error) {
	self.checkDBConnection()
	return self.mgoCollection.Find(nil), nil
}

func (self *Collection) subDocumentType(docType reflect.Type, fieldName string, subDocSelectors []string) (reflect.Type, error) {
	if fieldName == "" {
		return docType, errs.Format("Collection '%s', selector '%s': Empty field name", self.Name, strings.Join(subDocSelectors, "."))
	}

	switch docType.Kind() {
	case reflect.Struct:
		bsonName := strings.ToLower(fieldName)
		field := utils.FindFlattenedStructField(docType, MatchBsonField(bsonName))
		if field != nil {
			return field.Type, nil
		}
		return nil, errs.Format("Collection '%s', selector '%s': Struct has no field '%s'", self.Name, strings.Join(subDocSelectors, "."), fieldName)

	case reflect.Array, reflect.Slice:
		_, numberErr := strconv.Atoi(fieldName)
		if numberErr == nil || fieldName == "$" {
			return docType, nil
		}
		return docType.Elem(), nil

	case reflect.Ptr, reflect.Interface:
		return self.subDocumentType(docType.Elem(), fieldName, subDocSelectors)
	}

	return nil, errs.Format("Collection '%s', selector '%s': Can't select sub-document '%s' of type '%s'", self.Name, strings.Join(subDocSelectors, "."), fieldName, docType.String())
}

func (self *Collection) ValidateSelector(subDocSelectors ...string) (err error) {
	return nil // todo remove

	if len(subDocSelectors) == 0 {
		return nil
	}
	docType := self.DocumentType
	for _, selector := range subDocSelectors {
		fields := strings.Split(selector, ".")
		if selector == "" || len(fields) == 0 {
			return errs.Format("Invalid empty selector in '%s'", strings.Join(subDocSelectors, "."))
		}
		for _, field := range fields {
			docType, err = self.subDocumentType(docType, field, subDocSelectors)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// return error?
func (self *Collection) NewDocument(subDocSelectors ...string) interface{} {
	if len(subDocSelectors) > 0 {
		panic("Sub document selectors are not implemented")
	}
	docType := self.DocumentType
	for _, selector := range subDocSelectors {
		for _, field := range strings.Split(selector, ".") {
			var err error
			docType, err = self.subDocumentType(docType, field, subDocSelectors)
			errs.AsPanic(err)
		}
	}
	doc := reflect.New(docType).Interface()
	self.InitDocument(doc)
	return doc
}

func (self *Collection) InitDocument(doc interface{}, subDocSelectors ...string) {
	if len(subDocSelectors) > 0 {
		panic("Sub document selectors are not implemented")
	}
	switch s := doc.(type) {
	case Document:
		if len(subDocSelectors) > 0 {
			panic("Can't initialize mongo.Document with subDocSelectors")
		}
		s.Init(self, doc)

	case SubDocument:
		if len(subDocSelectors) == 0 {
			panic("Need subDocSelectors to initialize mongo.SubDocument")
		}
		selector := strings.Join(subDocSelectors, ".")
		s.Init(self, selector, doc)
	}
}

func (self *Collection) Ref(id bson.ObjectId) Ref {
	return Ref{id, self.Name}
}

func (self *Collection) DocumentWithID(id bson.ObjectId, subDocSelectors ...string) (document interface{}, err error) {
	if len(subDocSelectors) > 0 {
		panic("Sub document selectors are not implemented")
	}
	if id == "" {
		return nil, errs.Format("mongo.Collection '%s': Can't get document with empty id", self.Name)
	}
	if err = self.ValidateSelector(subDocSelectors...); err != nil {
		return nil, err
	}

	self.checkDBConnection()
	document = self.NewDocument(subDocSelectors...)
	q := self.mgoCollection.FindId(id)
	if len(subDocSelectors) == 0 {
		err = q.One(document)
	} else {
		err = q.Select(strings.Join(subDocSelectors, ".")).One(document)
	}
	if err != nil {
		return nil, err
	}
	// document has to be initialized again,
	// because mgo zeros the struct while unmarshalling.
	// Newly created slice elements need to be initialized too
	self.InitDocument(document)
	return document, nil
}

func (self *Collection) TryDocumentWithID(id bson.ObjectId, subDocSelectors ...string) (document interface{}, found bool, err error) {
	if len(subDocSelectors) > 0 {
		panic("Sub document selectors are not implemented")
	}
	if id == "" {
		return nil, false, nil
	}
	document, err = self.DocumentWithID(id, subDocSelectors...)
	if err == mgo.ErrNotFound {
		return nil, false, nil
	}
	return document, err == nil, err
}

func (self *Collection) DocumentWithIDIterator(id bson.ObjectId, subDocSelectors ...string) model.Iterator {
	if len(subDocSelectors) > 0 {
		panic("Sub document selectors are not implemented")
	}
	return model.NewObjectOrErrorIterator(self.DocumentWithID(id, subDocSelectors...))
}

func (self *Collection) TryDocumentWithIDIterator(id bson.ObjectId, subDocSelectors ...string) model.Iterator {
	if len(subDocSelectors) > 0 {
		panic("Sub document selectors are not implemented")
	}
	document, ok, err := self.TryDocumentWithID(id, subDocSelectors...)
	if err != nil {
		return model.NewErrorIterator(err)
	}
	if !ok {
		return model.NewObjectIterator()
	}
	return model.NewObjectIterator(document)
}

func (self *Collection) FilterReferenced(refs []Ref) Query {
	return self.FilterRef("_id", refs...)
}

func (self *Collection) Count() (n int, err error) {
	return self.mgoCollection.Count()
}

// Inserts document regardless if it's already in the collection
// If document has a DocumentBase, the ID will be updated to the
// newly created one
func (self *Collection) Insert(document interface{}) (id bson.ObjectId, err error) {
	self.checkDBConnection()
	// Need to set a valid ID, even if Upsert() returns another ID	
	id = bson.NewObjectId()
	if doc, ok := document.(Document); ok {
		doc.SetObjectId(id)
	}
	change, err := self.mgoCollection.Upsert(bson.M{"_id": id}, document)
	if err != nil {
		return id, err
	}
	id = change.UpsertedId.(bson.ObjectId)
	if doc, ok := document.(Document); ok {
		doc.SetObjectId(id)
	}
	return id, nil
}

func (self *Collection) Update(id bson.ObjectId, document interface{}) (err error) {
	self.checkDBConnection()
	err = self.mgoCollection.Update(bson.M{"_id": id}, document)
	debug.Nop()
	return err
}

func (self *Collection) Remove(ids ...bson.ObjectId) error {
	self.checkDBConnection()
	return self.mgoCollection.Remove(bson.M{"_id": bson.M{"$in": ids}})
}

func (self *Collection) RemoveAllNotIn(ids ...bson.ObjectId) error {
	self.checkDBConnection()
	return self.mgoCollection.Remove(bson.M{"_id": bson.M{"$nin": ids}})
}

//func (self *Collection) EnsureIndex(unique bool, keyNodes ...oldmodel.Node) error {
//	keys := make([]string, len(keyNodes))
//	for i, keyNode := range keyNodes {
//		keys[i] = NodePath(keyNode)
//	}
//	index := mgo.Index{
//		Key:        keys,
//		Unique:     unique,
//		Background: true,
//	}
//	return self.mgoCollection.EnsureIndex(index)
//}
