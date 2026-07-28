package p_website

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/plugins/p_filesystem"
	"github.com/lariv-in/lariv/views"
)

// createBlankPagePatcher creates a blank HTML VNode when CreateNewPage is checked.
type createBlankPagePatcher struct{}

func (createBlankPagePatcher) Patch(_ views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	if formErrors == nil {
		formErrors = map[string]error{}
	}
	if formData == nil {
		formData = map[string]any{}
	}

	createNew, _ := formData["CreateNewPage"].(bool)
	delete(formData, "CreateNewPage")
	newPageName, _ := formData["NewPageName"].(string)
	delete(formData, "NewPageName")

	if !createNew {
		pageID, hasPageID := formData["PageID"]
		if !hasPageID || pageID == nil || pageID == "" || pageID == uint(0) {
			formErrors["PageID"] = errors.New("template page is required")
		}
		return formData, formErrors
	}

	newPageName = strings.TrimSpace(newPageName)
	if newPageName == "" {
		formErrors["NewPageName"] = errors.New("filename is required")
		return formData, formErrors
	}
	if !IsEditableHTMLName(newPageName) {
		formErrors["NewPageName"] = fmt.Errorf("filename must end in .html, .htm, or .tmpl (got %q)", filepath.Ext(newPageName))
		return formData, formErrors
	}

	db, err := getters.DBFromContext(r.Context())
	if err != nil {
		formErrors["NewPageName"] = err
		return formData, formErrors
	}

	parent, err := p_filesystem.EnsureDirectoryPath(db, Config.NewPageRootDir)
	if err != nil {
		formErrors["NewPageName"] = fmt.Errorf("new page root directory: %w", err)
		return formData, formErrors
	}

	node, err := p_filesystem.CreateVNodeFromReader(db, newPageName, strings.NewReader(blankPageStarterHTML), parent)
	if err != nil {
		formErrors["NewPageName"] = err
		return formData, formErrors
	}
	formData["PageID"] = node.ID
	return formData, formErrors
}
