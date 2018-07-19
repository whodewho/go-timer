BIN := `pwd | awk -F"/" '{print $$NF}'`

default: online

online:
	go build -tags $(BIN) -i -o ./bin/$(BIN)
