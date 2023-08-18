package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/mistralmail/mistralmail"
	mistralmailbackend "github.com/mistralmail/mistralmail/backend"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	config  *mistralmail.Config
	backend *mistralmailbackend.Backend
)

func main() {

	var err error

	config = mistralmail.BuildConfigFromEnv()

	backend, err = mistralmailbackend.New(config.DatabaseURL)
	if err != nil {
		log.Fatalf("couldn't create backend: %v", err)
	}
	defer backend.Close()

	var rootCmd = &cobra.Command{Use: "mistralmail-cli"}

	var createUserCmd = &cobra.Command{
		Use:   "create-user",
		Short: "Create a new user",
		Run:   handleCreateUserCommand,
	}

	rootCmd.AddCommand(createUserCmd)
	err = rootCmd.Execute()
	log.Fatalf("%v", err)

}

func handleCreateUserCommand(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Printf("Password: ")
	passwordBytes, _ := term.ReadPassword(int(syscall.Stdin))
	password := string(passwordBytes)
	password = strings.TrimSpace(password)

	fmt.Printf("\nRepeat password: ")
	password2Bytes, _ := term.ReadPassword(int(syscall.Stdin))
	password2 := string(password2Bytes)
	password2 = strings.TrimSpace(password2)

	if password != password2 {
		log.Fatalf("passwords don't match")
	}

	user, err := backend.CreateNewUser(email, password)
	if err != nil {
		log.Fatalf("couldn't create new user: %v", err)
	}
	log.Printf("Successfully created new user %s with id %d", user.Email, user.ID)
}
