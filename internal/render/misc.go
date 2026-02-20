package render

import "os"

func loadFromFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func saveToFile(data []byte, saveTo string) error {
	return os.WriteFile(saveTo, data, 0o644)
}
