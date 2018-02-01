package provider

type Provider interface {
	FeatureTables() []string
	CollectionFeatureIds(collName string) ([]int, error)
	GetFeature(collectionName string, id int) ([]byte, error)
}
