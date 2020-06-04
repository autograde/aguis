module github.com/autograde/aguis

go 1.14

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/autograde/aguis/kit v0.0.0-20200424153704-be31622bfc7a
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v0.0.0-20170803041405-316b4ba9c289
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/google/go-cmp v0.4.0
	github.com/google/go-github/v30 v30.1.0
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/gorilla/sessions v1.2.0
	github.com/gosimple/slug v1.9.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/hashicorp/go-retryablehttp v0.6.6 // indirect
	github.com/jinzhu/gorm v1.9.12
	github.com/labstack/echo-contrib v0.9.0
	github.com/labstack/echo/v4 v4.1.16
	github.com/markbates/goth v1.64.1
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.6.0
	github.com/prometheus/common v0.10.0 // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/tebeka/selenium v0.9.9
	github.com/urfave/cli v1.22.2
	github.com/xanzy/go-gitlab v0.32.1
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200602180216-279210d13fed // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200602225109-6fdc65e7d980 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	golang.org/x/tools v0.0.0-20200321224714-0d839f3cf2ed // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/genproto v0.0.0-20200603110839-e855014d5736 // indirect
	google.golang.org/grpc v1.29.1
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/autograde/aguis/kit => ./kit
