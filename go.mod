module github.com/SamHennessy/hlive

go 1.17

// Until the client download bug is fixed
// https://github.com/mxschmitt/playwright-go/issues/236
replace github.com/mxschmitt/playwright-go => github.com/SamHennessy/playwright-go v0.1400.1-0.20211116005112-e0b9afafb423

require (
	github.com/cornelk/hashmap v1.0.1
	github.com/go-test/deep v1.0.7
	github.com/gorilla/websocket v1.4.2
	github.com/mxschmitt/playwright-go v0.1400.1-0.20211028151004-65c211c7fa66
	github.com/rs/xid v1.3.0
	github.com/rs/zerolog v1.25.0
	github.com/teris-io/shortid v0.0.0-20201117134242-e59966efd125
)

require (
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964 // indirect
	github.com/dchest/siphash v1.1.0 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
)
