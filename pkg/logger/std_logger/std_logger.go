package std_logger

import "log"

type stdLogger struct {}

func New() *stdLogger {
	return &stdLogger{}
}


func (l *stdLogger) Debug(i ...any) {
	log.Println(i...)
}

func (l *stdLogger) Debugf(format string, args ...any) {
	log.Printf(format+"\n", args...)
}

func (l *stdLogger) Print(i ...any) {
	log.Println(i...)
}

func (l *stdLogger) Info(i ...any) {
	log.Println(i...)
}

func (l *stdLogger) Infof(format string, args ...any) {
	log.Printf(format+"\n", args...)
}

func (l *stdLogger) Warn(i ...any) {
	log.Println(i...)
}

func (l *stdLogger) Warnf(format string, args ...any) {
	log.Printf(format+"\n", args...)
}

func (l *stdLogger) Error(i ...any) {
	log.Println(i...)
}

func (l *stdLogger) Errorf(format string, args ...any) {
	log.Printf(format+"\n", args...)
}

func (l *stdLogger) Fatal(i ...any) {
	log.Fatalln(i...)
}

func (l *stdLogger) Fatalf(format string, args ...any) {
	log.Fatalf(format+"\n", args...)
}

func (l *stdLogger) Panic(i ...any) {
	log.Panicln(i...)
}

func (l *stdLogger) Panicf(format string, args ...any) {
	log.Panicf(format+"\n", args...)
}
