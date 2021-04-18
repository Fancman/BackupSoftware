package helper

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/segmentio/ksuid"
)

// https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html
func GenKsuid() string {
	return ksuid.New().String()
}

func GetAppDir() (string, error) {
	path, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return path, nil
}

// Reads lines from text file
func ReadFileLines(path string) ([]string, error) {
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
func GetDrives() (r []string) {
	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		err := DriveExists(string(drive))
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
func DriveExists(drive_letter string) error {
	f, err := os.Open(drive_letter + ":\\")
	if err == nil {
		f.Close()
		return nil
	}
	return err
}

func RemoveDriveLetter(name string) string {
	name = strings.TrimSpace(name)
	name = path.Clean(name)
	// Remove drive letter
	if len(name) == 2 && name[1] == ':' {
		name = ""
	} else if len(name) > 2 && name[1] == ':' {
		name = name[2:]
	}

	return name
}

// Get ksuid form .drive file by drive letter
func GetKsuidFromDrive(drive_letter string) string {
	if Exists(drive_letter + ":/.drive") {
		// ak ma .drive subor a nie je zapisane v db
		lines, err := ReadFileLines(drive_letter + ":/.drive")

		if err != nil {
			fmt.Printf("Error while reading a file: %s", err)
			return ""
		}

		return lines[0]
	}
	return ""
}

// File exists?
/*func Exists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}*/

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
