package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/LordFarquaadtheCreator/cover-letter-writter/internal/generate"
	"github.com/LordFarquaadtheCreator/cover-letter-writter/internal/history"
	"github.com/LordFarquaadtheCreator/cover-letter-writter/internal/profile"
)

type deps struct {
	Profiles *profile.Store
	History  *history.Store
}

// --- Tool inputs ---

type CreateProfileInput struct {
	Label     string `json:"label,omitempty" jsonschema:"Optional short label for the profile (e.g. 'default', 'tech', 'design')"`
	Name      string `json:"name" jsonschema:"required,Full name as it should appear on the cover letter"`
	Address   string `json:"address,omitempty" jsonschema:"Address line shown in the contact column"`
	Email     string `json:"email" jsonschema:"required,Email shown in the contact column"`
	Phone     string `json:"phone,omitempty" jsonschema:"Phone shown in the contact column"`
	OutputDir string `json:"outputDir,omitempty" jsonschema:"Default output directory for PDFs. Defaults to ~/Downloads."`
	Filename  string `json:"filename,omitempty" jsonschema:"Default PDF filename. Defaults to <Name>NoSpacesCoverLetter.pdf."`
}

type GetProfileInput struct {
	ID string `json:"id" jsonschema:"required,Profile ID"`
}

type UpdateProfileInput struct {
	ID        string `json:"id" jsonschema:"required,Profile ID to update"`
	Label     string `json:"label,omitempty"`
	Name      string `json:"name,omitempty"`
	Address   string `json:"address,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	OutputDir string `json:"outputDir,omitempty"`
	Filename  string `json:"filename,omitempty"`
}

type DeleteProfileInput struct {
	ID string `json:"id" jsonschema:"required,Profile ID to delete"`
}

type ListHistoryInput struct {
	ProfileID string `json:"profileId,omitempty" jsonschema:"Filter history by profile ID"`
}

// --- Tool outputs ---

type CreateProfileOutput struct {
	Profile profile.Profile `json:"profile"`
}

type GetProfileOutput struct {
	Profile profile.Profile `json:"profile"`
}

type ListProfilesOutput struct {
	Profiles []profile.Profile `json:"profiles"`
}

type UpdateProfileOutput struct {
	Profile profile.Profile `json:"profile"`
}

type DeleteProfileOutput struct {
	Deleted bool `json:"deleted"`
}

type ListHistoryOutput struct {
	Entries []history.Entry `json:"entries"`
}

// Run starts the stdio MCP server.
func Run(dataDir string) error {
	d := deps{
		Profiles: profile.NewStore(dataDir),
		History:  history.NewStore(dataDir),
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "cover-letter-writter", Version: "1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_profile",
		Description: "Create a user profile storing contact info (name, address, email, phone) used to generate cover letters. Returns the profile with its ID. Call this once per persona/role-target. The profile persists on disk across server restarts.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateProfileInput) (*mcp.CallToolResult, CreateProfileOutput, error) {
		return handleCreateProfile(ctx, req, args, d)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_profile",
		Description: "Fetch a single profile by ID. Returns the full profile (name, address, email, phone, output defaults).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetProfileInput) (*mcp.CallToolResult, GetProfileOutput, error) {
		return handleGetProfile(ctx, req, args, d)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_profiles",
		Description: "List all stored profiles sorted by creation time. Use this to discover profile IDs before calling generate_cover_letter.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, ListProfilesOutput, error) {
		return handleListProfiles(ctx, req, d)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_profile",
		Description: "Update fields on an existing profile. Only non-empty fields are applied. Returns the updated profile.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateProfileInput) (*mcp.CallToolResult, UpdateProfileOutput, error) {
		return handleUpdateProfile(ctx, req, args, d)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_profile",
		Description: "Delete a profile by ID. History entries for the profile are preserved.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DeleteProfileInput) (*mcp.CallToolResult, DeleteProfileOutput, error) {
		return handleDeleteProfile(ctx, req, args, d)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_cover_letter",
		Description: "Generate a styled PDF cover letter. REQUIRES a profile_id (from create_profile or list_profiles) and body text. The body is the cover letter content — the agent drafts or refines this. Output is saved to the profile's outputDir (or override), defaulting to ~/Downloads. Returns the absolute file path. Each generation is recorded in history.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args generate.Input) (*mcp.CallToolResult, generate.Output, error) {
		return handleGenerate(ctx, req, args, d)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_history",
		Description: "List past cover letter generations, newest first. Optionally filter by profile_id. Each entry includes the body text, output path, and timestamp — useful for reviewing what was sent to a given employer.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListHistoryInput) (*mcp.CallToolResult, ListHistoryOutput, error) {
		return handleListHistory(ctx, req, args, d)
	})

	return server.Run(context.Background(), &mcp.StdioTransport{})
}

// --- Handlers ---

func handleCreateProfile(ctx context.Context, req *mcp.CallToolRequest, args CreateProfileInput, d deps) (*mcp.CallToolResult, CreateProfileOutput, error) {
	p, err := d.Profiles.Create(profile.Profile{
		Label:     args.Label,
		Name:      args.Name,
		Address:   args.Address,
		Email:     args.Email,
		Phone:     args.Phone,
		OutputDir: args.OutputDir,
		Filename:  args.Filename,
	})
	if err != nil {
		return nil, CreateProfileOutput{}, err
	}
	return jsonResult(CreateProfileOutput{Profile: p})
}

func handleGetProfile(ctx context.Context, req *mcp.CallToolRequest, args GetProfileInput, d deps) (*mcp.CallToolResult, GetProfileOutput, error) {
	p, err := d.Profiles.Get(args.ID)
	if err != nil {
		return nil, GetProfileOutput{}, err
	}
	return jsonResult(GetProfileOutput{Profile: p})
}

func handleListProfiles(ctx context.Context, req *mcp.CallToolRequest, d deps) (*mcp.CallToolResult, ListProfilesOutput, error) {
	ps, err := d.Profiles.List()
	if err != nil {
		return nil, ListProfilesOutput{}, err
	}
	return jsonResult(ListProfilesOutput{Profiles: ps})
}

func handleUpdateProfile(ctx context.Context, req *mcp.CallToolRequest, args UpdateProfileInput, d deps) (*mcp.CallToolResult, UpdateProfileOutput, error) {
	p, err := d.Profiles.Update(args.ID, profile.Profile{
		Label:     args.Label,
		Name:      args.Name,
		Address:   args.Address,
		Email:     args.Email,
		Phone:     args.Phone,
		OutputDir: args.OutputDir,
		Filename:  args.Filename,
	})
	if err != nil {
		return nil, UpdateProfileOutput{}, err
	}
	return jsonResult(UpdateProfileOutput{Profile: p})
}

func handleDeleteProfile(ctx context.Context, req *mcp.CallToolRequest, args DeleteProfileInput, d deps) (*mcp.CallToolResult, DeleteProfileOutput, error) {
	if err := d.Profiles.Delete(args.ID); err != nil {
		return nil, DeleteProfileOutput{}, err
	}
	return jsonResult(DeleteProfileOutput{Deleted: true})
}

func handleGenerate(ctx context.Context, req *mcp.CallToolRequest, args generate.Input, d deps) (*mcp.CallToolResult, generate.Output, error) {
	if args.ProfileID == "" {
		return nil, generate.Output{}, fmt.Errorf("profileId is required — call list_profiles or create_profile first")
	}
	if args.Body == "" {
		return nil, generate.Output{}, fmt.Errorf("body is required")
	}
	p, err := d.Profiles.Get(args.ProfileID)
	if err != nil {
		return nil, generate.Output{}, err
	}
	out, err := generate.Run(p, args.To, args.Body, args.OutputDir, args.Filename)
	if err != nil {
		return nil, generate.Output{}, err
	}
	if _, err := d.History.Add(history.Entry{
		ProfileID:   p.ID,
		ProfileName: p.Name,
		Body:        args.Body,
		OutputPath:  out.OutputPath,
		Filename:    out.Filename,
	}); err != nil {
		// history is best-effort — don't fail the generation
		_ = err
	}
	return jsonResult(out)
}

func handleListHistory(ctx context.Context, req *mcp.CallToolRequest, args ListHistoryInput, d deps) (*mcp.CallToolResult, ListHistoryOutput, error) {
	es, err := d.History.List(args.ProfileID)
	if err != nil {
		return nil, ListHistoryOutput{}, err
	}
	return jsonResult(ListHistoryOutput{Entries: es})
}

// jsonResult marshals the structured output as pretty JSON in the text content.
func jsonResult[T any](out T) (*mcp.CallToolResult, T, error) {
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, out, fmt.Errorf("marshal result: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
	}, out, nil
}
