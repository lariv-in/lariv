package lago

func BuildAllRegistries() {
	_ = RegistryCommand.All()
	_ = RegistryDBInit.All()
	_ = RegistryGenerator.All()
	_ = RegistryLayer.All()
	_ = RegistryPage.All()
	_ = RegistryPlugin.All()
	_ = RegistryRoute.All()
	_ = RegistryView.All()
	_ = RegistryConfig.All()
}
