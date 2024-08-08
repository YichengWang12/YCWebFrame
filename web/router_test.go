package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func Test_router_AddRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		// 通配符测试用例
		{
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc/*",
		},
		// 参数路由
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},
		// 正则路由
		{
			method: http.MethodDelete,
			path:   "/reg/:id(.*)",
		},
	}

	mockHandler := func(ctx *Context) {}
	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: {
				path: "/",
				children: map[string]*node{
					"user": {
						path: "user",
						children: map[string]*node{
							"home": {path: "home", handler: mockHandler, typ: nodeTypeStatic},
						},
						handler: mockHandler,
						typ:     nodeTypeStatic,
					},
					"order": {
						path: "order",
						children: map[string]*node{
							"detail": {path: "detail", handler: mockHandler, typ: nodeTypeStatic},
						},
						starChild: &node{path: "*", handler: mockHandler, typ: nodeTypeAny},
						typ:       nodeTypeStatic,
					},
					"param": {
						path: "param",
						paramChild: &node{
							path:      ":id",
							paramName: "id",
							starChild: &node{
								path:    "*",
								handler: mockHandler,
								typ:     nodeTypeAny,
							},
							children: map[string]*node{"detail": {path: "detail", handler: mockHandler, typ: nodeTypeStatic}},
							handler:  mockHandler,
							typ:      nodeTypeParam,
						},
					},
				},
				starChild: &node{
					path: "*",
					children: map[string]*node{
						"abc": {
							path:      "abc",
							starChild: &node{path: "*", handler: mockHandler, typ: nodeTypeAny},
							handler:   mockHandler,
							typ:       nodeTypeStatic,
						},
					},
					starChild: &node{path: "*", handler: mockHandler, typ: nodeTypeAny},
					handler:   mockHandler,
					typ:       nodeTypeAny,
				},
				handler: mockHandler,
				typ:     nodeTypeStatic,
			},
			http.MethodPost: {
				path: "/",
				children: map[string]*node{
					"order": {path: "order", children: map[string]*node{
						"create": {path: "create", handler: mockHandler, typ: nodeTypeStatic},
					}},
					"login": {path: "login", handler: mockHandler, typ: nodeTypeStatic},
				},
				typ: nodeTypeStatic,
			},
			//http.MethodDelete: {
			//	path: "/",
			//	children: map[string]*node{
			//		"reg": {
			//			path: "reg",
			//			typ:  nodeTypeStatic,
			//			regChild: &node{
			//				path:      ":id(.*)",
			//				paramName: "id",
			//				typ:       nodeTypeReg,
			//				handler:   mockHandler,
			//			},
			//		},
			//	},
			//	regChild: &node{
			//		path:      ":name(^.+$)",
			//		paramName: "name",
			//		typ:       nodeTypeReg,
			//		children: map[string]*node{
			//			"abc": {
			//				path:    "abc",
			//				handler: mockHandler,
			//			},
			//		},
			//	},
			//},
		},
	}
	msg, ok := wantRouter.equal(r)
	assert.True(t, ok, msg)

	//illegal case
	r = newRouter()
	//empty string
	assert.PanicsWithValue(t, "web: path is empty", func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	})
	//path does not start with '/'
	assert.PanicsWithValue(t, "web: path must starts with '/'", func() {
		r.addRoute(http.MethodGet, "a/b/c", mockHandler)
	})
	//path ends with '/'
	assert.PanicsWithValue(t, "web: path cannot end with '/'", func() {
		r.addRoute(http.MethodGet, "/a/b/c/", mockHandler)
	})
	//duplicate '/'
	assert.PanicsWithValue(t, "web : multiple duplicate '/'", func() {
		r.addRoute(http.MethodGet, "/a//b//c", mockHandler)
	})
	//duplicate root router
	r.addRoute(http.MethodGet, "/", mockHandler)
	assert.PanicsWithValue(t, "web: route conflict[/]", func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	})
	//duplicate normal router
	r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	assert.PanicsWithValue(t, "web: duplicate path [/a/b/c]", func() {
		r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	})

	// 同时注册通配符路由，参数路由，正则路由
	assert.PanicsWithValue(t, "web: illegal route : already has '*' route here :id", func() {
		r.addRoute(http.MethodGet, "/a/*", mockHandler)
		r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: illegal route : already has '*' route here :id(.*)", func() {
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: illegal route : already has '*' route here :id", func() {
		r.addRoute(http.MethodGet, "/*", mockHandler)
		r.addRoute(http.MethodGet, "/:id", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: illegal route : already has param route here [*]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: illegal route : already has param route here [*]", func() {
		r.addRoute(http.MethodGet, "/:id", mockHandler)
		r.addRoute(http.MethodGet, "/*", mockHandler)
	})
	r = newRouter()
	//assert.PanicsWithValue(t, "web: illegal route : already has param route here :id(.*)", func() {
	//	r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
	//	r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
	//})
	//r = newRouter()
	//assert.PanicsWithValue(t, "web: illegal route : already has regexp route here [*]", func() {
	//	r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
	//	r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
	//})
	//r = newRouter()
	//assert.PanicsWithValue(t, "web: illegal route : already has regexp route here :id", func() {
	//	r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
	//	r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
	//})
	// 参数冲突
	assert.PanicsWithValue(t, "web: illegal route : already has param route here :id", func() {
		r.addRoute(http.MethodGet, "/a/b/c/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/c/:name", mockHandler)
	})

}

func (r router) equal(y router) (string, bool) {
	for k, v := range r.trees {
		yv, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("web:router tree does not exist method: [%s]", k), false
		}
		str, ok := v.equal(yv)
		if !ok {
			return k + "-" + str, ok
		}
	}
	return "", true
}

func (n *node) equal(y *node) (string, bool) {
	if y == nil {
		return "the target node is nil", false
	}
	if n.path != y.path {
		return fmt.Sprintf("the path %s does not euqal to path %s, x %s, y %s",
			n.path, y.path, n.path, y.path), false

	}
	nhv := reflect.ValueOf(n.handler)
	yhv := reflect.ValueOf(y.handler)
	if nhv != yhv {
		return fmt.Sprintf("the handler function %s does not equal to function %s, x %s, y %s",
			n.path, y.path, nhv.Type().String(), yhv.Type().String()), false
	}

	if n.paramName != y.paramName {
		return fmt.Sprintf("the parameter name %s does not equal to parameter name %s, x %s, y %s",
			n.path, y.path, n.paramName, y.paramName), false
	}

	if n.typ != y.typ {
		return fmt.Sprintf("the node type %s does not equal to type %s, x %d, y %d",
			n.path, y.path, n.typ, y.typ), false
	}
	if len(n.children) != len(y.children) {
		return fmt.Sprintf("the amount of child node is not equal x %s, y %s", n.path, y.path), false
	}
	if len(n.children) == 0 {
		return "", true
	}

	if n.starChild != nil {
		str, ok := n.starChild.equal(y.starChild)
		if !ok {
			return fmt.Sprintf("%s '*' nodes do not match %s", n.path, str), false
		}
	}
	if n.paramChild != nil {
		str, ok := n.paramChild.equal(y.paramChild)
		if !ok {
			return fmt.Sprintf("%s path param do not match %s", n.path, str), false
		}
	}

	if n.regChild != nil {
		str, ok := n.regChild.equal(y.regChild)
		if !ok {
			return fmt.Sprintf("%s path param do not match %s", n.path, str), false
		}
	}

	for k, v := range n.children {
		yv, ok := y.children[k]
		if !ok {
			return fmt.Sprintf("%s the child node %s does not exist", n.path, k), false
		}
		str, ok := v.equal(yv)
		if !ok {
			return n.path + "-" + str, ok
		}
	}
	return "", true
}

func Test_router_FindRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodGet,
			path:   "/user/*/home",
		},
		{
			method: http.MethodPost,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},
		// 正则
		{
			method: http.MethodDelete,
			path:   "/reg/:id(.*)",
		},
		{
			method: http.MethodDelete,
			path:   "/:id([0-9]+)/home",
		},
	}

	mockHandler := func(ctx *Context) {}

	testCases := []struct {
		name   string
		method string
		path   string
		found  bool
		mi     *matchInfo
	}{
		{
			name:   "method not found",
			method: http.MethodHead,
		},
		{
			name:   "path not found",
			method: http.MethodGet,
			path:   "/abc",
		},
		{
			name:   "root",
			method: http.MethodGet,
			path:   "/",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "/",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "user",
			method: http.MethodGet,
			path:   "/user",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "user",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "no handler",
			method: http.MethodPost,
			path:   "/order",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path: "order",
				},
			},
		},
		{
			name:   "two layer",
			method: http.MethodPost,
			path:   "/order/create",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "create",
					handler: mockHandler,
				},
			},
		},
		// 通配符匹配
		{
			// 命中/order/*
			name:   "star match",
			method: http.MethodPost,
			path:   "/order/delete",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		{
			// 命中通配符在中间的
			// /user/*/home
			name:   "star in middle",
			method: http.MethodGet,
			path:   "/user/Tom/home",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "home",
					handler: mockHandler,
				},
			},
		},
		{
			// 比 /order/* 多了一段
			name:   "overflow",
			method: http.MethodPost,
			path:   "/order/delete/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		// 参数匹配
		{
			// 命中 /param/:id
			name:   ":id",
			method: http.MethodGet,
			path:   "/param/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    ":id",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /param/:id/*
			name:   ":id*",
			method: http.MethodGet,
			path:   "/param/123/abc",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /param/:id/detail
			name:   ":id*",
			method: http.MethodGet,
			path:   "/param/123/detail",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "detail",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /reg/:id(.*)
			name:   ":id(.*)",
			method: http.MethodDelete,
			path:   "/reg/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    ":id(.*)",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /:id([0-9]+)/home
			name:   ":id([0-9]+)",
			method: http.MethodDelete,
			path:   "/123/home",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    ":id(.*)",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 未命中 /:id([0-9]+)/home
			name:   "not :id([0-9]+)",
			method: http.MethodDelete,
			path:   "/abc/home",
		},
	}
	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mi, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.found, found)
			if !found {
				return
			}
			assert.Equal(t, tc.mi.pathParams, mi.pathParams)
			n := mi.n
			wantVal := reflect.ValueOf(tc.mi.n.handler)
			nVal := reflect.ValueOf(n.handler)
			assert.Equal(t, wantVal, nVal)
		})
	}
}

func Test_router_findRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodGet,
			path:   "/user/*/home",
		},
		{
			method: http.MethodPost,
			path:   "/order/*",
		},
		// 参数路由
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},

		// 正则
		{
			method: http.MethodDelete,
			path:   "/reg/:id(.*)",
		},
		{
			method: http.MethodDelete,
			path:   "/:id([0-9]+)/home",
		},
	}

	mockHandler := func(ctx *Context) {}

	testCases := []struct {
		name   string
		method string
		path   string
		found  bool
		mi     *matchInfo
	}{
		{
			name:   "method not found",
			method: http.MethodHead,
		},
		{
			name:   "path not found",
			method: http.MethodGet,
			path:   "/abc",
		},
		{
			name:   "root",
			method: http.MethodGet,
			path:   "/",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "/",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "user",
			method: http.MethodGet,
			path:   "/user",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "user",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "no handler",
			method: http.MethodPost,
			path:   "/order",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path: "order",
				},
			},
		},
		{
			name:   "two layer",
			method: http.MethodPost,
			path:   "/order/create",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "create",
					handler: mockHandler,
				},
			},
		},
		// 通配符匹配
		{
			// 命中/order/*
			name:   "star match",
			method: http.MethodPost,
			path:   "/order/delete",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		{
			// 命中通配符在中间的
			// /user/*/home
			name:   "star in middle",
			method: http.MethodGet,
			path:   "/user/Tom/home",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "home",
					handler: mockHandler,
				},
			},
		},
		{
			// 比 /order/* 多了一段
			name:   "overflow",
			method: http.MethodPost,
			path:   "/order/delete/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		// 参数匹配
		{
			// 命中 /param/:id
			name:   ":id",
			method: http.MethodGet,
			path:   "/param/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    ":id",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /param/:id/*
			name:   ":id*",
			method: http.MethodGet,
			path:   "/param/123/abc",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /param/:id/detail
			name:   ":id*",
			method: http.MethodGet,
			path:   "/param/123/detail",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "detail",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /reg/:id(.*)
			name:   ":id(.*)",
			method: http.MethodDelete,
			path:   "/reg/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    ":id(.*)",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /:id([0-9]+)/home
			name:   ":id([0-9]+)",
			method: http.MethodDelete,
			path:   "/123/home",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    ":id(.*)",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 未命中 /:id([0-9]+)/home
			name:   "not :id([0-9]+)",
			method: http.MethodDelete,
			path:   "/abc/home",
		},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mi, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.found, found)
			if !found {
				return
			}
			assert.Equal(t, tc.mi.pathParams, mi.pathParams)
			n := mi.n
			wantVal := reflect.ValueOf(tc.mi.n.handler)
			nVal := reflect.ValueOf(n.handler)
			assert.Equal(t, wantVal, nVal)
		})
	}

}

func Test_findRoute_Middleware(t *testing.T) {
	var mdlBuilder = func(i byte) Middleware {
		return func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				ctx.RespData = append(ctx.RespData, i)
				next(ctx)
			}
		}
	}
	mdlsRoute := []struct {
		method string
		path   string
		mdls   []Middleware
	}{
		{
			method: http.MethodGet,
			path:   "/a/b",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b')},
		},
		{
			method: http.MethodGet,
			path:   "/a/*",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('*')},
		},
		{
			method: http.MethodGet,
			path:   "/a/b/*",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b'), mdlBuilder('*')},
		},
		{
			method: http.MethodPost,
			path:   "/a/b/*",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b'), mdlBuilder('*')},
		},
		{
			method: http.MethodPost,
			path:   "/a/*/c",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('*'), mdlBuilder('c')},
		},
		{
			method: http.MethodPost,
			path:   "/a/b/c",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b'), mdlBuilder('c')},
		},
		{
			method: http.MethodDelete,
			path:   "/*",
			mdls:   []Middleware{mdlBuilder('*')},
		},
		{
			method: http.MethodDelete,
			path:   "/",
			mdls:   []Middleware{mdlBuilder('/')},
		},
	}
	r := newRouter()
	for _, mdlRoute := range mdlsRoute {
		r.addRoute(mdlRoute.method, mdlRoute.path, nil, mdlRoute.mdls...)
	}
	testCases := []struct {
		name   string
		method string
		path   string
		// 我们借助 ctx 里面的 RespData 字段来判断 middleware 有没有按照预期执行
		wantResp string
	}{
		{
			name:   "static, not match",
			method: http.MethodGet,
			path:   "/a",
		},
		{
			name:     "static, match",
			method:   http.MethodGet,
			path:     "/a/c",
			wantResp: "a*",
		},
		{
			name:     "static and star",
			method:   http.MethodGet,
			path:     "/a/b",
			wantResp: "a*ab",
		},
		{
			name:     "static and star",
			method:   http.MethodGet,
			path:     "/a/b/c",
			wantResp: "a*abab*",
		},
		{
			name:     "abc",
			method:   http.MethodPost,
			path:     "/a/b/c",
			wantResp: "a*cab*abc",
		},
		{
			name:     "root",
			method:   http.MethodDelete,
			path:     "/",
			wantResp: "/",
		},
		{
			name:     "root star",
			method:   http.MethodDelete,
			path:     "/a",
			wantResp: "/*",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mi, _ := r.findRoute(tc.method, tc.path)
			mdls := mi.mdls
			var root HandleFunc = func(ctx *Context) {
				// 使用 string 可读性比较高
				assert.Equal(t, tc.wantResp, string(ctx.RespData))
			}
			for i := len(mdls) - 1; i >= 0; i-- {
				root = mdls[i](root)
			}
			// 开始调度
			root(&Context{
				RespData: make([]byte, 0, len(tc.wantResp)),
			})
		})
	}

}

func handleFunc1(ctx *Context) {
	// 模拟处理逻辑
}

func handleFunc2(ctx *Context) {
	// 模拟处理逻辑
}

func setupRouter() *router {
	r := newRouter()
	r.addRoute("GET", "/static/path", handleFunc1)
	r.addRoute("GET", "/param/:id", handleFunc2)
	r.addRoute("GET", "/regexp/:name([a-z]+)", handleFunc1)
	r.addRoute("GET", "/wildcard/*", handleFunc2)
	return &r
}

func BenchmarkStaticRoute(b *testing.B) {
	r := setupRouter()
	for i := 0; i < b.N; i++ {
		_, _ = r.findRoute("GET", "/static/path")
	}
}

func BenchmarkParamRoute(b *testing.B) {
	r := setupRouter()
	for i := 0; i < b.N; i++ {
		_, _ = r.findRoute("GET", "/param/123")
	}
}

func BenchmarkRegexpRoute(b *testing.B) {
	r := setupRouter()
	for i := 0; i < b.N; i++ {
		_, _ = r.findRoute("GET", "/regexp/abc")
	}
}

func BenchmarkWildcardRoute(b *testing.B) {
	r := setupRouter()
	for i := 0; i < b.N; i++ {
		_, _ = r.findRoute("GET", "/wildcard/anything")
	}
}
