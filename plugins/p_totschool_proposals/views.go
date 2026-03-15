package p_totschool_proposals

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alnah/go-md2pdf"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
	"gorm.io/gorm"
)

func ProposalQueryPatcher(v views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	user := r.Context().Value("$user").(p_users.User)
	if !(user.IsSuperuser || user.Role.Name == "totschool_admin") {
		query = query.Where("CreatedByID = ?", user.ID)
	}
	return query
}

// ProposalFormPatcher enriches form data for CRUD handlers:
// - sets CreatedByID from the authenticated user
// - flattens questionnaire answers into the Answers JSON field
func ProposalFormPatcher(v views.View, r *http.Request, formData map[string]any) map[string]any {
	user := r.Context().Value("$user").(p_users.User)
	formData["CreatedByID"] = user.ID

	var items []QAItem
	for i := 0; i < len(QUESTIONS); i++ {
		key := fmt.Sprintf("answers[%d]", i)
		raw := formData[key]
		answer, _ := raw.(string)
		items = append(items, QAItem{
			Question: QUESTIONS[i],
			Answer:   answer,
		})
		delete(formData, key)
	}

	var p Proposal
	_ = p.SetAnswers(items)
	formData["Answers"] = p.Answers

	return formData
}


func generateHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)
		user := r.Context().Value("$user").(p_users.User)

		var proposal Proposal
		if err := db.Where("id = ?", idStr).First(&proposal).Error; err != nil {
			http.NotFound(w, r)
			return
		}

		answersText, err := proposal.FormatAnswersForAI()
		if err != nil || answersText == "" || len(answersText) < 10 {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(`<div class="alert alert-error">No answers provided. Please fill in the questionnaire first.</div>`))
				return
			}
			http.Error(w, "No answers provided. Please fill in the questionnaire first.", http.StatusBadRequest)
			return
		}

		advisorName := user.Name
		currentDate := time.Now().Format("January 02, 2006")
		clientName := proposal.Title

		systemPrompt := fmt.Sprintf(`You are a Wealth Strategist and Family Financial Advisor generating a comprehensive Family Financial Goals Accomplishment Presentation.

ADVISOR DETAILS:
- Name: %s
- Date: %s

REPORT STRUCTURE (Generate each section with markdown headers):

## 1. COVER PAGE
- Advisor name and title (Wealth Strategist, Family Financial Advisor)
- Client family name
- Purpose: "Family Financial Goals Accomplishment Presentation - For Dream Fulfillment, Goals Accomplishment, and Absolute Financial Freedom"

## 2. TABLE OF CONTENTS
- List all sections with page references

## 3. FAMILY BASIC UNDERSTANDING
- Family head details (name, age)
- Spouse details (name, age)
- Children details (names, ages)
- Residence type (owned/rented, location)
- Family values and concerns
- Interests and commitments

## 4. PRESENT PROFESSIONAL SITUATION
- Occupation and organization
- Industry
- Monthly/annual income
- Spouse income (if any)
- Monthly/annual expenditure

## 5. FAMILY DETAILED UNDERSTANDING
- Who is the bread earner
- Family dependency analysis
- Financial decision-making responsibility
- Family values and lifestyle assessment

## 6. CHILDREN'S LIFE PLAN AND PROJECTION
- For each child: education milestones, future expenses, provisions needed
- Higher education cost estimates
- Timeline for graduation/post-graduation expenses

## 7. SWOT ANALYSIS
- STRENGTHS: Income stability, assets, savings habits, low liabilities
- WEAKNESSES: Gaps in financial planning, low safety buckets
- OPPORTUNITIES: Savings potential, course correction possibilities
- THREATS: Life uncertainties, potential loan impacts

## 8. DREAMS AND GOALS - FAMILY DREAM PROFILE
- Financial freedom goals
- Property investments and plans
- Dream home aspirations with cost estimates
- Dream car goals
- Quality family time and vacation plans
- Dream income and pension targets

## 9. KEY CONCERN AREAS
- Identified gaps in current financial planning
- Areas needing immediate attention

## 10. RECOMMENDATIONS AND SOLUTIONS
- Specific insurance recommendations
- Investment strategies
- Risk coverage analysis
- Premium estimates
- Key benefits of proposed solutions

## 11. RETIREMENT PLANNING
- Pension goals
- Post-retirement income strategies
- Lifestyle maintenance plans

## 12. ACTION PLAN - HOW TO GET STARTED
- Immediate steps
- Timeline for implementation
- Review schedule

RULES:
1. Output in markdown format with clear section headers
2. Use the client's actual data from the questionnaire
3. Be professional, thorough, and personalized
4. Include specific numbers and calculations where possible
5. Make recommendations actionable and realistic
6. Address all family members by name where applicable
7. Consider Indian financial context (LIC, mutual funds, property investment)`, advisorName, currentDate)

		userPrompt := fmt.Sprintf(`Generate a comprehensive Family Financial Goals Accomplishment Presentation for:

CLIENT: %s

QUESTIONNAIRE RESPONSES:
%s

Create a detailed, personalized financial proposal following the report structure provided.`, clientName, answersText)

		Generate(db, proposal.ID, userPrompt, systemPrompt)

		lago.NewRedirectView("proposals.DetailRoute", map[string]getters.Getter{
			"id": getters.GetterStatic(fmt.Sprintf("%d", proposal.ID)),
		}).ServeHTTP(w, r)
	})
}

func cancelHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)

		var proposal Proposal
		if err := db.Where("id = ?", idStr).First(&proposal).Error; err != nil {
			http.NotFound(w, r)
			return
		}

		if proposal.GenerationID != nil {
			CancelGeneration(db, proposal.ID)
		}

		lago.NewRedirectView("proposals.DetailRoute", map[string]getters.Getter{
			"id": getters.GetterStatic(idStr),
		}).ServeHTTP(w, r)
	})
}

func aiEditFormHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)

		var proposal Proposal
		if err := db.Where("id = ?", idStr).First(&proposal).Error; err != nil {
			http.NotFound(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), "proposal", getters.MapFromStruct(&proposal))
		v.RenderPage(w, r.WithContext(ctx))
	})
}

func aiEditHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)

		var proposal Proposal
		if err := db.Where("id = ?", idStr).First(&proposal).Error; err != nil {
			http.NotFound(w, r)
			return
		}

		content := r.FormValue("generated_content")
		instructions := r.FormValue("instructions")
		if content == "" || instructions == "" {
			http.Error(w, "Missing content or instructions", http.StatusBadRequest)
			return
		}

		systemPrompt := `You are an expert proposal writer and editor. Your task is to edit or rewrite the given proposal according to the user's instructions.

Rules:
1. Only output the edited proposal content - no formatting outside of what is requested.
2. If the user asks for markdown, provide it. The input is likely in markdown.
3. Preserve the general structure of the markdown unless instructed otherwise
4. Maintain a professional tone unless instructed otherwise
5. Keep all factual information unchanged unless specifically asked to modify them
6. Ensure the output is valid markdown
7. DO NOT surround your response in a Markdown code block. Output the markdown string directly.`

		userPrompt := fmt.Sprintf("Here is the current proposal markdown:\n\n---\n%s\n---\n\nPlease edit this proposal according to these instructions: %s\n\nOutput only the edited markdown, nothing else.", content, instructions)

		Generate(db, proposal.ID, userPrompt, systemPrompt)

		lago.NewRedirectView("proposals.DetailRoute", map[string]getters.Getter{
			"id": getters.GetterStatic(idStr),
		}).ServeHTTP(w, r)
	})
}

func exportPdfHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)

		var proposal Proposal
		if err := db.Where("id = ?", idStr).First(&proposal).Error; err != nil {
			http.NotFound(w, r)
			return
		}

		if proposal.GeneratedContent == "" {
			http.Error(w, "No proposal content to export. Please generate the proposal first.", http.StatusUnprocessableEntity)
			return
		}

		conv, err := md2pdf.NewConverter()
		if err != nil {
			http.Error(w, "PDF converter unavailable", http.StatusInternalServerError)
			return
		}
		defer conv.Close()

		result, err := conv.Convert(r.Context(), md2pdf.Input{
			Markdown: proposal.GeneratedContent,
		})
		if err != nil {
			http.Error(w, "PDF generation failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, proposal.Title))
		w.Write(result.PDF)
	})
}

func init() {
	lago.RegistryView.Register("proposals.ListView", p_users.AuthenticationMiddleware(
		views.ListView[Proposal]("proposals")(lago.GetPageView("proposals.ProposalTable")).WithQueryPatcher(ProposalQueryPatcher)))

	lago.RegistryView.Register("proposals.DetailView", p_users.AuthenticationMiddleware(
		views.DetailView[Proposal]("proposal")(
			lago.GetPageView("proposals.ProposalDetail"))))

	lago.RegistryView.Register("proposals.CreateView", p_users.AuthenticationMiddleware(
		views.CreateView[Proposal](lago.GetterRoutePath("proposals.DetailRoute", map[string]getters.Getter{"id": getters.GetterKey("$id")}))(lago.GetPageView("proposals.ProposalCreateForm")).WithFormPatcher(ProposalFormPatcher)))

	lago.RegistryView.Register("proposals.UpdateView", p_users.AuthenticationMiddleware(
		views.DetailView[Proposal]("proposal")(
			views.UpdateView[Proposal](lago.GetterRoutePath("proposals.DetailRoute", map[string]getters.Getter{"id": getters.GetterKey("$id")}))(lago.GetPageView("proposals.ProposalUpdateForm")).WithFormPatcher(ProposalFormPatcher))))

	lago.RegistryView.Register("proposals.DeleteView", p_users.AuthenticationMiddleware(
		views.DetailView[Proposal]("proposal")(
			views.DeleteView[Proposal](lago.GetterRoutePath("proposals.ListRoute", nil))(
				lago.GetPageView("proposals.ProposalDeleteForm")))))

	lago.RegistryView.Register("proposals.GenerateView", p_users.AuthenticationMiddleware(
		lago.GetPageView("proposals.ProposalDetail").WithMethod(http.MethodPost, generateHandler)))

	lago.RegistryView.Register("proposals.CancelView", p_users.AuthenticationMiddleware(
		lago.GetPageView("proposals.ProposalDetail").WithMethod(http.MethodPost, cancelHandler)))

	lago.RegistryView.Register("proposals.AiEditFormView", p_users.AuthenticationMiddleware(
		lago.GetPageView("proposals.AiEditModal").WithMethod(http.MethodGet, aiEditFormHandler)))

	lago.RegistryView.Register("proposals.AiEditView", p_users.AuthenticationMiddleware(
		lago.GetPageView("proposals.AiEditModal").WithMethod(http.MethodPost, aiEditHandler)))

	lago.RegistryView.Register("proposals.ExportPdfView", p_users.AuthenticationMiddleware(
		lago.GetPageView("proposals.ProposalDetail").WithMethod(http.MethodGet, exportPdfHandler)))
}
