package components

import (
	"context"
	"fmt"

	"github.com/lariv-in/getters"
	"github.com/nyaruka/phonenumbers"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldPhone struct {
	Page
	Getter  getters.Getter
	Classes string
}

func (e FieldPhone) Build(ctx context.Context) Node {
	value := e.Getter(ctx)
	v, ok := value.(*phonenumbers.PhoneNumber)
	if !ok {
		vStr, ok := value.(string)
		if !ok {
			return ContainerError{Error: getters.GetterStatic(fmt.Errorf("Invalid value for a phone number: %s", value))}.Build(ctx)
		}
		val, err := phonenumbers.Parse(vStr, "IN")
		if err != nil {
			return ContainerError{Error: getters.GetterStatic(err)}.Build(ctx)
		}
		v = val
	}
	if v == nil {
		return ContainerError{Error: getters.GetterStatic(fmt.Errorf("Invalid value for a phone number"))}.Build(ctx)
	}
	return Div(Class(fmt.Sprintf("text-xl font-semibold text-primary %s", e.Classes)), Text(phonenumbers.Format(v, phonenumbers.E164 )))
}
