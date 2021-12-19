module qsvencc-client

go 1.17

replace local.packages/qsvencc-tcp => ../..

require (
	github.com/jessevdk/go-flags v1.5.0
	local.packages/qsvencc-tcp v0.0.0-00010101000000-000000000000
)

require golang.org/x/sys v0.0.0-20210320140829-1e4c9ba3b0c4 // indirect
