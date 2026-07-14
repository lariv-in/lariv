// Package docs contains explanations and code examples for plugin-specific files in Lago.
//
// # Lago Directory Structure
//
// This document outlines the standard file and directory structure of a Lago-based web application.
// Lago is designed to be highly modular and plugin-centric, isolating layouts, routes, models, views,
// and database migrations within individual plugin directories.
//
// # Project Structure Overview
//
// A typical application using the Lago framework follows this layout:
//
//	<project root>
//	├── config.toml
//	├── go.mod
//	├── go.sum
//	├── main.go
//	└── plugins/
//	    └── <plugin name>/
//	        ├── templates/
//	        │   └── <go html/template compatible files for rendering html directly>
//	        ├── migrations/
//	        │   └── <migration sql files created by goose>
//	        ├── app.go
//	        ├── config.go
//	        ├── pages.go
//	        ├── migrations.go
//	        ├── routes.go
//	        ├── models.go
//	        ├── commands.go
//	        └── views.go
//
// # Root Level Components
//
//   - config.toml: The main configuration file for the application. It contains settings like debug mode, database type, server listen address, and plugin-specific options (e.g., mail server settings, PWA settings, user authentication configs).
//   - go.mod & go.sum: Standard Go module definition files managing project dependencies.
//   - main.go: The application entry point. It handles loading configurations, registering the list of active plugins, and booting the Cobra CLI wrapper by calling lago.Start(...).
//
// # The plugins Directory
//
// All functionality in a Lago application is packaged inside plugins. Each directory inside plugins/ represents a self-contained feature module.
//
// Below is a detailed breakdown of each file and folder inside a plugin directory.
//
//   - templates/ (Directory): Stores Go HTML template files. Refer to [components.TemplateComponent] and [components.TemplateFSComponent].
//   - migrations/ (Directory): Contains sequential database schema migration files (Up/Down SQL).
//   - app.go: The core definition and entrypoint of the plugin. See [github.com/lariv-in/lago/docs/app] for example plugin setups.
//   - config.go: Defines configuration schemas and settings specific to the plugin. See [github.com/lariv-in/lago/docs/config] for configuration examples.
//   - pages.go: Declares pages, dashboards, navigation menus, and widgets used in the UI layer. See [github.com/lariv-in/lago/docs/pages] for page examples.
//   - migrations.go: Integrates the embedded migration SQL scripts with the framework's migration system. See [github.com/lariv-in/lago/docs/migrations] for migrations examples.
//   - routes.go: Sets up routing endpoints, custom endpoints, API handlers, or patches existing routes in the system. See [github.com/lariv-in/lago/docs/routes] for routing examples.
//   - models.go: Defines the database models and entities used by GORM or other database layers. See [github.com/lariv-in/lago/docs/models] for model examples.
//   - commands.go: Exposes plugin-specific commands to the main application's command-line interface. See [github.com/lariv-in/lago/docs/commands] for CLI command examples.
//   - views.go: Declares view configurations and transactional view pipelines for rendering HTML layouts. See [github.com/lariv-in/lago/docs/views] for view definition examples.
//
// # Architectural & Utility Components
//
// Beyond plugin-specific files, the Lago framework utilizes several foundational middleware and UI layout concepts:
//
//   - layers.go: Explains the view request-handling layers (middleware) that compile into transactional request pipelines. See [github.com/lariv-in/lago/docs/layers] for middleware layer examples.
//   - querypatchers.go: Documents query modifiers applied to GORM database operations. See [github.com/lariv-in/lago/docs/querypatchers] for query patcher examples.
//   - components.go: Describes the visual UI component library powering Lago pages. See [github.com/lariv-in/lago/docs/components] for UI component examples.
package docs
