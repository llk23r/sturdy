package graphql

import (
	"getsturdy.com/api/pkg/di"
	"getsturdy.com/api/pkg/graphql/resolvers"
	"getsturdy.com/api/pkg/installations/graphql"
)

func Module(c *di.Container) {
	c.Register(graphql.New)
	c.Register(New, new(resolvers.InstallationsRootResolver))
}
