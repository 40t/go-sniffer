IMG ?= harbor.cestc.com/public-release/go-sniffer:v0.0.1
REMOTEIMG ?= 10.32.226.224:85/public-release/go-sniffer:v0.0.1
TARGET = go-sniffer
all: $(TARGET)

$(TARGET):
	go build -ldflags "-s -w" -o bin/$@

init:
	swag init

run:
	go run main.go -config config.json

docker:
#	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o traffic-validate
#	upx traffic-validate
	docker build --platform linux/amd64 -t $(IMG) .

push:
	docker push $(IMG)

kind:
	kind load docker-image --name networkpolicy $(IMG)
	kind load docker-image --name networkpolicy $(IMG)

ci: docker kind

cd:
	kubectl apply -f ./build

del:
	kubectl delete -f ./build

roll:
	kubectl -n gb patch deployment go-sniffer --patch "{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{\"date\":\"`date +'%s'`\"}}}}}"

tag:
	docker tag $(IMG) $(REMOTEIMG)
	docker push $(REMOTEIMG)

deploy: ci cd

upload: docker tag