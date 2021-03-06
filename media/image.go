package media

import (
	"bytes"
	_ "code.google.com/p/go.image/bmp"
	_ "code.google.com/p/go.image/tiff"
	"github.com/ungerik/go-start/model"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/png"
	// "github.com/ungerik/go-start/view"
)

type HorAlignment int

const (
	HorCenter HorAlignment = iota
	Left
	Right
)

type VerAlignment int

const (
	VerCenter VerAlignment = iota
	Top
	Bottom
)

// NewImage creates a new Image and saves the original version to Config.Backend.
// GIF, TIFF, BMP images will be read, but written as PNG.
func NewImage(filename string, data []byte) (*Image, error) {
	i, t, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if t == "gif" || t == "bmp" || t == "tiff" {
		var buf bytes.Buffer
		err = png.Encode(&buf, i)
		if err != nil {
			return nil, err
		}
		data = buf.Bytes()
		filename += ".png"
		i, t, err = image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	}
	width := i.Bounds().Dx()
	height := i.Bounds().Dy()
	version := newImageVersion(
		ValidUrlFilename(filename),
		"image/"+t,
		image.Rect(0, 0, width, height),
		width,
		height,
		i.ColorModel() == color.GrayModel || i.ColorModel() == color.Gray16Model,
	)
	err = version.SaveImageData(data)
	if err != nil {
		return nil, err
	}
	image := &Image{Versions: []ImageVersion{version}}
	image.Init()
	return image, nil
}

type Image struct {
	ID          model.String `bson:",omitempty"`
	Description model.String
	Link        model.Url
	Versions    []ImageVersion
}

func (self *Image) Init() {
	for i := range self.Versions {
		self.Versions[i].image = self
	}
}

func (self *Image) Filename() string {
	return self.Versions[0].Filename.Get()
}

func (self *Image) ContentType() string {
	return self.Versions[0].ContentType.Get()
}

func (self *Image) Width() int {
	return self.Versions[0].Width.GetInt()
}

func (self *Image) Height() int {
	return self.Versions[0].Height.GetInt()
}

func (self *Image) Rectangle() image.Rectangle {
	return self.Versions[0].SourceRect.Rectangle()
}

func (self *Image) Grayscale() bool {
	return self.Versions[0].Grayscale.Get()
}

// AspectRatio returns Width / Height
func (self *Image) AspectRatio() float64 {
	return self.Versions[0].AspectRatio()
}

func (self *Image) sourceRectTouchOriginalFromOutside(width, height int, horAlign HorAlignment, verAlign VerAlignment) (r image.Rectangle) {
	var offset image.Point
	aspectRatio := float64(width) / float64(height)
	if aspectRatio > self.AspectRatio() {
		// Wider than original
		// so touchOriginalFromOutside means
		// that the source rect is as high as the original
		r.Max.X = int(float64(self.Height()) * aspectRatio)
		r.Max.Y = self.Height()
		switch horAlign {
		case HorCenter:
			offset.X = (self.Width() - r.Max.X) / 2
		case Right:
			offset.X = self.Width() - r.Max.X
		}
	} else {
		// Heigher than original,
		// so touchOriginalFromOutside means
		// that the source rect is as wide as the original
		r.Max.X = self.Width()
		r.Max.Y = int(float64(self.Width()) / aspectRatio)
		switch verAlign {
		case VerCenter:
			offset.Y = (self.Height() - r.Max.Y) / 2
		case Bottom:
			offset.Y = self.Height() - r.Max.Y
		}
	}
	return r.Add(offset)
}

func (self *Image) sourceRectTouchOriginalFromInside(width, height int, horAlign HorAlignment, verAlign VerAlignment) (r image.Rectangle) {
	var offset image.Point
	aspectRatio := float64(width) / float64(height)
	if aspectRatio > self.AspectRatio() {
		// Wider than original
		// so touchOriginalFromInside means
		// that the source rect is as wide as the original
		r.Max.X = self.Width()
		r.Max.Y = int(float64(self.Width()) / aspectRatio)
		switch verAlign {
		case VerCenter:
			offset.Y = (self.Height() - r.Max.Y) / 2
		case Bottom:
			offset.Y = self.Height() - r.Max.Y
		}
	} else {
		// Heigher than original,
		// so touchOriginalFromInside means
		// that the source rect is as high as the original
		r.Max.X = int(float64(self.Height()) * aspectRatio)
		r.Max.Y = self.Height()
		switch horAlign {
		case HorCenter:
			offset.X = (self.Width() - r.Max.X) / 2
		case Right:
			offset.X = self.Width() - r.Max.X
		}
	}
	return r.Add(offset)
}

// SourceRectVersion searches and returns an existing matching version,
// or a new one will be created and saved.
func (self *Image) VersionSourceRect(sourceRect image.Rectangle, width, height int, grayscale bool, outsideColor color.Color) (im *ImageVersion, err error) {
	if self.Grayscale() {
		grayscale = true // Ignore color requests when original image is grayscale
	}

	// Search for exact match
	for i := range self.Versions {
		v := &self.Versions[i]
		match := v.SourceRect.Rectangle() == sourceRect &&
			v.Width.GetInt() == width &&
			v.Height.GetInt() == height &&
			v.OutsideColor.EqualsColor(outsideColor) &&
			v.Grayscale.Get() == grayscale
		if match {
			return v, nil
		}
	}

	// No exact match, create version
	origImage, err := self.Versions[0].LoadImage()
	if err != nil {
		return nil, err
	}

	var versionImage image.Image
	if sourceRect.In(self.Rectangle()) {
		versionImage = ResizeImage(origImage, sourceRect, width, height)
		if grayscale && !self.Grayscale() {
			var grayVersion image.Image = image.NewGray(versionImage.Bounds())
			draw.Draw(grayVersion.(draw.Image), versionImage.Bounds(), versionImage, image.ZP, draw.Src)
			versionImage = grayVersion
		}
	} else {
		if grayscale {
			versionImage = image.NewGray(image.Rect(0, 0, width, height))
		} else {
			versionImage = image.NewRGBA(image.Rect(0, 0, width, height))
		}
		// Fill version with outsideColor
		draw.Draw(versionImage.(draw.Image), versionImage.Bounds(), image.NewUniform(outsideColor), image.ZP, draw.Src)
		// Where to draw the source image into the version image
		var destRect image.Rectangle
		if !(sourceRect.Min.X < 0 || sourceRect.Min.Y < 0) {
			panic("touching from outside means that sourceRect x or y must be negative")
		}
		sourceW := float64(sourceRect.Dx())
		sourceH := float64(sourceRect.Dy())
		destRect.Min.X = int(float64(-sourceRect.Min.X) / sourceW * float64(width))
		destRect.Min.Y = int(float64(-sourceRect.Min.Y) / sourceH * float64(height))
		destRect.Max.X = destRect.Min.X + int(float64(self.Width())/sourceW*float64(width))
		destRect.Max.Y = destRect.Min.Y + int(float64(self.Height())/sourceH*float64(height))
		destImage := ResizeImage(origImage, origImage.Bounds(), destRect.Dx(), destRect.Dy())
		draw.Draw(versionImage.(draw.Image), destRect, destImage, image.ZP, draw.Src)
	}

	// Save new image version
	self.Versions = append(self.Versions, newImageVersion(self.Filename(), self.ContentType(), sourceRect, width, height, grayscale))
	version := &self.Versions[len(self.Versions)-1]
	err = version.SaveImage(versionImage)
	if err != nil {
		return nil, err
	}
	err = Config.Backend.SaveImage(self)
	if err != nil {
		return nil, err
	}
	return version, nil
}

func (self *Image) Version(width, height int, horAlign HorAlignment, verAlign VerAlignment, grayscale bool) (im *ImageVersion, err error) {
	return self.VersionSourceRect(self.sourceRectTouchOriginalFromInside(width, height, horAlign, verAlign), width, height, grayscale, color.RGBA{})
}

func (self *Image) VersionCentered(width, height int, grayscale bool) (im *ImageVersion, err error) {
	return self.Version(width, height, HorCenter, VerCenter, grayscale)
}

func (self *Image) VersionTouchOrigFromOutside(width, height int, horAlign HorAlignment, verAlign VerAlignment, grayscale bool, outsideColor color.Color) (im *ImageVersion, err error) {
	return self.VersionSourceRect(self.sourceRectTouchOriginalFromOutside(width, height, horAlign, verAlign), width, height, grayscale, outsideColor)
}

func (self *Image) VersionTouchOrigFromOutsideCentered(width, height int, grayscale bool, outsideColor color.Color) (im *ImageVersion, err error) {
	return self.VersionTouchOrigFromOutside(width, height, HorCenter, VerCenter, grayscale, outsideColor)
}

func (self *Image) Thumbnail(size int) (im *ImageVersion, err error) {
	return self.VersionCentered(size, size, self.Grayscale())
}
