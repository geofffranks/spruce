package goptions

import (
	"fmt"
	"os"
	"time"
)

func ExampleFlagSet_PrintHelp() {
	options := struct {
		Server   string        `goptions:"-s, --server, obligatory, description='Server to connect to'"`
		Password string        `goptions:"-p, --password, description='Don\\'t prompt for password'"`
		Timeout  time.Duration `goptions:"-t, --timeout, description='Connection timeout in seconds'"`
		Help     Help          `goptions:"-h, --help, description='Show this help'"`

		Verbs
		Execute struct {
			Command string   `goptions:"--command, mutexgroup='input', description='Command to exectute', obligatory"`
			Script  *os.File `goptions:"--script, mutexgroup='input', description='Script to exectute', rdonly"`
		} `goptions:"execute"`
		Delete struct {
			Path  string `goptions:"-n, --name, obligatory, description='Name of the entity to be deleted'"`
			Force bool   `goptions:"-f, --force, description='Force removal'"`
		} `goptions:"delete"`
	}{ // Default values goes here
		Timeout: 10 * time.Second,
	}

	args := []string{"--help"}
	fs := NewFlagSet("goptions", &options)
	err := fs.Parse(args)
	if err == ErrHelpRequest {
		fs.PrintHelp(os.Stdout)
		return
	} else if err != nil {
		fmt.Printf("Failure: %s", err)
	}

	// Output:
	// Usage: goptions [global options] <verb> [verb options]
	//
	// Global options:
	//         -s, --server   Server to connect to (*)
	//         -p, --password Don't prompt for password
	//         -t, --timeout  Connection timeout in seconds (default: 10s)
	//         -h, --help     Show this help
	//
	// Verbs:
	//     delete:
	//         -n, --name     Name of the entity to be deleted (*)
	//         -f, --force    Force removal
	//     execute:
	//             --command  Command to exectute (*)
	//             --script   Script to exectute
}

func ExampleVerbs() {
	options := struct {
		ImportantFlag string        `goptions:"-f, --flag, description='Important flag, obligatory'"`
		Password      string        `goptions:"-p, --password, description='Don\\'t prompt for password'"`
		Timeout       time.Duration `goptions:"-t, --timeout, description='Connection timeout in seconds'"`
		Help          Help          `goptions:"-h, --help, description='Show this help'"`

		Verb    Verbs
		Execute struct {
			Command string   `goptions:"--command, mutexgroup='input', description='Command to exectute', obligatory"`
			Script  *os.File `goptions:"--script, mutexgroup='input', description='Script to exectute', rdonly"`
		} `goptions:"execute"`
		Delete struct {
			Path  string `goptions:"-n, --name, obligatory, description='Name of the entity to be deleted'"`
			Force bool   `goptions:"-f, --force, description='Force removal'"`
		} `goptions:"delete"`
	}{ // Default values goes here
		Timeout: 10 * time.Second,
	}

	args := []string{"delete", "-n", "/usr/bin"}
	fs := NewFlagSet("goptions", &options)
	_ = fs.Parse(args)
	// Error handling omitted
	fmt.Printf("Selected verb: %s", options.Verb)

	// Output:
	// Selected verb: delete
}

func ExampleRemainder() {
	options := struct {
		Username  string `goptions:"-u, --user, obligatory, description='Name of the user'"`
		Remainder Remainder
	}{}

	args := []string{"-u", "surma", "some", "more", "args"}
	fs := NewFlagSet("goptions", &options)
	_ = fs.Parse(args)
	// Error handling omitted
	fmt.Printf("Remainder: %#v", options.Remainder)

	// Output:
	// Remainder: goptions.Remainder{"some", "more", "args"}
}
