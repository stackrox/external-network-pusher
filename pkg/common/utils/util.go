package utils

// ToCompoundName accepts two names and returns a compound name.
// If second name is empty, the first name is returned as the compound name.
func ToCompoundName(nameA, nameB string) string {
	if nameB == "" {
		return nameA
	}
	return nameA + "-" + nameB
}
