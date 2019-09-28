module go-home.io/x/server

require (
	bou.ke/monkey v1.0.2-0.20190527161844-ca6af776195d
	github.com/corona10/goimagehash v0.2.0
	github.com/creasty/defaults v1.2.1
	github.com/disintegration/imaging v1.5.0
	github.com/docker/docker v1.4.2-0.20190818020526-0c46a20f9471
	github.com/dsnet/compress v0.0.0-20171208185109-cc9eb1d7ad76 // indirect
	github.com/enr/go-commons v0.0.0-20150504121636-bcd3f40eeea8 // indirect
	github.com/fatih/color v1.7.0
	github.com/fortytw2/leaktest v1.2.0
	github.com/go-playground/locales v0.12.1 // indirect
	github.com/go-playground/universal-translator v0.16.0 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/mux v1.6.2
	github.com/gorilla/websocket v1.3.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/mholt/archiver v1.1.3-0.20180818163635-089cf0d1f58c
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646 // indirect
	github.com/nwaples/rardecode v0.0.0-20171029023500-e06696f847ae // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pierrec/lz4 v2.0.3+incompatible // indirect
	github.com/pkg/errors v0.8.0
	github.com/rakyll/statik v0.1.5
	github.com/stretchr/testify v1.2.2
	github.com/ulikunitz/xz v0.5.4 // indirect
	github.com/vkorn/go-bintray v0.0.0-20180801131521-627b4bc5e556
	go-home.io/x/server/plugins v0.0.0-20181025030525-18e916b213bc
	golang.org/x/crypto v0.0.0-20180904163835-0709b304e793
	golang.org/x/image v0.0.0-20180708004352-c73c2afc3b81 // indirect
	golang.org/x/sys v0.0.0-20180905080454-ebe1bf3edb33 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1
	gopkg.in/go-playground/validator.v9 v9.21.0
	gopkg.in/robfig/cron.v2 v2.0.0-20150107220207-be2e0b0deed5
	gopkg.in/yaml.v2 v2.2.1
)

replace go-home.io/x/server/plugins => ./plugins

go 1.13
