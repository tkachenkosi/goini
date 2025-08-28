TARGET = dist/main
SOURCES = *.go

.PHONY: build run vet lint clean
.DEFAULT_GOAL = run

build:
	@echo Компиляция программы!
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o ${TARGET} .
${TARGET}: main.go
	@echo Компиляция программы вторым способом.
	go build -o ${TARGET} main.go
run:
	go run .
vet:
	@echo Запускаем $@
	@echo go vet
lint:
	@echo Запускаем $@
	golangci-lint run
clean:
ifneq (,$(wildcard $(TARGET)))
	 @echo Удаляем файл $(TARGET)
	 rm -f ${TARGET}
endif
	go clean
