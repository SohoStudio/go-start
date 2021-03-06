package view

import (
	"fmt"
	"reflect"

	"github.com/ungerik/go-start/reflection"
)

/*
RenderView implements all View methods for a View.Render compatible function.

Example:

	renderView := RenderView(
		func(ctx *Context) error {
			writer.Write([]byte("<html><body>Any Content</body></html>"))
			return nil
		},
	)
*/
type RenderView func(ctx *Context) error

func (self RenderView) Init(thisView View) {
}

func (self RenderView) ID() string {
	return ""
}

func (self RenderView) IterateChildren(callback IterateChildrenCallback) {
}

func (self RenderView) Render(ctx *Context) error {
	return self(ctx)
}

func RenderViewBindURLArgs(renderFunc interface{}) RenderView {
	v := reflect.ValueOf(renderFunc)
	t := v.Type()
	if t.Kind() != reflect.Func {
		panic(fmt.Errorf("RenderViewBindURLArgs: renderFunc must be a function, got %T", renderFunc))
	}
	if t.NumIn() == 0 {
		panic(fmt.Errorf("RenderViewBindURLArgs: renderFunc has no arguments, needs at least one *view.Response"))
	}
	if t.In(0) != reflect.TypeOf((*Response)(nil)) {
		panic(fmt.Errorf("RenderViewBindURLArgs: renderFunc's first argument must type must be *view.Response, got %s", t.In(0)))
	}
	if t.NumOut() != 1 {
		panic(fmt.Errorf("RenderViewBindURLArgs: renderFunc must have one result, got %d", t.NumOut()))
	}
	if t.Out(0) != reflection.TypeOfError {
		panic(fmt.Errorf("RenderViewBindURLArgs: renderFunc's result must be of type error, got %s", t.Out(0)))
	}
	return RenderView(
		func(ctx *Context) error {
			if len(ctx.URLArgs) != t.NumIn()-1 {
				panic(fmt.Errorf("RenderViewBindURLArgs: number of response URL args does not match renderFunc's arg count"))
			}
			args := make([]reflect.Value, t.NumIn())
			args[0] = reflect.ValueOf(ctx)
			for i, urlArg := range ctx.URLArgs {
				val, err := reflection.StringToValueOfType(urlArg, t.In(i+1))
				if err != nil {
					return err
				}
				args[i+1] = reflect.ValueOf(val)
			}
			return v.Call(args)[0].Interface().(error)
		},
	)
}
