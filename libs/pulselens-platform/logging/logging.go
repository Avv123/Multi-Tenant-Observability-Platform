package logging

import (
	"log"
	"os"
)

func Initialize() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC)
}

func Infof(format string, args ...any) {
	log.Printf("INFO "+format, args...)
}

func Errorf(format string, args ...any) {
	log.Printf("ERROR "+format, args...)
}

func Fatalf(format string, args ...any) {
	log.Fatalf("FATAL "+format, args...)
}
