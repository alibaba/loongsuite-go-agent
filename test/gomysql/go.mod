module go_mysql

go 1.24.0

replace github.com/alibaba/loongsuite-go/test/verifier => ../../test/verifier

replace github.com/alibaba/loongsuite-go/pkg => ../../pkg

require (
	github.com/alibaba/loongsuite-go/test/verifier v0.0.0-00010101000000-000000000000
	github.com/go-mysql-org/go-mysql v1.11.0
)
