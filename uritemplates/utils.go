package uritemplates

func Expand(path string, values map[string]string) (string, error) {
	template, err := Parse(path)
	if err != nil {
		return "", err
	}
	return template.Expand(values)
}
