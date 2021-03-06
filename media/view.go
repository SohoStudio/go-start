package media

import (
	"io"

	"github.com/ungerik/go-start/view"
)

var View = view.NewViewURLWrapper(view.RenderView(
	func(ctx *view.Context) error {
		reader, contentType, err := Config.Backend.ImageVersionReader(ctx.URLArgs[0])
		if err != nil {
			if _, ok := err.(ErrInvalidImageID); ok {
				return view.NotFound(ctx.URLArgs[0] + "/" + ctx.URLArgs[1] + " not found")
			}
			return err
		}
		_, err = io.Copy(ctx.Response, reader)
		if err != nil {
			return err
		}
		err = reader.Close()
		if err != nil {
			return err
		}
		ctx.Response.Header().Set("Content-Type", contentType)
		return nil
	},
))

func ViewPath(name string) view.ViewPath {
	return view.ViewPath{Name: name, Args: 2, View: View}
}
