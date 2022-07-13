package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func parseArgs() Arguments {
	id := flag.String("-id", "", "the id of the user")
	operation := flag.String("-operation", "", "operation type")
	item := flag.String("-item", "", "the user data")
	fileName := flag.String("-fileName", "", "the write path")
	flag.Parse()

	return Arguments{
		"id":        *id,
		"operation": *operation,
		"item":      *item,
		"fileName":  *fileName,
	}
}

func Perform(args Arguments, writer io.Writer) error {
	operations := map[string]func(Arguments, io.Writer) error{
		"add":      Add,
		"list":     List,
		"findById": FindById,
		"remove":   Remove,
	}

	if args["operation"] == "" {
		return fmt.Errorf("-operation flag has to be specified")
	}
	selectedOperation, ok := operations[args["operation"]]
	if !ok {
		return fmt.Errorf("Operation %v not allowed!", args["operation"])
	}
	fileName := args["fileName"]
	if fileName == "" {
		return fmt.Errorf("-fileName flag has to be specified")
	}
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	operationError := selectedOperation(args, writer)
	if operationError != nil {
		return operationError
	}

	return nil
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

func Add(args Arguments, writer io.Writer) error {
	item := args["item"]
	if item == "" {
		return fmt.Errorf("-item flag has to be specified")
	}
	var freshman User
	err := json.Unmarshal([]byte(args["item"]), &freshman)
	if err != nil {
		return err
	}

	users := getUsers(args["fileName"])

	id := freshman.Id
	for _, user := range users {
		if user.Id == id {
			writer.Write([]byte(fmt.Sprintf("Item with id %s already exists", id)))
		}
	}

	users = append(users, freshman)
	usersJson, err := json.Marshal(users)
	if err != nil {
		return err
	}
	err = os.WriteFile(args["fileName"], usersJson, 0644)
	if err != nil {
		return err
	}

	return nil
}

func List(args Arguments, writer io.Writer) error {
	data, err := ioutil.ReadFile(args["fileName"])
	if err != nil {
		return err
	}
	if len(data) == 0 {
		writer.Write([]byte{'[', ']'})
	} else {
		writer.Write(data)
	}
	return nil
}

func FindById(args Arguments, writer io.Writer) error {
	id := args["id"]
	if id == "" {
		return fmt.Errorf("-id flag has to be specified")
	}

	users := getUsers(args["fileName"])

	for _, user := range users {
		if user.Id == id {
			userAim, err := json.Marshal(user)
			if err != nil {
				return err
			}
			_, err = writer.Write(userAim)
			return err
		}
	}
	writer.Write([]byte(""))
	return nil
}

func Remove(args Arguments, writer io.Writer) error {
	id := args["id"]
	if id == "" {
		return fmt.Errorf("-id flag has to be specified")
	}

	users := getUsers(args["fileName"])

	var removeIndex int = -1
	for index, user := range users {
		if user.Id == id {
			removeIndex = index
		}
	}
	if removeIndex == -1 {
		writer.Write([]byte(fmt.Sprintf("Item with id %s not found", id)))
		return nil
	}

	users = append(users[:removeIndex], users[removeIndex+1:]...)
	usersJson, err := json.Marshal(users)
	if err != nil {
		return err
	}
	err = os.WriteFile(args["fileName"], usersJson, 0644)
	return err
}

func getUsers(fileName string) []User {
	usersJson, _ := ioutil.ReadFile(fileName)
	var users []User
	json.Unmarshal(usersJson, &users)
	return users
}
