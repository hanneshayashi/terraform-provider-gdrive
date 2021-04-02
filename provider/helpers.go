package provider

func contains(s string, slice []string) bool {
	for i := range slice {
		if s == slice[i] {
			return true
		}
	}
	return false
}

// func noDiff(_, _, _ string, _ *schema.ResourceData) bool {
// 	return true
// }
