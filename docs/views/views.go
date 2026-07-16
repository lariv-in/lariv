// Package views contains explanations and code examples for view controllers in Lariv.
//
// # View Controllers (views.go)
//
// Views coordinate middleware handler layers (views.Layer) and page layouts (components.PageInterface).
//
// # Custom View Implementations (Warning)
//
// Implementing a custom View struct or overriding ServeHTTP manually is almost never recommended.
// Instead, compose standard views.View structs with custom middleware handler layers.
//
// # View and Layers Composition Example
//
//	package myplugin
//
//	import (
//		"github.com/lariv-in/lariv/registry"
//		"github.com/lariv-in/lariv/views"
//	)
//
//	func pluginViews() lariv.PluginFeatures[*views.View] {
//		return lariv.PluginFeatures[*views.View]{
//			Entries: []registry.Pair[string, *views.View]{
//				{
//					Key: "blog.PostDetail",
//					Value: &views.View{
//						PageName:   "blog.detail",
//						PageLookup: pluginPageResolver,
//						Layers: []registry.Pair[string, views.Layer]{
//							// 1. PathLayer parses paths variables (e.g. {id}) to context
//							{
//								Key: "path_id",
//								Value: views.PathLayer{
//									Names: []string{"id"},
//								},
//							},
//							// 2. DetailLayer preloads database rows
//							{
//								Key: "db_detail",
//								Value: views.LayerDetail[BlogPost]{
//									PathParamKey: getters.Static("id"),
//									Key:          getters.Static("$post"),
//								},
//							},
//						},
//					},
//				},
//			},
//		}
//	}
//
// # Views Package Reference
//
// For transactional view rendering pipelines, refer to the [github.com/lariv-in/lariv/views] package documentation.
package views
