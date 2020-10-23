build:
	go mod download
	go build -o recmd-dmn

default: build

test:
	(cd dmn; go test)

clean:
	rm -f recmd-dmn

install:
	go install