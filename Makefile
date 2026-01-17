CC         = go
BUILD_PATH = ./bin
SRC        = ./cmd/sshman/main.go
TARGET     = sshman
BINS       = $(BUILD_PATH)/$(TARGET)
INST       = ~/.local/bin
CONF_DEST  = ~/.config/sshman/
CONF       = ./config/sshman.json

.PHONY: all clean build run install

all: run

clean:
	rm -rf $(BUILD_PATH)

build: clean
	mkdir -p $(CONF_DEST)
	mkdir -p $(BUILD_PATH)
	$(CC) build -tags sqlite_omit_load_extension -o $(BINS) $(SRC)

run:
	$(CC) run $(SRC) --about

install: build 
	mkdir -p $(INST)
	ditto $(BINS) $(INST)/$(TARGET)
	cp -vn $(CONF) $(CONF_DEST)

uninstall: clean
	sudo rm -f $(INST)/$(TARGET)
	rm -rf $(CONF_DEST)
