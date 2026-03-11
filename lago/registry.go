package lago

func BuildAllRegistries() {
	RegistryGenerator.Build()
	RegistryMiddleware.Build()
	RegistryPage.Build()
	RegistryPlugins.Build()
	RegistryRoute.Build()
	RegistryView.Build()
	RegistryConfig.Build()
}
