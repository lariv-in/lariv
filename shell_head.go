package lariv

import (
	"github.com/lariv-in/lariv/components"
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

func init() {
	_ = components.RegistryShellHeadNodes.Register("core.Title", html.TitleEl(gomponents.Text("Lariv")))
}
