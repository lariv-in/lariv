package p_totschool_proposals

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
	"gorm.io/gorm"
	"maragu.dev/gomponents"
	g "maragu.dev/gomponents/html"
)

func proposalScope(db *gorm.DB, user p_users.User) *gorm.DB {
	if user.IsSuperuser {
		return db
	}
	var roleName string
	db.Model(&p_users.Role{}).Where("id = ?", user.RoleID).Select("name").Scan(&roleName)
	if roleName == "totschool_admin" {
		return db
	}
	return db.Where("created_by_id = ?", user.ID)
}

func getProposalOr404(w http.ResponseWriter, r *http.Request, db *gorm.DB, idStr string, user p_users.User) *Proposal {
	var p Proposal
	err := proposalScope(db, user).Where("id = ?", idStr).First(&p).Error
	if err != nil {
		http.NotFound(w, r)
		return nil
	}
	return &p
}

func listHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := r.Context().Value("$db").(*gorm.DB)
		user := r.Context().Value("$user").(p_users.User)

		query := proposalScope(db, user).Model(&Proposal{})

		pageStr := r.URL.Query().Get("page")
		pageNum := 1
		if pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				pageNum = p
			}
		}
		pageSize := 12

		if title := r.URL.Query().Get("title"); title != "" {
			query = query.Where("title LIKE ?", "%"+title+"%")
		}
		if sort := r.URL.Query().Get("sort"); sort != "" {
			switch sort {
			case "title", "created_at", "updated_at",
				"title desc", "created_at desc", "updated_at desc":
				query = query.Order(sort)
			}
		}

		var total int64
		if err := query.Count(&total).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var results []Proposal
		err := query.Limit(pageSize).Offset((pageNum - 1) * pageSize).Order("created_at DESC").Find(&results).Error
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		numPages := int((total + int64(pageSize) - 1) / int64(pageSize))
		objectList := components.ObjectList[Proposal]{
			Items:    results,
			Number:   pageNum,
			NumPages: numPages,
			Total:    total,
		}

		ctx := context.WithValue(r.Context(), "proposals", objectList)
		queryMap := map[string]any{}
		for param, values := range r.URL.Query() {
			if len(values) > 0 && values[0] != "" {
				queryMap[param] = values[0]
			}
		}
		ctx = context.WithValue(ctx, "$get", queryMap)
		v.RenderPage(w, r.WithContext(ctx))
	})
}

func detailHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)
		user := r.Context().Value("$user").(p_users.User)

		proposal := getProposalOr404(w, r, db, idStr, user)
		if proposal == nil {
			return
		}

		ctx := r.Context()

		// If generation is in progress, signal pending to the template
		if proposal.GenerationID != nil {
			ctx = context.WithValue(ctx, "generation_pending", true)
		}

		proposalMap := getters.MapFromStruct(proposal)
		items, _ := proposal.ParseAnswers()
		proposalMap["Answers"] = items
		ctx = context.WithValue(ctx, "proposal", proposalMap)
		v.RenderPage(w, r.WithContext(ctx))
	})
}

func createHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := r.Context().Value("$db").(*gorm.DB)
		user := r.Context().Value("$user").(p_users.User)

		if r.Method == http.MethodGet {
			inMap := map[string]any{"Title": ""}
			for i := 0; i < len(QUESTIONS); i++ {
				inMap[fmt.Sprintf("answer_%d", i)] = ""
			}
			ctx := context.WithValue(r.Context(), "$in", inMap)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		title := r.FormValue("title")
		var answers []QAItem
		for i := 0; i < len(QUESTIONS); i++ {
			key := fmt.Sprintf("answers[%d]", i)
			answers = append(answers, QAItem{Question: QUESTIONS[i], Answer: r.FormValue(key)})
		}

		proposal := Proposal{
			Title:       title,
			CreatedByID: user.ID,
		}
		if err := proposal.SetAnswers(answers); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := db.Create(&proposal).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		redirectTo(w, r, fmt.Sprintf(AppUrl+"%d/", proposal.ID))
	})
}

func updateHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)
		user := r.Context().Value("$user").(p_users.User)

		proposal := getProposalOr404(w, r, db, idStr, user)
		if proposal == nil {
			return
		}

		if r.Method == http.MethodGet {
			proposalMap := getters.MapFromStruct(proposal)
			items, _ := proposal.ParseAnswers()
			for i := 0; i < len(QUESTIONS); i++ {
				val := ""
				if i < len(items) {
					val = items[i].Answer
				}
				proposalMap[fmt.Sprintf("answer_%d", i)] = val
			}
			ctx := context.WithValue(r.Context(), "$in", proposalMap)
			ctx = context.WithValue(ctx, "proposal", proposalMap)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		proposal.Title = r.FormValue("title")
		var answers []QAItem
		for i := 0; i < len(QUESTIONS); i++ {
			key := fmt.Sprintf("answers[%d]", i)
			answers = append(answers, QAItem{Question: QUESTIONS[i], Answer: r.FormValue(key)})
		}
		if err := proposal.SetAnswers(answers); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := db.Save(proposal).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		redirectTo(w, r, fmt.Sprintf(AppUrl+"%d/", proposal.ID))
	})
}

func deleteHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)
		user := r.Context().Value("$user").(p_users.User)

		proposal := getProposalOr404(w, r, db, idStr, user)
		if proposal == nil {
			return
		}

		if r.Method == http.MethodGet {
			ctx := context.WithValue(r.Context(), "proposal", getters.MapFromStruct(proposal))
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		if err := db.Delete(proposal).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		redirectTo(w, r, AppUrl)
	})
}

func redirectTo(w http.ResponseWriter, r *http.Request, url string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
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

		proposal := getProposalOr404(w, r, db, idStr, user)
		if proposal == nil {
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

		redirectTo(w, r, fmt.Sprintf(AppUrl+"%d/", proposal.ID))
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
		user := r.Context().Value("$user").(p_users.User)

		proposal := getProposalOr404(w, r, db, idStr, user)
		if proposal == nil {
			return
		}

		if proposal.GenerationID != nil {
			CancelGeneration(db, proposal.ID)
		}

		redirectTo(w, r, fmt.Sprintf(AppUrl+"%d/", proposal.ID))
	})
}

func aiEditFormHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)
		user := r.Context().Value("$user").(p_users.User)

		proposal := getProposalOr404(w, r, db, idStr, user)
		if proposal == nil {
			return
		}

		ctx := context.WithValue(r.Context(), "proposal", getters.MapFromStruct(proposal))
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
		user := r.Context().Value("$user").(p_users.User)

		proposal := getProposalOr404(w, r, db, idStr, user)
		if proposal == nil {
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

		redirectTo(w, r, fmt.Sprintf(AppUrl+"%s/", idStr))
	})
}

func exportPdfHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)
		user := r.Context().Value("$user").(p_users.User)

		proposal := getProposalOr404(w, r, db, idStr, user)
		if proposal == nil {
			return
		}

		if proposal.GeneratedContent == "" {
			http.Error(w, "No proposal content to export. Please generate the proposal first.", http.StatusUnprocessableEntity)
			return
		}

		page := g.HTML(
			g.Head(
				g.Meta(g.Charset("utf-8")),
				g.TitleEl(gomponents.Text(proposal.Title)),
				g.StyleEl(g.Type("text/css"), gomponents.Raw(`
body { font-family: Georgia, serif; max-width: 800px; margin: 40px auto; padding: 0 20px; line-height: 1.6; color: #333; }
h1, h2, h3 { margin-top: 1.5em; }
table { border-collapse: collapse; width: 100%; }
th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
@media print { body { margin: 0; } }
`)),
			),
			g.Body(gomponents.Raw(components.RenderMarkdown(proposal.GeneratedContent))),
		)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = page.Render(w)
	})
}

func init() {
	lago.RegistryView.Register("proposals.ListView", p_users.AuthMiddleware(
		lago.GetPageView("proposals.ProposalTable").WithMethod(http.MethodGet, listHandler)))

	lago.RegistryView.Register("proposals.DetailView", p_users.AuthMiddleware(
		lago.GetPageView("proposals.ProposalDetail").WithMethod(http.MethodGet, detailHandler)))

	lago.RegistryView.Register("proposals.CreateView", p_users.AuthMiddleware(
		lago.GetPageView("proposals.ProposalCreateForm").WithMethod(http.MethodGet, createHandler).WithMethod(http.MethodPost, createHandler)))

	lago.RegistryView.Register("proposals.UpdateView", p_users.AuthMiddleware(
		lago.GetPageView("proposals.ProposalUpdateForm").WithMethod(http.MethodGet, updateHandler).WithMethod(http.MethodPost, updateHandler)))

	lago.RegistryView.Register("proposals.DeleteView", p_users.AuthMiddleware(
		lago.GetPageView("proposals.ProposalDeleteForm").WithMethod(http.MethodGet, deleteHandler).WithMethod(http.MethodPost, deleteHandler)))

	lago.RegistryView.Register("proposals.GenerateView", p_users.AuthMiddleware(
		lago.GetPageView("proposals.ProposalDetail").WithMethod(http.MethodPost, generateHandler)))

	lago.RegistryView.Register("proposals.CancelView", p_users.AuthMiddleware(
		lago.GetPageView("proposals.ProposalDetail").WithMethod(http.MethodPost, cancelHandler)))

	lago.RegistryView.Register("proposals.AiEditFormView", p_users.AuthMiddleware(
		lago.GetPageView("proposals.AiEditModal").WithMethod(http.MethodGet, aiEditFormHandler)))

	lago.RegistryView.Register("proposals.AiEditView", p_users.AuthMiddleware(
		lago.GetPageView("proposals.AiEditModal").WithMethod(http.MethodPost, aiEditHandler)))

	lago.RegistryView.Register("proposals.ExportPdfView", p_users.AuthMiddleware(
		lago.GetPageView("proposals.ProposalDetail").WithMethod(http.MethodGet, exportPdfHandler)))
}
