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
	onLoadHandler       goja.Callable
	windowLoadListeners []goja.Callable
}

// collectTableRows returns all tr elements in a table node in WHATWG 4.9.1 order:
// thead rows first, then tbody/direct tr rows in tree order, then tfoot rows.
// collectSectionRows returns all direct tr children of a tbody/thead/tfoot node.
func collectSectionRows(sectionNode *dom.Node) []*dom.Node {
	var rows []*dom.Node
	for _, child := range sectionNode.Children {
		if child.Type == dom.Element && child.TagName == "tr" {
			rows = append(rows, child)
		}
	}
	return rows
}

func collectTableRows(tableNode *dom.Node) []*dom.Node {
	var rows []*dom.Node
	collectTRs := func(section *dom.Node) {
		for _, child := range section.Children {
			if child.Type == dom.Element && child.TagName == "tr" {
				rows = append(rows, child)
			}
		}
	}
	for _, child := range tableNode.Children {
		if child.Type == dom.Element && child.TagName == "thead" {
			collectTRs(child)
		}
	}
	for _, child := range tableNode.Children {
		if child.Type == dom.Element {
			switch child.TagName {
			case "tbody":
				collectTRs(child)
			case "tr":
				rows = append(rows, child)
			}
		}
	}
	for _, child := range tableNode.Children {
		if child.Type == dom.Element && child.TagName == "tfoot" {
			collectTRs(child)
		}
	}
	return rows
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

	docObj.DefineAccessorProperty("body",

		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			bodyNode := dom.FindElementsByTagName(rt.document, dom.TagBody)
			if bodyNode == nil {
				return goja.Null()
			}
			return rt.wrapElement(bodyNode)
		}),
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				newBodyNode := unwrapNode(rt, call.Arguments[0])
				if newBodyNode == nil || (newBodyNode.TagName != "body" && newBodyNode.TagName != "frameset") {
					panic(rt.vm.NewTypeError("HierarchyRequestError: The new body element must be a body or frameset element."))
				}
				existingBody := dom.FindElementsByTagName(rt.document, dom.TagBody)
				if existingBody != nil && existingBody.Parent != nil {
					parent := existingBody.Parent
					parent.RemoveChild(existingBody)
					parent.AppendChild(newBodyNode)
				} else {
					htmlNode := dom.FindElementsByTagName(rt.document, "html")
					if htmlNode != nil {
						htmlNode.AppendChild(newBodyNode)
					} else {
						rt.document.AppendChild(newBodyNode)
					}
				}
				if rt.onReflow != nil {
					rt.onReflow()
				}
			}

			return goja.Undefined()
		}),
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

	window.DefineAccessorProperty("onload",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if rt.onLoadHandler == nil {
				return goja.Null()
			}
			return rt.vm.ToValue(rt.onLoadHandler)
		}),
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				if callback, ok := goja.AssertFunction(call.Arguments[0]); ok {
					rt.onLoadHandler = callback
				} else {
					rt.onLoadHandler = nil
				}
			}
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	window.Set("addEventListener", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		eventType := call.Arguments[0].String()
		callback, ok := goja.AssertFunction(call.Arguments[1])
		if !ok {
			return goja.Undefined()
		}
		if eventType == "load" {
			rt.windowLoadListeners = append(rt.windowLoadListeners, callback)
		}
		return goja.Undefined()
	})

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

	tagName := strings.ToUpper(node.TagName)

	elem := newElement(rt, node)
	obj := rt.vm.NewObject()

	attrsObj := rt.vm.NewObject()
	for name, value := range node.Attributes {
		attrsObj.Set(name, value)
	}
	obj.Set("attributes", attrsObj)

	// Static properties
	obj.Set("tagName", tagName)
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

			if tagName == "BASE" {
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

	if tagName == "TITLE" {
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

	if tagName == "A" {
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

	if tagName == "BLOCKQUOTE" || tagName == "Q" ||
		tagName == "INS" || tagName == "DEL" {
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
	if tagName == "TABLE" {
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
				allRows := collectTableRows(node)
				var rows []any
				for _, row := range allRows {
					rows = append(rows, rt.wrapElement(row))
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

			allRows := collectTableRows(node)
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

		obj.Set("deleteRow", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			index := int64(-1)
			if len(call.Arguments) > 0 {
				index = call.Argument(0).ToInteger()
			}

			allRows := collectTableRows(node)

			if index == -1 && len(allRows) > 0 {
				index = int64(len(allRows) - 1)
			}

			if index >= 0 && index < int64(len(allRows)) {
				targetRow := allRows[index]
				targetRow.Parent.RemoveChild(targetRow)
				if rt.onReflow != nil {
					rt.onReflow()
				}
				return goja.Undefined()
			}

			return goja.Undefined()
		}))

	}

	// HTMLTableSectionElement properties (WHATWG 4.9.5-4.9.7)
	if tagName == "TBODY" || tagName == "THEAD" || tagName == "TFOOT" {
		// rows - HTMLCollection of tr elements within this section only
		obj.DefineAccessorProperty("rows",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				sectionRows := collectSectionRows(node)
				var rows []any
				for _, row := range sectionRows {
					rows = append(rows, rt.wrapElement(row))
				}
				return rt.vm.NewArray(rows...)
			}),
			nil,
			goja.FLAG_FALSE, goja.FLAG_TRUE)

		// insertRow(index) - creates a new tr and inserts it at index within this section.
		// index -1 or omitted appends at end. Out-of-range returns undefined (spec: IndexSizeError).
		obj.Set("insertRow", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			index := int64(-1)
			if len(call.Arguments) > 0 {
				index = call.Argument(0).ToInteger()
			}

			sectionRows := collectSectionRows(node)
			newRow := dom.NewElement("tr", map[string]string{})
			newRow.Parent = node

			if index == -1 || index == int64(len(sectionRows)) {
				node.Children = append(node.Children, newRow)
			} else if index >= 0 && index < int64(len(sectionRows)) {
				targetRow := sectionRows[index]
				for i, child := range node.Children {
					if child == targetRow {
						node.Children = append(
							node.Children[:i],
							append([]*dom.Node{newRow}, node.Children[i:]...)...)
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

		// deleteRow(index) - removes the tr at index within this section.
		// index -1 removes the last row. Out-of-range does nothing (spec: IndexSizeError).
		obj.Set("deleteRow", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			index := int64(-1)
			if len(call.Arguments) > 0 {
				index = call.Argument(0).ToInteger()
			}

			sectionRows := collectSectionRows(node)

			if index == -1 && len(sectionRows) > 0 {
				index = int64(len(sectionRows) - 1)
			}

			if index >= 0 && index < int64(len(sectionRows)) {
				sectionRows[index].Parent.RemoveChild(sectionRows[index])
				if rt.onReflow != nil {
					rt.onReflow()
				}
			}

			return goja.Undefined()
		}))
	}

	// HTMLTableRowElement properties (WHATWG 4.9.8)
	if tagName == "TR" {
		// tr.cells - returns HTMLCollection of td/th elements in document order
		obj.DefineAccessorProperty("cells",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				var cells []any
				for _, child := range node.Children {
					if child.Type == dom.Element && (child.TagName == "td" || child.TagName == "th") {
						cells = append(cells, rt.wrapElement(child))
					}
				}
				return rt.vm.NewArray(cells...)
			}),
			nil,
			goja.FLAG_FALSE, goja.FLAG_TRUE)

		// tr.rowIndex - returns the position of the row in the table's rows collection, or -1
		obj.DefineAccessorProperty("rowIndex",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				// Walk up to find the parent <table>
				var tableNode *dom.Node
				for p := node.Parent; p != nil; p = p.Parent {
					if p.Type == dom.Element && p.TagName == "table" {
						tableNode = p
						break
					}
				}
				if tableNode == nil {
					return rt.vm.ToValue(-1)
				}
				allRows := collectTableRows(tableNode)
				for i, row := range allRows {
					if row == node {
						return rt.vm.ToValue(i)
					}
				}
				return rt.vm.ToValue(-1)
			}),
			nil,
			goja.FLAG_FALSE, goja.FLAG_TRUE)

		// tr.sectionRowIndex - position of the row within its parent section
		obj.DefineAccessorProperty("sectionRowIndex",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				p := node.Parent
				if p == nil {
					return rt.vm.ToValue(-1)
				}
				indexCount := 0
				for _, child := range p.Children {
					if child.Type == dom.Element && child.TagName == "tr" {
						if child == node {
							return rt.vm.ToValue(indexCount)
						}
						indexCount++
					}
				}
				return rt.vm.ToValue(-1)
			}),
			nil,
			goja.FLAG_FALSE, goja.FLAG_TRUE)

		// tr.insertCell(index) - inserts a new td cell at the given index, returns the new cell
		obj.Set("insertCell", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			index := int64(-1)
			if len(call.Arguments) > 0 {
				index = call.Argument(0).ToInteger()
			}

			// Collect current cells (td/th only)
			var cells []*dom.Node
			for _, child := range node.Children {
				if child.Type == dom.Element && (child.TagName == "td" || child.TagName == "th") {
					cells = append(cells, child)
				}
			}

			newCell := dom.NewElement("td", map[string]string{})
			newCell.Parent = node

			if index == -1 || index == int64(len(cells)) {
				// Append at end
				node.Children = append(node.Children, newCell)
			} else if index >= 0 && index < int64(len(cells)) {
				// Insert before the cell at the given index
				targetCell := cells[index]
				for i, child := range node.Children {
					if child == targetCell {
						node.Children = append(
							node.Children[:i],
							append([]*dom.Node{newCell}, node.Children[i:]...)...)
						break
					}
				}
			} else {
				return goja.Undefined()
			}

			if rt.onReflow != nil {
				rt.onReflow()
			}
			return rt.wrapElement(newCell)
		}))

		// tr.deleteCell(index) - removes the cell at the given index
		obj.Set("deleteCell", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			index := int64(-1)
			if len(call.Arguments) > 0 {
				index = call.Argument(0).ToInteger()
			}

			// Collect current cells (td/th only)
			var cells []*dom.Node
			for _, child := range node.Children {
				if child.Type == dom.Element && (child.TagName == "td" || child.TagName == "th") {
					cells = append(cells, child)
				}
			}

			if index == -1 && len(cells) > 0 {
				index = int64(len(cells) - 1)
			}

			if index >= 0 && index < int64(len(cells)) {
				node.RemoveChild(cells[index])
				if rt.onReflow != nil {
					rt.onReflow()
				}
			}

			return goja.Undefined()
		}))
	}

	if tagName == "TD" || tagName == "TH" {
		obj.DefineAccessorProperty("colSpan",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				colspanAttr := node.Attributes["colspan"]
				if colspanAttr == "" {
					return rt.vm.ToValue(1)
				}
				colspan, err := strconv.Atoi(colspanAttr)
				if err != nil {
					return rt.vm.ToValue(1)
				}

				if colspan < 1 {
					return rt.vm.ToValue(1)
				}

				return rt.vm.ToValue(colspan)
			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					v, err := strconv.Atoi(call.Arguments[0].String())
					if err == nil {
						if v < 1 {
							v = 1
						} else if v > 1000 {
							v = 1000
						}
						node.Attributes["colspan"] = strconv.Itoa(v)
					}
					if rt.onReflow != nil {
						rt.onReflow()
					}
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)

		obj.DefineAccessorProperty("rowSpan",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				rowspanAttr := node.Attributes["rowspan"]
				if rowspanAttr == "" {
					return rt.vm.ToValue(1)
				}
				rowspan, err := strconv.Atoi(rowspanAttr)
				if err != nil {
					return rt.vm.ToValue(1)
				}
				return rt.vm.ToValue(rowspan)
			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					v, err := strconv.Atoi(call.Arguments[0].String())
					if err == nil {
						if v < 0 {
							v = 0
						} else if v > 65534 {
							v = 65534
						}
						node.Attributes["rowspan"] = strconv.Itoa(v)
					}
					if rt.onReflow != nil {
						rt.onReflow()
					}
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)

		obj.DefineAccessorProperty("cellIndex",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if node.Parent == nil || (node.Parent.TagName != "tr") {
					return rt.vm.ToValue(-1)
				}
				idx := 0
				for _, sibling := range node.Parent.Children {
					if sibling == node {
						return rt.vm.ToValue(idx)
					}
					if sibling.Type == dom.Element && (sibling.TagName == "td" || sibling.TagName == "th") {
						idx++
					}
				}
				return rt.vm.ToValue(-1)
			}),
			nil,
			goja.FLAG_FALSE, goja.FLAG_TRUE)

		obj.DefineAccessorProperty("headers",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				headersAttr := node.Attributes["headers"]
				if headersAttr == "" {
					return rt.vm.ToValue("")
				}
				return rt.vm.ToValue(headersAttr)
			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					node.Attributes["headers"] = call.Arguments[0].String()
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)

		// scope - enumerated attribute, limited to known values per WHATWG 4.9.11
		// valid values: "row", "col", "rowgroup", "colgroup"; invalid/missing → ""
		obj.DefineAccessorProperty("scope",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				switch node.Attributes["scope"] {
				case "row", "col", "rowgroup", "colgroup":
					return rt.vm.ToValue(node.Attributes["scope"])
				default:
					return rt.vm.ToValue("")
				}
			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					node.Attributes["scope"] = call.Arguments[0].String()
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)

	}

	if tagName == "OL" {
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

	if tagName == "THEAD" || tagName == "TFOOT" || tagName == "TBODY" {
		obj.DefineAccessorProperty("rows",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				var rows []any
				for _, child := range node.Children {
					if child.Type == dom.Element && child.TagName == "tr" {
						rows = append(rows, rt.wrapElement(child))
					}
				}
				return rt.vm.NewArray(rows...)
			}),
			nil,
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

func (rt *JSRuntime) FireLoad() {
	// Path 1: window.onload / body.onload (same slot — script wins over inline attribute)
	if rt.onLoadHandler != nil {
		rt.onLoadHandler(goja.Undefined())
	} else {
		// Fall back to <body onload="..."> only if no script set window.onload
		bodyNode := dom.FindElementsByTagName(rt.document, dom.TagBody)
		if bodyNode != nil {
			rt.ExecuteInlineEvent(bodyNode, "load")
		}
	}

	// Path 2: window.addEventListener('load', fn) — independent list, always fires
	for _, listener := range rt.windowLoadListeners {
		listener(goja.Undefined())
	}
}
