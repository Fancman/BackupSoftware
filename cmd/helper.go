package cmd

import (
	"bufio"
	"os"

	"github.com/segmentio/ksuid"
)

// https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html
func gen_ksuid() string {
	return ksuid.New().String()
}

func get_appdata_dir() (string, error) {
	path, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return path, nil
}

// Reads lines from text file
func read_file_lines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// Returns list of available drives
func get_drives() (r []string) {
	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		f, err := os.Open(string(drive) + ":\\")
		if err == nil {
			r = append(r, string(drive))
			f.Close()
		}
	}
	return r
}

// File exists?
func file_exists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
