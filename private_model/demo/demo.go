package main

import "github.com/livegoplayer/go_db_helper/private_model"

var APPROOT = "D:\\files\\workspace\\go\\filestore-server"

func main() {
	GetDbProjectPath()
}

func GetDbProjectPath() {
	private_model.Parse(APPROOT)
}
