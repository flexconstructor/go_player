package main

import (
	"log"
	"github.com/flexconstructor/GoPlayer"
)
type GPLogger struct {

}

func NewGPLogger()(GoPlayer.Logger){
	return &GPLogger{}
}

func (l *GPLogger)Debug(arg0 interface{}, args ...interface{}) {
	log.Println("DEBUG: ",arg0, args...)
}

func (l *GPLogger)Trace(arg0 interface{}, args ...interface{}) {
	log.Println("TRACE:", arg0, args...)
}

func (l *GPLogger)Info(arg0 interface{}, args ...interface{}) {
	log.Println("INFO: ",arg0, args...)
}

func (l *GPLogger)Warn(arg0 interface{}, args ...interface{}) error {
	return log.Panicln("WARN: ",arg0, args...)
}

// Package-level wrapper for Global logger.
func (l *GPLogger)Error(arg0 interface{}, args ...interface{}) error {
	return log.Panicln("ERROR: ",arg0, args...)
}

func (l *GPLogger)Critical(arg0 interface{}, args ...interface{}) error {
	return log.Fatal("FATAL: ",arg0, args...)
}


