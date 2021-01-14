module main

go 1.15

replace pkg/k8sDiscovery => ./pkg/k8sDiscovery

require (
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.2.0
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/zhiminwen/quote v0.0.0-20200612004834-54f3725dbd6a
	k8s.io/apimachinery v0.17.12
	pkg/k8sDiscovery v0.0.0-00010101000000-000000000000
)
