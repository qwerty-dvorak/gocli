package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"
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

type Flag struct {
	Base
	Name	string
	Value   bool
}

var haderror bool

func printCommandUsage() {
	fmt.Printf("%sUsage (to run on text): ./main file <filename>%s\n", cyan, reset)
	fmt.Printf("%sUsage (to run in admin prompt): ./main admin%s\n", cyan, reset)
	fmt.Printf("%sUsage (to run in flag prompt): ./main flag%s\n", cyan, reset)
	fmt.Printf("%sUsage (to run in whitelist prompt): ./main whitelist%s\n", cyan, reset)
}

func printAdminUsage() {
	fmt.Printf("%sAvailable commands:%s\n", yellow, reset)
	fmt.Printf("  Type %s'q'%s to exit\n", green, reset)
	fmt.Printf("  Type %s'h'%s for help\n", green, reset)
	fmt.Printf("  Type %s'add <email>'%s to add an admin\n", green, reset)
	fmt.Printf("  Type %s'delete <email>'%s to delete an admin\n", green, reset)
	fmt.Printf("  Type %s'modify <email>'%s to modify an admin\n", green, reset)
}

func printFlagUsage(){
	fmt.Printf("%sAvailable commands:%s\n", yellow, reset)
	fmt.Printf("  Type %s'q'%s to exit\n", green, reset)
	fmt.Printf("  Type %s'h'%s for help\n", green, reset)
	fmt.Printf("  Type %s'see'%s to see all flags\n", green, reset)
	fmt.Printf("  Type %s'set <flag>'%s to set flag to true\n", green, reset)
	fmt.Printf("  Type %s'reset <flag>'%s to set flag to false\n", green, reset)
}

func printWhitelist(){
	fmt.Printf("%sAvailable commands:%s\n", yellow, reset)
	fmt.Printf("  Type %s'q'%s to exit\n", green, reset)
	fmt.Printf("  Type %s'h'%s for help\n", green, reset)
	fmt.Printf("  Type %s'add'%s to seed csv\n", green, reset)
}

func main() {
	args := os.Args
	if len(args) > 3 {
		fmt.Printf("%sError: Too many arguments provided%s\n", red, reset)
		printCommandUsage()
		os.Exit(64)
	} else if len(args) ==3 {
		db, err := basic.NewSession()
		if err != nil {
			fmt.Printf("%sCould not connect to database: %v%s\n", red, err, reset)
			os.Exit(74)
		}
		fmt.Printf("%sConnected to database%s\n", magenta, reset)
		runFile(args[2],db);
	}else if len(args) == 2 {
		if (args[1] =="admin"){
			db, err := basic.NewSession()
			if err != nil {
				fmt.Printf("%sCould not connect to database: %v%s\n", red, err, reset)
				os.Exit(74)
			}
			fmt.Printf("%sConnected to database%s\n", magenta, reset)
			printAdminUsage()
			runPrompt1(db)
		} else if (args[1]=="flag"){
			db, err := basic.NewSession()
			if err != nil {
				fmt.Printf("%sCould not connect to database: %v%s\n", red, err, reset)
				os.Exit(74)
			}
			fmt.Printf("%sConnected to database%s\n", magenta, reset)
			printFlagUsage()
			runPrompt2(db)
		} else if (args[1]=="whitelist"){
			db, err := basic.NewSession()
			if err != nil {
				fmt.Printf("%sCould not connect to database: %v%s\n", red, err, reset)
				os.Exit(74)
			}
			fmt.Printf("%sConnected to database%s\n", magenta, reset)
			printWhitelist()
			runPrompt3(db)
		} else {
			fmt.Printf("%sWrong argument%s\n", red, reset)
			printCommandUsage()
		}
	} else {
		fmt.Printf("%sWrong argument%s\n", red, reset)
		printCommandUsage()
	}
}


func runPrompt1(db *sql.DB) {
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
			printAdminUsage()
			continue
		}
		run1(line, db)
		if haderror {
			haderror = false
		}
	}
}

func runPrompt2(db *sql.DB) {
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
			printFlagUsage()
			continue
		}
		run2(line, db)
		if haderror {
			haderror = false
		}
	}
}

func runPrompt3(db *sql.DB) {
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
			printWhitelist()
			continue
		}
		run3(line, db)
		if haderror {
			haderror = false
		}
	}
}

func runFile(filename string, db *sql.DB) {
	re := regexp.MustCompile(`^(.*?)(\.[^.]*$|$)`)
    matches := re.FindStringSubmatch(filename)
    if len(matches) < 3 {
        fmt.Printf("%sInvalid filename format: %s%s\n", red, filename, reset)
        os.Exit(64)
    }
    name := matches[1]
    extension := matches[2]
	commandRe := regexp.MustCompile(`^(.*?)(_.*)?$`)
    commandMatches := commandRe.FindStringSubmatch(name)
    if len(commandMatches) < 2 {
        fmt.Printf("%sInvalid command format in filename: %s%s\n", red, filename, reset)
        os.Exit(64)
    }
    commandName := commandMatches[1]
	commandObject := commandMatches[2]
	fmt.Printf("%sRunning command %s (name: %s, extension: %s)%s\n", cyan, commandName, name, extension, reset)
    //fmt.Printf("%sRunning file %s (name: %s, extension: %s)%s\n", cyan, filename, name, extension, reset)
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("%sCould not open file %s%s\n", red, filename, reset)
		os.Exit(74)
	}
	defer file.Close()


	scanner := bufio.NewScanner(file)
	if (commandObject == "_admin"){
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}
			line =commandName + " " + line
			fmt.Printf("%s>%s %s\n", blue, reset, line)
			run1(line, db)
			if haderror {
				haderror = false
			}
		}
	}else if (commandObject == "_flag"){
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}
			line =commandName + " " + line
			fmt.Printf("%s>%s %s\n", blue, reset, line)
			run2(line, db)
			if haderror {
				haderror = false
			}
		}
	} else {
		fmt.Printf("%sInvalid command object in filename: %s%s\n", red, filename, reset)
		os.Exit(64)
	}
}



func run1(source string, db *sql.DB) {
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
		printAdminUsage()
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

func GetAdminId(userID string, db *sql.DB) (string,error) {
	var adminID string
	err := db.QueryRow(`SELECT id FROM admins WHERE user_id = $1`, userID).Scan(&adminID)
	if err != nil {
		if err == sql.ErrNoRows {
			return adminID, errors.New("Admin not found")
		}
		return adminID, err
	}
	return adminID, nil
}

func DeleteAdmin(email string, db *sql.DB) error {
	user, err := CheckUser(email, db)
	if err != nil {
		return err
	}
	admin, err := GetAdminId(user.ID, db)
	if err != nil {
		return err
	}
	_, er := db.Exec("DELETE FROM qr_data where admin_id = $1", admin)
	if er != nil {
		return er
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

func printFlagDetails(db *sql.DB) {
	headers := []string{"Flag", "Value"}
	var flags []Flag
	rows, err := db.Query("SELECT name, value FROM flags")
	if err != nil {
		fmt.Printf("%sError: %v%s\n", red, err, reset)
		haderror = true
		return
	}
	defer rows.Close()
	for rows.Next() {
		var flag Flag
		err = rows.Scan(&flag.Name, &flag.Value)
		if err != nil {
			fmt.Printf("%sError: %v%s\n", red, err, reset)
			haderror = true
			return
		}
		flags = append(flags, flag)
	}
	detailWidth := len(headers[0])
	valueWidth := len(headers[1])
	for _, row := range flags {
		if len(row.Name) > detailWidth {
			detailWidth = len(row.Name)
		}
		if len(fmt.Sprintf("%t", row.Value)) > valueWidth {
			valueWidth = len(fmt.Sprintf("%t", row.Value))
		}
	}

	fmt.Printf("%sDetails of the flags are as follows:%s\n", cyan, reset)
	fmt.Printf("+-%s-+-%s-+\n", strings.Repeat("-", detailWidth), strings.Repeat("-", valueWidth))
	printTableRow(detailWidth, valueWidth, headers[0], headers[1], magenta+bold)
	for _, row := range flags {
		printTableRow(detailWidth, valueWidth, row.Name, fmt.Sprintf("%t", row.Value), "")
	}
}

func run2(source string, db *sql.DB) {
	words := strings.Fields(source)
	if len(words) == 0 {
		fmt.Printf("%sError: empty command%s\n", red, reset)
		haderror = true
		return
	}

	firstWord := words[0]
	switch firstWord {
	case "see":
		printFlagDetails(db)
	case "set":
		if len(words) < 2 {
			fmt.Printf("%sError: missing flag name%s\n", red, reset)
			haderror = true
			return
		}
		flag := words[1]
		err := SetFlag(flag, db)
		if err != nil {
			fmt.Printf("%sError setting flag %v%s\n", red, err, reset)
			haderror = true
		} else {
			fmt.Printf("%sFlag set successfully%s\n", green, reset)
		}
	case "reset":
		if len(words) < 2 {
			fmt.Printf("%sError: missing flag name%s\n", red, reset)
			haderror = true
			return
		}
		flag := words[1]
		err := ResetFlag(flag, db)
		if err != nil {
			fmt.Printf("%sError resetting flag %v%s\n", red, err, reset)
			haderror = true
		} else {
			fmt.Printf("%sFlag reset successfully%s\n", green, reset)
		}
	default:
		fmt.Printf("%sError: unknown command %s%s\n", red, firstWord, reset)
		printFlagUsage()
		haderror = true
	}
}

func SetFlag(flag string, db *sql.DB) error {
	var existingFlag Flag
	err := db.QueryRow(`SELECT name, value FROM flags WHERE name = $1`, flag).
		Scan(&existingFlag.Name, &existingFlag.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("flag not found")
		}
		return err
	}

	_, err = db.Exec(`
		UPDATE flags SET value = $1 WHERE name = $2
	`, true, flag)
	if err != nil {
		return err
	}
	return nil
}

func ResetFlag(flag string, db *sql.DB) error {
	var existingFlag Flag
	err := db.QueryRow(`SELECT name, value FROM flags WHERE name = $1`, flag).
		Scan(&existingFlag.Name, &existingFlag.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("flag not found")
		}
		return err
	}

	_, err = db.Exec(`
		UPDATE flags SET value = $1 WHERE name = $2
	`, false, flag)
	if err != nil {
		return err
	}
	return nil
}

func run3(source string, db *sql.DB) {
	words := strings.Fields(source)
	if len(words) == 0 {
		fmt.Printf("%sError: empty command%s\n", red, reset)
		haderror = true
		return
	}

	firstWord := words[0]
	switch firstWord {
	case "add":
		err := AddWhitelist(db)
		if err != nil {
			fmt.Printf("%sError adding whitelist: %v%s\n", red, err, reset)
			haderror = true
		} else {
			fmt.Printf("%sWhitelist added successfully%s\n", green, reset)
		}
	default:
		fmt.Printf("%sError: unknown command %s%s\n", red, firstWord, reset)
		printWhitelist()
		haderror = true
	}
}

func AddWhitelist(db *sql.DB) error {
    file, err := os.Open("whitelist.csv")
    if err != nil {
        fmt.Printf("%sCould not open file whitelist.csv%s\n", red, reset)
        return err
    }

    defer file.Close()
    scanner := bufio.NewScanner(file)

    // Skip the first header line
    if scanner.Scan() {
        header := scanner.Text()
        fmt.Printf("%sSkipping header: %s%s\n", cyan, header, reset)
    }

    for scanner.Scan() {
        line := scanner.Text()
        if strings.TrimSpace(line) == "" {
            continue
        }
        words := strings.Split(line, ",")
        if len(words) < 3 {
            fmt.Printf("%sError: invalid format in whitelist.csv%s\n", red, reset)
            return errors.New("invalid format")
        }

        name := words[0]
        email := words[2]

		var existingUser User
		err := db.QueryRow(`SELECT id, email, name FROM whitelists WHERE email = $1`, email).
			Scan(&existingUser.ID, &existingUser.Email, &existingUser.Name)
		if err != nil {
			if err != sql.ErrNoRows {
				return err
			}
		}

		if existingUser.Email == email {
			fmt.Printf("%sUser already exists in whitelist: %s%s\n", yellow, email, reset)
			return nil
		}

		existingUser.Email = email
		existingUser.Name = name
		id := uuid.New().String()

		_, err = db.Exec(`INSERT INTO whitelists (id, name, email) VALUES ($1, $2, $3)`, id, name, email)
		if err != nil {
			fmt.Printf("%sError inserting into whitelist: %v%s\n", red, err, reset)
			return err
		}
		
        fmt.Printf("%sProcessing: Name=%s, Email=%s%s\n", green, name, email, reset)
    }

    if err := scanner.Err(); err != nil {
        fmt.Printf("%sError reading file: %v%s\n", red, err, reset)
        return err
    }
    return nil
}