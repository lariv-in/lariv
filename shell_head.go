package lago

import (
	"github.com/lariv-in/lago/components"
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

func init() {
	_ = components.RegistryShellHeadNodes.Register("core.Title", html.TitleEl(gomponents.Text("Lago")))
}
