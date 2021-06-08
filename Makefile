hello:
	echo "Hello"

build:
	docker build -t unzipper .

run:
	go run main.go