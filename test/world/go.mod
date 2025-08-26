module world

go 1.23.4

toolchain go1.23.9

replace github.com/alibaba/loongsuite-go-agent => ../../

require (
	dubbo.apache.org/dubbo-go/v3 v3.3.0
	github.com/alibaba/sentinel-golang v1.0.4
	github.com/apache/rocketmq-client-go/v2 v2.1.2
	github.com/cloudwego/eino v0.3.51
	github.com/cloudwego/eino-ext/components/model/ark v0.1.16
	github.com/cloudwego/eino-ext/components/model/claude v0.1.1
	github.com/cloudwego/eino-ext/components/model/ollama v0.0.0-20250710065240-482d48888f25
	github.com/cloudwego/eino-ext/components/model/openai v0.0.0-20250710065240-482d48888f25
	github.com/cloudwego/eino-ext/components/model/qwen v0.0.0-20250718041314-444cfd7822ec
	github.com/cloudwego/hertz v0.9.3
	github.com/gin-gonic/gin v1.10.0
	github.com/go-kratos/kratos/v2 v2.7.2
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gofiber/fiber/v2 v2.52.5
	github.com/gorilla/mux v1.8.1
	github.com/labstack/echo/v4 v4.12.0
	github.com/mark3labs/mcp-go v0.20.0
	github.com/segmentio/kafka-go v0.4.48
	github.com/sirupsen/logrus v1.9.3
	github.com/valyala/fasthttp v1.57.0
	go.mongodb.org/mongo-driver v1.11.6
	go.opentelemetry.io/otel v1.24.0
	go.opentelemetry.io/otel/trace v1.24.0
	go.uber.org/zap v1.26.0
	google.golang.org/grpc v1.64.1
	gorm.io/driver/mysql v1.5.7
	gorm.io/gorm v1.25.7
)

require (
	cloud.google.com/go/compute/metadata v0.5.0 // indirect
	github.com/RoaringBitmap/roaring v1.2.3 // indirect
	github.com/Workiva/go-datastructures v1.0.52 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/anthropics/anthropic-sdk-go v1.4.0 // indirect
	github.com/apache/dubbo-getty v1.4.10 // indirect
	github.com/apache/dubbo-go-hessian2 v1.12.5 // indirect
	github.com/aws/aws-sdk-go-v2 v1.33.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.3 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.29.1 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.54 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.24 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.28 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.28 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.9 // indirect
	github.com/aws/smithy-go v1.22.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.2.0 // indirect
	github.com/bytedance/go-tagexpr/v2 v2.9.2 // indirect
	github.com/bytedance/gopkg v0.1.0 // indirect
	github.com/bytedance/sonic v1.13.2 // indirect
	github.com/bytedance/sonic/loader v0.2.4 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cloudwego/base64x v0.1.5 // indirect
	github.com/cloudwego/eino-ext/libs/acl/openai v0.0.0-20250626133421-3c142631c961 // indirect
	github.com/cloudwego/netpoll v0.6.2 // indirect
	github.com/creasty/defaults v1.5.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dubbogo/gost v1.14.0 // indirect
	github.com/dubbogo/grpc-go v1.42.10 // indirect
	github.com/dubbogo/triple v1.2.2-rc4 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/evanphx/json-patch v0.5.2 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/getkin/kin-openapi v0.118.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-kratos/aegis v0.2.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.21.1 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/go-playground/form/v4 v4.2.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.20.0 // indirect
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/goph/emperror v0.17.2 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/henrylee2cn/ameda v1.4.10 // indirect
	github.com/henrylee2cn/goutil v0.0.0-20210127050712-89660552f6f8 // indirect
	github.com/influxdata/tdigest v0.0.1 // indirect
	github.com/invopop/yaml v0.1.0 // indirect
	github.com/jinzhu/copier v0.3.5 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/k0kubun/pp v3.0.1+incompatible // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/knadh/koanf v1.5.0 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20230326075908-cb1d2100619a // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/meguminnnnnnnnn/go-openai v0.0.0-20250620092828-0d508a1dcdde // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/montanaflynn/stats v0.6.6 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	github.com/nikolalohinski/gonja v1.5.3 // indirect
	github.com/nyaruka/phonenumbers v1.0.55 // indirect
	github.com/ollama/ollama v0.5.12 // indirect
	github.com/openai/openai-go v1.10.1 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml v1.9.3 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/power-devops/perfstat v0.0.0-20221212215047-62379fc7944b // indirect
	github.com/prometheus/client_golang v1.13.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/shirou/gopsutil/v3 v3.23.6 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/slongfield/pyfmt v0.0.0-20220222012616-ea85ff4c361f // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/volcengine/volc-sdk-golang v1.0.23 // indirect
	github.com/volcengine/volcengine-go-sdk v1.1.21 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/yargevad/filepathx v1.0.0 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.10.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/sdk v1.16.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/arch v0.15.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240722135656-d784300faade // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	stathat.com/c/consistent v1.0.0 // indirect
)
