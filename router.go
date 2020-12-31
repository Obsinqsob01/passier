package passier

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

// Context is the request object in order to use this router
type Context struct {
	W      http.ResponseWriter
	R      *http.Request
	Params url.Values
}

// JSON sends json response
func (c *Context) JSON(data interface{}) error {
	c.W.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(c.W).Encode(data)
}

// Handle it's type to handle I guess...
type Handle func(*Context) error

type node struct {
	children     []*node
	component    string
	isNamedParam bool
	methods      map[string]Handle
}

func (n *node) traverse(components []string, params url.Values) (*node, string) {
	component := components[0]
	if len(n.children) > 0 {
		for _, child := range n.children {
			if component == child.component || child.isNamedParam {
				if child.isNamedParam && params != nil {
					params.Add(child.component[1:], component)
				}
				next := components[1:]
				if len(next) > 0 {
					return child.traverse(next, params)
				}

				return child, component
			}
		}
	}

	return n, component
}

func (n *node) addNode(method, path string, handler Handle) {
	components := strings.Split(path, "/")[1:]
	count := len(components)

	for {
		aNode, component := n.traverse(components, nil)

		if aNode.component == component && count == 1 {
			aNode.methods[method] = handler
			return
		}

		newNode := node{
			component:    component,
			isNamedParam: false,
			methods:      make(map[string]Handle),
		}

		if len(component) > 0 && component[0] == ':' {
			newNode.isNamedParam = true
		}

		if count == 1 {
			newNode.methods[method] = handler
		}

		aNode.children = append(aNode.children, &newNode)
		count--
		if count == 0 {
			break
		}
	}
}

// Router ... duh
type Router struct {
	tree        *node
	rootHandler Handle
}

// VerifyPath verifies that path start is right
func (r *Router) VerifyPath(path byte) {
	if path != '/' {
		panic("Path has to start with a ./")
	}
}

// GET handles routes for get method
func (r *Router) GET(path string, handler Handle) {
	r.VerifyPath(path[0])
	r.tree.addNode(http.MethodGet, path, handler)
}

// POST handles routes for post method
func (r *Router) POST(path string, handler Handle) {
	r.VerifyPath(path[0])
	r.tree.addNode(http.MethodPost, path, handler)
}

// PUT handles routes for put method
func (r *Router) PUT(path string, handler Handle) {
	r.VerifyPath(path[0])
	r.tree.addNode(http.MethodPut, path, handler)
}

// DELETE handles routes for get method
func (r *Router) DELETE(path string, handler Handle) {
	r.VerifyPath(path[0])
	r.tree.addNode(http.MethodDelete, path, handler)
}

// PATCH handles routes for get method
func (r *Router) PATCH(path string, handler Handle) {
	r.VerifyPath(path[0])
	r.tree.addNode(http.MethodPatch, path, handler)
}

func (r *Router) handleContext(w http.ResponseWriter, req *http.Request, values url.Values) *Context {
	return &Context{w, req, values}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	params := req.Form
	node, _ := r.tree.traverse(strings.Split(req.URL.Path, "/")[1:], params)

	context := r.handleContext(w, req, params)

	if handler := node.methods[req.Method]; handler != nil {
		handler(context)
	} else {
		r.rootHandler(context)
	}
}

// New Creates a new instance of router
func New(rootHandler Handle) *Router {
	node := node{component: "/", isNamedParam: false, methods: make(map[string]Handle)}
	return &Router{tree: &node, rootHandler: rootHandler}
}
