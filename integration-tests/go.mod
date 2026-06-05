module github.com/smap/shared-libs/integration-tests

go 1.25.0

require (
	github.com/smap-hcmut/shared-libs/go v0.0.0
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/smap-hcmut/shared-libs/go => ../go
