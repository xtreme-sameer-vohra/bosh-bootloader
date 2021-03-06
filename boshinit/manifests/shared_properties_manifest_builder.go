package manifests

type SharedPropertiesManifestBuilder struct {
}

func NewSharedPropertiesManifestBuilder() *SharedPropertiesManifestBuilder {
	return &SharedPropertiesManifestBuilder{}
}

func (SharedPropertiesManifestBuilder) AWS(manifestProperties ManifestProperties) AWSProperties {
	return AWSProperties{
		AccessKeyId:           manifestProperties.AccessKeyID,
		SecretAccessKey:       manifestProperties.SecretAccessKey,
		DefaultKeyName:        manifestProperties.DefaultKeyName,
		DefaultSecurityGroups: []string{manifestProperties.SecurityGroup},
		Region:                manifestProperties.Region,
	}
}

func (SharedPropertiesManifestBuilder) NTP() []string {
	return []string{"0.pool.ntp.org", "1.pool.ntp.org"}
}
