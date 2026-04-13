package p_totschool_tally

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
)

func tallyCommonFields() []components.PageInterface {
	return []components.PageInterface{
		components.ContainerRow{
			Classes: "grid grid-cols-1 md:grid-cols-2 gap-4",
			Children: []components.PageInterface{
				components.InputNumber[int]{Name: "Visits", Label: "Visits", Required: true, Getter: getters.Key[int]("$in.Visits")},
				components.InputNumber[int]{Name: "Appointments", Label: "Appointments", Required: true, Getter: getters.Key[int]("$in.Appointments")},
				components.InputNumber[int]{Name: "Leads", Label: "Leads", Required: true, Getter: getters.Key[int]("$in.Leads")},
				components.InputNumber[int]{Name: "Presentations", Label: "Presentations", Required: true, Getter: getters.Key[int]("$in.Presentations")},
				components.InputNumber[int]{Name: "Demos", Label: "Demonstrations", Required: true, Getter: getters.Key[int]("$in.Demos")},
				components.InputNumber[int]{Name: "Letters", Label: "Follow Up Letters Sent", Required: true, Getter: getters.Key[int]("$in.Letters")},
				components.InputNumber[int]{Name: "FollowUps", Label: "Follow Ups", Required: true, Getter: getters.Key[int]("$in.FollowUps")},
				components.InputNumber[int]{Name: "Proposals", Label: "Proposals Given", Required: true, Getter: getters.Key[int]("$in.Proposals")},
				components.InputNumber[int]{Name: "Policies", Label: "Policies Sold", Required: true, Getter: getters.Key[int]("$in.Policies")},
				components.InputNumber[int]{Name: "Premium", Label: "Premium", Required: true, Getter: getters.Key[int]("$in.Premium")},
			},
		},
	}
}
