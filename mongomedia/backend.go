package mongomedia

import (
	"github.com/ungerik/go-start/media"
	"github.com/ungerik/go-start/mgo"
	"github.com/ungerik/go-start/mgo/bson"
	"github.com/ungerik/go-start/mongo"
	"io"
)

type Backend struct {
	gridFS *mgo.GridFS
	images *mongo.Collection
}

func (self *Backend) LoadImage(id string) (*media.Image, error) {
	doc, err := self.images.DocumentWithID(bson.ObjectIdHex(id))
	if err != nil {
		return nil, err
	}
	image := &doc.(*ImageDoc).Image
	image.Init()
	return image, nil
}

func (self *Backend) SaveImage(image *media.Image) error {
	if image.ID == "" {
		doc := self.images.NewDocument().(*ImageDoc)
		doc.Image = *image
		id, err := self.images.Insert(doc)
		if err != nil {
			return err
		}
		image.ID.Set(id.Hex())
		return nil
	}

	id := bson.ObjectIdHex(image.ID.Get())
	doc := self.images.NewDocument().(*ImageDoc)
	doc.SetObjectId(id)
	doc.Image = *image
	doc.Image.ID = ""
	return self.images.Update(id, doc)
}

func (self *Backend) ImageVersionReader(id string) (reader io.ReadCloser, ctype string, err error) {
	file, err := self.gridFS.OpenId(bson.ObjectIdHex(id))
	if err == mgo.NotFound {
		return nil, "", media.ErrInvalidImageID(id)
	} else if err != nil {
		return nil, "", err
	}
	return file, file.ContentType(), nil
}

func (self *Backend) ImageVersionWriter(version *media.ImageVersion) (writer io.WriteCloser, err error) {
	if version.ID != "" {
		err = self.gridFS.RemoveId(version.ID)
		if err != nil {
			return nil, err
		}
		version.ID = ""
	}
	file, err := self.gridFS.Create("")
	if err != nil {
		return nil, err
	}
	id := file.Id().(bson.ObjectId).Hex()
	file.SetName(id + "/" + version.Filename.Get())
	file.SetMeta(version)
	version.ID.Set(id)
	return file, err
}
