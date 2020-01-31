TEST?=$$(go list ./... |grep -v 'vendor')

default: test
# /new/app --title Sunnyside Gardens Yoga --street 48-21 39th Ave --city Sunnyside --county Queens --country USA --postalcode 11104
# /new/event --title Marathon Training Day --street Doughboy Plaza Woodside Ave. --city Woodside --county Queens --state New York  --country USA --postalcode 11377
# /datetime/event --title Harvest Festival --start Oct. 28, 2019 9:30am --end Nov. 2, 2019 5:30pm
clean:
	rm -rf bin/post
	rm -rf bin/get 
	rm -rf bin/update
	 
.PHONY: clean

build: 
	GOOS=linux GOARCH=amd64 go build -o  bin/datetime ./cmd/post/date/datetime.go
	GOOS=linux GOARCH=amd64 go build -o  bin/delete ./cmd/delete/deleteevent.go
	GOOS=linux GOARCH=amd64 go build -o  bin/update ./cmd/put/updateevent.go
	GOOS=linux GOARCH=amd64 go build -o  bin/post ./cmd/post/slack/newevent.go
	GOOS=linux GOARCH=amd64 go build -o  bin/listactive ./cmd/list/slack/active.go
	GOOS=linux GOARCH=amd64 go build -o  bin/query ./cmd/list/query/query.go
	GOOS=linux GOARCH=amd64 go build -o  bin/listall ./cmd/list/listall.go
	
.PHONY: build

deploy:
	serverless deploy -v > resp_deploy.txt

.PHONY: deploy 

post:
	
.PHONY: post

remove:
	serverless remove -v 

.PHONY: remove 


test: 
	docker-compose down
	docker-compose up -d --build --force-recreate
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test -v
	docker-compose down

.PHONY: test

list:
	curl -v  https://bbcrs2mxqk.execute-api.us-east-1.amazonaws.com/dev/event/list
	
.PHONY: list


update:
	curl -v  https://bbcrs2mxqk.execute-api.us-east-1.amazonaws.com/dev/event/update?--title+No+Title 

.PHONY: update 
	
