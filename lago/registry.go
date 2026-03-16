package lago

func BuildAllRegistries() {
	RegistryCommand.Build()
	RegistryGenerator.Build()
	RegistryMiddleware.Build()
	RegistryPage.Build()
	RegistryPlugin.Build()
	RegistryRoute.Build()
	RegistryView.Build()
	RegistryConfig.Build()
}
