package lago

func BuildAllRegistries() {
	RegistryGenerator.Build()
	RegistryMiddleware.Build()
	RegistryPage.Build()
	RegistryPlugin.Build()
	RegistryRoute.Build()
	RegistryView.Build()
	RegistryConfig.Build()
}
