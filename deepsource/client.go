// DeepSource SDK
package deepsource

import (
	"context"
	"fmt"

	"github.com/deepsourcelabs/cli/deepsource/analyzers"
	"github.com/deepsourcelabs/cli/deepsource/auth"
	"github.com/deepsourcelabs/cli/deepsource/issues"
	"github.com/deepsourcelabs/cli/deepsource/repository"
	"github.com/deepsourcelabs/cli/deepsource/transformers"
	"github.com/deepsourcelabs/cli/types"
	"github.com/deepsourcelabs/graphql"

	analyzerMutation "github.com/deepsourcelabs/cli/deepsource/analyzers/mutations"
	analyzerQuery "github.com/deepsourcelabs/cli/deepsource/analyzers/queries"
	authMutation "github.com/deepsourcelabs/cli/deepsource/auth/mutations"
	issuesQuery "github.com/deepsourcelabs/cli/deepsource/issues/queries"
	repoQuery "github.com/deepsourcelabs/cli/deepsource/repository/queries"
	transformerQuery "github.com/deepsourcelabs/cli/deepsource/transformers/queries"
)

var defaultHostName = "deepsource.icu"

type ClientOpts struct {
	Token    string
	HostName string
}

type Client struct {
	gql   *graphql.Client
	token string
}

// Returns a GraphQL client which can be used to interact with the GQL APIs
func (c Client) GQL() *graphql.Client {
	return c.gql
}

// Returns the PAT which is required for authentication and thus, interacting with the APIs
func (c Client) GetToken() string {
	return c.token
}

// Returns a new GQLClient
func New(cp ClientOpts) (*Client, error) {
	apiClientURL := getAPIClientURL(cp.HostName)
	gql := graphql.NewClient(apiClientURL)
	return &Client{
		gql:   gql,
		token: cp.Token,
	}, nil
}

// // Formats and returns the DeepSource Public API client URL
func getAPIClientURL(hostName string) string {
	apiClientURL := fmt.Sprintf("https://api.%s/graphql/", defaultHostName)

	// Check if the domain is different from the default domain (In case of Enterprise users)
	if hostName != defaultHostName {
		apiClientURL = fmt.Sprintf("http://%s/graphql/", hostName)
	}
	return apiClientURL
}

// Registers the device and allots it a device code which is further used for fetching
// the PAT and other authentication data
func (c Client) RegisterDevice(ctx context.Context) (*auth.Device, error) {
	req := authMutation.RegisterDeviceRequest{}
	res, err := req.Do(ctx, c)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Logs in the client using the deviceCode and the user Code and returns the PAT and data which is required for authentication
func (c Client) Login(ctx context.Context, deviceCode, description string) (*auth.PAT, error) {
	req := authMutation.RequestPATRequest{
		Params: authMutation.RequestPATParams{
			DeviceCode:  deviceCode,
			Description: description,
		},
	}

	res, err := req.Do(ctx, c)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Refreshes the authentication credentials. Takes the refreshToken as a parameter.
func (c Client) RefreshAuthCreds(ctx context.Context, token string) (*auth.PAT, error) {
	req := authMutation.RefreshTokenRequest{
		Params: authMutation.RefreshTokenParams{
			Token: token,
		},
	}
	res, err := req.Do(ctx, c)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Returns the list of Analyzers supported by DeepSource along with their meta like shortcode, metaschema.
func (c Client) GetSupportedAnalyzers(ctx context.Context) ([]analyzers.Analyzer, error) {
	req := analyzerQuery.AnalyzersRequest{}
	res, err := req.Do(ctx, c)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Returns the list of Transformers supported by DeepSource along with their meta like shortcode.
func (c Client) GetSupportedTransformers(ctx context.Context) ([]transformers.Transformer, error) {
	req := transformerQuery.TransformersRequest{}
	res, err := req.Do(ctx, c)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Returns the activation status of the repository whose data is sent as parameters.
// Owner : The username of the owner of the repository
// repoName : The name of the repository whose activation status has to be queried
// provider : The VCS provider which hosts the repo (GITHUB/GITLAB/BITBUCKET)
func (c Client) GetRepoStatus(ctx context.Context, owner, repoName, provider string) (*repository.Meta, error) {
	req := repoQuery.RepoStatusRequest{
		Params: repoQuery.RepoStatusParams{
			Owner:    owner,
			RepoName: repoName,
			Provider: provider,
		},
	}

	res, err := req.Do(ctx, c)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Returns the list of issues for a certain repository whose data is sent as parameters.
// Owner : The username of the owner of the repository
// repoName : The name of the repository whose activation status has to be queried
// provider : The VCS provider which hosts the repo (GITHUB/GITLAB/BITBUCKET)
// limit : The amount of issues to be listed. The default limit is 30 while the maximum limit is currently 100.
func (c Client) GetIssues(ctx context.Context, owner, repoName, provider string, limit int) ([]issues.Issue, error) {
	req := issuesQuery.IssuesListRequest{
		Params: issuesQuery.IssuesListParams{
			Owner:    owner,
			RepoName: repoName,
			Provider: provider,
			Limit:    limit,
		},
	}
	res, err := req.Do(ctx, c)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Returns the list of issues reported for a certain file in a certain repository whose data is sent as parameters.
// Owner : The username of the owner of the repository
// repoName : The name of the repository whose activation status has to be queried
// provider : The VCS provider which hosts the repo (GITHUB/GITLAB/BITBUCKET)
// filePath : The relative path of the file. Eg: "tests/mock.py" if a file `mock.py` is present in `tests` directory which in turn is present in the root dir
// limit : The amount of issues to be listed. The default limit is 30 while the maximum limit is currently 100.
func (c Client) GetIssuesForFile(ctx context.Context, owner, repoName, provider, filePath string, limit int) ([]issues.Issue, error) {
	req := issuesQuery.FileIssuesListRequest{
		Params: issuesQuery.FileIssuesListParams{
			Owner:    owner,
			RepoName: repoName,
			Provider: provider,
			FilePath: filePath,
			Limit:    limit,
		},
	}

	res, err := req.Do(ctx, c)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// SyncAnalyzer syncs the local custom Analyzer with DeepSource.
func (c Client) SyncAnalyzer(ctx context.Context, analyzerTOMLData *types.AnalyzerTOML, issuesData *[]types.AnalyzerIssue) (*analyzers.SyncResponse, error) {
	req := analyzerMutation.SyncAnalyzerRequest{
		Input: analyzerMutation.SyncAnalyzerInput{
			Analyzer: analyzerMutation.AnalyzerInput{
				Name:             analyzerTOMLData.Name,
				Version:          analyzerTOMLData.AnalyzerVersion,
				Shortcode:        analyzerTOMLData.Shortcode,
				Description:      analyzerTOMLData.Description,
				Tags:             analyzerTOMLData.Tags,
				RepositoryUrl:    analyzerTOMLData.Repository,
				DocumentationUrl: analyzerTOMLData.DocumentationURL,
				BugTrackerUrl:    analyzerTOMLData.BugTrackerURL,
				AnalysisCommand:  analyzerTOMLData.Analysis.Command,
				// TODO(SNT): Activate this when Autofix is supported for custom Analyzers.
				// AutofixCommand:   analyzerTOMLData.AutofixCommand,
			},
		},
	}

	// Making a duplicate instance of the issuesData slice since the category needs to be modified to
	// match the enum values expected by the syncAnalyzer mutation.
	issuesList := *issuesData

	// Assigning the enum value of issue category.
	for i := 0; i < len(issuesList); i++ {
		issuesList[i].Category = types.IssueCategoryEnumMap[issuesList[i].Category]
	}

	// Appending issues to the parameters.
	req.Input.Issues = append(req.Input.Issues, issuesList...)

	res, err := req.Do(ctx, c)
	if err != nil {
		return &analyzers.SyncResponse{}, err
	}
	return res, nil
}
