module 01cloud-api

go 1.22


require (
	cloud.google.com/go/pubsub v1.3.1
	cloud.google.com/go/storage v1.11.0
	github.com/DATA-DOG/go-sqlmock v1.4.1
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/ashwanthkumar/slack-go-webhook v0.0.0-20200209025033-430dd4e66960
	github.com/aws/aws-sdk-go v1.34.27
	// github.com/berrybytes/01cloud-store v0.0.0-00010101000000-000000000000
	github.com/berrybytes/dockerhub-go v1.0.3
	github.com/cloudflare/cloudflare-go v0.16.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v1.4.2-0.20200203170920-46ec8731fbce
	github.com/gabriel-vasile/mimetype v1.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.5.2
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/gosimple/slug v1.9.0
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/jinzhu/gorm v1.9.12
	github.com/joho/godotenv v1.4.0
	github.com/jpillora/go-tld v1.2.1
	github.com/ktrysmt/go-bitbucket v0.9.56
	github.com/parnurzeal/gorequest v0.2.16
	github.com/prometheus/client_golang v1.6.0
	github.com/prometheus/common v0.9.1
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.8.2
	github.com/swaggo/http-swagger v1.0.0
	github.com/swaggo/swag v1.7.0
	github.com/tidwall/gjson v1.6.0
	github.com/xanzy/go-gitlab v0.50.2
	github.com/zclconf/go-cty v1.2.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/oauth2 v0.7.0
	google.golang.org/api v0.32.0
	google.golang.org/grpc v1.36.1
	google.golang.org/protobuf v1.28.0
	gopkg.in/stretchr/testify.v1 v1.2.2
	helm.sh/helm/v3 v3.1.2
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.3
)

// require github.com/berrybytes/01cloud-store v0.0.0-00010101000000-000000000000

require github.com/berrybytes/01cloud-store v0.0.0-00010101000000-000000000000

require (
	cloud.google.com/go v0.66.0 // indirect
	cloud.google.com/go/firestore v1.2.0 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/Microsoft/hcsshim v0.8.7 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/agext/levenshtein v1.2.1 // indirect
	github.com/apparentlymart/go-textseg v1.0.0 // indirect
	github.com/apparentlymart/go-textseg/v12 v12.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/containerd/continuity v0.0.0-20200107194136-26c1120b8d41 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/spec v0.20.3 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.8 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.3.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/lib/pq v1.1.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/procfs v0.0.11 // indirect
	github.com/rainycape/unidecode v0.0.0-20150907023854-cb7f23ec59be // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14 // indirect
	github.com/tidwall/match v1.0.1 // indirect
	github.com/tidwall/pretty v1.0.1 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.0.2 // indirect
	github.com/xdg-go/stringprep v1.0.2 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.mongodb.org/mongo-driver v1.9.1 // indirect
	go.opencensus.io v0.22.4 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
	golang.org/x/tools v0.6.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20200914193844-75d14daec038 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog v1.0.0 // indirect
	moul.io/http2curl v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
)
