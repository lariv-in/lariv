package forms

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/forms/"

// FormFieldsObjectListContextKey is set on form detail GET (AttachFormFieldsObjectListContext) for the fields DataTable.
const FormFieldsObjectListContextKey = "form_fields_table"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("forms", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "clipboard-document-list",
		URL:         u,
		VerboseName: "Forms",
	}); err != nil {
		log.Panic(err)
	}
}
