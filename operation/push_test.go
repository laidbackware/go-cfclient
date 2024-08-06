package operation

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/laidbackware/go-cfclient/v3/client"
	"github.com/laidbackware/go-cfclient/v3/config"
	"github.com/laidbackware/go-cfclient/v3/testutil"

	"github.com/stretchr/testify/require"
)

func TestAppPush(t *testing.T) {
	serverURL := testutil.SetupFakeAPIServer()
	defer testutil.Teardown()

	g := testutil.NewObjectJSONGenerator(8723)
	org := g.Organization()
	space := g.Space()
	job := g.Job("COMPLETE")
	app := g.Application()
	pkg := g.Package("READY")
	build := g.Build("STAGED")
	droplet := g.Droplet()
	dropletAssoc := g.DropletAssociation()

	fakeAppZipReader := strings.NewReader("blah zip zip")
	var numOfInstances uint = 2
	manifest := &AppManifest{
		Name:       app.Name,
		Buildpacks: []string{"java-buildpack-offline"},

		AppManifestProcess: AppManifestProcess{
			HealthCheckType:         "http",
			HealthCheckHTTPEndpoint: "/health",
			Instances:               &numOfInstances,
			Memory:                  "1G",
		},
		Routes: &AppManifestRoutes{
			{
				Route: "https://spring-music.cf.apps.example.org",
			},
		},
		Services: &AppManifestServices{{Name: "spring-music-sql"}},
		Stack:    "cflinuxfs3",
	}

	testutil.SetupMultiple([]testutil.MockRoute{
		{
			Method:   http.MethodGet,
			Endpoint: "/v3/organizations",
			Output:   g.SinglePaged(org.JSON),
			Status:   http.StatusOK,
		},
		{
			Method:   http.MethodGet,
			Endpoint: "/v3/spaces",
			Output:   g.SinglePaged(space.JSON),
			Status:   http.StatusOK,
		},
		{
			Method:           http.MethodPost,
			Endpoint:         fmt.Sprintf("/v3/spaces/%s/actions/apply_manifest", space.GUID),
			Output:           g.SinglePaged(space.JSON),
			Status:           http.StatusAccepted,
			RedirectLocation: fmt.Sprintf("%s/v3/jobs/%s", serverURL, job.GUID),
		},
		{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("/v3/jobs/%s", job.GUID),
			Output:   g.Single(job.JSON),
			Status:   http.StatusOK,
		},
		{
			Method:   http.MethodGet,
			Endpoint: "/v3/apps",
			Output:   g.SinglePaged(app.JSON),
			Status:   http.StatusOK,
		},
		{
			Method:   http.MethodPost,
			Endpoint: "/v3/packages",
			Output:   g.Single(pkg.JSON),
			Status:   http.StatusCreated,
		},
		{
			Method:   http.MethodPost,
			Endpoint: fmt.Sprintf("/v3/packages/%s/upload", pkg.GUID),
			Output:   g.Single(pkg.JSON),
			Status:   http.StatusOK,
		},
		{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("/v3/packages/%s", pkg.GUID),
			Output:   g.Single(pkg.JSON),
			Status:   http.StatusOK,
		},
		{
			Method:   http.MethodPost,
			Endpoint: "/v3/builds",
			Output:   g.Single(build.JSON),
			Status:   http.StatusCreated,
		},
		{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("/v3/builds/%s", build.GUID),
			Output:   g.Single(build.JSON),
			Status:   http.StatusOK,
		},
		{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("/v3/packages/%s/droplets", pkg.GUID),
			Output:   g.SinglePaged(droplet.JSON),
			Status:   http.StatusOK,
		},
		{
			Method:   http.MethodPatch,
			Endpoint: fmt.Sprintf("/v3/apps/%s/relationships/current_droplet", app.GUID),
			Output:   g.Single(dropletAssoc.JSON),
			Status:   http.StatusOK,
		},
		{
			Method:   http.MethodPost,
			Endpoint: fmt.Sprintf("/v3/apps/%s/actions/start", app.GUID),
			Output:   g.Single(app.JSON),
			Status:   http.StatusOK,
		},
	}, t)

	c, _ := config.New(serverURL, config.Token("", "fake-refresh-token"), config.SkipTLSValidation())
	cf, err := client.New(c)
	require.NoError(t, err)

	pusher := NewAppPushOperation(cf, org.Name, space.Name)
	// Invalid strategy
	strategy := StrategyMode(10)
	pusher.WithStrategy(strategy)
	_, err = pusher.Push(context.Background(), manifest, fakeAppZipReader)
	require.NoError(t, err)
}
