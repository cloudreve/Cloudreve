package email

import (
	"context"
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"github.com/cloudreve/Cloudreve/v4/ent"
	"github.com/cloudreve/Cloudreve/v4/pkg/setting"
)

type CommonContext struct {
	SiteBasic *setting.SiteBasic
	Logo      *setting.Logo
	SiteUrl   string
}

// ResetContext used for variables in reset email
type ResetContext struct {
	*CommonContext
	User *ent.User
	Url  string
}

// NewResetEmail generates reset email from template
func NewResetEmail(ctx context.Context, settings setting.Provider, user *ent.User, url string) (string, string, error) {
	templates := settings.ResetEmailTemplate(ctx)
	if len(templates) == 0 {
		return "", "", fmt.Errorf("reset email template not configured")
	}

	selected := selectTemplate(templates, user)
	resetCtx := ResetContext{
		CommonContext: commonContext(ctx, settings),
		User:          user,
		Url:           url,
	}

	tmpl, err := template.New("reset").Parse(selected.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse email template: %w", err)
	}

	var res strings.Builder
	err = tmpl.Execute(&res, resetCtx)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute email template: %w", err)
	}

	return fmt.Sprintf("[%s] %s", resetCtx.SiteBasic.Name, selected.Title), res.String(), nil
}

// ActivationContext used for variables in activation email
type ActivationContext struct {
	*CommonContext
	User *ent.User
	Url  string
}

// NewActivationEmail generates activation email from template
func NewActivationEmail(ctx context.Context, settings setting.Provider, user *ent.User, url string) (string, string, error) {
	templates := settings.ActivationEmailTemplate(ctx)
	if len(templates) == 0 {
		return "", "", fmt.Errorf("activation email template not configured")
	}

	selected := selectTemplate(templates, user)
	activationCtx := ActivationContext{
		CommonContext: commonContext(ctx, settings),
		User:          user,
		Url:           url,
	}

	tmpl, err := template.New("activation").Parse(selected.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse email template: %w", err)
	}

	var res strings.Builder
	err = tmpl.Execute(&res, activationCtx)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute email template: %w", err)
	}

	return fmt.Sprintf("[%s] %s", activationCtx.SiteBasic.Name, selected.Title), res.String(), nil
}

func commonContext(ctx context.Context, settings setting.Provider) *CommonContext {
	logo := settings.Logo(ctx)
	siteUrl := settings.SiteURL(ctx)
	res := &CommonContext{
		SiteBasic: settings.SiteBasic(ctx),
		Logo:      settings.Logo(ctx),
		SiteUrl:   siteUrl.String(),
	}

	// Add site url if logo is not an url
	if !strings.HasPrefix(logo.Light, "http") {
		logoPath, _ := url.Parse(logo.Light)
		res.Logo.Light = siteUrl.ResolveReference(logoPath).String()
	}

	if !strings.HasPrefix(logo.Normal, "http") {
		logoPath, _ := url.Parse(logo.Normal)
		res.Logo.Normal = siteUrl.ResolveReference(logoPath).String()
	}

	return res
}

func selectTemplate(templates []setting.EmailTemplate, u *ent.User) setting.EmailTemplate {
	selected := templates[0]
	if u != nil {
		for _, t := range templates {
			if strings.EqualFold(t.Language, u.Settings.Language) {
				selected = t
				break
			}
		}
	}

	return selected
}
