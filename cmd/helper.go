package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/segmentio/ksuid"
)

// https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html
func GenKsuid() string {
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
		err := drive_exists(string(drive))
		if err == nil {
			r = append(r, string(drive))
		}
	}
	return r
}

func isCommandAvailable(name string) bool {
	cmd := exec.Command(name)

	err := cmd.Run()

	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// Does drive exist?
func drive_exists(drive_letter string) error {
	f, err := os.Open(drive_letter + ":\\")
	if err == nil {
		f.Close()
		return nil
	}
	return err
}

// Get ksuid form .drive file by drive letter
func get_ksuid_from_drive(drive_letter string) (string, error) {
	if FileExists(drive_letter + ":/.drive") {
		// ak ma .drive subor a nie je zapisane v db
		lines, err := read_file_lines(drive_letter + ":/.drive")

		if err != nil {
			fmt.Printf("Error while reading a file: %s", err)
			return "", err
		}

		return lines[0], nil
	}
	return "", errors.New(".drive file doesn't exist on drive")
}

// File exists?
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
