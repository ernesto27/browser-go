package js

import (
	"browser/dom"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/dop251/goja"
)

type JSRuntime struct {
	vm                  *goja.Runtime
	document            *dom.Node
	onReflow            func()
	onAlert             func(message string)
	Events              *EventManager
	onConfirm           func(string) bool
	currentURL          string
	onReload            func()
	onPrompt            func(message, defaultValue string) *string
	elementCache        map[*dom.Node]*goja.Object
	onTitleChange       func(string)
	beforeUnloadHandler goja.Callable
}

func NewJSRuntime(document *dom.Node, onReflow func()) *JSRuntime {
	rt := &JSRuntime{
		vm:           goja.New(),
		document:     document,
		onReflow:     onReflow,
		Events:       NewEventManager(),
		elementCache: make(map[*dom.Node]*goja.Object),
	}
	rt.setupGlobals()
	return rt
}

func (rt *JSRuntime) setupGlobals() {
	console := rt.vm.NewObject()
	console.Set("log", func(call goja.FunctionCall) goja.Value {
		for _, arg := range call.Arguments {
			fmt.Print(arg.String(), " ")
		}
		fmt.Println()
		return goja.Undefined()
	})
	rt.vm.Set("console", console)

	doc := newDocument(rt, rt.document)
	docObj := rt.vm.NewObject()
	docObj.Set("getElementById", doc.GetElementById)

	// document.documentElement
	docObj.DefineAccessorProperty("documentElement",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			for _, child := range rt.document.Children {
				if child.Type == dom.Element {
					return rt.wrapElement(child)
				}
			}
			return goja.Null()
		}),
		nil,
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	docObj.DefineAccessorProperty("head", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
		headNode := dom.FindElementsByTagName(rt.document, dom.TagHead)
		if headNode == nil {
			return goja.Null()
		}
		return rt.wrapElement(headNode)
	}),
		nil,
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	docObj.DefineAccessorProperty("title",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			return rt.vm.ToValue(dom.FindTitle(rt.document))
		}),
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				newTitle := call.Arguments[0].String()
				titleNode := dom.FindElementsByTagName(rt.document, dom.TagTitle)
				if titleNode != nil {
					titleNode.Children = nil
					titleNode.AppendChild(dom.NewText(newTitle))
				}
				if rt.onTitleChange != nil {
					rt.onTitleChange(newTitle)
				}
			}
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	docObj.DefineAccessorProperty("body", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
		bodyNode := dom.FindElementsByTagName(rt.document, dom.TagBody)
		if bodyNode == nil {
			return goja.Null()
		}
		return rt.wrapElement(bodyNode)
	}),
		nil,
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	docObj.DefineAccessorProperty("baseURI",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			baseHref := dom.FindBaseHref(rt.document)
			if baseHref != "" {
				return rt.vm.ToValue(baseHref)
			}
			return rt.vm.ToValue(rt.currentURL)
		}),
		nil,
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	rt.vm.Set("document", docObj)

	rt.vm.Set("alert", func(call goja.FunctionCall) goja.Value {
		message := ""
		if len(call.Arguments) > 0 {
			message = call.Arguments[0].String()
		}
		if rt.onAlert != nil {
			rt.onAlert(message)
		}

		return goja.Undefined()
	})

	rt.vm.Set("confirm", func(call goja.FunctionCall) goja.Value {
		message := ""
		if len(call.Arguments) > 0 {
			message = call.Arguments[0].String()
		}

		result := false
		if rt.onConfirm != nil {
			result = rt.onConfirm(message)
		}

		return rt.vm.ToValue(result)
	})

	rt.vm.Set("prompt", func(call goja.FunctionCall) goja.Value {
		message := ""
		defaultValue := ""

		if len(call.Arguments) > 0 {
			message = call.Arguments[0].String()
		}

		if len(call.Arguments) > 1 {
			defaultValue = call.Arguments[1].String()
		}

		if rt.onPrompt != nil {
			result := rt.onPrompt(message, defaultValue)
			if result == nil {
				return goja.Null()
			}
			return rt.vm.ToValue(*result)
		}

		return goja.Null()
	})

	docObj.Set("createElement", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Null()
		}

		tagName := call.Arguments[0].String()
		newNode := dom.NewElement(tagName, nil)
		return rt.wrapElement(newNode)
	})

	docObj.Set("createTextNode", func(call goja.FunctionCall) goja.Value {
		text := ""
		if len(call.Arguments) > 0 {
			text = call.Arguments[0].String()
		}

		newNode := dom.NewText(text)
		return rt.wrapElement(newNode)
	})

	window := rt.vm.NewObject()
	location := rt.vm.NewObject()

	location.Set("href", rt.currentURL)

	location.Set("reload", func(call goja.FunctionCall) goja.Value {
		if rt.onReload != nil {
			rt.onReload()
		}
		return goja.Undefined()
	})

	window.Set("location", location)

	window.DefineAccessorProperty("onbeforeunload",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if rt.beforeUnloadHandler == nil {
				return goja.Null()
			}
			return rt.vm.ToValue(rt.beforeUnloadHandler)
		}),
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				if callback, ok := goja.AssertFunction(call.Arguments[0]); ok {
					rt.beforeUnloadHandler = callback
				} else {
					rt.beforeUnloadHandler = nil
				}
			}
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	rt.vm.Set("window", window)

}

func (rt *JSRuntime) Execute(code string) error {
	_, err := rt.vm.RunString(code)
	if err != nil {
		fmt.Println("JS error: ", err)
	}
	return err
}

// FindScripts extracts JavaScript code from <script> tags
func FindScripts(node *dom.Node) []string {
	var scripts []string
	findScriptsRecursive(node, &scripts)
	return scripts
}

func findScriptsRecursive(node *dom.Node, scripts *[]string) {
	if node == nil {
		return
	}

	if node.Type == dom.Element && node.TagName == "script" {
		// Get inline script content
		for _, child := range node.Children {
			if child.Type == dom.Text && child.Text != "" {
				*scripts = append(*scripts, child.Text)
			}
		}
	}

	for _, child := range node.Children {
		findScriptsRecursive(child, scripts)
	}
}

func (rt *JSRuntime) wrapElement(node *dom.Node) goja.Value {
	if node == nil {
		return goja.Null()
	}

	// Check cache first
	if cached, ok := rt.elementCache[node]; ok {
		return cached
	}

	elem := newElement(rt, node)
	obj := rt.vm.NewObject()

	attrsObj := rt.vm.NewObject()
	for name, value := range node.Attributes {
		attrsObj.Set(name, value)
	}
	obj.Set("attributes", attrsObj)

	// Static properties
	obj.Set("tagName", strings.ToUpper(node.TagName))
	obj.Set("id", node.Attributes["id"])
	obj.Set("className", node.Attributes["class"])

	// Methods
	obj.Set("getAttribute", elem.GetAttribute)
	obj.Set("setAttribute", elem.SetAttribute)

	// Dynamic property: textContent (getter/setter)
	obj.DefineAccessorProperty("textContent",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			return rt.vm.ToValue(elem.GetTextContent())
		}),
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				elem.SetTextContent(call.Arguments[0].String())
			}
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	// parentElement - only returns Element nodes, not Document
	obj.DefineAccessorProperty("parentElement",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if node.Parent == nil || node.Parent.Type != dom.Element {
				return goja.Null()
			}
			return rt.wrapElement(node.Parent)
		}),
		nil,
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	obj.DefineAccessorProperty("children",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			var elements []any
			for _, child := range node.Children {
				if child.Type == dom.Element {
					elements = append(elements, rt.wrapElement(child))
				}
			}
			arr := rt.vm.NewArray(elements...)
			return arr
		}),
		nil,
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	obj.Set("addEventListener", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}

		eventType := call.Arguments[0].String()

		callback, ok := goja.AssertFunction(call.Arguments[1])
		if !ok {
			return goja.Undefined()
		}

		rt.Events.AddEventListener(node, eventType, callback)
		return goja.Undefined()
	})

	obj.DefineAccessorProperty("innerHTML",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			return rt.vm.ToValue(elem.GetInnerHTML())
		}),
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				elem.SetInnerHTML(call.Arguments[0].String())
			}
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	obj.DefineAccessorProperty("innerText",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			return rt.vm.ToValue(node.InnerText())
		}),
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				node.SetInnerText(call.Arguments[0].String())
				if rt.onReflow != nil {
					rt.onReflow()
				}
			}
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	obj.DefineAccessorProperty("href",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			href := node.Attributes["href"]
			if href == "" {
				return goja.Undefined()
			}

			if strings.ToUpper(node.TagName) == "BASE" {
				return rt.vm.ToValue(href)
			}

			baseHref := dom.FindBaseHref(rt.document)
			if baseHref != "" {
				baseURL, err := url.Parse(baseHref)
				if err == nil {
					refURL, err := url.Parse(href)
					if err == nil {
						resolved := baseURL.ResolveReference(refURL)
						return rt.vm.ToValue(resolved.String())
					}
				}
			}

			return rt.vm.ToValue(href)
		}),
		nil,
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	obj.DefineAccessorProperty("className",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if node.Attributes == nil {
				return rt.vm.ToValue("")
			}
			return rt.vm.ToValue(node.Attributes["class"])
		}),
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				if node.Attributes == nil {
					node.Attributes = make(map[string]string)
				}
				node.Attributes["class"] = call.Arguments[0].String()
				if rt.onReflow != nil {
					rt.onReflow()
				}
			}
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	if strings.ToUpper(node.TagName) == "TITLE" {
		obj.DefineAccessorProperty("text",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				return rt.vm.ToValue(elem.GetTextContent())
			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					elem.SetTextContent(call.Arguments[0].String())
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)
	}

	if strings.ToUpper(node.TagName) == "A" {
		relList := dom.NewDOMTokenList(node, "rel")
		relListObj := rt.vm.NewObject()

		relListObj.DefineAccessorProperty("length",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				return rt.vm.ToValue(relList.Length())
			}),
			nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

		relListObj.Set("item", rt.vm.ToValue(func(index int) string {
			return relList.Item(index)
		}))

		relListObj.Set("contains", rt.vm.ToValue(func(token string) bool {
			return relList.Contains(token)
		}))

		relListObj.Set("add", rt.vm.ToValue(func(token string) {
			relList.Add(token)
		}))

		relListObj.Set("remove", rt.vm.ToValue(func(token string) {
			relList.Remove(token)
		}))

		relListObj.Set("toggle", rt.vm.ToValue(func(token string) bool {
			return relList.Toggle(token)
		}))

		obj.Set("relList", relListObj)

		// .text property (alias for innerText)
		obj.DefineAccessorProperty("text",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				return rt.vm.ToValue(elem.GetTextContent())
			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					elem.SetTextContent(call.Arguments[0].String())
					if rt.onReflow != nil {
						rt.onReflow()
					}
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)

		// Helper to parse resolved href
		getURL := func() *url.URL {
			href := node.Attributes["href"]
			if href == "" {
				return nil
			}
			parsed, err := url.Parse(href)
			if err != nil {
				return nil
			}
			if !parsed.IsAbs() {
				baseHref := dom.FindBaseHref(rt.document)
				if baseHref != "" {
					baseURL, err := url.Parse(baseHref)
					if err == nil {
						parsed = baseURL.ResolveReference(parsed)
					}
				}
			}
			return parsed
		}

		obj.DefineAccessorProperty("protocol",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if u := getURL(); u != nil {
					return rt.vm.ToValue(u.Scheme + ":")
				}
				return rt.vm.ToValue(":")
			}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

		obj.DefineAccessorProperty("username",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if u := getURL(); u != nil && u.User != nil {
					return rt.vm.ToValue(u.User.Username())
				}
				return rt.vm.ToValue("")
			}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

		obj.DefineAccessorProperty("password",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if u := getURL(); u != nil && u.User != nil {
					pass, _ := u.User.Password()
					return rt.vm.ToValue(pass)
				}
				return rt.vm.ToValue("")
			}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

		obj.DefineAccessorProperty("host",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if u := getURL(); u != nil {
					return rt.vm.ToValue(u.Host)
				}
				return rt.vm.ToValue("")
			}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

		obj.DefineAccessorProperty("hostname",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if u := getURL(); u != nil {
					return rt.vm.ToValue(u.Hostname())
				}
				return rt.vm.ToValue("")
			}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

		obj.DefineAccessorProperty("port",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if u := getURL(); u != nil {
					return rt.vm.ToValue(u.Port())
				}
				return rt.vm.ToValue("")
			}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

		obj.DefineAccessorProperty("pathname",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if u := getURL(); u != nil {
					return rt.vm.ToValue(u.Path)
				}
				return rt.vm.ToValue("")
			}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

		obj.DefineAccessorProperty("search",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if u := getURL(); u != nil && u.RawQuery != "" {
					return rt.vm.ToValue("?" + u.RawQuery)
				}
				return rt.vm.ToValue("")
			}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

		obj.DefineAccessorProperty("hash",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if u := getURL(); u != nil && u.Fragment != "" {
					return rt.vm.ToValue("#" + u.Fragment)
				}
				return rt.vm.ToValue("")
			}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

		obj.DefineAccessorProperty("origin",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if u := getURL(); u != nil {
					return rt.vm.ToValue(u.Scheme + "://" + u.Host)
				}
				return rt.vm.ToValue("")
			}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)
	}

	if strings.ToUpper(node.TagName) == "BLOCKQUOTE" || strings.ToUpper(node.TagName) == "Q" ||
		strings.ToUpper(node.TagName) == "INS" || strings.ToUpper(node.TagName) == "DEL" {
		obj.DefineAccessorProperty("cite",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				cite := node.Attributes["cite"]
				if cite == "" {
					return goja.Undefined()
				}

				// Resolve URL relative to document base
				baseHref := dom.FindBaseHref(rt.document)
				if baseHref != "" {
					baseURL, err := url.Parse(baseHref)
					if err == nil {
						refURL, err := url.Parse(cite)
						if err == nil {
							resolved := baseURL.ResolveReference(refURL)
							return rt.vm.ToValue(resolved.String())
						}
					}
				}

				return rt.vm.ToValue(cite)

			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					node.Attributes["cite"] = call.Arguments[0].String()
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)
	}

	// HTMLTableElement.caption property (WHATWG 4.9.1)
	if strings.ToUpper(node.TagName) == "TABLE" {
		obj.DefineAccessorProperty("caption",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				// Return first caption child, or null
				for _, child := range node.Children {
					if child.Type == dom.Element && child.TagName == "caption" {
						return rt.wrapElement(child)
					}
				}
				return goja.Null()
			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					// Remove existing caption first
					for _, child := range node.Children {
						if child.Type == dom.Element && child.TagName == "caption" {
							node.RemoveChild(child)
							break
						}
					}

					// If new value is not null, insert as first child
					if !goja.IsNull(call.Arguments[0]) && !goja.IsUndefined(call.Arguments[0]) {
						newCaption := unwrapNode(rt, call.Arguments[0])
						if newCaption != nil {
							newCaption.Parent = node
							node.Children = append([]*dom.Node{newCaption}, node.Children...)
						}
					}

					if rt.onReflow != nil {
						rt.onReflow()
					}
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)

		// HTMLTableElement.createCaption() (WHATWG 4.9.1)
		obj.Set("createCaption", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			// Return existing caption if one exists
			for _, child := range node.Children {
				if child.Type == dom.Element && child.TagName == "caption" {
					return rt.wrapElement(child)
				}
			}
			// Create new caption and insert as first child
			newCaption := dom.NewElement("caption", map[string]string{})
			newCaption.Parent = node
			node.Children = append([]*dom.Node{newCaption}, node.Children...)
			if rt.onReflow != nil {
				rt.onReflow()
			}
			return rt.wrapElement(newCaption)
		}))

		// HTMLTableElement.deleteCaption() (WHATWG 4.9.1)
		obj.Set("deleteCaption", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			for _, child := range node.Children {
				if child.Type == dom.Element && child.TagName == "caption" {
					node.RemoveChild(child)
					if rt.onReflow != nil {
						rt.onReflow()
					}
					break
				}
			}
			return goja.Undefined()
		}))

		obj.DefineAccessorProperty("tHead",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				for _, child := range node.Children {
					if child.Type == dom.Element && child.TagName == "thead" {
						return rt.wrapElement(child)
					}
				}
				return goja.Null()
			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					// Remove existing thead first
					for _, child := range node.Children {
						if child.Type == dom.Element && child.TagName == "thead" {
							node.RemoveChild(child)
							break
						}
					}

					// Find insertion index: after all caption and colgroup elements
					insertIdx := 0
					for _, child := range node.Children {
						if child.Type == dom.Element && (child.TagName == "caption" || child.TagName == "colgroup") {
							insertIdx++
						} else {
							break
						}
					}

					// If new value is not null, insert at computed index
					if !goja.IsNull(call.Arguments[0]) && !goja.IsUndefined(call.Arguments[0]) {
						newTHead := unwrapNode(rt, call.Arguments[0])
						if newTHead != nil {
							newTHead.Parent = node
							node.Children = append(
								node.Children[:insertIdx],
								append([]*dom.Node{newTHead}, node.Children[insertIdx:]...)...)
						}
					}

					if rt.onReflow != nil {
						rt.onReflow()
					}
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)

		// HTMLTableElement.createTHead() (WHATWG 4.9.1)
		obj.Set("createTHead", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			// Return existing thead if one exists
			for _, child := range node.Children {
				if child.Type == dom.Element && child.TagName == "thead" {
					return rt.wrapElement(child)
				}
			}
			// Create new thead and insert after caption/colgroup
			newTHead := dom.NewElement("thead", map[string]string{})
			newTHead.Parent = node
			insertIdx := 0
			for _, child := range node.Children {
				if child.Type == dom.Element && (child.TagName == "caption" || child.TagName == "colgroup") {
					insertIdx++
				} else {
					break
				}
			}
			node.Children = append(
				node.Children[:insertIdx],
				append([]*dom.Node{newTHead}, node.Children[insertIdx:]...)...)
			if rt.onReflow != nil {
				rt.onReflow()
			}
			return rt.wrapElement(newTHead)
		}))

		// HTMLTableElement.deleteTHead() (WHATWG 4.9.1)
		obj.Set("deleteTHead", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			for _, child := range node.Children {
				if child.Type == dom.Element && child.TagName == "thead" {
					node.RemoveChild(child)
					if rt.onReflow != nil {
						rt.onReflow()
					}
					break
				}
			}
			return goja.Undefined()
		}))

		obj.DefineAccessorProperty("tFoot",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				for _, child := range node.Children {
					if child.Type == dom.Element && child.TagName == "tfoot" {
						return rt.wrapElement(child)
					}
				}
				return goja.Null()
			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					for _, child := range node.Children {
						if child.Type == dom.Element && child.TagName == "tfoot" {
							node.RemoveChild(child)
							break
						}
					}

					if !goja.IsNull(call.Arguments[0]) && !goja.IsUndefined(call.Arguments[0]) {
						newTFoot := unwrapNode(rt, call.Arguments[0])
						if newTFoot != nil {
							newTFoot.Parent = node
							node.Children = append(node.Children, newTFoot)
						}
					}

					if rt.onReflow != nil {
						rt.onReflow()
					}
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)

		// HTMLTableElement.createTFoot() (WHATWG 4.9.1)
		obj.Set("createTFoot", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			// Return existing tfoot if one exists
			for _, child := range node.Children {
				if child.Type == dom.Element && child.TagName == "tfoot" {
					return rt.wrapElement(child)
				}
			}
			// Create new tfoot and append at end
			newTFoot := dom.NewElement("tfoot", map[string]string{})
			newTFoot.Parent = node
			node.Children = append(node.Children, newTFoot)
			if rt.onReflow != nil {
				rt.onReflow()
			}
			return rt.wrapElement(newTFoot)
		}))

		// HTMLTableElement.deleteTFoot() (WHATWG 4.9.1)
		obj.Set("deleteTFoot", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			for _, child := range node.Children {
				if child.Type == dom.Element && child.TagName == "tfoot" {
					node.RemoveChild(child)
					if rt.onReflow != nil {
						rt.onReflow()
					}
					break
				}
			}
			return goja.Undefined()
		}))

		// HTMLTableElement.tBodies (WHATWG 4.9.1) - returns HTMLCollection of tbody elements
		obj.DefineAccessorProperty("tBodies",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				var tbodies []any
				for _, child := range node.Children {
					if child.Type == dom.Element && child.TagName == "tbody" {
						tbodies = append(tbodies, rt.wrapElement(child))
					}
				}
				return rt.vm.NewArray(tbodies...)
			}),
			nil,
			goja.FLAG_FALSE, goja.FLAG_TRUE)

		// HTMLTableElement.rows (WHATWG 4.9.1) - returns HTMLCollection of all tr elements
		// Order: thead rows first, then tbody/direct tr rows in tree order, then tfoot rows
		obj.DefineAccessorProperty("rows",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				var rows []any

				collectTRs := func(section *dom.Node) {
					for _, child := range section.Children {
						if child.Type == dom.Element && child.TagName == "tr" {
							rows = append(rows, rt.wrapElement(child))
						}
					}
				}

				// Phase 1: thead rows
				for _, child := range node.Children {
					if child.Type == dom.Element && child.TagName == "thead" {
						collectTRs(child)
					}
				}

				// Phase 2: tbody rows and direct tr children
				for _, child := range node.Children {
					if child.Type == dom.Element {
						switch child.TagName {
						case "tbody":
							collectTRs(child)
						case "tr":
							rows = append(rows, rt.wrapElement(child))
						}
					}
				}

				// Phase 3: tfoot rows
				for _, child := range node.Children {
					if child.Type == dom.Element && child.TagName == "tfoot" {
						collectTRs(child)
					}
				}

				return rt.vm.NewArray(rows...)
			}),
			nil,
			goja.FLAG_FALSE, goja.FLAG_TRUE)

		obj.Set("createTBody", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			newTBody := dom.NewElement("tbody", map[string]string{})
			newTBody.Parent = node

			insertIdx := len(node.Children)
			for i := len(node.Children) - 1; i >= 0; i-- {
				if node.Children[i].Type == dom.Element && node.Children[i].TagName == "tbody" {
					insertIdx = i + 1
					break
				}
			}

			node.Children = append(
				node.Children[:insertIdx],
				append([]*dom.Node{newTBody}, node.Children[insertIdx:]...)...)

			if rt.onReflow != nil {
				rt.onReflow()
			}
			return rt.wrapElement(newTBody)
		}))

		obj.Set("insertRow", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			index := int64(-1)
			if len(call.Arguments) > 0 {
				index = call.Argument(0).ToInteger()
			}

			// Collect all rows using same 3-phase ordering as table.rows
			var allRows []*dom.Node
			collectRows := func(section *dom.Node) {
				for _, child := range section.Children {
					if child.Type == dom.Element && child.TagName == "tr" {
						allRows = append(allRows, child)
					}
				}
			}
			for _, child := range node.Children {
				if child.Type == dom.Element && child.TagName == "thead" {
					collectRows(child)
				}
			}
			for _, child := range node.Children {
				if child.Type == dom.Element {
					switch child.TagName {
					case "tbody":
						collectRows(child)
					case "tr":
						allRows = append(allRows, child)
					}
				}
			}
			for _, child := range node.Children {
				if child.Type == dom.Element && child.TagName == "tfoot" {
					collectRows(child)
				}
			}

			newRow := dom.NewElement("tr", map[string]string{})

			if index == -1 || index == int64(len(allRows)) {
				// Append to parent of last row, or create tbody if no rows
				if len(allRows) == 0 {
					newTBody := dom.NewElement("tbody", map[string]string{})
					newTBody.Parent = node
					node.Children = append(node.Children, newTBody)
					newRow.Parent = newTBody
					newTBody.Children = append(newTBody.Children, newRow)
				} else {
					lastRow := allRows[len(allRows)-1]
					parent := lastRow.Parent
					newRow.Parent = parent
					parent.Children = append(parent.Children, newRow)
				}
			} else if index >= 0 && index < int64(len(allRows)) {
				// Insert before the row at the given index
				targetRow := allRows[index]
				parent := targetRow.Parent
				newRow.Parent = parent
				for i, child := range parent.Children {
					if child == targetRow {
						parent.Children = append(
							parent.Children[:i],
							append([]*dom.Node{newRow}, parent.Children[i:]...)...)
						break
					}
				}
			} else {
				return goja.Undefined()
			}

			if rt.onReflow != nil {
				rt.onReflow()
			}
			return rt.wrapElement(newRow)
		}))
	}

	if strings.ToUpper(node.TagName) == "OL" {
		obj.DefineAccessorProperty("start",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				startAttr := node.Attributes["start"]
				if startAttr == "" {
					return rt.vm.ToValue(1)
				}

				start, err := strconv.Atoi(startAttr)
				if err != nil {
					return rt.vm.ToValue(1)
				}
				return rt.vm.ToValue(start)
			}),

			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					node.Attributes["start"] = call.Arguments[0].String()
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE,
		)

		obj.DefineAccessorProperty("reversed",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				_, exists := node.Attributes["reversed"]
				return rt.vm.ToValue(exists)
			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					if call.Arguments[0].ToBoolean() {
						node.Attributes["reversed"] = ""
					} else {
						delete(node.Attributes, "reversed")
					}
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)

		// type property - kind of list marker (1, a, A, i, I)
		obj.DefineAccessorProperty("type",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				typeAttr := node.Attributes["type"]
				if typeAttr == "" {
					return rt.vm.ToValue("1")
				}
				return rt.vm.ToValue(typeAttr)
			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					node.Attributes["type"] = call.Arguments[0].String()
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)
	}

	obj.DefineAccessorProperty("title",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			title := node.Attributes["title"]
			if title == "" {
				return goja.Undefined()
			}
			return rt.vm.ToValue(title)
		}),
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				if node.Attributes == nil {
					node.Attributes = make(map[string]string)
				}
				node.Attributes["title"] = call.Arguments[0].String()
			}
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	obj.Set("appendChild", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}

		childNode := unwrapNode(rt, call.Arguments[0])
		if childNode == nil {
			return goja.Undefined()
		}

		node.AppendChild(childNode)

		if rt.onReflow != nil {
			rt.onReflow()
		}

		return call.Arguments[0]
	})

	obj.Set("removeChild", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}

		childNode := unwrapNode(rt, call.Arguments[0])
		if childNode == nil {
			return goja.Undefined()
		}

		node.RemoveChild(childNode)

		if rt.onReflow != nil {
			rt.onReflow()
		}

		return call.Arguments[0]
	})

	obj.Set("remove", func(call goja.FunctionCall) goja.Value {
		node.Remove()

		if rt.onReflow != nil {
			rt.onReflow()
		}
		return goja.Undefined()
	})

	classList := rt.vm.NewObject()
	classList.Set("add", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			elem.ClassListAdd(call.Arguments[0].String())
		}
		return goja.Undefined()
	})

	classList.Set("remove", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			elem.ClassListRemove(call.Arguments[0].String())
		}
		return goja.Undefined()
	})

	obj.Set("classList", classList)

	obj.Set("_elem", elem)

	// HTMLStyleElement.disabled property (spec 4.2.6)
	if node.TagName == "style" {
		obj.DefineAccessorProperty("disabled",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				return rt.vm.ToValue(node.Disabled)
			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					node.Disabled = call.Arguments[0].ToBoolean()
					if rt.onReflow != nil {
						rt.onReflow()
					}
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)
	}

	if node.TagName == "data" {
		obj.DefineAccessorProperty("value",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if val, ok := node.Attributes["value"]; ok {
					return rt.vm.ToValue(val)
				}
				return rt.vm.ToValue("")
			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					node.Attributes["value"] = call.Arguments[0].String()
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)
	}

	// HTMLTimeElement.dateTime property (WHATWG 4.5.14)
	// HTMLModElement.dateTime property (WHATWG 4.7.1, 4.7.2)
	if node.TagName == "time" || node.TagName == "ins" || node.TagName == "del" {
		obj.DefineAccessorProperty("dateTime",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if val, ok := node.Attributes["datetime"]; ok {
					return rt.vm.ToValue(val)
				}
				return rt.vm.ToValue("")
			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					node.Attributes["datetime"] = call.Arguments[0].String()
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)
	}

	obj.DefineAccessorProperty("lang",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if val, ok := node.Attributes["lang"]; ok {
				return rt.vm.ToValue(val)
			}
			return rt.vm.ToValue("")
		}),
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				node.Attributes["lang"] = call.Arguments[0].String()
			}
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	// Cache before returning
	rt.elementCache[node] = obj

	return obj
}

func (rt *JSRuntime) DispatchClick(node *dom.Node) bool {
	inlinePrevented := rt.ExecuteInlineEvent(node, "click")
	listenerPrevented := rt.Events.Dispatch(rt, node, "click")

	return inlinePrevented || listenerPrevented
}

func (rt *JSRuntime) SetAlertHandler(handler func(message string)) {
	rt.onAlert = handler
}

func (rt *JSRuntime) SetConfirmHandler(handler func(string) bool) {
	rt.onConfirm = handler
}

func (rt *JSRuntime) SetCurrentURL(url string) {
	rt.currentURL = url
}

func (rt *JSRuntime) SetReloadHandler(handler func()) {
	rt.onReload = handler
}

func (rt *JSRuntime) SetPromptHandler(handler func(message, defaultValue string) *string) {
	rt.onPrompt = handler
}

func (rt *JSRuntime) SetTitleChangeHandler(handler func(string)) {
	rt.onTitleChange = handler
}

func (rt *JSRuntime) ExecuteInlineEvent(node *dom.Node, eventType string) bool {
	if node == nil || node.Type != dom.Element {
		return false
	}

	attrName := "on" + eventType

	code, exists := node.Attributes[attrName]
	if !exists || code == "" {
		return false
	}

	err := rt.Execute(code)
	if err != nil {
		fmt.Printf("Error executing inline %s: %v\n", eventType, err)
		return false
	}

	return true
}

func (rt *JSRuntime) CheckBeforeUnload() bool {
	fmt.Println("CheckBeforeUnload called")

	// Check window.onbeforeunload (set via JavaScript)
	if rt.beforeUnloadHandler != nil {
		result, err := rt.beforeUnloadHandler(goja.Undefined())
		if err != nil {
			return true
		}
		if result != nil && !goja.IsUndefined(result) && !goja.IsNull(result) {
			if rt.onConfirm != nil {
				return rt.onConfirm("Changes you made may not be saved. Leave anyway?")
			}
		}
	}

	// Check <body onbeforeunload="..."> attribute
	fmt.Println("  Checking body onbeforeunload attribute")
	bodyNode := dom.FindElementsByTagName(rt.document, dom.TagBody)
	if bodyNode != nil {
		code, ok := bodyNode.Attributes["onbeforeunload"]
		if ok && code != "" {
			// Wrap in function since inline handlers are implicitly functions
			wrappedCode := "(function() { " + code + " })()"
			result, err := rt.vm.RunString(wrappedCode)
			if err != nil {
				return true
			}
			if result != nil && !goja.IsUndefined(result) && !goja.IsNull(result) {
				if rt.onConfirm != nil {
					return rt.onConfirm("Changes you made may not be saved. Leave anyway?")
				}
			}
		}
	}

	fmt.Println("  No beforeunload handler, allowing navigation")
	return true
}
