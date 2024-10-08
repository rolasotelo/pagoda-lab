package handlers

import (
	"context"
	"net/http"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/page"
	"github.com/mikestefanello/pagoda/pkg/services"
)

type Dashboard struct {
	*services.TemplateRenderer
}

func init() {
	Register(new(Dashboard))
}

func (d *Dashboard) Init(c *services.Container) error {
	d.TemplateRenderer = c.TemplateRenderer
	return nil
}

func (d *Dashboard) Routes(g *echo.Group) {
	g.GET("/dashboard", d.Page).Name = "dashboard"
}

func (d *Dashboard) Page(ctx echo.Context) error {
	p := page.New(ctx)
	p.Layout = "main"
	p.Name = "dashboard"
	p.Title = "Terraform Workspaces Dashboard"

	token := os.Getenv("TFE_TOKEN")
	if token == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "TFE_TOKEN environment variable not set")
	}

	config := &tfe.Config{
		Token: token,
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create TFE client")
	}

	workspaces, err := listWorkspaces(client)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list workspaces")
	}

	p.Data = echo.Map{
		"Workspaces": workspaces,
	}

	return d.RenderPage(ctx, p)
}

func listWorkspaces(client *tfe.Client) ([]*tfe.Workspace, error) {
	options := tfe.WorkspaceListOptions{
		ListOptions: tfe.ListOptions{
			PageNumber: 0,
			PageSize:   100,
		},
	}

	var allWorkspaces []*tfe.Workspace
	for {
		workspaces, err := client.Workspaces.List(context.Background(), "", &options)
		if err != nil {
			return nil, err
		}
		allWorkspaces = append(allWorkspaces, workspaces.Items...)
		if workspaces.CurrentPage >= workspaces.TotalPages {
			break
		}
		options.PageNumber = workspaces.NextPage
	}

	return allWorkspaces, nil
}
