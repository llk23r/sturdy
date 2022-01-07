package enterprise

import "mash/pkg/graphql/resolvers"

var Resolver = &FeaturesRootResolver{}

type FeaturesRootResolver struct{}

func (r *FeaturesRootResolver) Features() []resolvers.Feature {
	return []resolvers.Feature{
		resolvers.FeatureBuildkite,
		resolvers.FeatureGitHub,
	}
}
