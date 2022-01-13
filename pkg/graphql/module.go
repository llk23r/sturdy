package graphql

import (
	graphql_author "mash/pkg/author/graphql"
	graphql_change "mash/pkg/change/graphql"
	graphql_acl "mash/pkg/codebase/acl/graphql"
	graphql_codebase "mash/pkg/codebase/graphql"
	graphql_comments "mash/pkg/comments/graphql"
	"mash/pkg/di"
	graphql_features "mash/pkg/features/graphql"
	graphql_file "mash/pkg/file/graphql"
	graphql_buildkite "mash/pkg/integrations/buildkite/graphql"
	graphql_ci "mash/pkg/integrations/graphql"
	graphql_notification "mash/pkg/notification/graphql"
	graphql_onboarding "mash/pkg/onboarding/graphql"
	graphql_organization "mash/pkg/organization/graphql"
	graphql_pki "mash/pkg/pki/graphql"
	graphql_presence "mash/pkg/presence/graphql"
	graphql_review "mash/pkg/review/graphql"
	graphql_servicetokens "mash/pkg/servicetokens/graphql"
	graphql_suggestion "mash/pkg/suggestions/graphql"
	graphql_user "mash/pkg/user/graphql"
	graphql_view "mash/pkg/view/graphql"
	graphql_workspace_activity "mash/pkg/workspace/activity/graphql"
	graphql_workspace "mash/pkg/workspace/graphql"
	graphql_workspace_watchers "mash/pkg/workspace/watchers/graphql"
)

func Module(c *di.Container) {
	c.Register(NewRootResolver)
	c.Register(graphql_acl.NewResolver)
	c.Register(graphql_author.NewResolver)
	c.Register(graphql_change.NewResolver)
	c.Register(graphql_comments.NewResolver)
	c.Register(graphql_change.NewFileDiffRootResolver)
	c.Register(graphql_suggestion.New)
	c.Register(graphql_notification.NewResolver)
	c.Register(graphql_user.NewResolver)
	c.Register(graphql_view.NewViewStatusRootResolver)
	c.Register(graphql_view.NewResolver)
	c.Register(graphql_workspace_watchers.NewRootResolver)
	c.Register(graphql_workspace.NewResolver)
	c.Register(graphql_workspace_activity.New)
	c.Register(graphql_review.New)
	c.Register(graphql_file.NewFileRootResolver)
	c.Register(graphql_presence.NewRootResolver)
	c.Register(graphql_ci.NewRootResolver)
	c.Register(graphql_onboarding.NewRootResolver)
	c.Register(graphql_pki.NewResolver)
	c.Register(graphql_servicetokens.New)
	c.Register(graphql_buildkite.New)
	c.Register(graphql_organization.New)

	c.Import(graphql_codebase.Module)
	c.Import(graphql_features.Module)
}