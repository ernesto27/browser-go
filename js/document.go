package js

import (
	"browser/dom"

	"github.com/dop251/goja"
)

type Document struct {
	rt   *JSRuntime
	root *dom.Node
}

func newDocument(rt *JSRuntime, root *dom.Node) *Document {
	return &Document{
		rt:   rt,
		root: root,
	}
}

func (d *Document) GetElementById(id string) goja.Value {
	node := dom.FindByID(d.root, id)
	if node == nil {
		return goja.Null()
	}

	return d.rt.wrapElement(node)
}
