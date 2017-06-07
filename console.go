package godless

import (
	"errors"
	"fmt"
)

func TerminalConsole(addr string) error {
	fmt.Printf("The console should send user commands to '%v'.\n", addr)
	fmt.Println("Data found in the APIResponse should be formatted in a human readable manner.")
	fmt.Println("")
	fmt.Println("Some functions that will help:")
	fmt.Println("\tClient.SendQuery(query *Query) (APIResponse, error)")
	fmt.Println("\tCompileQuery(source string) *Query")
	fmt.Println("")
	fmt.Println("Use this terminal package:")
	fmt.Println("\tgithub.com/jroimartin/gocui")
	fmt.Println("")
	fmt.Println("Have a lot of fun!")
	fmt.Println("")

	return errors.New("not implemented")
}
