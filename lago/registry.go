package lago

func BuildAllRegistries() {
	RegistryCommand.Build()
	RegistryGenerator.Build()
	RegistryLayer.Build()
	RegistryPage.Build()
	RegistryPlugin.Build()
	RegistryRoute.Build()
	RegistryView.Build()
	RegistryConfig.Build()
}
