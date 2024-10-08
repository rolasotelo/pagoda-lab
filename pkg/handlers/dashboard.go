package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/page"
	"github.com/mikestefanello/pagoda/pkg/services"
)

const defaultOrganization = "rolasotelo" // Add this constant at the top of the file

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
		log.Printf("Failed to create TFE client: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create TFE client: %v", err))
	}

	workspaces, err := listWorkspaces(client, defaultOrganization)
	if err != nil {
		log.Printf("Failed to list workspaces: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to list workspaces: %v", err))
	}

	p.Data = echo.Map{
		"Workspaces": workspaces,
	}

	return d.RenderPage(ctx, p)
}

func listWorkspaces(client *tfe.Client, organization string) ([]*tfe.Workspace, error) {
	options := tfe.WorkspaceListOptions{
		ListOptions: tfe.ListOptions{
			PageNumber: 0,
			PageSize:   100,
		},
	}

	var allWorkspaces []*tfe.Workspace
	for {
		log.Printf("Fetching workspaces page %d for organization %s", options.PageNumber, organization)
		workspaces, err := client.Workspaces.List(context.Background(), organization, &options)
		if err != nil {
			log.Printf("Error fetching workspaces: %v", err)
			return nil, fmt.Errorf("error fetching workspaces: %w", err)
		}
		allWorkspaces = append(allWorkspaces, workspaces.Items...)
		if len(workspaces.Items) < options.PageSize {
			break
		}
		options.PageNumber++
	}

	log.Printf("Total workspaces fetched for organization %s: %d", organization, len(allWorkspaces))
	return allWorkspaces, nil
}
