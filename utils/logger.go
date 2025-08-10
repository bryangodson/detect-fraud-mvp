package utils

import "log"

// Wrap the standard log so in future we can change format or add file output
func Info(v ...interface{})  { log.Println(v...) }
func Error(v ...interface{}) { log.Println(v...) }
