package getters

import (
	"context"

	"maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

// FormAttr returns attributes for FormComponent.Attr: optional HTTP method and
// Alpine @submit.prevent from onSubmit (see [FormSubmit], [FormSubmitGet], [FormSubmitCloseModal]).
// Pass an empty method to omit; pass nil onSubmit to omit @submit.prevent.
func FormAttr(method string, onSubmit Getter[string]) Getter[gomponents.Node] {
	return func(ctx context.Context) (gomponents.Node, error) {
		var group gomponents.Group
		if method != "" {
			group = append(group, ghtml.Method(method))
		}
		if onSubmit != nil {
			expr, err := onSubmit(ctx)
			if err != nil {
				return nil, err
			}
			if expr != "" {
				group = append(group, gomponents.Attr("@submit.prevent", expr))
			}
		}
		if len(group) == 0 {
			return nil, nil
		}
		return group, nil
	}
}
