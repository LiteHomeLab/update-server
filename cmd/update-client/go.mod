module github.com/LiteHomeLab/update-server/cmd/update-client

go 1.23

require (
	github.com/LiteHomeLab/update-server/clients/go/client v0.0.0
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/LiteHomeLab/update-server/clients/go/client => ../../clients/go/client
