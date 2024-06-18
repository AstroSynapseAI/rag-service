package templates

import (
	"fmt"
	"os"
)

func Load(name string) (string, error) {
	path := "./engine/templates/" + name

	tmplContent, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading search context:", err)
		return "", err
	}

	return string(tmplContent), nil
}
