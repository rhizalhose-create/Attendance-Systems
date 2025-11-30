// config/config.go
package config

import "gorm.io/gorm"



func GetDB() *gorm.DB {
    return DB
}


func InitDatabase() {
    ConnectDB()
}

func ShowTableStructure() {
    DisplayTableStructure()
}   

func ShowUserStats() {
    DisplayUserStats()
}
