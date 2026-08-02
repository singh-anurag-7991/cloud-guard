package scanner

import ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

// tagValue returns the value of the named tag, or "" if it is absent.
// AWS returns tag keys and values as pointers, and untagged resources are
// extremely common, so every field needs a nil check.
func tagValue(tags []ec2types.Tag, key string) string {
	for _, t := range tags {
		if t.Key != nil && *t.Key == key && t.Value != nil {
			return *t.Value
		}
	}
	return ""
}
