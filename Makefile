CC         = go
BUILD_PATH = ./bin
SRC        = ./cmd/sshman/main.go
TARGET     = sshman
BINS       = $(BUILD_PATH)/$(TARGET)
INST       = /usr/local/bin
CONF_DEST  = ~/.config/sshman/
CONF       = ./config/sshman.json

.PHONY: all clean build run install

all: run

clean:
	rm -rf $(BUILD_PATH)

build: clean
	mkdir -p $(CONF_DEST)
	mkdir -p $(BUILD_PATH)
	$(CC) build -o $(BINS) $(SRC)

run:
	$(CC) run $(SRC) --help

install: build 
	sudo cp -v $(BINS) $(INST)
	cp -vn $(CONF) $(CONF_DEST)

uninstall: clean
	sudo rm -f $(INST)/$(TARGET)
	rm -rf $(CONF_DEST)
