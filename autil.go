package go_binance

import (
	"encoding/json"
	"fmt"
	"github.com/redlon23/go-binance/models"
	"log"
	"os"
)

func GetKey(data []byte) string {
	k := new(models.ListenKey)
	if err := json.Unmarshal(data, &k); err != nil {
		log.Println(err)
		return ""
	}
	return k.Key
}

func ParseMetaInformation(wsMessage []byte) (*models.StreamMetaMessage, error) {
	meta := new(models.StreamMetaMessage)
	err := json.Unmarshal(wsMessage, &meta)
	if err != nil {
		log.Println(err)
		return meta, err
	}
	return meta, nil
}

// CheckLogsFolder Checks if log folder exists, if not creates one
func CheckLogsFolder() {
	if _, err := os.Stat("./logs"); os.IsNotExist(err) {
		fmt.Println("No logs folder in the current directory, creating one...")
		err := os.Mkdir("logs", 066)
		if err != nil {
			fmt.Println("Failed to create a folder!")
		}
	} else {
		fmt.Println("There is already a logs folder in the current directory!")
	}
}
