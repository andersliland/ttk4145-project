package utilities

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

func SaveBackup(filename string, cabOrders [NumFloors]bool) error {
	data, err := json.Marshal(cabOrders)
	if err != nil {
		log.Println("json.Marshal() error: Failed to marshal backup")
		return err
	}
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		log.Println("ioutil.WriteFile() error: Failed to save backup")
		return err
	}
	return nil
}

func LoadBackup(filename string, cabOrders *[NumFloors]bool) error {
	if _, fileNotFound := os.Stat(filename); fileNotFound == nil {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Println("loadFromDisk() error: Failed to read file")
		}
		if err := json.Unmarshal(data, &cabOrders); err != nil {
			log.Println("loadFromDisk() error: Failed to unmarshal")
		}
		return nil
	} else {
		log.Println("\t\t\t cabOrders backupfile not found")
		return fileNotFound
	}
}
