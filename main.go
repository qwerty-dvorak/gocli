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

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	magenta = "\033[35m"
	cyan   = "\033[36m"
	bold   = "\033[1m"
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

func printCommandUsage() {
	fmt.Printf("%sUsage (to run on text): ./main <filename>%s\n", cyan, reset)
	fmt.Printf("%sUsage (to run in prompt): ./main%s\n", cyan, reset)
}

func printPromptUsage() {
	fmt.Printf("%sAvailable commands:%s\n", yellow, reset)
	fmt.Printf("  Type %s'q'%s to exit\n", green, reset)
	fmt.Printf("  Type %s'h'%s for help\n", green, reset)
	fmt.Printf("  Type %s'add <email>'%s to add an admin\n", green, reset)
	fmt.Printf("  Type %s'delete <email>'%s to delete an admin\n", green, reset)
	fmt.Printf("  Type %s'modify <email>'%s to modify an admin\n", green, reset)
}

func main() {
	args := os.Args
	if len(args) > 2 {
		fmt.Printf("%sError: Too many arguments provided%s\n", red, reset)
		printCommandUsage()
		os.Exit(64)
	} else {
		db, err := basic.NewSession()
		if err != nil {
			fmt.Printf("%sCould not connect to database: %v%s\n", red, err, reset)
			os.Exit(74)
		}
		fmt.Printf("%sConnected to database%s\n", magenta, reset)
		if len(args) == 2 {
			runFile(args[1], db)
		} else {
			fmt.Printf("%sRunning in prompt mode%s\n", cyan, reset)
			printPromptUsage()
			runPrompt(db)
		}
	}
}

func runPrompt(db *sql.DB) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s>%s ", blue, reset)
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println()
			break
		}
		line = strings.TrimSpace(line)
		if line == "exit" || line == "quit" || line == "q" {
			break
		}
		if line == "h" || line == "help" {
			printPromptUsage()
			continue
		}
		run(line, db)
		if haderror {
			haderror = false
		}
	}
}

func runFile(filename string, db *sql.DB) {
	fmt.Printf("%sRunning file %s%s\n", cyan, filename, reset)
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("%sCould not open file %s%s\n", red, filename, reset)
		os.Exit(74)
	}
	defer file.Close()

	source, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("%sCould not read file %s%s\n", red, filename, reset)
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
		fmt.Printf("%sError: empty command%s\n", red, reset)
		haderror = true
		return
	}

	firstWord := words[0]
	switch firstWord {
	case "add":
		if len(words) < 2 {
			fmt.Printf("%sError: missing email for add command%s\n", red, reset)
			haderror = true
			return
		}
		email := words[1]
		err := AddAdmin(email, db)
		if err != nil {
			fmt.Printf("%sError adding admin: %v%s\n", red, err, reset)
			haderror = true
		} else {
			fmt.Printf("%sAdmin added successfully%s\n", green, reset)
		}
	case "delete":
		if len(words) < 2 {
			fmt.Printf("%sError: missing email for delete command%s\n", red, reset)
			haderror = true
			return
		}
		email := words[1]
		err := DeleteAdmin(email, db)
		if err != nil {
			fmt.Printf("%sError deleting admin: %v%s\n", red, err, reset)
			haderror = true
		} else {
			fmt.Printf("%sAdmin deleted successfully%s\n", green, reset)
		}
	case "modify":
		if len(words) < 2 {
			fmt.Printf("%sError: missing email for modify command%s\n", red, reset)
			haderror = true
			return
		}
		email := words[1]
		err := ModifyAdmin(email, db)
		if err != nil {
			fmt.Printf("%sError modifying admin: %v%s\n", red, err, reset)
			haderror = true
		} else {
			fmt.Printf("%sAdmin modified successfully%s\n", green, reset)
		}
	default:
		fmt.Printf("%sError: unknown command %s%s\n", red, firstWord, reset)
		printPromptUsage()
		haderror = true
	}
}

func CheckUser(email string, db *sql.DB) (*User, error) {
	query := `SELECT id, email, name FROM users WHERE email=$1`
	row := db.QueryRow(query, email)
	var user User
	err := row.Scan(&user.ID, &user.Email, &user.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("error scanning row")
	}
	return &user, nil
}

func askForAccess(accessType string) int {
    var input string
    for {
        fmt.Printf("%sGrant %s access? (y/n): %s", yellow, accessType, reset)
        fmt.Scanln(&input)
        input = strings.ToLower(strings.TrimSpace(input))
        if input == "y" ||  input == "t" {
            return 1
        } else if input == "n" || input == "f" {
            return 0
        } else if input == "" {
            return -1
        } else if input == "exit" || input == "quit" || input == "q" {
            os.Exit(0)
        }
        fmt.Printf("%sInvalid input. Please enter 'y' for yes or 'n' for no.%s\n", red, reset)
    }
}

func NewAdmin(user User) Admin {
    return Admin{
        Base:                     NewBase(),
        CheckinAccess:            false,
        AnticheatAccess:          false,
        QrmgmtAccess:             false,
        QuestionManagementAccess: false,
        CommunicationAccess:      false,
        UserID:                   user.ID,
        User:                     user,
    }
}

func AddAdmin(email string, db *sql.DB) error {
	user, err := CheckUser(email, db)
	if err != nil {
		return err
	}

	var existingAdminID string
	err = db.QueryRow("SELECT id FROM admins WHERE user_id = $1", user.ID).Scan(&existingAdminID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if existingAdminID != "" {
		return errors.New("admin already exists")
	}

	admin := NewAdmin(*user)

	checkinAccess := askForAccess("Checkin")
    if checkinAccess != -1 {
        admin.CheckinAccess = checkinAccess == 1
    }
    anticheatAccess := askForAccess("Anticheat")
    if anticheatAccess != -1 {
        admin.AnticheatAccess = anticheatAccess == 1
    }
    qrmgmtAccess := askForAccess("Qr Management")
    if qrmgmtAccess != -1 {
        admin.QrmgmtAccess = qrmgmtAccess == 1
    }
    questionManagementAccess := askForAccess("Question Management")
    if questionManagementAccess != -1 {
        admin.QuestionManagementAccess = questionManagementAccess == 1
    }
    communicationAccess := askForAccess("Communication")
    if communicationAccess != -1 {
        admin.CommunicationAccess = communicationAccess == 1
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
		return err
	}

	return nil
}

func DeleteAdmin(email string, db *sql.DB) error {
	user, err := CheckUser(email, db)
	if err != nil {
		return err
	}
	result, err := db.Exec("DELETE FROM admins WHERE user_id = $1", user.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("admin not found")
	}
	return nil
}

func printAdminDetails(user User, existingAdmin Admin) {
	headers := []string{"Detail", "Value"}
	rows := [][]string{
		{"Name", user.Name},
		{"Checkin Access", fmt.Sprintf("%t", existingAdmin.CheckinAccess)},
		{"Anticheat Access", fmt.Sprintf("%t", existingAdmin.AnticheatAccess)},
		{"QR Management Access", fmt.Sprintf("%t", existingAdmin.QrmgmtAccess)},
		{"Question Management Access", fmt.Sprintf("%t", existingAdmin.QuestionManagementAccess)},
		{"Communication Access", fmt.Sprintf("%t", existingAdmin.CommunicationAccess)},
	}

	detailWidth := len(headers[0])
	valueWidth := len(headers[1])
	for _, row := range rows {
		if len(row[0]) > detailWidth {
			detailWidth = len(row[0])
		}
		if len(row[1]) > valueWidth {
			valueWidth = len(row[1])
		}
	}

	fmt.Printf("%sDetails of the admin are as follows:%s\n", cyan, reset)
	fmt.Printf("+-%s-+-%s-+\n", strings.Repeat("-", detailWidth), strings.Repeat("-", valueWidth))
	printTableRow(detailWidth, valueWidth, headers[0], headers[1], magenta+bold)
	for _, row := range rows {
		printTableRow(detailWidth, valueWidth, row[0], row[1], "")
	}
}

func printTableRow(detailWidth, valueWidth int, detail, value string, colorCode string) {
	fmt.Printf("| %s%-*s%s | %s%-*s%s |\n",
		colorCode, detailWidth, detail, reset,
		colorCode, valueWidth, value, reset)
	fmt.Printf("+-%s-+-%s-+\n", strings.Repeat("-", detailWidth), strings.Repeat("-", valueWidth))
}

func ModifyAdmin(email string, db *sql.DB) error {
	user, err := CheckUser(email, db)
	if err != nil {
		return err
	}

	var existingAdmin Admin
	err = db.QueryRow(`SELECT id, checkin_access, anticheat_access, qrmgmt_access,
		question_management_access, communication_access, user_id, created_at, updated_at
		FROM admins WHERE user_id = $1`, user.ID).
		Scan(&existingAdmin.ID, &existingAdmin.CheckinAccess, &existingAdmin.AnticheatAccess,
			&existingAdmin.QrmgmtAccess, &existingAdmin.QuestionManagementAccess,
			&existingAdmin.CommunicationAccess, &existingAdmin.UserID,
			&existingAdmin.CreatedAt, &existingAdmin.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("admin not found")
		}
		return err
	}

	printAdminDetails(*user, existingAdmin)

	checkinAccess := askForAccess("Checkin")
    if checkinAccess != -1 {
        existingAdmin.CheckinAccess = checkinAccess == 1
    }
    anticheatAccess := askForAccess("Anticheat")
    if anticheatAccess != -1 {
        existingAdmin.AnticheatAccess = anticheatAccess == 1
    }
    qrmgmtAccess := askForAccess("Qr Management")
    if qrmgmtAccess != -1 {
        existingAdmin.QrmgmtAccess = qrmgmtAccess == 1
    }
    questionManagementAccess := askForAccess("Question Management")
    if questionManagementAccess != -1 {
        existingAdmin.QuestionManagementAccess = questionManagementAccess == 1
    }
    communicationAccess := askForAccess("Communication")
    if communicationAccess != -1 {
        existingAdmin.CommunicationAccess = communicationAccess == 1
    }

	_, err = db.Exec(`
		UPDATE admins SET checkin_access = $1, anticheat_access = $2, qrmgmt_access = $3,
		question_management_access = $4, communication_access = $5, updated_at = $6
		WHERE user_id = $7
	`, existingAdmin.CheckinAccess, existingAdmin.AnticheatAccess, existingAdmin.QrmgmtAccess,
	existingAdmin.QuestionManagementAccess, existingAdmin.CommunicationAccess, existingAdmin.UpdatedAt, existingAdmin.UserID)
	if err != nil {
		return err
	}
	printAdminDetails(*user, existingAdmin)
	return nil
}