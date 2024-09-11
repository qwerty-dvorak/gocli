package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/qwerty-dvorak/gocli/basic"
)

type Base struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewBase() Base {
	return Base{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

type User struct {
	Base
	Email string
	Name  string
}

type Admin struct {
	Base
	CheckinAccess            bool
	AnticheatAccess          bool
	QrmgmtAccess             bool
	QuestionManagementAccess bool
	CommunicationAccess      bool
	UserID                   string
	User                     User
}

var haderror bool

func main() {
	args := os.Args
	db, err := init.NewSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to database: %v\n", err)
		os.Exit(74)
	}
	if len(args) > 2 {
		fmt.Printf("Hello, %s!\n", args[1])
	} else if len(args) == 2 {
		runFile(args[1], db)
	} else {
		runPrompt(db)
	}
}

func runFile(filename string, db *sql.DB) {
	fmt.Printf("Running file %s\n", filename)
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open file %s\n", filename)
		os.Exit(74)
	}
	defer file.Close()

	source, err := io.ReadAll(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read file %s\n", filename)
		os.Exit(74)
	}

	run(string(source), db)
	if !haderror {
		os.Exit(65)
	}
}

func run(source string, db *sql.DB) {
	words := strings.Fields(source)
	if len(words) == 0 {
		fmt.Fprintf(os.Stderr, "Error: empty command\n")
		haderror = true
		return
	}

	firstWord := words[0]
	if firstWord == "adduser" {
		name := words[1]
		AddAdmin(name, db)
	} else {
		fmt.Fprintf(os.Stderr, "Error: unknown command %s\n", firstWord)
		haderror = true
	}
}

func AddAdmin(name string, db *sql.DB) error {
	user, err := CheckUser(name, db)
	if err != nil {
		return err
	}
	admin := Admin{
		Base:                     NewBase(),
		CheckinAccess:            false,
		AnticheatAccess:          false,
		QrmgmtAccess:             false,
		QuestionManagementAccess: false,
		CommunicationAccess:      false,
		UserID:                   user.ID,
	}

	_, err = db.Exec(`
        INSERT INTO admins (id, checkin_access, anticheat_access, qrmgmt_access, 
                            question_management_access, communication_access, user_id, 
                            created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `, admin.ID, admin.CheckinAccess, admin.AnticheatAccess, admin.QrmgmtAccess,
		admin.QuestionManagementAccess, admin.CommunicationAccess, admin.UserID,
		admin.CreatedAt, admin.UpdatedAt)
	if err != nil {
		fmt.Println("Error adding admin")
		return err
	}
	fmt.Println("Admin added successfully")
	return nil
}

func CheckUser(name string, db *sql.DB) (*User, error) {
	query := `SELECT id, email, name, created_at, updated_at FROM users WHERE name=$1`
	row := db.QueryRow(query, name)

	var user User
	err := row.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func runPrompt(db *sql.DB) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println()
			break
		}
		line = strings.TrimSpace(line)
		if line == "" || line == "exit" || line == "quit" {
			break
		}
		run(line, db)
		if haderror {
			haderror = false
		}
	}
}
