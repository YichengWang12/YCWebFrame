package web

import (
	"fmt"
	"regexp"
	"strings"
)

type router struct {
	// trees 是按照 HTTP 方法来组织的
	trees map[string]*node
}

type nodeType int

const (
	// 静态路由
	nodeTypeStatic = iota
	// 正则路由
	//nodeTypeReg
	// 路径参数路由
	nodeTypeParam
	// 通配符路由
	nodeTypeAny
)

type node struct {
	typ nodeType

	path string
	// children 子节点
	// 子节点的 path => node
	children map[string]*node
	// handler 命中路由之后执行的逻辑
	handler HandleFunc

	mdls        []Middleware
	matchedMdls []Middleware

	route string

	// 通配符 * 表达的节点，任意匹配
	starChild *node

	paramChild *node
	// 正则路由和参数路由都会使用这个字段
	paramName string

	// 正则表达式
	regChild *node
	regExpr  *regexp.Regexp
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
	mdls       []Middleware
}

func newRouter() router {
	return router{
		trees: map[string]*node{},
	}
}

// addRoute register route
// method should be HTTP method
// Do not support multiple '/' and path must starts with '/', cannot end with'/'
func (r *router) addRoute(method string, path string, handler HandleFunc, ms ...Middleware) {
	// if path is empty
	if path == "" {
		panic("web: path is empty")
	}

	if path[0] != '/' {
		panic("web: path must starts with '/'")
	}

	if path != "/" && path[len(path)-1] == '/' {
		panic("web: path cannot end with '/'")
	}

	root, ok := r.trees[method]

	if !ok {
		root = &node{path: "/"}
		r.trees[method] = root
	}

	if path == "/" {
		if root.handler != nil {
			panic("web: route conflict[/]")
		}
		root.handler = handler
		root.mdls = ms
		return
	}

	segs := strings.Split(path[1:], "/")

	for _, s := range segs {
		if s == "" {
			panic("web : multiple duplicate '/'")
		}
		root = root.childOrCreate(s)
	}

	if root.handler != nil {
		panic(fmt.Sprintf("web: duplicate path [%s]", path))
	}
	root.handler = handler
	root.route = path
	root.mdls = ms
}

// findRoute find the requested route node
// notice: node will be considered registered only it has handler (not nil)
//func (r *router) findRoute1(method string, path string) (*matchInfo, bool) {
//	root, ok := r.trees[method]
//	if !ok {
//		return nil, false
//	}
//	if path == "/" {
//		return &matchInfo{n: root, mdls: root.mdls}, true
//	}
//	segs := strings.Split(strings.Trim(path, "/"), "/")
//	mi := &matchInfo{}
//	for _, s := range segs {
//		var child *node
//		child, ok = root.childOf(s)
//		if !ok {
//			if root.typ == nodeTypeAny {
//				mi.n = root
//				mi.mdls = r.findMdls(root, segs)
//				return mi, true
//			}
//			return nil, false
//		}
//		if child.paramName != "" {
//			mi.addValue(child.paramName, s)
//		}
//		root = child
//	}
//	mi.n = root
//	mi.mdls = r.findMdls(root, segs)
//	return mi, true
//}

func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	if path == "/" {
		return &matchInfo{n: root, mdls: root.mdls}, true
	}

	segs := strings.Split(strings.Trim(path, "/"), "/")
	mi := &matchInfo{}
	cur := root
	for _, s := range segs {
		var matchParam bool
		cur, matchParam, ok = cur.childOf(s)
		if !ok {
			return nil, false
		}
		if matchParam {
			mi.addValue(root.path[1:], s)
		}
	}
	mi.n = cur
	mi.mdls = r.findMdls(root, segs)
	return mi, true
}

func (r *router) findMdls(root *node, segs []string) []Middleware {
	queue := []*node{root}
	res := make([]Middleware, 0, 16)
	for i := 0; i < len(segs); i++ {
		seg := segs[i]
		var children []*node
		for _, cur := range queue {
			if len(cur.mdls) > 0 {
				res = append(res, cur.mdls...)
			}
			children = append(children, cur.childrenOf(seg)...)
		}
		queue = children
	}

	for _, cur := range queue {
		if len(cur.mdls) > 0 {
			res = append(res, cur.mdls...)
		}
	}
	return res
}

func (n *node) childrenOf(path string) []*node {
	res := make([]*node, 0, 4)
	var static *node
	if n.children != nil {
		static = n.children[path]
	}
	if n.starChild != nil {
		res = append(res, n.starChild)
	}
	if n.paramChild != nil {
		res = append(res, n.paramChild)
	}
	//if n.regChild != nil {
	//	res = append(res, n.regChild)
	//}
	if static != nil {
		res = append(res, static)
	}
	return res
}

func (n *node) childOf(path string) (*node, bool, bool) {
	if n.children == nil {
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}
	res, ok := n.children[path]
	if !ok {
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}
	return res, false, ok
}

func (n *node) childOfNonStatic(path string) (*node, bool) {
	if n.regChild != nil {
		if n.regChild.regExpr.Match([]byte(path)) {
			return n.regChild, true
		}
	}
	if n.paramChild != nil {
		return n.paramChild, true
	}
	return n.starChild, n.starChild != nil
}
func (n *node) childOrCreate(path string) *node {
	if path == "*" {
		if n.paramChild != nil {
			panic(fmt.Sprintf("web: illegal route : already has param route here [%s]", path))
		}
		if n.regChild != nil {
			panic(fmt.Sprintf("web: illegal route : already has param route here [%s]", path))
		}
		if n.starChild == nil {
			n.starChild = &node{path: path, typ: nodeTypeAny}
		}
		return n.starChild
	}
	if path[0] == ':' {
		paramName, _, _ := n.parseParam(path)
		//if isReg {
		//	return n.childOrCreateReg(path, expr, paramName)
		//}
		return n.childOrCreateParam(path, paramName)

	}
	if n.children == nil {
		n.children = make(map[string]*node)
	}
	child, ok := n.children[path]
	if !ok {
		child = &node{path: path, typ: nodeTypeStatic}
		n.children[path] = child
	}
	return child
}

func (n *node) childOrCreateParam(path string, paramName string) *node {
	if n.regChild != nil {
		panic(fmt.Sprintf("web: illegal route : already has regexp route here %s", path))
	}
	if n.starChild != nil {
		panic(fmt.Sprintf("web: illegal route : already has '*' route here %s", path))
	}
	if n.paramChild != nil {
		if n.paramChild.path != path {
			panic(fmt.Sprintf("web: illegal route : already has param route here %s", n.paramChild.path))
		}

	} else {
		n.paramChild = &node{path: path, paramName: paramName, typ: nodeTypeParam}
	}
	return n.paramChild

}

//func (n *node) childOrCreateReg(path string, expr string, paramName string) *node {
//
//	if n.starChild != nil {
//		panic(fmt.Sprintf("web: illegal route : already has '*' route here %s", path))
//	}
//	if n.paramChild != nil {
//		panic(fmt.Sprintf("web: illegal route : already has param route here %s", path))
//	}
//	if n.regChild != nil {
//		if n.regChild.regExpr.String() != expr || n.paramName != paramName {
//			panic(fmt.Sprintf("web: illegal route : already has regexp route here %s", path))
//		}
//	} else {
//		regExpr, err := regexp.Compile(expr)
//		if err != nil {
//			panic(fmt.Errorf("web: illegal route : invalid regexp %w", err))
//		}
//		n.regChild = &node{path: path, regExpr: regExpr, paramName: paramName, typ: nodeTypeReg}
//	}
//	return n.regChild
//
//}

func (m *matchInfo) addValue(key string, value string) {
	if m.pathParams == nil {
		m.pathParams = map[string]string{key: value}
	}
	m.pathParams[key] = value
}

func (n *node) parseParam(path string) (string, string, bool) {
	path = path[1:]
	segs := strings.SplitN(path, "(", 2)
	if len(segs) == 2 {
		expr := segs[1]
		if strings.HasSuffix(expr, ")") {
			return segs[0], expr[:len(expr)-1], true
		}
	}
	return path, "", false
}
