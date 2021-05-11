package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/Fancman/BackupSoftware/cmd"
	"github.com/Fancman/BackupSoftware/database"
	helper "github.com/Fancman/BackupSoftware/helpers"
)

var db = &database.SQLite{}
var out io.Writer = os.Stdout

func TestConnection_01(t *testing.T) {
	err := db.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		t.Errorf("Cannot connect to db: %v", err)
	}
}

func TestAddDrive_02(t *testing.T) {
	cmd := cmd.TestRootCmd()

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"add-drive"})
	cmd.Execute()
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "hi-via-args" {
		t.Fatalf("expected \"%s\" got \"%s\"", "hi-via-args", string(out))
	}

	cmd.Execute()
}
