package provider

type Provider interface {
	FeatureTables() []string
	CollectionFeatureIds(collName string) ([]int, error)
}
