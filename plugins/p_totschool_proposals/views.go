package p_totschool_proposals

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/alnah/go-md2pdf"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// proposalDB returns $db from context; logs and returns nil if missing or wrong type.
func proposalDB(r *http.Request, op string) *gorm.DB {
	raw := r.Context().Value("$db")
	db, ok := raw.(*gorm.DB)
	if !ok || db == nil {
		slog.Error(op+": missing or invalid $db in context",
			"dbType", fmt.Sprintf("%T", raw))
		return nil
	}
	return db
}

type proposalQueryPatcher struct{}

func (proposalQueryPatcher) Patch(v views.View, r *http.Request, query gorm.ChainInterface[Proposal]) gorm.ChainInterface[Proposal] {
	rawUser := r.Context().Value("$user")
	user, ok := rawUser.(p_users.User)
	if !ok {
		slog.Error("proposalQueryPatcher: missing or invalid $user in context",
			"pageName", v.PageName,
			"userType", fmt.Sprintf("%T", rawUser))
		return query
	}
	role, _ := r.Context().Value("$role").(string)
	if user.IsSuperuser || role == "totschool_admin" {
		return query
	}
	return query.Where("created_by_id = ?", user.ID)
}

// proposalDetailCtxLayer enriches the detail view context for a proposal.
// It expects LayerDetail to have already loaded the Proposal under "proposal"
// and sets GenerationPending from GenerationID.
type proposalDetailCtxLayer struct{}

func (proposalDetailCtxLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		raw := ctx.Value("proposal")
		proposal, ok := raw.(Proposal)
		if !ok {
			slog.Error("proposalDetailCtxLayer: missing or invalid proposal in context",
				"proposalType", fmt.Sprintf("%T", raw))
			next.ServeHTTP(w, r)
			return
		}

		if proposal.GenerationID != nil {
			ctx = context.WithValue(ctx, "GenerationPending", true)
		} else {
			ctx = context.WithValue(ctx, "GenerationPending", false)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type proposalFormPatcher struct{}

func (proposalFormPatcher) Patch(v views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	rawUser := r.Context().Value("$user")
	user, ok := rawUser.(p_users.User)
	if !ok {
		slog.Error("proposalFormPatcher: missing or invalid $user in context",
			"pageName", v.PageName,
			"userType", fmt.Sprintf("%T", rawUser))
		return formData, formErrors
	}
	formData["CreatedByID"] = user.ID
	return formData, formErrors
}

func redirectProposalDetail(w http.ResponseWriter, r *http.Request, idStr string) bool {
	url, err := getters.IfOr(lago.RoutePath("proposals.DetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(idStr)),
	}), r.Context(), "")
	if err != nil || url == "" {
		http.NotFound(w, r)
		return false
	}
	lago.Redirect(w, r, url)
	return true
}

func generateHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			slog.Warn("generateHandler: method not allowed",
				"method", r.Method,
				"path", r.URL.Path,
				"pageName", v.PageName)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		db := proposalDB(r, "generateHandler")
		if db == nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		rawUser := r.Context().Value("$user")
		user, ok := rawUser.(p_users.User)
		if !ok {
			slog.Error("generateHandler: missing or invalid $user in context",
				"pageName", v.PageName,
				"userType", fmt.Sprintf("%T", rawUser))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		proposal, err := gorm.G[Proposal](db).Where("id = ?", idStr).First(r.Context())
		if err != nil {
			slog.Error("generateHandler: proposal not found or DB error",
				"error", err,
				"id", idStr,
				"pageName", v.PageName)
			http.NotFound(w, r)
			return
		}

		answersText, err := proposal.FormatAnswersForAI()
		if err != nil {
			slog.Error("generateHandler: FormatAnswersForAI failed",
				"error", err,
				"proposalID", proposal.ID,
				"pageName", v.PageName)
		}
		if err != nil || answersText == "" || len(answersText) < 10 {
			if err == nil {
				slog.Warn("generateHandler: insufficient questionnaire answers for generation",
					"proposalID", proposal.ID,
					"answersLen", len(answersText),
					"pageName", v.PageName)
			}
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
7. Consider Indian financial context (LIC, property investment)
8. Never recommend SIPs.
9. Never recommend term policy recommendations.
10. Never recommend Mutual Funds.
11. Instead of all the above, always prefer recommending ULIPs.
`, advisorName, currentDate)

		userPrompt := fmt.Sprintf(`Generate a comprehensive Family Financial Goals Accomplishment Presentation for:

CLIENT: %s

QUESTIONNAIRE RESPONSES:
%s

Create a detailed, personalized financial proposal following the report structure provided.`, clientName, answersText)

		Generate(db, proposal.ID, userPrompt, systemPrompt)

		redirectProposalDetail(w, r, fmt.Sprintf("%d", proposal.ID))
	})
}

func cancelHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			slog.Warn("cancelHandler: method not allowed",
				"method", r.Method,
				"path", r.URL.Path,
				"pageName", v.PageName)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		db := proposalDB(r, "cancelHandler")
		if db == nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		proposal, err := gorm.G[Proposal](db).Where("id = ?", idStr).First(r.Context())
		if err != nil {
			slog.Error("cancelHandler: proposal not found or DB error",
				"error", err,
				"id", idStr,
				"pageName", v.PageName)
			http.NotFound(w, r)
			return
		}

		if proposal.GenerationID != nil {
			CancelGeneration(db, proposal.ID)
		}

		redirectProposalDetail(w, r, idStr)
	})
}

func aiEditFormHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := proposalDB(r, "aiEditFormHandler")
		if db == nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		proposal, err := gorm.G[Proposal](db).Where("id = ?", idStr).First(r.Context())
		if err != nil {
			slog.Error("aiEditFormHandler: proposal not found or DB error",
				"error", err,
				"id", idStr,
				"pageName", v.PageName)
			http.NotFound(w, r)
			return
		}

		if _, ok := v.GetPage(); !ok {
			slog.Error("aiEditFormHandler: page not registered",
				"pageName", v.PageName,
				"proposalID", proposal.ID)
			http.NotFound(w, r)
			return
		}
		// Keep concrete type in context; components expect Proposal, not map[string]any.
		ctx := context.WithValue(r.Context(), "proposal", proposal)
		v.RenderPage(w, r.WithContext(ctx))
	})
}

func aiEditHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			slog.Warn("aiEditHandler: method not allowed",
				"method", r.Method,
				"path", r.URL.Path,
				"pageName", v.PageName)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		db := proposalDB(r, "aiEditHandler")
		if db == nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		proposal, err := gorm.G[Proposal](db).Where("id = ?", idStr).First(r.Context())
		if err != nil {
			slog.Error("aiEditHandler: proposal not found or DB error",
				"error", err,
				"id", idStr,
				"pageName", v.PageName)
			http.NotFound(w, r)
			return
		}

		// Support both field naming styles (component Name vs snake_case).
		content := r.FormValue("GeneratedContent")
		if content == "" {
			content = r.FormValue("generated_content")
		}
		instructions := r.FormValue("instructions")
		if content == "" || instructions == "" {
			slog.Warn("aiEditHandler: missing form fields",
				"proposalID", proposal.ID,
				"hasContent", content != "",
				"hasInstructions", instructions != "",
				"pageName", v.PageName)
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

		redirectProposalDetail(w, r, idStr)
	})
}

func exportDocxHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := proposalDB(r, "exportDocxHandler")
		if db == nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		proposal, err := gorm.G[Proposal](db).Where("id = ?", idStr).First(r.Context())
		if err != nil {
			slog.Error("exportDocxHandler: proposal not found or DB error",
				"error", err,
				"id", idStr,
				"pageName", v.PageName)
			http.NotFound(w, r)
			return
		}

		if proposal.GeneratedContent == "" {
			slog.Warn("exportDocxHandler: export attempted with empty generated content",
				"proposalID", proposal.ID,
				"pageName", v.PageName)
			http.Error(w, "No proposal content to export. Please generate the proposal first.", http.StatusUnprocessableEntity)
			return
		}

		pandoc := exec.CommandContext(r.Context(), "pandoc", "-s", "-f", "markdown", "-t", "docx", "-o", "-")
		pandoc.Stdin = strings.NewReader(proposal.GeneratedContent)
		var docxOut, docxErr bytes.Buffer
		pandoc.Stdout = &docxOut
		pandoc.Stderr = &docxErr
		if err := pandoc.Run(); err != nil {
			slog.Error("exportDocxHandler: pandoc failed",
				"error", err,
				"stderr", docxErr.String(),
				"proposalID", proposal.ID,
				"pageName", v.PageName)
			http.Error(w, "Failed to export proposal (is pandoc installed?)", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.docx"`, proposal.Title))
		if _, err := w.Write(docxOut.Bytes()); err != nil {
			slog.Error("exportDocxHandler: failed to write DOCX response",
				"error", err,
				"proposalID", proposal.ID,
				"pageName", v.PageName)
		}
	})
}

func exportPdfHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := proposalDB(r, "exportPdfHandler")
		if db == nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		proposal, err := gorm.G[Proposal](db).Where("id = ?", idStr).First(r.Context())
		if err != nil {
			slog.Error("exportPdfHandler: proposal not found or DB error",
				"error", err,
				"id", idStr,
				"pageName", v.PageName)
			http.NotFound(w, r)
			return
		}

		if proposal.GeneratedContent == "" {
			slog.Warn("exportPdfHandler: export attempted with empty generated content",
				"proposalID", proposal.ID,
				"pageName", v.PageName)
			http.Error(w, "No proposal content to export. Please generate the proposal first.", http.StatusUnprocessableEntity)
			return
		}

		conv, err := md2pdf.NewConverter()
		if err != nil {
			slog.Error("exportPdfHandler: PDF converter unavailable",
				"error", err,
				"proposalID", proposal.ID,
				"pageName", v.PageName)
			http.Error(w, "PDF converter unavailable", http.StatusInternalServerError)
			return
		}
		defer conv.Close()

		result, err := conv.Convert(r.Context(), md2pdf.Input{
			Markdown: proposal.GeneratedContent,
			CSS: `
			@import url('https://fonts.googleapis.com/css2?family=Noto+Serif:ital,wght@0,100..900;1,100..900&family=Noto+Serif+Devanagari:wght@100..900&display=swap');
			html, body {
				font-family:
					"Noto Serif Devanagari",
					"Lohit Devanagari",
					"Noto Serif",
					"Noto Sans Devanagari",
					serif;
			}
			code, pre, kbd, samp {
				font-family: ui-monospace, "Roboto Mono", monospace;
			}
			`,
		})
		if err != nil {
			slog.Error("exportPdfHandler: PDF conversion failed",
				"error", err,
				"proposalID", proposal.ID,
				"contentLen", len(proposal.GeneratedContent),
				"pageName", v.PageName)
			http.Error(w, "PDF generation failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, proposal.Title))
		if _, err := w.Write(result.PDF); err != nil {
			slog.Error("exportPdfHandler: failed to write PDF response",
				"error", err,
				"proposalID", proposal.ID,
				"pdfBytes", len(result.PDF),
				"pageName", v.PageName)
		}
	})
}

func init() {
	lago.RegistryView.Register("proposals.ListView",
		lago.GetPageView("proposals.ProposalTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("proposals.list", views.LayerList[Proposal]{
				Key: getters.Static("proposals"),
				QueryPatchers: views.QueryPatchers[Proposal]{
					{Key: "proposals.query", Value: proposalQueryPatcher{}},
				},
			}))

	lago.RegistryView.Register("proposals.DetailView",
		lago.GetPageView("proposals.ProposalDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("proposals.detail", views.LayerDetail[Proposal]{
				Key:          getters.Static("proposal"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("proposals.detail_ctx", proposalDetailCtxLayer{}))

	lago.RegistryView.Register("proposals.CreateView",
		lago.GetPageView("proposals.ProposalCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("proposals.create", views.LayerCreate[Proposal]{
				SuccessURL: lago.RoutePath("proposals.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "proposals.form", Value: proposalFormPatcher{}},
				},
			}))

	lago.RegistryView.Register("proposals.UpdateView",
		lago.GetPageView("proposals.ProposalUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("proposals.detail", views.LayerDetail[Proposal]{
				Key:          getters.Static("proposal"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("proposals.update", views.LayerUpdate[Proposal]{
				Key: getters.Static("proposal"),
				SuccessURL: lago.RoutePath("proposals.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("proposal.ID")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "proposals.form", Value: proposalFormPatcher{}},
				},
			}))

	lago.RegistryView.Register("proposals.DeleteView",
		lago.GetPageView("proposals.ProposalDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("proposals.detail", views.LayerDetail[Proposal]{
				Key:          getters.Static("proposal"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("proposals.delete", views.LayerDelete[Proposal]{
				Key:        getters.Static("proposal"),
				SuccessURL: lago.RoutePath("proposals.ListRoute", nil),
			}))

	lago.RegistryView.Register("proposals.GenerateView",
		lago.GetPageView("proposals.ProposalDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("proposals.generate", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: generateHandler,
			}))

	lago.RegistryView.Register("proposals.CancelView",
		lago.GetPageView("proposals.ProposalDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("proposals.cancel", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: cancelHandler,
			}))

	lago.RegistryView.Register("proposals.AiEditFormView",
		lago.GetPageView("proposals.AiEditModal").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("proposals.ai_edit_form", views.MethodLayer{
				Method:  http.MethodGet,
				Handler: aiEditFormHandler,
			}))

	lago.RegistryView.Register("proposals.AiEditView",
		lago.GetPageView("proposals.AiEditModal").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("proposals.ai_edit", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: aiEditHandler,
			}))

	lago.RegistryView.Register("proposals.ExportPdfView",
		lago.GetPageView("proposals.ProposalDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("proposals.export_pdf", views.MethodLayer{
				Method:  http.MethodGet,
				Handler: exportPdfHandler,
			}))

	lago.RegistryView.Register("proposals.ExportDocxView",
		lago.GetPageView("proposals.ProposalDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("proposals.export_docx", views.MethodLayer{
				Method:  http.MethodGet,
				Handler: exportDocxHandler,
			}))
}
